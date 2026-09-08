#!/usr/bin/env bash
# Download Go and Node.js (JavaScript runtime) for local development.
#
# - Node.js: official release archives from GitHub (nodejs/node releases).
# - Go: official binary tarballs from https://go.dev/dl/ (the Go project does not
#   publish Linux/macOS binary archives on github.com/golang/go releases).
#
# Usage:
#   ./scripts/download-go-node.sh
#   DEST=~/opt/toolchains GO_VER=1.24.4 NODE_VER=22.12.0 ./scripts/download-go-node.sh
#
# After extraction, add to PATH e.g.:
#   export PATH="$PWD/.tools/go/bin:$PWD/.tools/node/bin:$PATH"

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${DEST:-$ROOT/.tools}"
GO_VER="${GO_VER:-1.24.4}"
NODE_VER="${NODE_VER:-22.12.0}"

mkdir -p "$DEST/downloads"

case "$(uname -s)" in
  Linux)
    GO_OS_ARCH="linux-amd64"
    NODE_OS="linux"
    ;;
  Darwin)
    case "$(uname -m)" in
      arm64) GO_OS_ARCH="darwin-arm64" ;;
      *) GO_OS_ARCH="darwin-amd64" ;;
    esac
    NODE_OS="darwin"
    ;;
  *)
    echo "Unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64) NODE_ARCH="x64" ;;
  aarch64|arm64) NODE_ARCH="arm64" ;;
  *)
    echo "Unsupported arch: $(uname -m)" >&2
    exit 1
    ;;
esac

GO_TGZ="go${GO_VER}.${GO_OS_ARCH}.tar.gz"
GO_URL="https://go.dev/dl/${GO_TGZ}"
NODE_TARXZ="node-v${NODE_VER}-${NODE_OS}-${NODE_ARCH}.tar.xz"
NODE_URL="https://github.com/nodejs/node/releases/download/v${NODE_VER}/${NODE_TARXZ}"

echo "Downloading Go ${GO_VER} from go.dev ..."
curl -fsSL -o "$DEST/downloads/$GO_TGZ" "$GO_URL"
echo "Extracting Go to $DEST/go ..."
rm -rf "$DEST/go"
mkdir -p "$DEST"
tar -C "$DEST" -xzf "$DEST/downloads/$GO_TGZ"

echo "Downloading Node.js ${NODE_VER} from GitHub (nodejs/node releases) ..."
curl -fsSL -L -o "$DEST/downloads/$NODE_TARXZ" "$NODE_URL"
echo "Extracting Node to $DEST/node ..."
rm -rf "$DEST/node"
mkdir -p "$DEST"
tar -C "$DEST" -xJf "$DEST/downloads/$NODE_TARXZ"
mv "$DEST/node-v${NODE_VER}-${NODE_OS}-${NODE_ARCH}" "$DEST/node"

echo "Done."
echo "  Go:   $DEST/go/bin/go"
echo "  Node: $DEST/node/bin/node"
echo "  npm:  $DEST/node/bin/npm"
echo "Add to PATH:"
echo "  export PATH=\"$DEST/go/bin:$DEST/node/bin:\$PATH\""
