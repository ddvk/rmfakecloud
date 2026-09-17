(function (global) {
  "use strict";
  var h = global.RM.h;
  var useState = global.RM.useState;
  var useEffect = global.RM.useEffect;
  var useRef = global.RM.useRef;
  global.RMShells = global.RMShells || {};

  var KNOWN = { remarkable: 1, googledocs: 1, icloud: 1, os: 1 };

  function normalizeChromeStyle(style) {
    if (style && KNOWN[style]) return style;
    return "remarkable";
  }

  function isMobile(props) {
    if (props && props.formFactor === "mobile") return true;
    if (props && props.forceMobile) return true;
    return global.detectFormFactor && global.detectFormFactor() === "mobile";
  }

  function filterNav(nav) {
    var auth = global.useAuthState ? global.useAuthState() : { state: { user: null } };
    var user = auth.state && auth.state.user;
    var isAdmin = document.body.dataset.admin === "true";
    if (!isAdmin && user) {
      var roles = user.Roles || user.roles || [];
      if (roles.indexOf && roles.indexOf("Admin") >= 0) isAdmin = true;
      if (user.isAdmin) isAdmin = true;
    }
    return (nav || []).filter(function (it) {
      return !it.adminOnly || isAdmin;
    });
  }

  function pathActive(href) {
    var path = location.pathname;
    return path === href || (href !== "/" && path.indexOf(href) === 0);
  }

  function currentLabel(nav) {
    var items = filterNav(nav);
    for (var i = 0; i < items.length; i++) {
      if (pathActive(items[i].href)) return items[i].label || items[i].id;
    }
    if (location.pathname === "/profile") return "Account";
    if (location.pathname === "/help") return "Help";
    if (location.pathname === "/" || location.pathname === "") return "Home";
    return document.title || "rmfakecloud";
  }

  function NavLinks(props) {
    var nav = filterNav(props.nav);
    return h(
      "ul",
      { className: props.className || "shell-nav-list" },
      nav.map(function (it) {
        var active = pathActive(it.href);
        return h(
          "li",
          { key: it.id, className: "shell-nav-item" + (active ? " is-active" : "") },
          h(
            "a",
            {
              href: it.href,
              className: "shell-nav-link" + (active ? " is-active" : ""),
              "aria-current": active ? "page" : undefined,
              onClick: props.onNavigate,
            },
            h("span", { className: "icon icon-" + (it.icon || "file"), "aria-hidden": "true" }),
            h("span", { className: "shell-nav-label" }, it.label || it.id)
          )
        );
      })
    );
  }

  function accountLabel() {
    var auth = global.useAuthState ? global.useAuthState() : { state: { user: null } };
    var user = auth.state && auth.state.user;
    if (user && typeof user === "object") {
      return user.Email || user.email || user.UserID || user.userId || user.Name || user.name || "";
    }
    return document.body.dataset.email || document.body.dataset.user || "Account";
  }

  function UserSlot(props) {
    var openState = useState(false);
    var open = openState[0];
    var setOpen = openState[1];
    useEscClose(open, setOpen);
    useEffect(
      function () {
        if (!open) return;
        function onDoc(e) {
          var t = e.target;
          if (t && t.closest && t.closest(".shell-user-slot")) return;
          setOpen(false);
        }
        document.addEventListener("click", onDoc);
        return function () {
          document.removeEventListener("click", onDoc);
        };
      },
      [open]
    );
    var email = accountLabel() || "Account";
    return h(
      "div",
      { className: "shell-user-slot" + (props.className ? " " + props.className : "") + (open ? " is-open" : "") },
      h(
        "button",
        {
          type: "button",
          className: "user-menu-toggle",
          "aria-haspopup": "menu",
          "aria-expanded": open ? "true" : "false",
          onClick: function (e) {
            if (e && e.stopPropagation) e.stopPropagation();
            setOpen(!open);
          },
        },
        h("span", { className: "icon icon-person", "aria-hidden": "true" }),
        h("span", { className: "user-menu-name" }, String(email))
      ),
      open
        ? h(
            "ul",
            { className: "user-menu-list", role: "menu" },
            h("li", { role: "none" }, h("a", { href: "/profile", role: "menuitem" }, "Settings")),
            h("li", { role: "none" }, h("a", { href: "/help", role: "menuitem" }, "Help")),
            h(
              "li",
              { role: "none" },
              h(
                "form",
                { className: "inline-form", method: "post", action: "/logout" },
                h("button", { type: "submit", className: "btn btn-link", role: "menuitem" }, "Log out")
              )
            )
          )
        : null
    );
  }

  function attachMain(el, mainSlot) {
    if (!el || !mainSlot) return;
    if (mainSlot.parentNode !== el) {
      el.appendChild(mainSlot);
    }
  }

  function useEscClose(open, setOpen) {
    useEffect(
      function () {
        if (!open) return;
        function onKey(e) {
          if (e.key === "Escape") setOpen(false);
        }
        document.addEventListener("keydown", onKey);
        return function () {
          document.removeEventListener("keydown", onKey);
        };
      },
      [open]
    );
  }

  function useChromeMobileClass(on) {
    useEffect(
      function () {
        if (!on) return;
        document.body.classList.add("chrome-mobile");
        return function () {
          document.body.classList.remove("chrome-mobile");
        };
      },
      [on]
    );
  }

  /* ----- Layout 1: remarkable ----- */
  function Remarkable(props) {
    var mobile = isMobile(props);
    useChromeMobileClass(mobile);
    var openState = useState(false);
    var open = openState[0];
    var setOpen = openState[1];
    useEscClose(open, setOpen);
    var close = function () {
      setOpen(false);
    };
    var bar = h(
      "header",
      { className: "shell-rm-bar" },
      mobile
        ? h(
            "button",
            {
              type: "button",
              className: "shell-hamburger shell-rm-hamburger",
              "aria-label": open ? "Close menu" : "Open menu",
              "aria-expanded": open ? "true" : "false",
              onClick: function () {
                setOpen(!open);
              },
            },
            "☰"
          )
        : null,
      h("a", { className: "shell-brand shell-rm-brand", href: "/" }, "rmfakecloud"),
      !mobile
        ? h(NavLinks, { nav: props.nav, className: "shell-rm-top-list" })
        : null,
      h(UserSlot)
    );
    var content = h("div", {
      className: "shell-content shell-rm-content",
      ref: function (el) {
        attachMain(el, props.mainSlot);
      },
    });

    if (!mobile) {
      return h("div", { className: "shell shell-remarkable" }, bar, content);
    }

    return h(
      "div",
      { className: "shell shell-remarkable is-mobile" },
      bar,
      open
        ? h("div", {
            className: "shell-drawer-scrim",
            onClick: close,
          })
        : null,
      h(
        "nav",
        {
          className: "shell-rm-drawer" + (open ? " is-open" : ""),
          "aria-modal": open ? "true" : undefined,
          "aria-hidden": open ? "false" : "true",
          "aria-label": "Library",
        },
        h("div", { className: "shell-rm-drawer-title" }, "Library"),
        h(NavLinks, {
          nav: props.nav,
          className: "shell-rm-nav-list",
          onNavigate: close,
        })
      ),
      content
    );
  }

  /* ----- Layout 2: googledocs ----- */
  function GoogleDocs(props) {
    var mobile = isMobile(props);
    useChromeMobileClass(mobile);
    var title = currentLabel(props.nav);
    var openState = useState(false);
    var open = openState[0];
    var setOpen = openState[1];
    useEscClose(open, setOpen);
    var close = function () {
      setOpen(false);
    };

    if (mobile) {
      return h(
        "div",
        { className: "shell shell-googledocs is-mobile" },
        h(
          "header",
          { className: "shell-gd-appbar" },
          h(
            "button",
            {
              type: "button",
              className: "shell-hamburger shell-gd-hamburger",
              "aria-label": open ? "Close menu" : "Open menu",
              "aria-expanded": open ? "true" : "false",
              onClick: function () {
                setOpen(!open);
              },
            },
            "☰"
          ),
          h("span", { className: "shell-gd-title" }, title),
          h(UserSlot)
        ),
        open
          ? h("div", {
              className: "shell-drawer-scrim",
              onClick: close,
            })
          : null,
        h(
          "nav",
          {
            className: "shell-gd-drawer" + (open ? " is-open" : ""),
            "aria-modal": open ? "true" : undefined,
            "aria-hidden": open ? "false" : "true",
            "aria-label": "Main",
          },
          h("a", { className: "shell-brand shell-gd-drawer-brand", href: "/", onClick: close }, "rmfakecloud"),
          h(NavLinks, {
            nav: props.nav,
            className: "shell-gd-drawer-list",
            onNavigate: close,
          })
        ),
        h("div", {
          className: "shell-content shell-gd-content",
          ref: function (el) {
            attachMain(el, props.mainSlot);
          },
        })
      );
    }

    return h(
      "div",
      { className: "shell shell-googledocs" },
      h(
        "header",
        { className: "shell-gd-appbar" },
        h("a", { className: "shell-brand shell-gd-mark", href: "/" }, "rmfakecloud"),
        h("span", { className: "shell-gd-title" }, title),
        h("div", { className: "shell-gd-actions" }, h(UserSlot))
      ),
      h(
        "div",
        { className: "shell-gd-body" },
        h(
          "aside",
          { className: "shell-gd-rail", "aria-label": "Main" },
          h(NavLinks, { nav: props.nav, className: "shell-gd-rail-list" })
        ),
        h("div", {
          className: "shell-content shell-gd-content",
          ref: function (el) {
            attachMain(el, props.mainSlot);
          },
        })
      )
    );
  }

  /* ----- Layout 3: icloud ----- */
  function ICloud(props) {
    var mobile = isMobile(props);
    useChromeMobileClass(mobile);
    var title = currentLabel(props.nav);
    var openState = useState(false);
    var open = openState[0];
    var setOpen = openState[1];
    useEscClose(open, setOpen);
    var close = function () {
      setOpen(false);
    };

    if (mobile) {
      return h(
        "div",
        { className: "shell shell-icloud is-mobile" },
        h(
          "header",
          { className: "shell-ic-top" },
          h(
            "button",
            {
              type: "button",
              className: "shell-hamburger shell-ic-hamburger",
              "aria-label": open ? "Close menu" : "Open menu",
              "aria-expanded": open ? "true" : "false",
              onClick: function () {
                setOpen(!open);
              },
            },
            "☰"
          ),
          h("h1", { className: "shell-ic-page-title" }, title),
          h(UserSlot, { className: "shell-ic-account" })
        ),
        open
          ? h("div", {
              className: "shell-drawer-scrim shell-ic-scrim",
              onClick: close,
            })
          : null,
        h(
          "nav",
          {
            className: "shell-ic-drawer" + (open ? " is-open" : ""),
            "aria-modal": open ? "true" : undefined,
            "aria-hidden": open ? "false" : "true",
            "aria-label": "Main",
          },
          h(NavLinks, {
            nav: props.nav,
            className: "shell-ic-drawer-list",
            onNavigate: close,
          })
        ),
        h(
          "div",
          { className: "shell-ic-canvas" },
          h("div", {
            className: "shell-content shell-ic-panel",
            ref: function (el) {
              attachMain(el, props.mainSlot);
            },
          })
        )
      );
    }

    return h(
      "div",
      { className: "shell shell-icloud" },
      h(
        "aside",
        { className: "shell-ic-sidebar", "aria-label": "Main" },
        h("a", { className: "shell-brand shell-ic-brand", href: "/" }, "rmfakecloud"),
        h(NavLinks, { nav: props.nav, className: "shell-ic-nav-list" }),
        h(UserSlot, { className: "shell-ic-account" })
      ),
      h(
        "div",
        { className: "shell-ic-main" },
        h("header", { className: "shell-ic-heading" }, h("h1", { className: "shell-ic-page-title" }, title)),
        h(
          "div",
          { className: "shell-ic-canvas" },
          h("div", {
            className: "shell-content shell-ic-panel",
            ref: function (el) {
              attachMain(el, props.mainSlot);
            },
          })
        )
      )
    );
  }

  /* ----- Layout 4: os (folder window + Drive icons) ----- */
  function isDriveRoot() {
    var p = location.pathname;
    return p === "/" || p === "";
  }

  function DriveGrid(props) {
    var items = filterNav(props.nav);
    return h(
      "div",
      { className: "shell-os-drives", role: "list" },
      items.map(function (it) {
        return h(
          "a",
          {
            key: it.id,
            href: it.href,
            className: "shell-os-drive",
            role: "listitem",
          },
          h("span", { className: "shell-os-drive-icon icon icon-lg icon-" + (it.icon || "folder"), "aria-hidden": "true" }),
          h("span", { className: "shell-os-drive-label" }, it.label || it.id)
        );
      })
    );
  }

  function OS(props) {
    var mobile = isMobile(props);
    useChromeMobileClass(mobile);
    var atRoot = isDriveRoot();
    var title = atRoot ? "Drive" : currentLabel(props.nav);
    var contentRef = useRef(null);

    useEffect(
      function () {
        if (atRoot) return;
        attachMain(contentRef.current, props.mainSlot);
      },
      [props.mainSlot, atRoot]
    );

    if (mobile) {
      if (atRoot) {
        return h(
          "div",
          { className: "shell shell-os is-mobile" },
          h(
            "header",
            { className: "shell-os-mobile-bar" },
            h("h1", { className: "shell-os-mobile-title" }, "Drive"),
            h(UserSlot)
          ),
          h(DriveGrid, { nav: props.nav })
        );
      }
      return h(
        "div",
        { className: "shell shell-os is-mobile" },
        h(
          "header",
          { className: "shell-os-mobile-bar" },
          h("a", { className: "shell-os-back", href: "/" }, "← Drives"),
          h("span", { className: "shell-os-mobile-title" }, title),
          h(UserSlot)
        ),
        h("div", {
          className: "shell-content shell-os-mobile-content",
          ref: function (el) {
            contentRef.current = el;
            attachMain(el, props.mainSlot);
          },
        })
      );
    }

    return h(
      "div",
      { className: "shell shell-os" },
      h(
        "div",
        { className: "shell-os-desktop" },
        h(
          "div",
          { className: "shell-os-window", role: "region", "aria-label": "Folder window" },
          h(
            "div",
            { className: "shell-os-titlebar" },
            h(
              "div",
              { className: "shell-os-traffic", "aria-hidden": "true" },
              h("span", { className: "shell-os-dot shell-os-dot-close" }),
              h("span", { className: "shell-os-dot shell-os-dot-min" }),
              h("span", { className: "shell-os-dot shell-os-dot-max" })
            ),
            h("span", { className: "shell-os-window-title" }, title),
            h(UserSlot, { className: "shell-os-title-user" })
          ),
          h(
            "div",
            { className: "shell-os-toolbar" },
            atRoot
              ? h("span", { className: "shell-os-path" }, "Drive")
              : h(
                  "span",
                  { className: "shell-os-path" },
                  h("a", { href: "/" }, "Drive"),
                  h("span", { "aria-hidden": "true" }, " / "),
                  h("span", null, title)
                )
          ),
          h(
            "div",
            { className: "shell-os-body" },
            h(
              "aside",
              { className: "shell-os-sidebar", "aria-label": "Favorites" },
              h("div", { className: "shell-os-sidebar-label" }, "Favorites"),
              h(NavLinks, { nav: props.nav, className: "shell-os-fav-list" })
            ),
            atRoot
              ? h("div", { className: "shell-content shell-os-pane" }, h(DriveGrid, { nav: props.nav }))
              : h("div", {
                  className: "shell-content shell-os-pane",
                  ref: function (el) {
                    contentRef.current = el;
                    attachMain(el, props.mainSlot);
                  },
                })
          )
        )
      )
    );
  }

  global.RMShells.normalizeChromeStyle = normalizeChromeStyle;
  global.RMShells.remarkable = Remarkable;
  global.RMShells.googledocs = GoogleDocs;
  global.RMShells.icloud = ICloud;
  global.RMShells.os = OS;
})(typeof window !== "undefined" ? window : globalThis);
