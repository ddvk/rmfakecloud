(function () {
  "use strict";

  var API_BASE = "/ui/api/screenshare/";
  var STATUS = {
    WAITING: "waiting",
    CONNECTING: "connecting",
    STREAMING: "streaming",
    ERROR: "error",
  };
  var BACKDROP_PRESETS = {
    white: "#FFFFFF",
    "off-white": "#F9F6F1",
    gray: "#2D2D2D",
    black: "#000000",
  };

  var status = STATUS.WAITING;
  var errorMsg = "";
  var poppedOut = false;
  var manualRotation = 0;
  var backdrop = localStorage.getItem("screenshare-backdrop") || "black";
  var customColor = localStorage.getItem("screenshare-custom-color") || "#808080";
  var showControls = false;
  var pinnedPosition = null;
  var controlsPosition = "right";

  var pc = null;
  var dc = null;
  var roomId = null;
  var tabletClientId = null;
  var disconnected = false;
  var pollTimer = null;
  var canvas = null;
  var statusEl = null;
  var controlsEl = null;
  var optionsPanel = null;
  var stageEl = null;

  function setBackdropPref(val) {
    localStorage.setItem("screenshare-backdrop", val);
  }
  function setCustomColorPref(val) {
    localStorage.setItem("screenshare-custom-color", val);
  }
  function resolveBackdrop(pref) {
    return BACKDROP_PRESETS[pref] || pref;
  }
  function isLightColor(hex) {
    var c = String(hex || "").replace("#", "");
    if (c.length !== 6) return false;
    var r = parseInt(c.substring(0, 2), 16);
    var g = parseInt(c.substring(2, 4), 16);
    var b = parseInt(c.substring(4, 6), 16);
    return (r * 299 + g * 587 + b * 114) / 1000 > 128;
  }

  function api(path, opts) {
    opts = opts || {};
    var headers = Object.assign({ "Content-Type": "application/json" }, opts.headers || {});
    return fetch(API_BASE + path, Object.assign({ credentials: "same-origin", headers: headers }, opts));
  }

  function setStatus(next, msg) {
    status = next;
    if (typeof msg === "string") errorMsg = msg;
    renderUI();
    if (status === STATUS.WAITING) startPolling();
  }

  function cleanup() {
    dc = null;
    if (pc) {
      pc.close();
      pc = null;
    }
    roomId = null;
  }

  function updateCursor(lastPenX, lastPenY) {
    var cursor = document.getElementById("pen-cursor");
    if (!cursor || !canvas) return;
    if (lastPenX === 0 && lastPenY === 0) {
      cursor.style.display = "none";
      return;
    }
    var rect = canvas.getBoundingClientRect();
    var cw = canvas.width;
    var ch = canvas.height;
    var rot = ((manualRotation % 360) + 360) % 360;
    var isLandscape = rot === 90 || rot === 270;
    var renderedW = isLandscape ? rect.height : rect.width;
    var renderedH = isLandscape ? rect.width : rect.height;
    var scaleX = renderedW / cw;
    var scaleY = renderedH / ch;

    var cx = (lastPenX - cw / 2) * scaleX;
    var cy = (lastPenY - ch / 2) * scaleY;

    var rad = (rot * Math.PI) / 180;
    var rx = cx * Math.cos(rad) - cy * Math.sin(rad);
    var ry = cx * Math.sin(rad) + cy * Math.cos(rad);

    var screenX = rect.left + rect.width / 2 + rx;
    var screenY = rect.top + rect.height / 2 + ry;

    cursor.style.display = "block";
    cursor.style.left = screenX - 4 + "px";
    cursor.style.top = screenY - 4 + "px";
  }

  function setupPeerConnection(iceServers) {
    if (pc) pc.close();

    var config = {};
    if (iceServers && iceServers.length) {
      config.iceServers = iceServers.map(function (s) {
        return {
          urls: s.url || s.urls,
          username: s.username,
          credential: s.credential,
        };
      });
    }

    pc = new RTCPeerConnection(config);

    pc.ondatachannel = function (event) {
      var channel = event.channel;
      channel.binaryType = "arraybuffer";
      dc = channel;

      var screenWidth = 0;
      var screenHeight = 0;
      var ctx = null;
      var lastPenX = 0;
      var lastPenY = 0;
      var rotation = 0;
      var frameQueue = [];
      var rafPending = false;
      var pendingBuffer = null;
      var pendingExpected = 0;
      var pendingReceived = 0;

      channel.onopen = function () {
        var header = new TextEncoder().encode("reMarkable");
        var buf = new ArrayBuffer(header.length + 2);
        new Uint8Array(buf).set(header);
        channel.send(buf);
        pc.getStats().then(function (stats) {
          stats.forEach(function (s) {
            if (s.type === "candidate-pair" && s.state === "succeeded") {
              var remote = stats.get(s.remoteCandidateId);
              console.log("[screenshare] peer:", remote && remote.address, s.nominated ? "(active)" : "");
            }
          });
        });
      };

      channel.onmessage = function (e) {
        var data = e.data;
        if (!(data instanceof ArrayBuffer)) return;
        var bytes = new Uint8Array(data);
        if (bytes.length === 0) return;
        var pako = window.pako;
        if (!pako) {
          console.error("[screenshare] window.pako is required");
          return;
        }

        if (pendingBuffer && pendingExpected > 0) {
          var chunk = new Uint8Array(data);
          var remaining = pendingExpected - pendingReceived;
          var take = Math.min(chunk.byteLength, remaining);
          pendingBuffer.set(chunk.subarray(0, take), pendingReceived);
          pendingReceived += take;
          if (pendingReceived < pendingExpected) return;
        }

        var msgType = pendingBuffer ? 0x00 : bytes[0];

        if (msgType === 0x68 && bytes.length >= 7) {
          var view = new DataView(data);
          screenWidth = view.getUint16(3, false);
          screenHeight = view.getUint16(5, false);
          if (canvas) {
            canvas.width = screenWidth;
            canvas.height = screenHeight;
            ctx = canvas.getContext("2d", { willReadFrequently: true });
          }
          setStatus(STATUS.STREAMING);
          return;
        }

        if (msgType === 0x67) return;

        if (msgType === 0x66 && bytes.length >= 5) {
          var rv = new DataView(data, 1);
          var rawDeg = rv.getUint32(0, false);
          var newRotation = (360 - rawDeg) % 360;
          if (newRotation !== rotation) {
            var prev = rotation;
            rotation = newRotation;
            var delta = ((newRotation - prev + 540) % 360) - 180;
            manualRotation = manualRotation + delta;
            renderUI();
          }
          return;
        }

        if (msgType === 0x64 && bytes.length === 5 && ctx) {
          lastPenX = (bytes[1] << 8) | bytes[2];
          lastPenY = (bytes[3] << 8) | bytes[4];
          updateCursor(lastPenX, lastPenY);
          return;
        }

        if ((pendingBuffer && pendingReceived >= pendingExpected) || (msgType === 0x00 && bytes.length > 7 && ctx)) {
          var frameBytes;
          if (pendingBuffer && pendingReceived >= pendingExpected) {
            frameBytes = pendingBuffer;
            pendingBuffer = null;
            pendingExpected = 0;
            pendingReceived = 0;
          } else {
            var declaredSize = new DataView(data, 3, 4).getUint32(0, false);
            var expectedLen = 7 + declaredSize;
            if (bytes.length < expectedLen) {
              pendingBuffer = new Uint8Array(expectedLen);
              pendingBuffer.set(new Uint8Array(data), 0);
              pendingExpected = expectedLen;
              pendingReceived = bytes.length;
              return;
            }
            frameBytes = new Uint8Array(bytes);
          }

          frameQueue.push(frameBytes);
          if (!rafPending) {
            rafPending = true;
            requestAnimationFrame(function () {
              rafPending = false;
              while (frameQueue.length > 0) {
                var frame = frameQueue.shift();
                try {
                  var fv = new DataView(frame.buffer, frame.byteOffset, frame.byteLength);
                  var rectCount = fv.getUint16(1, false);
                  var deflatedSize = fv.getUint32(3, false);
                  if (7 + deflatedSize > frame.byteLength) continue;
                  var compressed = new Uint8Array(
                    frame.buffer.slice(frame.byteOffset + 7, frame.byteOffset + 7 + deflatedSize)
                  );
                  var raw = pako.inflate(compressed);
                  if (!raw || raw.length < 12) continue;

                  var rawView = new DataView(raw.buffer);
                  var pos = 0;
                  for (var ri = 0; ri < rectCount; ri++) {
                    if (pos + 12 > raw.byteLength) break;
                    var regionX = rawView.getUint16(pos, false);
                    var regionY = rawView.getUint16(pos + 2, false);
                    var regionW = rawView.getUint16(pos + 4, false);
                    var regionH = rawView.getUint16(pos + 6, false);
                    var pxDataLen = rawView.getUint32(pos + 8, false);
                    if (pos + 12 + pxDataLen > raw.byteLength) break;

                    var pxView = new DataView(raw.buffer, pos + 12, pxDataLen);
                    var imgData = new ImageData(regionW, regionH);
                    var out = imgData.data;
                    var pixels = regionW * regionH;
                    for (var i = 0; i < pixels; i++) {
                      var val = pxView.getUint16(i * 2, true);
                      var r5 = (val >> 11) & 0x1f;
                      var g6 = (val >> 5) & 0x3f;
                      var b5 = val & 0x1f;
                      out[i * 4] = (r5 << 3) | (r5 >> 2);
                      out[i * 4 + 1] = (g6 << 2) | (g6 >> 4);
                      out[i * 4 + 2] = (b5 << 3) | (b5 >> 2);
                      out[i * 4 + 3] = 255;
                    }
                    pos += 12 + pxDataLen;

                    if (
                      regionX === 0 &&
                      regionY === 0 &&
                      regionW >= screenWidth * 0.9 &&
                      regionH >= screenHeight * 0.9
                    ) {
                      screenWidth = regionW;
                      screenHeight = regionH;
                      canvas.width = screenWidth;
                      canvas.height = screenHeight;
                    }
                    ctx.putImageData(imgData, regionX, regionY);
                  }
                } catch (err) {
                  console.error("[screenshare] frame error:", err);
                  pendingBuffer = null;
                  pendingExpected = 0;
                  pendingReceived = 0;
                }
              }
              updateCursor(lastPenX, lastPenY);
            });
          }
        }
      };

      channel.onclose = function () {
        if (!disconnected) setStatus(STATUS.WAITING);
      };
    };

    // Don't trickle ICE candidates. The tablet crashes on mDNS candidates.
    pc.onicecandidate = function () {};

    pc.onconnectionstatechange = function () {
      if (pc.connectionState === "failed" || pc.connectionState === "disconnected") {
        cleanup();
        if (!disconnected) setStatus(STATUS.WAITING);
      }
    };

    return pc;
  }

  async function joinRoom() {
    try {
      setStatus(STATUS.CONNECTING);

      var offerRes = await api("offer");
      if (!offerRes.ok) throw new Error("Failed to get offer from device");

      var data = await offerRes.json();
      roomId = data.roomId;

      var peer = setupPeerConnection(data.iceServers);

      var msgs = data.messages || [];
      var offerMsg = msgs.find(function (m) {
        var p = m.payload;
        if (p && p.type === "webtrc" && p.payload) p = p.payload;
        return p && p.type === "offer";
      });

      if (!offerMsg) throw new Error("No offer received from device");

      if (offerMsg.clientId) tabletClientId = offerMsg.clientId;
      var offerPayload = offerMsg.payload;
      if (offerPayload.type === "webtrc") offerPayload = offerPayload.payload;

      var sdp = offerPayload.description || offerPayload.sdp;
      await peer.setRemoteDescription(new RTCSessionDescription({ type: "offer", sdp: sdp }));

      for (var mi = 0; mi < msgs.length; mi++) {
        var msg = msgs[mi];
        var inner = msg.payload;
        if (!inner) continue;
        if (inner.type === "webtrc" && inner.payload) inner = inner.payload;
        if (inner.type === "candidate") {
          await peer.addIceCandidate(
            new RTCIceCandidate({
              candidate: inner.candidate,
              sdpMid: inner.mid || "0",
            })
          );
        }
      }

      var answer = await peer.createAnswer();
      await peer.setLocalDescription(answer);

      if (peer.iceGatheringState !== "complete") {
        await new Promise(function (resolve) {
          function check() {
            if (peer.iceGatheringState === "complete") {
              peer.removeEventListener("icegatheringstatechange", check);
              resolve();
            }
          }
          peer.addEventListener("icegatheringstatechange", check);
          setTimeout(resolve, 5000);
        });
      }

      await api("room/" + data.roomId + "/answer", {
        method: "POST",
        body: JSON.stringify({
          targetClientId: offerMsg.clientId,
          payload: {
            type: "webtrc",
            payload: { type: "answer", description: peer.localDescription.sdp },
          },
        }),
      });
    } catch (e) {
      setStatus(STATUS.ERROR, e.message || String(e));
    }
  }

  function startPolling() {
    if (pollTimer) clearInterval(pollTimer);
    if (status !== STATUS.WAITING) return;
    pollTimer = setInterval(async function () {
      if (status !== STATUS.WAITING) {
        clearInterval(pollTimer);
        pollTimer = null;
        return;
      }
      var r = await api("room").catch(function () {
        return null;
      });
      if (r && r.ok) {
        clearInterval(pollTimer);
        pollTimer = null;
        joinRoom();
      }
    }, 2000);
  }

  function disconnect() {
    if (dc && dc.readyState === "open") {
      var buf = new ArrayBuffer(4);
      new DataView(buf).setInt32(0, 0x65, false);
      dc.send(buf);
    }
    cleanup();
    poppedOut = false;
    showControls = false;
    disconnected = true;
    setStatus(STATUS.ERROR, "Disconnected. Start a new screenshare session from the tablet to reconnect.");
  }

  function reconnect() {
    disconnected = false;
    errorMsg = "";
    setStatus(STATUS.WAITING);
  }

  function updateControlsPosition() {
    if (!poppedOut || !canvas || !canvas.width || !canvas.height) return;
    var rot = ((manualRotation % 360) + 360) % 360;
    var isRotated = rot % 180 !== 0;
    var aspect = isRotated ? canvas.height / canvas.width : canvas.width / canvas.height;
    var vpAspect = (window.innerWidth - 50) / window.innerHeight;
    controlsPosition = aspect < vpAspect ? "right" : "top";
    renderUI();
  }

  function applyCanvasStyle() {
    if (!canvas) return;
    var rot = ((manualRotation % 360) + 360) % 360;
    var isRotated = rot % 180 !== 0;
    canvas.style.background = "#000";
    canvas.style.borderRadius = poppedOut ? "0" : "4px";
    canvas.style.transform = "rotate(" + manualRotation + "deg)";
    canvas.style.transition = "transform 0.3s ease";
    canvas.style.display = status === STATUS.STREAMING ? "block" : "none";

    if (poppedOut) {
      var pad = controlsPosition === "right" ? 56 : 0;
      var topPad = controlsPosition === "top" ? 48 : 0;
      if (isRotated) {
        canvas.style.maxWidth = "calc(100vh - " + topPad + "px)";
        canvas.style.maxHeight = "calc(100vw - " + pad + "px)";
      } else {
        canvas.style.maxWidth = "calc(100vw - " + pad + "px)";
        canvas.style.maxHeight = "calc(100vh - " + topPad + "px)";
      }
      canvas.style.width = "auto";
      canvas.style.height = "auto";
    } else if (isRotated) {
      canvas.style.height = "70vh";
      canvas.style.width = "auto";
      canvas.style.maxWidth = "";
      canvas.style.maxHeight = "";
    } else {
      canvas.style.width = "100%";
      canvas.style.maxWidth = "min(600px, 100%)";
      canvas.style.height = "auto";
      canvas.style.maxHeight = "";
    }
  }

  function renderUI() {
    if (statusEl) {
      if (status === STATUS.WAITING) {
        statusEl.textContent = "Waiting for reMarkable to start screen sharing...";
        statusEl.hidden = false;
      } else if (status === STATUS.CONNECTING) {
        statusEl.textContent = "Connecting to reMarkable...";
        statusEl.hidden = false;
      } else if (status === STATUS.ERROR) {
        statusEl.textContent = errorMsg || "Error";
        statusEl.hidden = false;
      } else {
        statusEl.hidden = true;
        statusEl.textContent = "";
      }
    }

    applyCanvasStyle();

    if (stageEl) {
      if (poppedOut) {
        stageEl.style.position = "fixed";
        stageEl.style.inset = "0";
        stageEl.style.zIndex = "9999";
        stageEl.style.background = resolveBackdrop(backdrop);
        stageEl.style.display = status === STATUS.STREAMING ? "flex" : "none";
        stageEl.style.flexDirection = "column";
        stageEl.style.alignItems = "center";
        stageEl.style.justifyContent = "center";
        stageEl.style.paddingTop = controlsPosition === "top" ? "40px" : "0";
        stageEl.style.paddingRight = controlsPosition === "right" ? "50px" : "0";
      } else {
        stageEl.style.position = "";
        stageEl.style.inset = "";
        stageEl.style.zIndex = "";
        stageEl.style.background = "";
        stageEl.style.display = status === STATUS.STREAMING ? "flex" : "none";
        stageEl.style.flexDirection = "column";
        stageEl.style.alignItems = "center";
        stageEl.style.justifyContent = "";
        stageEl.style.paddingTop = "";
        stageEl.style.paddingRight = "";
      }
    }

    var light = poppedOut && isLightColor(resolveBackdrop(backdrop));
    if (controlsEl) {
      controlsEl.classList.toggle("ss-fullscreen", poppedOut);
      controlsEl.classList.toggle("ss-light", light);
      controlsEl.classList.toggle("ss-pos-right", controlsPosition === "right");
      controlsEl.classList.toggle("ss-pos-top", controlsPosition === "top");
      controlsEl.style.display = status === STATUS.STREAMING || (status === STATUS.ERROR && disconnected) ? "" : "none";
    }

    function show(id, on) {
      var el = document.getElementById(id);
      if (el) el.hidden = !on;
    }

    var streaming = status === STATUS.STREAMING;
    show("ss-rotate-l", streaming && (!poppedOut || showControls));
    show("ss-rotate-r", streaming && (!poppedOut || showControls));
    show("ss-fullscreen", streaming && !poppedOut);
    show("ss-exit-fs", streaming && poppedOut);
    show("ss-options", streaming && poppedOut);
    show("ss-disconnect", streaming && (!poppedOut || showControls));
    show("ss-reconnect", status === STATUS.ERROR && disconnected);

    if (optionsPanel) {
      optionsPanel.hidden = !(poppedOut && showControls);
      if (pinnedPosition || controlsPosition) {
        optionsPanel.dataset.position = pinnedPosition || controlsPosition;
      }
    }

    document.querySelectorAll("[data-backdrop]").forEach(function (btn) {
      var name = btn.getAttribute("data-backdrop");
      var selected = backdrop === name;
      btn.setAttribute("aria-pressed", selected ? "true" : "false");
      btn.classList.toggle("is-selected", selected);
    });

    var customInput = document.getElementById("ss-custom-color");
    if (customInput) {
      customInput.value = customColor;
      var customSelected = !BACKDROP_PRESETS[backdrop];
      customInput.classList.toggle("is-selected", customSelected);
    }
  }

  function bindUI() {
    statusEl = document.getElementById("ss-status");
    canvas = document.getElementById("ss-canvas");
    controlsEl = document.getElementById("ss-controls");
    optionsPanel = document.getElementById("ss-options-panel");
    stageEl = document.getElementById("ss-stage") || (canvas && canvas.parentElement);

    var rotateL = document.getElementById("ss-rotate-l");
    var rotateR = document.getElementById("ss-rotate-r");
    var fullscreen = document.getElementById("ss-fullscreen");
    var exitFs = document.getElementById("ss-exit-fs");
    var disconnectBtn = document.getElementById("ss-disconnect");
    var reconnectBtn = document.getElementById("ss-reconnect");
    var optionsBtn = document.getElementById("ss-options");
    var customInput = document.getElementById("ss-custom-color");

    if (rotateL) {
      rotateL.addEventListener("click", function () {
        manualRotation -= 90;
        renderUI();
      });
    }
    if (rotateR) {
      rotateR.addEventListener("click", function () {
        manualRotation += 90;
        renderUI();
      });
    }
    if (fullscreen) {
      fullscreen.addEventListener("click", function () {
        poppedOut = true;
        updateControlsPosition();
        renderUI();
      });
    }
    if (exitFs) {
      exitFs.addEventListener("click", function () {
        poppedOut = false;
        showControls = false;
        renderUI();
      });
    }
    if (disconnectBtn) disconnectBtn.addEventListener("click", disconnect);
    if (reconnectBtn) reconnectBtn.addEventListener("click", reconnect);
    if (optionsBtn) {
      optionsBtn.addEventListener("click", function () {
        if (!showControls) pinnedPosition = controlsPosition;
        showControls = !showControls;
        renderUI();
      });
    }

    document.querySelectorAll("[data-backdrop]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        var name = btn.getAttribute("data-backdrop");
        backdrop = name;
        setBackdropPref(name);
        renderUI();
      });
    });

    if (customInput) {
      customInput.addEventListener("click", function () {
        backdrop = customColor;
        setBackdropPref(customColor);
        renderUI();
      });
      customInput.addEventListener("input", function (e) {
        customColor = e.target.value;
        setCustomColorPref(customColor);
        backdrop = customColor;
        setBackdropPref(customColor);
        renderUI();
      });
    }

    document.addEventListener("mousedown", function (e) {
      if (!showControls) return;
      if (controlsEl && controlsEl.contains(e.target)) return;
      if (optionsPanel && optionsPanel.contains(e.target)) return;
      showControls = false;
      renderUI();
    });

    window.addEventListener("resize", function () {
      if (poppedOut) updateControlsPosition();
    });
  }

  function init() {
    if (!document.getElementById("ss-canvas")) return;
    bindUI();
    renderUI();
    startPolling();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
