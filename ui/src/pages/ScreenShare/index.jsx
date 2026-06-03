
import { useState, useEffect, useRef, useCallback } from "react";
import { Container, Alert, Spinner, Button } from "react-bootstrap";
import { BsGearFill, BsFullscreenExit, BsArrowCounterclockwise, BsArrowClockwise } from "react-icons/bs";
import pako from "pako";
import constants from "../../common/constants";

const MSG = {
  FRAMEBUFFER: 0,
  CURSOR: 100,
  SHUTDOWN: 101,
  ROTATION: 102,
  PING: 103,
  HANDSHAKE: 104,
};

// Frame wire format: [type(1), rectCount(2 BE), deflatedSize(4 BE), zlib...]
const HEADER_SIZE = 7;
// Rect wire format (after inflate): [x(2), y(2), w(2), h(2), pixelDataLen(4), pixels...]
const RECT_HEADER_SIZE = 12;

// Reassembles framebuffer frames split across multiple data-channel messages;
// other message types pass through untouched.
class FrameAssembler {
  constructor() {
    this.reset();
  }

  reset() {
    this.buffer = null;
    this.expectedSize = 0;
    this.receivedSize = 0;
  }

  // Returns a complete message ArrayBuffer, or null while a frame is still arriving.
  process(data) {
    if (this.buffer) return this.appendChunk(data);
    if (data.byteLength < HEADER_SIZE) return data;
    const view = new DataView(data);
    if (view.getUint8(0) !== MSG.FRAMEBUFFER) return data;
    const total = HEADER_SIZE + view.getUint32(3, false);
    if (data.byteLength >= total) return data;
    this.expectedSize = total;
    this.buffer = new Uint8Array(total);
    this.buffer.set(new Uint8Array(data), 0);
    this.receivedSize = data.byteLength;
    return null;
  }

  appendChunk(data) {
    const chunk = new Uint8Array(data);
    const take = Math.min(chunk.byteLength, this.expectedSize - this.receivedSize);
    this.buffer.set(chunk.subarray(0, take), this.receivedSize);
    this.receivedSize += take;
    if (this.receivedSize < this.expectedSize) return null;
    const done = this.buffer.buffer;
    this.reset();
    return done;
  }
}

function decodeFramebuffer(data, view) {
  const rectCount = view.getUint16(1, false);
  const deflatedSize = view.getUint32(3, false);
  if (HEADER_SIZE + deflatedSize > data.byteLength) {
    throw new Error("deflated size exceeds message buffer");
  }
  const compressed = new Uint8Array(data.slice(HEADER_SIZE, HEADER_SIZE + deflatedSize));
  const raw = pako.inflate(compressed);
  if (!raw) throw new Error("inflate produced no output (truncated stream)");

  const rv = new DataView(raw.buffer, raw.byteOffset, raw.byteLength);
  const rects = [];
  let pos = 0;
  for (let i = 0; i < rectCount; i++) {
    if (pos + RECT_HEADER_SIZE > raw.byteLength) break;
    const x = rv.getUint16(pos, false);
    const y = rv.getUint16(pos + 2, false);
    const w = rv.getUint16(pos + 4, false);
    const h = rv.getUint16(pos + 6, false);
    const pxDataLen = rv.getUint32(pos + 8, false);
    if (pos + RECT_HEADER_SIZE + pxDataLen > raw.byteLength) break;

    const pxView = new DataView(raw.buffer, raw.byteOffset + pos + RECT_HEADER_SIZE, pxDataLen);
    const imgData = new ImageData(w, h);
    const out = imgData.data;
    const pixels = w * h;
    for (let p = 0; p < pixels; p++) {
      const val = pxView.getUint16(p * 2, true);
      const r5 = (val >> 11) & 0x1f;
      const g6 = (val >> 5) & 0x3f;
      const b5 = val & 0x1f;
      out[p * 4] = (r5 << 3) | (r5 >> 2);
      out[p * 4 + 1] = (g6 << 2) | (g6 >> 4);
      out[p * 4 + 2] = (b5 << 3) | (b5 >> 2);
      out[p * 4 + 3] = 255;
    }
    pos += RECT_HEADER_SIZE + pxDataLen;
    rects.push({ x, y, imageData: imgData });
  }
  return { type: MSG.FRAMEBUFFER, rects };
}

function parseMessage(data) {
  if (data.byteLength < 1) throw new Error("empty message");
  const view = new DataView(data);
  const type = view.getUint8(0);
  switch (type) {
    case MSG.HANDSHAKE:
      if (data.byteLength < HEADER_SIZE) throw new Error("short handshake");
      return { type, width: view.getUint16(3, false), height: view.getUint16(5, false) };
    case MSG.FRAMEBUFFER:
      return decodeFramebuffer(data, view);
    case MSG.CURSOR:
      return { type, x: view.getUint16(1, false), y: view.getUint16(3, false) };
    case MSG.ROTATION:
      return { type, degrees: (360 - view.getUint32(1, false)) % 360 };
    case MSG.SHUTDOWN:
    case MSG.PING:
      return { type };
    default:
      throw new Error(`unknown message type: ${type}`);
  }
}

const STATUS = {
  WAITING: "waiting",
  CONNECTING: "connecting",
  STREAMING: "streaming",
  ERROR: "error",
};

const BACKDROP_PRESETS = {
  white: "#FFFFFF",
  "off-white": "#F9F6F1",
  gray: "#2D2D2D",
  black: "#000000",
};

function getBackdrop() {
  return localStorage.getItem("screenshare-backdrop") || "black";
}

function setBackdropPref(val) {
  localStorage.setItem("screenshare-backdrop", val);
}

function getCustomColor() {
  return localStorage.getItem("screenshare-custom-color") || "#808080";
}

function setCustomColorPref(val) {
  localStorage.setItem("screenshare-custom-color", val);
}

function resolveBackdrop(pref) {
  return BACKDROP_PRESETS[pref] || pref;
}

function isLightColor(hex) {
  const c = hex.replace("#", "");
  const r = parseInt(c.substring(0, 2), 16);
  const g = parseInt(c.substring(2, 4), 16);
  const b = parseInt(c.substring(4, 6), 16);
  return (r * 299 + g * 587 + b * 114) / 1000 > 128;
}

function api(path, opts = {}) {
  return fetch(`${constants.ROOT_URL}/screenshare/${path}`, {
    headers: { "Content-Type": "application/json" },
    ...opts,
  });
}

export default function ScreenShare() {
  const [status, setStatus] = useState(STATUS.WAITING);
  const [errorMsg, setErrorMsg] = useState("");
  const [poppedOut, setPoppedOut] = useState(false);
  const [manualRotation, setManualRotation] = useState(0);
  const [backdrop, setBackdrop] = useState(getBackdrop);
  const [customColor, setCustomColor] = useState(getCustomColor);
  const [showControls, setShowControls] = useState(false);
  const [pinnedPosition, setPinnedPosition] = useState(null);
  const [controlsPosition, setControlsPosition] = useState("right");
  const controlsRef = useRef(null);
  const manualRotationRef = useRef(0);

  useEffect(() => {
    if (!showControls) return;
    const handler = (e) => {
      if (controlsRef.current && !controlsRef.current.contains(e.target)) {
        setShowControls(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [showControls]);

  useEffect(() => {
    if (!poppedOut) return;
    const check = () => {
      const canvas = videoRef.current;
      if (!canvas || !canvas.width || !canvas.height) return;
      const rot = ((manualRotation % 360) + 360) % 360;
      const isRotated = rot % 180 !== 0;
      const aspect = isRotated ? canvas.height / canvas.width : canvas.width / canvas.height;
      const vpAspect = (window.innerWidth - 50) / window.innerHeight;
      setControlsPosition(aspect < vpAspect ? "right" : "top");
    };
    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, [poppedOut, manualRotation]);
  const videoRef = useRef(null);
  const pcRef = useRef(null);
  const dcRef = useRef(null);
  const roomIdRef = useRef(null);
  const tabletClientIdRef = useRef(null);

  const cleanup = useCallback(() => {
    dcRef.current = null;
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }
    roomIdRef.current = null;
  }, []);

  const setupPeerConnection = useCallback(
    (iceServers) => {
      if (pcRef.current) pcRef.current.close();

      const config = {};
      if (iceServers?.length) {
        config.iceServers = iceServers.map((s) => ({
          urls: s.url || s.urls,
          username: s.username,
          credential: s.credential,
        }));
      }

      const pc = new RTCPeerConnection(config);
      pcRef.current = pc;

      pc.ondatachannel = (event) => {
        const dc = event.channel;
        dc.binaryType = "arraybuffer";
        dcRef.current = dc;

        let screenWidth = 0;
        let screenHeight = 0;
        const canvas = videoRef.current;
        let ctx = null;
        let lastPenX = 0;
        let lastPenY = 0;
        let rotation = 0;
        const assembler = new FrameAssembler();

        function updateCursor() {
          const cursor = document.getElementById("pen-cursor");
          if (!cursor || !canvas) return;
          if (lastPenX === 0 && lastPenY === 0) {
            cursor.style.display = "none";
            return;
          }
          const rect = canvas.getBoundingClientRect();
          const cw = canvas.width;
          const ch = canvas.height;
          const rot = ((manualRotationRef.current % 360) + 360) % 360;
          const isLandscape = rot === 90 || rot === 270;
          const renderedW = isLandscape ? rect.height : rect.width;
          const renderedH = isLandscape ? rect.width : rect.height;
          const scaleX = renderedW / cw;
          const scaleY = renderedH / ch;

          const cx = (lastPenX - cw / 2) * scaleX;
          const cy = (lastPenY - ch / 2) * scaleY;

          const rad = rot * Math.PI / 180;
          const rx = cx * Math.cos(rad) - cy * Math.sin(rad);
          const ry = cx * Math.sin(rad) + cy * Math.cos(rad);

          const screenX = rect.left + rect.width / 2 + rx;
          const screenY = rect.top + rect.height / 2 + ry;

          cursor.style.display = "block";
          cursor.style.left = (screenX - 4) + "px";
          cursor.style.top = (screenY - 4) + "px";
        }

        dc.onopen = () => {
          const header = new TextEncoder().encode("reMarkable");
          const buf = new ArrayBuffer(header.length + 2);
          new Uint8Array(buf).set(header);
          dc.send(buf);
          pc.getStats().then(stats => {
            stats.forEach(s => {
              if (s.type === "candidate-pair" && s.state === "succeeded") {
                const local = stats.get(s.localCandidateId);
                const remote = stats.get(s.remoteCandidateId);
                console.log("[screenshare] peer:", remote?.address, s.nominated ? "(active)" : "");
              }
            });
          });
        };

        dc.onmessage = (e) => {
          const data = e.data;
          if (!(data instanceof ArrayBuffer)) return;

          const complete = assembler.process(data);
          if (!complete) return;

          try {
            const msg = parseMessage(complete);
            switch (msg.type) {
              case MSG.HANDSHAKE: {
                screenWidth = msg.width;
                screenHeight = msg.height;
                if (canvas) {
                  canvas.width = screenWidth;
                  canvas.height = screenHeight;
                  ctx = canvas.getContext("2d", { willReadFrequently: true });
                }
                setStatus(STATUS.STREAMING);
                break;
              }
              case MSG.FRAMEBUFFER: {
                if (!ctx) break;
                for (const rect of msg.rects) {
                  const { width: w, height: h } = rect.imageData;
                  if (rect.x === 0 && rect.y === 0 &&
                      w >= screenWidth * 0.9 && h >= screenHeight * 0.9) {
                    screenWidth = w;
                    screenHeight = h;
                    canvas.width = screenWidth;
                    canvas.height = screenHeight;
                  }
                  ctx.putImageData(rect.imageData, rect.x, rect.y);
                }
                break;
              }
              case MSG.CURSOR: {
                if (!ctx) break;
                lastPenX = msg.x;
                lastPenY = msg.y;
                updateCursor();
                break;
              }
              case MSG.ROTATION: {
                const newRotation = msg.degrees;
                if (newRotation !== rotation) {
                  const prev = rotation;
                  rotation = newRotation;
                  setManualRotation(r => r + ((newRotation - prev + 540) % 360 - 180));
                }
                break;
              }
              case MSG.SHUTDOWN:
              case MSG.PING:
                break;
            }
          } catch (err) {
            console.error("[screenshare] message error:", err);
            assembler.reset();
          }
        };

        dc.onclose = () => {
          if (!disconnectedRef.current) {
            setStatus(STATUS.WAITING);
          }
        };
      };

      // Don't trickle ICE candidates. The tablet crashes on mDNS candidates.
      // The SDP answer already contains our ICE info for LAN connectivity.
      pc.onicecandidate = () => {};

      pc.oniceconnectionstatechange = () =>

      pc.onconnectionstatechange = () => {

        if (
          pc.connectionState === "failed" ||
          pc.connectionState === "disconnected"
        ) {
          cleanup();
          if (!disconnectedRef.current) {
            setStatus(STATUS.WAITING);
          }
        }
      };

      return pc;
    },
    [cleanup]
  );

  const joinRoom = useCallback(async () => {
    try {
      setStatus(STATUS.CONNECTING);

      const offerRes = await api("offer");
      if (!offerRes.ok) throw new Error("Failed to get offer from device");

      const data = await offerRes.json();
      roomIdRef.current = data.roomId;

      const pc = setupPeerConnection(data.iceServers);

      const msgs = data.messages || [];
      const offerMsg = msgs.find(m => {
        let p = m.payload;
        if (p && p.type === "webtrc" && p.payload) p = p.payload;
        return p && p.type === "offer";
      });

      if (!offerMsg) throw new Error("No offer received from device");

      if (offerMsg.clientId) tabletClientIdRef.current = offerMsg.clientId;
      let offerPayload = offerMsg.payload;
      if (offerPayload.type === "webtrc") offerPayload = offerPayload.payload;

      const sdp = offerPayload.description || offerPayload.sdp;
      await pc.setRemoteDescription(new RTCSessionDescription({ type: "offer", sdp }));

      for (const msg of msgs) {
        let inner = msg.payload;
        if (!inner) continue;
        if (inner.type === "webtrc" && inner.payload) inner = inner.payload;
        if (inner.type === "candidate") {
          await pc.addIceCandidate(new RTCIceCandidate({
            candidate: inner.candidate,
            sdpMid: inner.mid || "0",
          }));
        }
      }

      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);

      // Wait for ICE gathering to complete so relay candidates are in the SDP
      if (pc.iceGatheringState !== "complete") {
        await new Promise((resolve) => {
          const check = () => {
            if (pc.iceGatheringState === "complete") {
              pc.removeEventListener("icegatheringstatechange", check);
              resolve();
            }
          };
          pc.addEventListener("icegatheringstatechange", check);
          setTimeout(resolve, 5000);
        });
      }

      await api(`room/${data.roomId}/answer`, {
        method: "POST",
        body: JSON.stringify({
          targetClientId: offerMsg.clientId,
          payload: {
            type: "webtrc",
            payload: { type: "answer", description: pc.localDescription.sdp },
          },
        }),
      });
    } catch (e) {
      setErrorMsg(e.message);
      setStatus(STATUS.ERROR);
    }
  }, [setupPeerConnection]);

  useEffect(() => {
    if (status !== STATUS.WAITING) return;

    const check = setInterval(async () => {
      const r = await api("room").catch(() => null);
      if (r?.ok) {
        clearInterval(check);
        joinRoom();
      }
    }, 2000);

    return () => clearInterval(check);
  }, [status, joinRoom]);

  useEffect(() => { manualRotationRef.current = manualRotation; }, [manualRotation]);
  useEffect(() => cleanup, [cleanup]);

  const disconnectedRef = useRef(false);

  const disconnect = () => {
    if (dcRef.current && dcRef.current.readyState === "open") {
      const buf = new ArrayBuffer(4);
      new DataView(buf).setInt32(0, 0x65, false);
      dcRef.current.send(buf);
    }
    cleanup();
    setPoppedOut(false);
    setShowControls(false);
    disconnectedRef.current = true;
    setStatus(STATUS.ERROR);
    setErrorMsg("Disconnected. Start a new screenshare session from the tablet to reconnect.");
  };

  const reconnect = () => {
    disconnectedRef.current = false;
    setStatus(STATUS.WAITING);
  };


  return (
    <Container className="mt-4">
      {status === STATUS.ERROR && (
        <Alert variant="info">
          {errorMsg}
          {disconnectedRef.current && (
            <Button variant="outline-primary" size="sm" className="ms-3" onClick={reconnect}>
              Reconnect
            </Button>
          )}
        </Alert>
      )}

      {status === STATUS.WAITING && (
        <Alert variant="info">
          <Spinner animation="border" size="sm" className="me-2" />
          Waiting for reMarkable to start screen sharing...
        </Alert>
      )}

      {status === STATUS.CONNECTING && (
        <Alert variant="info">
          <Spinner animation="border" size="sm" className="me-2" />
          Connecting to reMarkable...
        </Alert>
      )}

      <div
        style={poppedOut ? {
          position: "fixed",
          inset: 0,
          zIndex: 9999,
          background: resolveBackdrop(backdrop),
          display: status === STATUS.STREAMING ? "flex" : "none",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          paddingTop: controlsPosition === "top" ? 40 : 0,
          paddingRight: controlsPosition === "right" ? 50 : 0,
        } : {
          display: status === STATUS.STREAMING ? "flex" : "none",
          flexDirection: "column",
          alignItems: "center",
        }}
      >
        {(() => {
          const isRotated = ((manualRotation % 360) + 360) % 360 % 180 !== 0;
          const canvasStyle = {
            background: "#000",
            borderRadius: poppedOut ? 0 : "4px",
            transform: `rotate(${manualRotation}deg)`,
            transition: "transform 0.3s ease",
          };
          if (poppedOut) {
            const pad = controlsPosition === "right" ? 56 : 0;
            const topPad = controlsPosition === "top" ? 48 : 0;
            if (isRotated) {
              canvasStyle.maxWidth = `calc(100vh - ${topPad}px)`;
              canvasStyle.maxHeight = `calc(100vw - ${pad}px)`;
            } else {
              canvasStyle.maxWidth = `calc(100vw - ${pad}px)`;
              canvasStyle.maxHeight = `calc(100vh - ${topPad}px)`;
            }
            canvasStyle.width = "auto";
          } else if (isRotated) {
            canvasStyle.height = "70vh";
            canvasStyle.width = "auto";
          } else {
            canvasStyle.width = "100%";
            canvasStyle.maxWidth = "min(600px, 100%)";
          }
          return <canvas ref={videoRef} style={canvasStyle} />;
        })()}
        <div
          id="pen-cursor"
          style={{
            display: "none",
            position: "fixed",
            width: 8,
            height: 8,
            borderRadius: "50%",
            backgroundColor: "red",
            pointerEvents: "none",
            zIndex: 10000,
          }}
        />
        {(() => {
          const light = poppedOut && isLightColor(resolveBackdrop(backdrop));
          const btnVar = poppedOut ? (light ? "outline-dark" : "outline-light") : "outline-secondary";
          if (poppedOut) {
            return (
              <div ref={controlsRef} style={{
                position: "fixed",
                top: 8,
                right: 8,
                zIndex: 10001,
                display: "flex",
                flexDirection: "column",
                alignItems: controlsPosition === "right" ? "center" : "flex-end",
                gap: 4,
                maxHeight: "calc(100vh - 16px)",
                overflowY: "auto",
              }}>
                <div style={{ display: "flex", flexDirection: controlsPosition === "right" ? "column" : "row", gap: 4 }}>
                  {controlsPosition === "right" ? (
                    <>
                      <Button variant={btnVar} size="sm" onClick={() => setPoppedOut(false)} title="Exit fullscreen">
                        <BsFullscreenExit />
                      </Button>
                      <Button variant={btnVar} size="sm" onClick={() => { setShowControls((s) => { if (!s) setPinnedPosition(controlsPosition); return !s; }); }} title="Options">
                        <BsGearFill />
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button variant={btnVar} size="sm" onClick={() => { setShowControls((s) => { if (!s) setPinnedPosition(controlsPosition); return !s; }); }} title="Options">
                        <BsGearFill />
                      </Button>
                      <Button variant={btnVar} size="sm" onClick={() => setPoppedOut(false)} title="Exit fullscreen">
                        <BsFullscreenExit />
                      </Button>
                    </>
                  )}
                </div>
                {showControls && (
                  <div style={{
                    position: "fixed",
                    top: (pinnedPosition || controlsPosition) === "right" ? 90 : 8,
                    right: (pinnedPosition || controlsPosition) === "right" ? 8 : 90,
                    zIndex: 10002,
                    display: "flex", flexDirection: "column", gap: 4,
                    background: light ? "rgba(255,255,255,0.95)" : "rgba(0,0,0,0.9)",
                    borderRadius: 6, padding: 8,
                    alignItems: "stretch", minWidth: 160,
                  }}>
                    <div style={{ display: "flex", gap: 4 }}>
                      <Button variant={btnVar} size="sm" onClick={() => setManualRotation((r) => r - 90)} title="Rotate counter-clockwise" style={{ flex: 1, whiteSpace: "nowrap" }}>
                        <BsArrowCounterclockwise /> Rotate L
                      </Button>
                      <Button variant={btnVar} size="sm" onClick={() => setManualRotation((r) => r + 90)} title="Rotate clockwise" style={{ flex: 1, whiteSpace: "nowrap" }}>
                        <BsArrowClockwise /> Rotate R
                      </Button>
                    </div>
                    <Button variant={btnVar} size="sm" onClick={disconnect} title="End screenshare session">
                      Disconnect
                    </Button>
                    <div style={{ display: "flex", gap: 4, alignItems: "center", justifyContent: "center", marginTop: 2 }}>
                      {Object.entries(BACKDROP_PRESETS).map(([name, color]) => (
                        <button
                          key={name}
                          title={name}
                          onClick={() => { setBackdrop(name); setBackdropPref(name); }}
                          style={{
                            width: 22, height: 22, borderRadius: 3,
                            border: backdrop === name ? "2px solid #0d6efd" : "1px solid #666",
                            background: color, cursor: "pointer", padding: 0,
                          }}
                        />
                      ))}
                      <label title="Custom color" style={{ position: "relative", width: 22, height: 22, cursor: "pointer" }}>
                        <span style={{
                          display: "block", width: 22, height: 22, borderRadius: 3,
                          border: !BACKDROP_PRESETS[backdrop] ? "2px solid #0d6efd" : "1px solid #666",
                          background: "conic-gradient(red, yellow, lime, aqua, blue, magenta, red)",
                        }} />
                        <input
                          type="color"
                          value={customColor}
                          onClick={() => { setBackdrop(customColor); setBackdropPref(customColor); }}
                          onInput={(e) => {
                            setCustomColor(e.target.value);
                            setCustomColorPref(e.target.value);
                            setBackdrop(e.target.value);
                            setBackdropPref(e.target.value);
                          }}
                          style={{ position: "absolute", inset: 0, opacity: 0, cursor: "pointer", width: "100%", height: "100%" }}
                        />
                      </label>
                    </div>
                  </div>
                )}
              </div>
            );
          }
          return (
            <div style={{ marginTop: 8, display: "flex", alignItems: "center", gap: 6 }}>
              <Button variant={btnVar} size="sm" onClick={() => setManualRotation((r) => r - 90)} title="Rotate left">
                <BsArrowCounterclockwise />
              </Button>
              <Button variant={btnVar} size="sm" onClick={() => setManualRotation((r) => r + 90)} title="Rotate right">
                <BsArrowClockwise />
              </Button>
              <Button variant={btnVar} size="sm" onClick={() => setPoppedOut(true)}>
                Fullscreen
              </Button>
              <Button variant={btnVar} size="sm" onClick={disconnect}>
                Disconnect
              </Button>
            </div>
          );
        })()}
      </div>
    </Container>
  );
}
