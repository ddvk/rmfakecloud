(function (global) {
"use strict";
/** Apply XSLT 1.0 in the browser; returns a Document or text depending on stylesheet output. */
async function fetchXML(url) {
  const r = await fetch(url, { credentials: "same-origin" });
  if (!r.ok) throw new Error(`Failed to fetch ${url}`);
  const text = await r.text();
  return new DOMParser().parseFromString(text, "application/xml");
}

function parseXMLString(xmlString) {
  return new DOMParser().parseFromString(xmlString, "application/xml");
}

function applyColorOverrides(themeDoc, overrides) {
  if (!overrides || !Object.keys(overrides).length) return themeDoc;
  const colors = themeDoc.querySelector("colors");
  if (!colors) return themeDoc;
  const clone = themeDoc.cloneNode(true);
  const c = clone.querySelector("colors");
  Object.entries(overrides).forEach(([key, val]) => {
    if (val) c.setAttribute(key, val);
  });
  return clone;
}

function transformWithXSLT(xmlDoc, xslDoc) {
  if (typeof XSLTProcessor === "undefined") {
    throw new Error("XSLTProcessor not supported");
  }
  const proc = new XSLTProcessor();
  proc.importStylesheet(xslDoc);
  return proc.transformToDocument(xmlDoc);
}

function transformToText(xmlDoc, xslDoc) {
  if (typeof XSLTProcessor === "undefined") {
    throw new Error("XSLTProcessor not supported");
  }
  const proc = new XSLTProcessor();
  proc.importStylesheet(xslDoc);
  const frag = proc.transformToFragment(xmlDoc, document);
  return frag?.textContent || "";
}

const STYLE_ID = "rm-shell-theme";

function injectThemeCSS(cssText) {
  let el = document.getElementById(STYLE_ID);
  if (!el) {
    el = document.createElement("style");
    el.id = STYLE_ID;
    document.head.appendChild(el);
  }
  el.textContent = cssText || "";
}

function defaultChromePrefs() {
  return {
    style: "remarkable",
    folderColor: "#d4b483",
    tabColor: "#c4a574",
    outlineColor: "#e8e3d9",
    drawerColor: "#4a4035",
    labelColor: "#f8f7f6",
  };
}

function normalizeChromeStyle(style) {
  if (style === "remarkable" || style === "googledocs" || style === "icloud" || style === "os") {
    return style;
  }
  return "remarkable";
}

function chromeFromElement(el) {
  const d = defaultChromePrefs();
  if (!el) return d;
  return {
    style: normalizeChromeStyle(el.getAttribute("style") || d.style),
    folderColor: el.getAttribute("folder-color") || d.folderColor,
    tabColor: el.getAttribute("tab-color") || d.tabColor,
    outlineColor: el.getAttribute("outline-color") || d.outlineColor,
    drawerColor: el.getAttribute("drawer-color") || d.drawerColor,
    labelColor: el.getAttribute("label-color") || d.labelColor,
  };
}

function chromeXML(chrome, indent) {
  const c = { ...defaultChromePrefs(), ...(chrome || {}) };
  c.style = normalizeChromeStyle(c.style);
  return `${indent}<chrome
${indent}  style="${c.style}"
${indent}  folder-color="${c.folderColor}"
${indent}  tab-color="${c.tabColor}"
${indent}  outline-color="${c.outlineColor}"
${indent}  drawer-color="${c.drawerColor}"
${indent}  label-color="${c.labelColor}"
${indent}/>`;
}

function defaultConnectPrefs() {
  return {
    gaugeType: "circle",
    trackColor: "#3a3f44",
    fillColor: "#EE7B30",
    textColor: "#f8f7f6",
    promptLocation: "above",
    promptStyle: "plain",
    promptBg: "#212529",
    promptFg: "#e8e3d9",
    promptBorder: "#e8e3d9",
    fontFamily: "system",
    fontSize: "lg",
    codeColor: "#f8f7f6",
  };
}

function connectFromElement(el) {
  const d = defaultConnectPrefs();
  if (!el) return d;
  return {
    gaugeType: el.getAttribute("gauge-type") || d.gaugeType,
    trackColor: el.getAttribute("track-color") || d.trackColor,
    fillColor: el.getAttribute("fill-color") || d.fillColor,
    textColor: el.getAttribute("text-color") || d.textColor,
    promptLocation: el.getAttribute("prompt-location") || d.promptLocation,
    promptStyle: el.getAttribute("prompt-style") || d.promptStyle,
    promptBg: el.getAttribute("prompt-bg") || d.promptBg,
    promptFg: el.getAttribute("prompt-fg") || d.promptFg,
    promptBorder: el.getAttribute("prompt-border") || d.promptBorder,
    fontFamily: el.getAttribute("font-family") || d.fontFamily,
    fontSize: el.getAttribute("font-size") || d.fontSize,
    codeColor: el.getAttribute("code-color") || d.codeColor,
  };
}

function parseLayoutDocument(layoutDoc) {
  const root = layoutDoc.documentElement;
  if (!root || root.nodeName.toLowerCase() !== "layout") {
    return defaultLayout();
  }
  const iconsEl = root.querySelector("icons");
  const navEl = root.querySelector("nav");
  const loginEl = root.querySelector("login");
  const connectEl = root.querySelector("connect");
  const chromeEl = root.querySelector("chrome");
  const icons = {};
  if (iconsEl) {
    ["brand", "documents", "integrations", "connect", "screenshare", "admin", "profile"].forEach((k) => {
      icons[k] = iconsEl.getAttribute(k) || k;
    });
  }
  const items = [];
  if (navEl) {
    navEl.querySelectorAll("item").forEach((item) => {
      items.push({
        id: item.getAttribute("id"),
        href: item.getAttribute("href") || "/" + item.getAttribute("id"),
        visible: item.getAttribute("visible") !== "false",
        adminOnly: item.getAttribute("admin-only") === "true",
        icon: item.getAttribute("icon") || icons[item.getAttribute("id")] || item.getAttribute("id"),
      });
    });
  }
  return {
    id: root.getAttribute("id") || "default",
    name: root.getAttribute("name") || "Default",
    icons,
    chrome: chromeFromElement(chromeEl),
    nav: {
      brandPosition: (navEl && navEl.getAttribute("brand-position")) || "start",
      userMenu: (navEl && navEl.getAttribute("user-menu")) || "end",
      items: items.length > 0 ? items : defaultLayout().nav.items,
    },
    login: {
      primaryButton: (loginEl && loginEl.getAttribute("primary-button")) || "end",
      passkeyButton: (loginEl && loginEl.getAttribute("passkey-button")) || "start",
      showBrand: !loginEl || loginEl.getAttribute("show-brand") !== "false",
    },
    connect: connectFromElement(connectEl),
  };
}

function defaultLayout() {
  return {
    id: "default",
    name: "Default",
    icons: {
      brand: "cloud",
      documents: "folder",
      integrations: "puzzle",
      connect: "link",
      screenshare: "display",
      admin: "gear",
      profile: "person",
    },
    chrome: defaultChromePrefs(),
    nav: {
      brandPosition: "start",
      userMenu: "end",
      items: [
        { id: "documents", visible: true, adminOnly: false, icon: "folder" },
        { id: "integrations", visible: true, adminOnly: false, icon: "puzzle" },
        { id: "connect", visible: true, adminOnly: false, icon: "link" },
        { id: "screenshare", visible: true, adminOnly: false, icon: "display" },
        { id: "admin", visible: true, adminOnly: true, icon: "gear" },
        { id: "help", visible: true, adminOnly: false, icon: "book" },
        { id: "profile", visible: true, adminOnly: false, icon: "person" },
      ],
    },
    login: {
      primaryButton: "end",
      passkeyButton: "start",
      showBrand: true,
    },
    connect: defaultConnectPrefs(),
    mobile: defaultMobileLayout(),
  };
}

function defaultMobileLayout() {
  return {
    chrome: {
      ...defaultChromePrefs(),
      style: "remarkable",
    },
    nav: {
      brandPosition: "start",
      userMenu: "end",
      items: [
        { id: "documents", visible: true, adminOnly: false },
        { id: "integrations", visible: true, adminOnly: false },
        { id: "connect", visible: true, adminOnly: false },
        { id: "screenshare", visible: true, adminOnly: false },
        { id: "admin", visible: true, adminOnly: true },
        { id: "help", visible: true, adminOnly: false },
        { id: "profile", visible: true, adminOnly: false },
      ],
    },
    login: {
      primaryButton: "end",
      passkeyButton: "start",
      showBrand: true,
    },
    connect: {
      ...defaultConnectPrefs(),
      gaugeType: "bar",
      promptStyle: "boxed",
      fontSize: "md",
      promptBorder: "#EE7B30",
    },
  };
}

function navItemsXML(items, indent) {
  return (items || [])
    .map(
      (it) =>
        `${indent}<item id="${it.id}" visible="${it.visible !== false}"${
          it.adminOnly ? ' admin-only="true"' : ""
        }/>`
    )
    .join("\n");
}

function connectXML(connect, indent) {
  const c = { ...defaultConnectPrefs(), ...(connect || {}) };
  return `${indent}<connect
${indent}  gauge-type="${c.gaugeType}"
${indent}  track-color="${c.trackColor}"
${indent}  fill-color="${c.fillColor}"
${indent}  text-color="${c.textColor}"
${indent}  prompt-location="${c.promptLocation}"
${indent}  prompt-style="${c.promptStyle}"
${indent}  prompt-bg="${c.promptBg}"
${indent}  prompt-fg="${c.promptFg}"
${indent}  prompt-border="${c.promptBorder}"
${indent}  font-family="${c.fontFamily}"
${indent}  font-size="${c.fontSize}"
${indent}  code-color="${c.codeColor}"
${indent}/>`;
}

function parseNav(navEl, fallbackItems) {
  const items = [];
  if (navEl) {
    navEl.querySelectorAll(":scope > item").forEach((item) => {
      items.push({
        id: item.getAttribute("id"),
        href: item.getAttribute("href") || "/" + item.getAttribute("id"),
        visible: item.getAttribute("visible") !== "false",
        adminOnly: item.getAttribute("admin-only") === "true",
      });
    });
  }
  return {
    brandPosition: (navEl && navEl.getAttribute("brand-position")) || "start",
    userMenu: (navEl && navEl.getAttribute("user-menu")) || "end",
    items: items.length > 0 ? items : fallbackItems,
  };
}

function parseLogin(loginEl) {
  return {
    primaryButton: (loginEl && loginEl.getAttribute("primary-button")) || "end",
    passkeyButton: (loginEl && loginEl.getAttribute("passkey-button")) || "start",
    showBrand: !loginEl || loginEl.getAttribute("show-brand") !== "false",
  };
}

/** Promote layout/mobile/* over desktop layout nodes when formFactor is mobile. */
function applyFormFactor(themeDoc, formFactor) {
  if (formFactor !== "mobile") return themeDoc;
  const mobile = themeDoc.querySelector("layout > mobile");
  if (!mobile) return themeDoc;
  const clone = themeDoc.cloneNode(true);
  const layout = clone.querySelector("layout");
  const mob = clone.querySelector("layout > mobile");
  if (!layout || !mob) return themeDoc;
  ["nav", "login", "connect", "chrome"].forEach((tag) => {
    const src = mob.querySelector(`:scope > ${tag}`);
    if (!src) return;
    const existing = Array.from(layout.children).find((n) => n.nodeName.toLowerCase() === tag);
    if (existing) {
      layout.replaceChild(src.cloneNode(true), existing);
    } else {
      layout.insertBefore(src.cloneNode(true), mob);
    }
  });
  return clone;
}

function themeXMLFromEditor(state) {
  const { id, name, published, colors, icons, nav, login, connect, chrome, mobile } = state;
  const colorAttrs = Object.entries(colors)
    .map(([k, v]) => `${k}="${v}"`)
    .join(" ");
  const iconAttrs = Object.entries(icons)
    .map(([k, v]) => `${k}="${v}"`)
    .join(" ");
  const mob = { ...defaultMobileLayout(), ...(mobile || {}) };
  mob.chrome = { ...defaultMobileLayout().chrome, ...(mobile?.chrome || {}) };
  mob.nav = { ...defaultMobileLayout().nav, ...(mobile?.nav || {}) };
  mob.login = { ...defaultMobileLayout().login, ...(mobile?.login || {}) };
  mob.connect = { ...defaultMobileLayout().connect, ...(mobile?.connect || {}) };

  return `<?xml version="1.0" encoding="UTF-8"?>
<theme id="${id}" name="${escapeXml(name)}" published="${published ? "true" : "false"}">
  <colors ${colorAttrs}/>
  <icons ${iconAttrs}/>
  <layout>
${chromeXML(chrome, "    ")}
    <nav brand-position="${nav.brandPosition || "start"}" user-menu="${nav.userMenu || "end"}">
${navItemsXML(nav.items, "      ")}
    </nav>
    <login primary-button="${login.primaryButton || "end"}" passkey-button="${login.passkeyButton || "start"}" show-brand="${
    login.showBrand !== false
  }"/>
${connectXML(connect, "    ")}
    <mobile>
${chromeXML(mob.chrome, "      ")}
      <nav brand-position="${mob.nav.brandPosition || "start"}" user-menu="${mob.nav.userMenu || "end"}">
${navItemsXML(mob.nav.items, "        ")}
      </nav>
      <login primary-button="${mob.login.primaryButton || "end"}" passkey-button="${mob.login.passkeyButton || "start"}" show-brand="${
    mob.login.showBrand !== false
  }"/>
${connectXML(mob.connect, "      ")}
    </mobile>
  </layout>
</theme>
`;
}

function escapeXml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

function editorStateFromXML(xmlString) {
  const doc = parseXMLString(xmlString);
  const theme = doc.documentElement;
  const colorsEl = doc.querySelector("colors");
  const iconsEl = doc.querySelector("icons");
  const navEl = doc.querySelector("layout > nav");
  const loginEl = doc.querySelector("layout > login");
  const connectEl = doc.querySelector("layout > connect");
  const chromeEl = doc.querySelector("layout > chrome");
  const mobileEl = doc.querySelector("layout > mobile");
  const colors = {};
  const icons = {};
  if (colorsEl) {
    [...colorsEl.attributes].forEach((a) => {
      colors[a.name] = a.value;
    });
  }
  if (iconsEl) {
    [...iconsEl.attributes].forEach((a) => {
      icons[a.name] = a.value;
    });
  }
  const fallbackItems = defaultLayout().nav.items.map(({ id, visible, adminOnly }) => ({
    id,
    visible,
    adminOnly,
  }));
  const mobileFallback = defaultMobileLayout();
  return {
    id: theme.getAttribute("id") || "untitled",
    name: theme.getAttribute("name") || "Untitled",
    published: theme.getAttribute("published") !== "false",
    colors: {
      background1: "#212529",
      background2: "#0f0f0f",
      foreground1: "#f8f7f6",
      foreground2: "#e8e3d9",
      foreground3: "#e4dbaf",
      action: "#EE7B30",
      accept: "#00e676",
      reject: "#ff1744",
      ...colors,
    },
    icons: {
      brand: "cloud",
      documents: "folder",
      integrations: "puzzle",
      connect: "link",
      screenshare: "display",
      admin: "gear",
      profile: "person",
      ...icons,
    },
    chrome: chromeFromElement(chromeEl),
    nav: parseNav(navEl, fallbackItems),
    login: parseLogin(loginEl),
    connect: connectFromElement(connectEl),
    mobile: {
      chrome: chromeFromElement(mobileEl && mobileEl.querySelector(":scope > chrome")),
      nav: parseNav(mobileEl && mobileEl.querySelector(":scope > nav"), mobileFallback.nav.items),
      login: parseLogin(mobileEl && mobileEl.querySelector(":scope > login")),
      connect: connectFromElement(mobileEl && mobileEl.querySelector(":scope > connect")),
    },
  };
}

global.RMXslt = {
  fetchXML: fetchXML,
  parseXMLString: parseXMLString,
  applyColorOverrides: applyColorOverrides,
  applyFormFactor: applyFormFactor,
  transformToText: transformToText,
  transformWithXSLT: transformWithXSLT,
  injectThemeCSS: injectThemeCSS,
  parseLayoutDocument: parseLayoutDocument,
  defaultLayout: defaultLayout,
  normalizeChromeStyle: normalizeChromeStyle,
  chromeStyles: ["remarkable", "googledocs", "icloud", "os"],
  themeXMLFromEditor: typeof themeXMLFromEditor !== "undefined" ? themeXMLFromEditor : null,
  editorStateFromXML: typeof editorStateFromXML !== "undefined" ? editorStateFromXML : null,
};
})(typeof window !== "undefined" ? window : globalThis);
