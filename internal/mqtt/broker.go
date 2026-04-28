package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/screenshare"

	"github.com/gorilla/websocket"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	log "github.com/sirupsen/logrus"
)

type Broker struct {
	server      *mqtt.Server
	port        string
	tlsConfig   *tls.Config
	authHook    *AuthHook
	aclHook     *ACLHook
	RoomManager *screenshare.RoomManager
}

type NotificationHub interface {
	NotifyScreenshare(uid, fromClientID string, payload interface{})
}

type AuthHook struct {
	mqtt.HookBase
	validateToken func(token string) (userID string, err error)
	server        *mqtt.Server
	roomManager   *screenshare.RoomManager
	iceServers    []interface{}
	hub           NotificationHub
}

type ACLHook struct {
	mqtt.HookBase
}

type ConnectionHook struct {
	mqtt.HookBase
	roomManager *screenshare.RoomManager
}

type participant struct {
	clientID string
	userID   string
}

type SignalingMessage struct {
	Type        string                 `json:"type"`
	Room        string                 `json:"room"`
	RoomId      string                 `json:"roomId"`
	UserData    string                 `json:"userdata"`
	AccessCodes []string               `json:"accessCodes"`
	ClientId    string                 `json:"clientId"`
	Payload     map[string]interface{} `json:"payload"`
	IceServers  map[string]interface{} `json:"iceServers"`
}

func NewBroker(port string, tlsConfig *tls.Config, validateToken func(token string) (string, error), iceServers []interface{}, roomManager *screenshare.RoomManager, hub NotificationHub) *Broker {
	return &Broker{
		port:        port,
		tlsConfig:   tlsConfig,
		authHook:    &AuthHook{validateToken: validateToken, roomManager: roomManager, iceServers: iceServers, hub: hub},
		aclHook:     &ACLHook{},
		RoomManager: roomManager,
	}
}

func (h *ConnectionHook) ID() string {
	return "rmfakecloud-connection"
}

func (h *ConnectionHook) Provides(b byte) bool {
	return b == mqtt.OnPacketRead ||
		b == mqtt.OnDisconnect
}

func (h *ConnectionHook) OnPacketRead(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	packetType := "unknown"
	switch pk.FixedHeader.Type {
	case packets.Connect:
		packetType = "CONNECT"
		log.Debugf("MQTT: CONNECT packet received client_id=%s remote=%s keepalive=%d clean=%t username=%s",
			cl.ID, cl.Net.Remote, pk.Connect.Keepalive, pk.Connect.Clean, string(pk.Connect.Username))
	case packets.Connack:
		packetType = "CONNACK"
	case packets.Publish:
		packetType = "PUBLISH"
	case packets.Subscribe:
		packetType = "SUBSCRIBE"
	case packets.Pingreq:
		packetType = "PINGREQ"
		if h.roomManager != nil {
			if roomID := h.roomManager.FindActiveRoom(string(cl.Properties.Username)); roomID != "" {
				h.roomManager.Keepalive(roomID)
			}
		}
	case packets.Disconnect:
		packetType = "DISCONNECT"
	}

	log.Debugf("MQTT: Packet type=%s(%d) client_id=%s", packetType, pk.FixedHeader.Type, cl.ID)
	return pk, nil
}

func (h *ConnectionHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	reason := "normal"
	if err != nil {
		reason = fmt.Sprintf("error: %v", err)
	} else if expire {
		reason = "expired"
	}

	remote := "unknown"
	if cl.Net.Conn != nil {
		remote = cl.Net.Conn.RemoteAddr().String()
	}

	hadConnect := len(cl.Properties.Username) > 0 || cl.Properties.ProtocolVersion > 0

	log.Warnf("MQTT: Disconnect client_id=%s remote=%s reason=%s had_connect=%t protocol_version=%d",
		cl.ID, remote, reason, hadConnect, cl.Properties.ProtocolVersion)

	if h.roomManager != nil {
		h.roomManager.RemoveParticipant(cl.ID)
	}
}

func (b *Broker) SetTLSConfig(tlsConfig *tls.Config) {
	b.tlsConfig = tlsConfig
}

func (b *Broker) Start() error {
	b.server = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	b.authHook.server = b.server

	connHook := &ConnectionHook{roomManager: b.RoomManager}
	if err := b.server.AddHook(connHook, nil); err != nil {
		return fmt.Errorf("failed to add connection hook: %w", err)
	}

	if err := b.server.AddHook(b.authHook, nil); err != nil {
		return fmt.Errorf("failed to add auth hook: %w", err)
	}

	if err := b.server.AddHook(b.aclHook, nil); err != nil {
		return fmt.Errorf("failed to add ACL hook: %w", err)
	}

	tlsConfig := b.tlsConfig
	if tlsConfig != nil {
		tlsConfig = tlsConfig.Clone()

		// Force TLS 1.2 to work around desktop app TLS 1.3 issue
		tlsConfig.MaxVersion = tls.VersionTLS12
		log.Infof("MQTT: Forcing TLS 1.2 maximum version")

		originalGetCertificate := tlsConfig.GetCertificate
		tlsConfig.GetCertificate = func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			log.Debugf("MQTT TLS: ClientHello from=%s server_name=%s protocols=%v cipher_count=%d",
				info.Conn.RemoteAddr(), info.ServerName, info.SupportedProtos, len(info.CipherSuites))

			if originalGetCertificate != nil {
				return originalGetCertificate(info)
			}
			if len(tlsConfig.Certificates) > 0 {
				return &tlsConfig.Certificates[0], nil
			}
			return nil, fmt.Errorf("no certificates configured")
		}

		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			log.Debugf("MQTT TLS: Handshake SUCCESS from=%s server_name=%s version=%d cipher=%d protocol=%s",
				state.ServerName, state.ServerName, state.Version, state.CipherSuite, state.NegotiatedProtocol)
			return nil
		}
	}

	tcpListener := listeners.NewTCP(listeners.Config{
		ID:        "tcp",
		Address:   ":" + b.port,
		TLSConfig: tlsConfig,
	})

	if err := b.server.AddListener(tcpListener); err != nil {
		return fmt.Errorf("failed to add TCP listener: %w", err)
	}

	if b.tlsConfig != nil {
		log.Infof("MQTT TCP listener with TLS on port %s", b.port)
	} else {
		log.Infof("MQTT TCP listener (plain) on port %s", b.port)
	}

	go func() {
		if err := b.server.Serve(); err != nil {
			log.Fatalf("MQTT server error: %v", err)
		}
	}()

	return nil
}

func (b *Broker) Stop() error {
	if b.server != nil {
		return b.server.Close()
	}
	return nil
}

func (b *Broker) PublishSignaling(userID, clientID string, payload []byte) {
	if b.server == nil {
		return
	}
	topic := fmt.Sprintf("user/%s/client/%s/signaling/screenshare", userID, clientID)
	b.server.Publish(topic, payload, false, 1)
}

func (b *Broker) HasConnectedClient(userID string) bool {
	if b.server == nil {
		return false
	}
	prefix := userID + "-"
	for _, cl := range b.server.Clients.GetAll() {
		if string(cl.Properties.Username) == userID || strings.HasPrefix(cl.ID, prefix) {
			return true
		}
	}
	return false
}

func (b *Broker) EstablishConnection(listenerID string, conn net.Conn) error {
	if b.server == nil {
		conn.Close()
		return fmt.Errorf("MQTT broker not started")
	}
	return b.server.EstablishConnection(listenerID, conn)
}

var ErrInvalidMessage = errors.New("message type not binary")

type WsConn struct {
	net.Conn
	C *websocket.Conn
	r io.Reader
}

func NewWsConn(wsConn *websocket.Conn) *WsConn {
	return &WsConn{Conn: wsConn.UnderlyingConn(), C: wsConn}
}

func (ws *WsConn) Read(p []byte) (int, error) {
	if ws.r == nil {
		op, r, err := ws.C.NextReader()
		if err != nil {
			return 0, err
		}
		if op != websocket.BinaryMessage {
			return 0, ErrInvalidMessage
		}
		ws.r = r
	}

	var n int
	for {
		if n == len(p) {
			return n, nil
		}
		br, err := ws.r.Read(p[n:])
		n += br
		if err != nil {
			ws.r = nil
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return n, err
		}
	}
}

func (ws *WsConn) Write(p []byte) (int, error) {
	err := ws.C.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (ws *WsConn) Close() error {
	return ws.Conn.Close()
}

func (h *AuthHook) ID() string {
	return "rmfakecloud-auth"
}

func (h *AuthHook) Provides(b byte) bool {
	return b == mqtt.OnConnectAuthenticate ||
		b == mqtt.OnConnect ||
		b == mqtt.OnDisconnect ||
		b == mqtt.OnSubscribe ||
		b == mqtt.OnUnsubscribe ||
		b == mqtt.OnPublish
}

func (h *AuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	log.Infof("MQTT: Authentication attempt from client_id=%s remote=%s", cl.ID, cl.Net.Remote)

	if h.validateToken == nil {
		log.Debug("MQTT: No token validator configured, allowing connection")
		return true
	}

	username := string(pk.Connect.Username)
	password := string(pk.Connect.Password)

	token := password
	if token == "" {
		token = username
	}

	userID, err := h.validateToken(token)
	if err != nil {
		log.Warnf("MQTT: Authentication FAILED client_id=%s remote=%s error=%v", cl.ID, cl.Net.Remote, err)
		return false
	}

	if userID == "" {
		log.Warnf("MQTT: Authentication FAILED client_id=%s remote=%s reason=empty_user_id", cl.ID, cl.Net.Remote)
		return false
	}

	cl.Properties.Username = []byte(userID)

	log.Infof("MQTT: Authentication SUCCESS client_id=%s user_id=%s remote=%s", cl.ID, userID, cl.Net.Remote)
	return true
}

func (h *AuthHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	userID := string(cl.Properties.Username)
	log.Debugf("MQTT: Client CONNECTED client_id=%s user_id=%s remote=%s clean_session=%t keepalive=%d",
		cl.ID, userID, cl.Net.Remote, pk.Connect.Clean, pk.Connect.Keepalive)
	return nil
}

func (h *AuthHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	userID := string(cl.Properties.Username)
	reason := "normal"
	if err != nil {
		reason = fmt.Sprintf("error: %v", err)
	} else if expire {
		reason = "expired"
	}
	log.Infof("MQTT: Client DISCONNECTED client_id=%s user_id=%s reason=%s", cl.ID, userID, reason)
}

func (h *AuthHook) OnSubscribe(cl *mqtt.Client, pk packets.Packet) packets.Packet {
	userID := string(cl.Properties.Username)
	for _, sub := range pk.Filters {
		log.Debugf("MQTT: Client SUBSCRIBE client_id=%s user_id=%s topic=%s qos=%d",
			cl.ID, userID, sub.Filter, sub.Qos)
	}
	return pk
}

func (h *AuthHook) OnUnsubscribe(cl *mqtt.Client, pk packets.Packet) packets.Packet {
	userID := string(cl.Properties.Username)
	for _, topic := range pk.Filters {
		log.Debugf("MQTT: Client UNSUBSCRIBE client_id=%s user_id=%s topic=%s",
			cl.ID, userID, topic.Filter)
	}
	return pk
}

func (h *AuthHook) handleSignalingMessage(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	switch msg.Type {
	case "create-room":
		h.handleCreateRoom(senderClientID, userID, msg, qos)
	case "join-auth-room", "join-active-room":
		h.handleJoinRoom(senderClientID, userID, msg, qos)
	case "broadcast":
		h.handleBroadcast(senderClientID, userID, msg, qos)
	case "direct":
		h.handleDirect(senderClientID, userID, msg, qos)
	default:
		log.Warnf("MQTT: Unknown signaling message type=%s from client_id=%s", msg.Type, senderClientID)
	}
}

func (h *AuthHook) handleCreateRoom(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	if h.roomManager == nil {
		return
	}

	// Only create MQTT room if no REST room exists (avoid duplicates on DualBroker)
	if existing := h.roomManager.FindActiveRoom(userID); existing != "" {
		log.Debugf("MQTT: skipping room creation, REST room already exists for user=%s", userID)
		// Still send room-created response so MQTT side is happy
		response := SignalingMessage{Type: "room-created", RoomId: existing}
		responseBytes, _ := json.Marshal(response)
		broadcastTopic := fmt.Sprintf("user/%s/signaling", userID)
		if h.server != nil {
			h.server.Publish(broadcastTopic, responseBytes, false, qos)
		}
		return
	}

	room := h.roomManager.CreateRoom(userID, senderClientID)

	response := SignalingMessage{
		Type:   "room-created",
		RoomId: room.RoomID,
	}

	broadcastTopic := fmt.Sprintf("user/%s/signaling", userID)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("MQTT: failed to marshal room-created: %v", err)
		return
	}

	if h.server != nil {
		h.server.Publish(broadcastTopic, responseBytes, false, qos)
	}
}

func (h *AuthHook) handleJoinRoom(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	if h.roomManager == nil {
		return
	}

	roomID := msg.RoomId
	if roomID == "" {
		roomID = h.roomManager.FindActiveRoom(userID)
	}

	if roomID == "" || !h.roomManager.RoomExists(roomID) {
		log.Infof("MQTT: room not found for join client=%s", senderClientID)
		response := SignalingMessage{Type: "room-not-found"}
		responseBytes, _ := json.Marshal(response)
		responseTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", userID, senderClientID, roomID)
		if h.server != nil {
			h.server.Publish(responseTopic, responseBytes, false, qos)
		}
		return
	}

	h.roomManager.AddParticipant(roomID, senderClientID, userID)

	iceServers := h.iceServers
	if iceServers == nil {
		iceServers = []interface{}{}
	}

	response := SignalingMessage{
		Type:   "room-joined",
		RoomId: roomID,
		IceServers: map[string]interface{}{
			"ice_servers": iceServers,
		},
	}
	h.sendResponse(senderClientID, userID, roomID, &response, qos)
}

func (h *AuthHook) getPeers(roomID, senderClientID string) []participant {
	clients := h.roomManager.GetClients(roomID)
	var peers []participant
	for _, c := range clients {
		if c.ClientID != senderClientID {
			peers = append(peers, participant{clientID: c.ClientID, userID: c.UserID})
		}
	}
	return peers
}

func (h *AuthHook) handleBroadcast(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	if h.roomManager == nil {
		return
	}

	roomID := msg.RoomId
	if roomID == "" {
		roomID = h.roomManager.FindActiveRoom(userID)
	}

	peers := h.getPeers(roomID, senderClientID)

	broadcastMsg := map[string]interface{}{
		"type":     "broadcast",
		"clientId": senderClientID,
		"payload":  msg.Payload,
	}

	msgBytes, err := json.Marshal(broadcastMsg)
	if err != nil {
		log.Errorf("MQTT: failed to marshal broadcast: %v", err)
		return
	}

	for _, peer := range peers {
		peerTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", peer.userID, peer.clientID, roomID)
		if h.server != nil {
			h.server.Publish(peerTopic, msgBytes, false, qos)
		}
	}


	if roomID != "" {
		payloadBytes, _ := json.Marshal(msg.Payload)
		h.roomManager.AddBroadcast(roomID, senderClientID, payloadBytes)
	}
}

func (h *AuthHook) handleDirect(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	targetClientID := msg.ClientId
	if targetClientID == "" {
		log.Warnf("MQTT: direct message missing clientId from sender=%s", senderClientID)
		return
	}

	roomID := msg.RoomId
	if roomID == "" && h.roomManager != nil {
		roomID = h.roomManager.FindActiveRoom(userID)
	}

	directMsg := map[string]interface{}{
		"type":     "direct",
		"clientId": senderClientID,
		"payload":  msg.Payload,
	}

	msgBytes, err := json.Marshal(directMsg)
	if err != nil {
		log.Errorf("MQTT: failed to marshal direct message: %v", err)
		return
	}

	peerTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", userID, targetClientID, roomID)
	if h.server != nil {
		h.server.Publish(peerTopic, msgBytes, false, qos)
	}


	if roomID != "" && h.roomManager != nil {
		payloadBytes, _ := json.Marshal(msg.Payload)
		h.roomManager.AddDirect(roomID, senderClientID, targetClientID, payloadBytes)
	}
}

func (h *AuthHook) sendResponse(clientID, userID, roomID string, response *SignalingMessage, qos byte) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("MQTT: failed to marshal response: %v", err)
		return
	}

	responseTopic := fmt.Sprintf("user/%s/client/%s/signaling/room/%s", userID, clientID, roomID)
	if h.server != nil {
		h.server.Publish(responseTopic, responseBytes, false, qos)
	}
}

func (h *AuthHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	userID := string(cl.Properties.Username)
	payloadSize := len(pk.Payload)
	log.Debugf("MQTT: Client PUBLISH client_id=%s user_id=%s topic=%s qos=%d size=%d retain=%t",
		cl.ID, userID, pk.TopicName, pk.FixedHeader.Qos, payloadSize, pk.FixedHeader.Retain)

	if log.GetLevel() >= log.DebugLevel && payloadSize > 0 && payloadSize < 1000 {
		log.Debugf("MQTT: Publish payload: %s", string(pk.Payload))
	}


	if roomID := h.roomManager.FindActiveRoom(userID); roomID != "" {
		h.roomManager.Keepalive(roomID)
	}

	if strings.HasPrefix(pk.TopicName, "remarkable/screenshare/signaling/user/") {
		parts := strings.Split(pk.TopicName, "/")
		if len(parts) >= 7 {
			topicUserID := parts[4]
			senderClientID := parts[6]

			var msg SignalingMessage
			if err := json.Unmarshal(pk.Payload, &msg); err != nil {
				log.Warnf("MQTT: Failed to parse signaling message from client_id=%s error=%v", senderClientID, err)
				return pk, nil
			}

			log.Debugf("MQTT: Screenshare signaling client_id=%s user_id=%s type=%s room=%s",
				senderClientID, topicUserID, msg.Type, msg.Room)

			h.handleSignalingMessage(senderClientID, topicUserID, &msg, pk.FixedHeader.Qos)
		}
	}

	// Only process messages from real MQTT clients (not inline), store only, don't re-publish.
	if cl.ID != "inline" && strings.HasPrefix(pk.TopicName, "user/") && strings.Contains(pk.TopicName, "/signaling") {
		var msg SignalingMessage
		if err := json.Unmarshal(pk.Payload, &msg); err == nil && (msg.Type == "direct" || msg.Type == "broadcast") {
			senderClientID := msg.ClientId
			if senderClientID == "" {
				senderClientID = cl.ID
			}
			roomID := h.roomManager.FindActiveRoom(userID)
			if roomID != "" {
				payloadBytes, _ := json.Marshal(msg.Payload)
				if msg.Type == "direct" {
					targetClientID := msg.ClientId
	
					parts := strings.Split(pk.TopicName, "/")
					if len(parts) >= 4 {
						targetClientID = parts[3]
					}
					h.roomManager.AddDirect(roomID, senderClientID, targetClientID, payloadBytes)
				} else {
					h.roomManager.AddBroadcast(roomID, senderClientID, payloadBytes)
				}
			}
		}
	}

	return pk, nil
}

func (h *ACLHook) ID() string {
	return "rmfakecloud-acl"
}

func (h *ACLHook) Provides(b byte) bool {
	return b == mqtt.OnACLCheck
}

func (h *ACLHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	userID := string(cl.Properties.Username)
	action := "READ"
	if write {
		action = "WRITE"
	}

	if userID == "" {
		log.Warnf("MQTT ACL: DENY client_id=%s action=%s topic=%s reason=no_user_id", cl.ID, action, topic)
		return false
	}

	allowed := h.checkAccess(userID, topic, write)

	if allowed {
		log.Debugf("MQTT ACL: ALLOW client_id=%s user_id=%s action=%s topic=%s", cl.ID, userID, action, topic)
	} else {
		log.Warnf("MQTT ACL: DENY client_id=%s user_id=%s action=%s topic=%s reason=not_authorized", cl.ID, userID, action, topic)
	}

	return allowed
}

func (h *ACLHook) checkAccess(userID, topic string, write bool) bool {
	userPrefix := "user/" + userID + "/"

	if strings.HasPrefix(topic, userPrefix) {
		log.Debugf("MQTT ACL: Matched user topic prefix user_id=%s topic=%s", userID, topic)
		return true
	}

	screenshareUserPrefix := "remarkable/screenshare/signaling/user/" + userID + "/"
	if strings.HasPrefix(topic, screenshareUserPrefix) {
		log.Debugf("MQTT ACL: Matched screenshare user topic user_id=%s topic=%s", userID, topic)
		return true
	}

	if !write && topic == "remarkable/screenshare/signaling" {
		log.Debugf("MQTT ACL: Matched public read topic user_id=%s topic=%s", userID, topic)
		return true
	}

	log.Debugf("MQTT ACL: No matching rule user_id=%s topic=%s write=%t", userID, topic, write)
	return false
}
