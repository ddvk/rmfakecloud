/**
 * Minimal React-like runtime (no JSX). Function components + hooks.
 * Exposed as window.RM
 */
(function (global) {
  "use strict";

  var hookIndex = 0;
  var hooks = [];
  var effects = [];
  var pendingEffects = [];
  var currentComponent = null;
  var renderScheduled = false;
  var rootContainer = null;
  var rootVNode = null;
  var contextStack = [];

  function scheduleRender() {
    if (renderScheduled || !rootContainer || !rootVNode) return;
    renderScheduled = true;
    Promise.resolve().then(function () {
      renderScheduled = false;
      flushRender();
    });
  }

  function flushRender() {
    if (!rootContainer || !rootVNode) return;
    hookIndex = 0;
    effects = [];
    hooks = rootContainer.__rmRootHooks || (rootContainer.__rmRootHooks = []);
    currentComponent = rootContainer;
    var vnode;
    if (rootVNode && rootVNode.__isComponentRoot) {
      vnode = rootVNode.type(rootVNode.props || {});
    } else if (typeof rootVNode === "function") {
      vnode = rootVNode();
    } else {
      vnode = rootVNode;
    }
    pendingEffects = pendingEffects.concat(effects);
    effects = [];
    if (rootContainer.__rmVNode) {
      diff(rootContainer, vnode, rootContainer.__rmVNode, 0);
    } else {
      while (rootContainer.firstChild) rootContainer.removeChild(rootContainer.firstChild);
      rootContainer.appendChild(createNode(vnode));
    }
    rootContainer.__rmVNode = vnode;
    runEffects();
  }

  function runEffects() {
    var list = pendingEffects.slice();
    pendingEffects = [];
    list.forEach(function (fx) {
      if (fx.cleanup) {
        try {
          fx.cleanup();
        } catch (e) {
          console.error(e);
        }
      }
      var cleanup = fx.fn();
      if (typeof cleanup === "function") fx.cleanup = cleanup;
    });
  }

  function h(type, props) {
    var children = [];
    for (var i = 2; i < arguments.length; i++) {
      var c = arguments[i];
      if (c == null || c === false) continue;
      if (Array.isArray(c)) children = children.concat(c);
      else children.push(c);
    }
    props = props || {};
    if (children.length) props.children = children.length === 1 ? children[0] : children;
    return { type: type, props: props, key: props.key };
  }

  function normalizeChildren(ch) {
    if (ch == null || ch === false) return [];
    if (!Array.isArray(ch)) return [ch];
    var out = [];
    ch.forEach(function (c) {
      if (c == null || c === false) return;
      if (Array.isArray(c)) out = out.concat(normalizeChildren(c));
      else out.push(c);
    });
    return out;
  }

  function createNode(vnode) {
    if (vnode == null || vnode === false) return document.createComment("");
    if (typeof vnode === "string" || typeof vnode === "number") {
      return document.createTextNode(String(vnode));
    }
    if (typeof vnode.type === "function") {
      return mountComponent(vnode);
    }
    var el = document.createElement(vnode.type);
    applyProps(el, {}, vnode.props || {});
    var kids = normalizeChildren(vnode.props && vnode.props.children);
    kids.forEach(function (kid) {
      el.appendChild(createNode(kid));
    });
    el.__rmVNode = vnode;
    return el;
  }

  function mountComponent(vnode) {
    var holder = document.createDocumentFragment();
    var wrap = document.createElement("div");
    wrap.style.display = "contents";
    wrap.__rmComponent = vnode.type;
    wrap.__rmProps = vnode.props || {};
    wrap.__rmHooks = [];
    var prevHooks = hooks;
    var prevIndex = hookIndex;
    var prevComp = currentComponent;
    hooks = wrap.__rmHooks;
    hookIndex = 0;
    currentComponent = wrap;
    var rendered = vnode.type(wrap.__rmProps);
    currentComponent = prevComp;
    hookIndex = prevIndex;
    hooks = prevHooks;
    var node = createNode(rendered);
    wrap.appendChild(node);
    wrap.__rmChildVNode = rendered;
    // Return wrap so we have a stable DOM anchor
    // effects collected during render are queued
    pendingEffects = pendingEffects.concat(effects);
    effects = [];
    return wrap;
  }

  function applyProps(el, oldProps, newProps) {
    oldProps = oldProps || {};
    newProps = newProps || {};
    var key;
    for (key in oldProps) {
      if (key === "children" || key === "key") continue;
      if (!(key in newProps)) {
        if (key === "className") el.removeAttribute("class");
        else if (key.startsWith("on") && typeof oldProps[key] === "function") {
          el.removeEventListener(key.slice(2).toLowerCase(), oldProps[key]);
        } else if (key === "style") el.removeAttribute("style");
        else if (key === "ref") {
          /* noop */
        } else el.removeAttribute(key);
      }
    }
    for (key in newProps) {
      if (key === "children" || key === "key") continue;
      var val = newProps[key];
      if (oldProps[key] === val) continue;
      if (key === "className") {
        el.setAttribute("class", val || "");
      } else if (key === "style" && val && typeof val === "object") {
        Object.keys(val).forEach(function (k) {
          el.style[k] = val[k];
        });
      } else if (key === "dangerouslySetInnerHTML" && val) {
        el.innerHTML = val.__html || "";
      } else if (key === "ref") {
        if (typeof val === "function") val(el);
        else if (val && typeof val === "object") val.current = el;
      } else if (key.startsWith("on") && typeof val === "function") {
        var ev = key.slice(2).toLowerCase();
        if (typeof oldProps[key] === "function") el.removeEventListener(ev, oldProps[key]);
        el.addEventListener(ev, val);
      } else if (key === "checked" || key === "value" || key === "selected") {
        el[key] = val;
      } else if (val === false || val == null) {
        el.removeAttribute(key);
      } else if (val === true) {
        el.setAttribute(key, "");
      } else {
        el.setAttribute(key, String(val));
      }
    }
  }

  function diff(parent, newVNode, oldVNode, index) {
    index = index || 0;
    // Simplified: replace children of parent for shell roots
    if (!oldVNode) {
      parent.appendChild(createNode(newVNode));
      return;
    }
    if (!newVNode) {
      if (parent.childNodes[index]) parent.removeChild(parent.childNodes[index]);
      return;
    }
    if (typeof newVNode === "string" || typeof newVNode === "number") {
      var text = String(newVNode);
      var cur = parent.childNodes[index];
      if (cur && cur.nodeType === 3) cur.nodeValue = text;
      else {
        var tn = document.createTextNode(text);
        if (cur) parent.replaceChild(tn, cur);
        else parent.appendChild(tn);
      }
      return;
    }
    if (typeof newVNode.type === "function") {
      var host = parent.childNodes[index];
      if (!host || host.__rmComponent !== newVNode.type) {
        var created = createNode(newVNode);
        if (host) parent.replaceChild(created, host);
        else parent.appendChild(created);
        return;
      }
      host.__rmProps = newVNode.props || {};
      var prevH = hooks;
      var prevI = hookIndex;
      var prevC = currentComponent;
      hooks = host.__rmHooks;
      hookIndex = 0;
      currentComponent = host;
      effects = [];
      var rendered = newVNode.type(host.__rmProps);
      currentComponent = prevC;
      hookIndex = prevI;
      hooks = prevH;
      pendingEffects = pendingEffects.concat(effects);
      effects = [];
      if (host.__rmChildVNode) {
        diff(host, rendered, host.__rmChildVNode, 0);
      } else {
        while (host.firstChild) host.removeChild(host.firstChild);
        host.appendChild(createNode(rendered));
      }
      host.__rmChildVNode = rendered;
      return;
    }
    var el = parent.childNodes[index];
    if (!el || el.nodeType !== 1 || el.tagName.toLowerCase() !== String(newVNode.type).toLowerCase()) {
      var fresh = createNode(newVNode);
      if (el) parent.replaceChild(fresh, el);
      else parent.appendChild(fresh);
      return;
    }
    applyProps(el, (oldVNode && oldVNode.props) || {}, newVNode.props || {});
    var newKids = normalizeChildren(newVNode.props && newVNode.props.children);
    var oldKids = normalizeChildren(oldVNode && oldVNode.props && oldVNode.props.children);
    var max = Math.max(newKids.length, oldKids.length);
    for (var i = 0; i < max; i++) {
      if (i >= newKids.length) {
        while (el.childNodes.length > newKids.length) el.removeChild(el.lastChild);
        break;
      }
      if (i >= oldKids.length) el.appendChild(createNode(newKids[i]));
      else diff(el, newKids[i], oldKids[i], i);
    }
    el.__rmVNode = newVNode;
  }

  function useState(initial) {
    var i = hookIndex++;
    if (hooks[i] === undefined) {
      hooks[i] = typeof initial === "function" ? initial() : initial;
    }
    var setState = (function (idx) {
      return function (next) {
        var cur = hooks[idx];
        var val = typeof next === "function" ? next(cur) : next;
        if (Object.is(val, cur)) return;
        hooks[idx] = val;
        scheduleRender();
      };
    })(i);
    return [hooks[i], setState];
  }

  function useReducer(reducer, initialArg, init) {
    var i = hookIndex++;
    if (hooks[i] === undefined) {
      hooks[i] = init ? init(initialArg) : initialArg;
    }
    var dispatch = (function (idx) {
      return function (action) {
        var next = reducer(hooks[idx], action);
        if (Object.is(next, hooks[idx])) return;
        hooks[idx] = next;
        scheduleRender();
      };
    })(i);
    return [hooks[i], dispatch];
  }

  function useRef(initial) {
    var i = hookIndex++;
    if (hooks[i] === undefined) hooks[i] = { current: initial };
    return hooks[i];
  }

  function depsChanged(a, b) {
    if (!a || !b || a.length !== b.length) return true;
    for (var i = 0; i < a.length; i++) if (!Object.is(a[i], b[i])) return true;
    return false;
  }

  function useEffect(fn, deps) {
    var i = hookIndex++;
    var prev = hooks[i];
    if (!prev || depsChanged(prev.deps, deps)) {
      var entry = { fn: fn, deps: deps, cleanup: prev && prev.cleanup };
      hooks[i] = entry;
      effects.push(entry);
    }
  }

  function useMemo(fn, deps) {
    var i = hookIndex++;
    var prev = hooks[i];
    if (!prev || depsChanged(prev.deps, deps)) {
      hooks[i] = { deps: deps, value: fn() };
    }
    return hooks[i].value;
  }

  function useCallback(fn, deps) {
    return useMemo(function () {
      return fn;
    }, deps);
  }

  function createContext(defaultValue) {
    var ctx = { _default: defaultValue, _id: Math.random() };
    ctx.Provider = function Provider(props) {
      var value = props.value !== undefined ? props.value : defaultValue;
      contextStack.push({ ctx: ctx, value: value });
      try {
        var kids = props.children;
        if (Array.isArray(kids)) return h("div", { style: { display: "contents" } }, kids);
        return kids || null;
      } finally {
        contextStack.pop();
      }
    };
    return ctx;
  }

  function useContext(ctx) {
    for (var i = contextStack.length - 1; i >= 0; i--) {
      if (contextStack[i].ctx === ctx) return contextStack[i].value;
    }
    return ctx._default;
  }

  function render(vnode, container) {
    if (typeof container === "string") container = document.querySelector(container);
    rootContainer = container;
    if (typeof vnode === "function") {
      rootVNode = { type: vnode, props: {}, __isComponentRoot: true };
    } else if (vnode && typeof vnode.type === "function") {
      rootVNode = Object.assign({ __isComponentRoot: true }, vnode);
    } else {
      rootVNode = vnode;
    }
    // Clear and mount
    while (container.firstChild) container.removeChild(container.firstChild);
    hookIndex = 0;
    hooks = container.__rmRootHooks || (container.__rmRootHooks = []);
    currentComponent = container;
    effects = [];
    var tree;
    if (rootVNode.__isComponentRoot) {
      hookIndex = 0;
      hooks = container.__rmRootHooks;
      tree = rootVNode.type(rootVNode.props || {});
    } else {
      tree = rootVNode;
    }
    pendingEffects = pendingEffects.concat(effects);
    container.appendChild(createNode(tree));
    container.__rmVNode = tree;
    runEffects();
  }

  function createRoot(container) {
    return {
      render: function (vnode) {
        render(vnode, container);
      },
    };
  }

  var api = {
    h: h,
    useState: useState,
    useEffect: useEffect,
    useReducer: useReducer,
    useRef: useRef,
    useMemo: useMemo,
    useCallback: useCallback,
    createContext: createContext,
    useContext: useContext,
    render: render,
    createRoot: createRoot,
    Fragment: function Fragment(props) {
      return h("div", { style: { display: "contents" } }, props && props.children);
    },
  };

  global.RM = api;
})(typeof window !== "undefined" ? window : globalThis);
