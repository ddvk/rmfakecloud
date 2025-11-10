package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	log "github.com/sirupsen/logrus"
)

type Broker struct {
	server    *mqtt.Server
	port      string
	tlsConfig *tls.Config
	authHook  *AuthHook
	aclHook   *ACLHook
}

type AuthHook struct {
	mqtt.HookBase
	validateToken func(token string) (userID string, err error)
	server        *mqtt.Server
	roomManager   *RoomManager
	iceServers    []interface{}
}

type ACLHook struct {
	mqtt.HookBase
}

type ConnectionHook struct {
	mqtt.HookBase
	roomManager *RoomManager
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*ScreenshareRoom
}

type ScreenshareRoom struct {
	name          string
	participants  map[string]*Participant
	creatorUserID string
	creatorDeviceID string
}

type Participant struct {
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

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*ScreenshareRoom),
	}
}

func (rm *RoomManager) GetOrCreateRoom(roomName, userID, deviceID string) *ScreenshareRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomName]
	if !exists {
		room = &ScreenshareRoom{
			name:            roomName,
			participants:    make(map[string]*Participant),
			creatorUserID:   userID,
			creatorDeviceID: deviceID,
		}
		rm.rooms[roomName] = room
		log.Infof("MQTT: Created new screenshare room=%s user_id=%s device_id=%s", roomName, userID, deviceID)
	}
	return room
}

func (rm *RoomManager) AddParticipant(roomName, clientID, userID string) bool {
	room := rm.GetOrCreateRoom(roomName, userID, clientID)
	wasNew := len(room.participants) == 0
	room.participants[clientID] = &Participant{
		clientID: clientID,
		userID:   userID,
	}
	log.Debugf("MQTT: Added participant to room=%s client_id=%s user_id=%s total_participants=%d",
		roomName, clientID, userID, len(room.participants))
	return wasNew
}

func (rm *RoomManager) RemoveParticipant(clientID string) (roomName, creatorUserID, creatorDeviceID string, wasLastParticipant bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for rName, room := range rm.rooms {
		if _, exists := room.participants[clientID]; exists {
			delete(room.participants, clientID)
			log.Debugf("MQTT: Removed participant from room=%s client_id=%s remaining=%d",
				rName, clientID, len(room.participants))

			if len(room.participants) == 0 {
				roomName = rName
				creatorUserID = room.creatorUserID
				creatorDeviceID = room.creatorDeviceID
				wasLastParticipant = true
				delete(rm.rooms, rName)
				log.Infof("MQTT: Removed empty room=%s", rName)
			}
			return
		}
	}
	return
}

func (rm *RoomManager) GetPeers(roomName, senderClientID string) []Participant {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomName]
	if !exists {
		return nil
	}

	var peers []Participant
	for clientID, participant := range room.participants {
		if clientID != senderClientID {
			peers = append(peers, *participant)
		}
	}
	return peers
}

func (rm *RoomManager) FindActiveRoom(userID string) string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for roomName := range rm.rooms {
		return roomName
	}
	return ""
}

func (rm *RoomManager) RoomExists(roomName string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	_, exists := rm.rooms[roomName]
	return exists
}

func NewBroker(port string, tlsConfig *tls.Config, validateToken func(token string) (string, error), iceServers []interface{}) *Broker {
	roomManager := NewRoomManager()
	return &Broker{
		port:      port,
		tlsConfig: tlsConfig,
		authHook:  &AuthHook{validateToken: validateToken, roomManager: roomManager, iceServers: iceServers},
		aclHook:   &ACLHook{},
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

func (b *Broker) Start() error {
	b.server = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	b.authHook.server = b.server

	connHook := &ConnectionHook{roomManager: b.authHook.roomManager}
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
	roomName := msg.Room
	if roomName == "" {
		roomName = "screenshare"
	}

	if h.roomManager != nil {
		h.roomManager.AddParticipant(roomName, senderClientID, userID)
	}

	response := SignalingMessage{
		Type:   "room-created",
		RoomId: roomName,
	}

	broadcastTopic := fmt.Sprintf("user/%s/signaling", userID)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("MQTT: Failed to marshal room-created response error=%v", err)
		return
	}

	log.Debugf("MQTT: Broadcasting room-created to all clients topic=%s room=%s", broadcastTopic, roomName)

	if h.server != nil {
		err := h.server.Publish(broadcastTopic, responseBytes, false, qos)
		if err != nil {
			log.Errorf("MQTT: Failed to broadcast room-created error=%v", err)
		}
	}
}

func (h *AuthHook) handleJoinRoom(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	roomName := msg.Room
	if roomName == "" {
		if h.roomManager != nil {
			roomName = h.roomManager.FindActiveRoom(userID)
		}
		if roomName == "" {
			roomName = "screenshare"
		}
	}

	if h.roomManager == nil || !h.roomManager.RoomExists(roomName) {
		log.Infof("MQTT: Join room failed, room does not exist room=%s client_id=%s", roomName, senderClientID)

		response := SignalingMessage{
			Type: "room-not-found",
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Errorf("MQTT: Failed to marshal room-not-found response error=%v", err)
			return
		}

		responseTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", userID, senderClientID, roomName)
		if h.server != nil {
			h.server.Publish(responseTopic, responseBytes, false, qos)
		}
		return
	}

	h.roomManager.AddParticipant(roomName, senderClientID, userID)

	iceServers := h.iceServers
	if iceServers == nil {
		iceServers = []interface{}{}
	}

	response := SignalingMessage{
		Type:   "room-joined",
		RoomId: roomName,
		IceServers: map[string]interface{}{
			"iceServers": iceServers,
		},
	}
	h.sendResponse(senderClientID, userID, roomName, &response, qos)
}

func (h *AuthHook) handleBroadcast(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	roomName := msg.Room
	if roomName == "" {
		if h.roomManager != nil {
			roomName = h.roomManager.FindActiveRoom(userID)
		}
	}

	var peers []Participant
	if h.roomManager != nil {
		peers = h.roomManager.GetPeers(roomName, senderClientID)
	}

	broadcastMsg := map[string]interface{}{
		"type":     "broadcast",
		"clientId": senderClientID,
		"payload":  msg.Payload,
	}

	msgBytes, err := json.Marshal(broadcastMsg)
	if err != nil {
		log.Errorf("MQTT: Failed to marshal broadcast message error=%v", err)
		return
	}

	log.Debugf("MQTT: Broadcasting message from client_id=%s to %d peers in room=%s",
		senderClientID, len(peers), roomName)

	for _, peer := range peers {
		peerTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", peer.userID, peer.clientID, roomName)
		if h.server != nil {
			err := h.server.Publish(peerTopic, msgBytes, false, qos)
			if err != nil {
				log.Errorf("MQTT: Failed to broadcast to peer=%s error=%v", peer.clientID, err)
			} else {
				log.Debugf("MQTT: Broadcast sent to peer=%s", peer.clientID)
			}
		}
	}
}

func (h *AuthHook) handleDirect(senderClientID, userID string, msg *SignalingMessage, qos byte) {
	targetClientID := msg.ClientId
	if targetClientID == "" {
		log.Warnf("MQTT: Direct message missing clientId from sender=%s", senderClientID)
		return
	}

	roomName := msg.Room
	if roomName == "" {
		if h.roomManager != nil {
			roomName = h.roomManager.FindActiveRoom(userID)
		}
	}

	directMsg := map[string]interface{}{
		"type":     "direct",
		"clientId": senderClientID,
		"payload":  msg.Payload,
	}

	msgBytes, err := json.Marshal(directMsg)
	if err != nil {
		log.Errorf("MQTT: Failed to marshal direct message error=%v", err)
		return
	}

	peerTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", userID, targetClientID, roomName)

	log.Debugf("MQTT: Sending direct message from client_id=%s to peer=%s room=%s",
		senderClientID, targetClientID, roomName)

	if h.server != nil {
		err := h.server.Publish(peerTopic, msgBytes, false, qos)
		if err != nil {
			log.Errorf("MQTT: Failed to send direct message to peer=%s error=%v", targetClientID, err)
		} else {
			log.Debugf("MQTT: Direct message sent to peer=%s", targetClientID)
		}
	}
}

func (h *AuthHook) sendResponse(clientID, userID, roomName string, response *SignalingMessage, qos byte) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("MQTT: Failed to marshal response error=%v", err)
		return
	}

	responseTopic := fmt.Sprintf("user/%s/client/%s/signaling/%s", userID, clientID, roomName)

	log.Debugf("MQTT: Sending response type=%s to client_id=%s topic=%s",
		response.Type, clientID, responseTopic)

	if h.server != nil {
		err := h.server.Publish(responseTopic, responseBytes, false, qos)
		if err != nil {
			log.Errorf("MQTT: Failed to send response to client=%s error=%v", clientID, err)
		} else {
			log.Debugf("MQTT: Response sent successfully to client=%s", clientID)
		}
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

	// Pattern: remarkable/screenshare/signaling/user/{userId}/client/{clientId}
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
