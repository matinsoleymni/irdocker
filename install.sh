
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAME="irdocker"
SRC_GO="${ROOT}/main.go"
COMP_SRC="${ROOT}/completions/irdocker.bash"

echo "🔧 Installing irdocker..."

# Install Go if not present
if ! command -v go &>/dev/null; then
  echo "📦 Go not found. Installing Go..."
  GO_VERSION="1.22.3"
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  echo "✅ Go installed."
fi

# Build irdocker
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go build -o irdocker .

sudo mv irdocker /usr/local/bin/irdocker
sudo chmod +x /usr/local/bin/irdocker

echo ""
echo "✅ irdocker installed successfully!"
echo "   Try: irdocker nginx"

# Install bash completion
COMPLETION_DIR="/etc/bash_completion.d"
COMPLETION_SRC="$SCRIPT_DIR/completions/irdocker.bash"
if [ -f "$COMPLETION_SRC" ]; then
  if [ -d "$COMPLETION_DIR" ]; then
    sudo cp "$COMPLETION_SRC" "$COMPLETION_DIR/"
    echo "✅ Bash completion installed to $COMPLETION_DIR/irdocker.bash"
  else
    echo "⚠️  Bash completion directory not found: $COMPLETION_DIR"
    echo "   You can manually source completions/irdocker.bash in your ~/.bashrc"
  fi
else
  echo "⚠️  Bash completion script not found: $COMPLETION_SRC"
fi
