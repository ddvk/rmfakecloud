import React from "react";

class ErrorBoundary extends React.Component {
  state = { hasError: false, errorMessage: null };

  // A fake logging service
  logErrorToServices = console.log;

  ensureMetaRefreshRedirect() {
    if (typeof document === "undefined") return;

    const metaId = "rmfakecloud-error-refresh";
    let meta = document.getElementById(metaId);
    if (!meta) {
      meta = document.createElement("meta");
      meta.id = metaId;
      document.head.appendChild(meta);
    }

    meta.setAttribute("http-equiv", "refresh");
    meta.setAttribute("content", "0;url=/");
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI.
    return { hasError: true, errorMessage: error?.toString() };
  }

  componentDidUpdate(prevProps, prevState) {
    if (!prevState.hasError && this.state.hasError) {
      this.ensureMetaRefreshRedirect();
    }
  }

  componentDidCatch(error, info) {
    this.logErrorToServices(error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      // You can render any custom fallback UI
      return (
        <div style={{ padding: 24, textAlign: "center" }}>
          <h1 style={{ marginBottom: 8 }}>Something went wrong</h1>
          {this.state.errorMessage ? <div style={{ color: "#666" }}>{this.state.errorMessage}</div> : null}
          <div style={{ marginTop: 16 }}>
            <a href="/">Go to home</a>
          </div>
          <div style={{ marginTop: 8, color: "#888", fontSize: 12 }}>
            Redirecting you to <code>/</code>...
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
