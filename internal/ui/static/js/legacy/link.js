(function (global) {
  "use strict";
  var RM = global.RM;
  function Link(props) {
    props = props || {};
    return RM.h(
      "a",
      {
        href: props.to || props.href || "#",
        className: props.className,
        "aria-current": props["aria-current"],
        onClick: props.onClick,
      },
      props.children
    );
  }
  global.Link = Link;
})(typeof window !== "undefined" ? window : globalThis);
