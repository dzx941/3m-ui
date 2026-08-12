#!/usr/bin/env sh
set -eu
umask 077

REPO="dzx941/3m-ui"
INSTALL_DIR="/usr/local/bin"
BIN_PATH="$INSTALL_DIR/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
LOG_DIR="/var/log/3m-ui"
MIHOMO_BIN="/usr/local/bin/mihomo"
SUBCONVERTER_DIR="/usr/local/subconverter"
SUBCONVERTER_BIN="/usr/local/bin/subconverter"
SERVICE_NAME="3m-ui"

need_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "Please run as root." >&2
        exit 1
    fi
}

command_exists() { command -v "$1" >/dev/null 2>&1; }

arch_name() {
    case "$(uname -m)" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        armv7l|armv7*) echo armv7 ;;
        *) echo "Unsupported architecture" >&2; exit 1 ;;
    esac
}

os_id() {
    if [ -r /etc/os-release ]; then
        . /etc/os-release
        echo "${ID:-unknown}"
    else
        echo unknown
    fi
}

install_prerequisites() {
    if command_exists apt-get; then
        apt-get update
        apt-get install -y ca-certificates curl gzip tar
    elif command_exists dnf; then
        dnf install -y ca-certificates curl gzip tar
    elif command_exists yum; then
        yum install -y ca-certificates curl gzip tar
    elif command_exists apk; then
        apk add --no-cache ca-certificates curl gzip tar
    elif command_exists pacman; then
        pacman -Sy --noconfirm ca-certificates curl gzip tar
    elif command_exists zypper; then
        zypper --non-interactive install ca-certificates curl gzip tar
    fi
}

asset_suffix() {
    case "${THREE_M_UI_STATIC:-auto}" in
        1|true|yes) echo "-static" ;;
        0|false|no) echo "" ;;
        *)
            if [ -r /etc/os-release ] && grep -q '^ID=alpine' /etc/os-release; then
                echo "-static"
            elif [ -e /lib/ld-musl-x86_64.so.1 ] || [ -e /lib/ld-musl-aarch64.so.1 ] || [ -e /lib/ld-musl-armhf.so.1 ]; then
                echo "-static"
            else
                echo ""
            fi
            ;;
    esac
}

download() {
    url="$1"
    output="$2"
    if command_exists curl; then
        curl -fsSL --retry 3 --connect-timeout 10 "$url" -o "$output"
    elif command_exists wget; then
        wget -qO "$output" "$url"
    else
        echo "curl/wget required" >&2
        exit 1
    fi
}

random_hex() {
    bytes="${1:-32}"
    if [ -r /dev/urandom ] && command_exists od; then
        dd if=/dev/urandom bs=1 count="$bytes" 2>/dev/null | od -An -tx1 | tr -d ' \n'
        return
    fi
    echo "Unable to access a secure random source." >&2
    exit 1
}

init_system() {
    if command_exists systemctl && [ -d /run/systemd/system ]; then
        echo systemd
        return
    fi
    if command_exists rc-service; then
        echo openrc
        return
    fi
    echo "No supported init system" >&2
    exit 1
}

mihomo_asset() {
    case "$(uname -m)" in
        x86_64|amd64) echo "mihomo-linux-amd64-compatible" ;;
        aarch64|arm64) echo "mihomo-linux-arm64" ;;
        armv7l|armv7*) echo "mihomo-linux-armv7" ;;
        *) return 1 ;;
    esac
}

latest_mihomo_tag() {
    tag=""
    if command_exists curl; then
        tag="$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/MetaCubeX/mihomo/releases/latest 2>/dev/null | sed 's#.*/##')" || true
    fi
    case "$tag" in
        v[0-9]*) echo "$tag" ;;
        *) echo "" ;;
    esac
}

install_mihomo() {
    if [ -x "$MIHOMO_BIN" ]; then
        echo "Mihomo already installed."
        return 0
    fi

    mkdir -p "$DATA_DIR/mihomo"
    asset="$(mihomo_asset)"
    tag="$(latest_mihomo_tag)"
    if [ -z "$tag" ]; then
        echo "Unable to determine the latest Mihomo release." >&2
        return 1
    fi

    tmp="$(mktemp)"
    stage="$(mktemp)"
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag}/${asset}-${tag}.gz"
    echo "Downloading Mihomo ${tag}..."

    if ! download "$url" "$tmp" || [ ! -s "$tmp" ]; then
        echo "Failed to download Mihomo: $url" >&2
        rm -f "$tmp" "$stage"
        return 1
    fi
    if ! gzip -dc "$tmp" > "$stage" || [ ! -s "$stage" ]; then
        echo "Failed to extract Mihomo." >&2
        rm -f "$tmp" "$stage"
        return 1
    fi
    chmod 0755 "$stage"
    if ! "$stage" -v >/dev/null 2>&1; then
        echo "Downloaded Mihomo binary failed executable validation." >&2
        rm -f "$tmp" "$stage"
        return 1
    fi
    install -m 0755 "$stage" "$MIHOMO_BIN"
    rm -f "$tmp" "$stage"
    echo "Mihomo installed successfully."
}

install_subconverter() {
    if [ -x "$SUBCONVERTER_BIN" ]; then
        echo "Subconverter already installed."
        return 0
    fi

    case "$(uname -m)" in
        x86_64|amd64) sub_arch="linux64" ;;
        aarch64|arm64) sub_arch="aarch64" ;;
        armv7l|armv7*) sub_arch="armv7" ;;
        *) echo "Unsupported subconverter architecture; skipping." >&2; return 0 ;;
    esac

    tmp="$(mktemp)"
    stage="$(mktemp -d)"
    url="https://github.com/asdlokj1qpi233/subconverter/releases/download/v0.9.9/subconverter_${sub_arch}.tar.gz"
    echo "Installing subconverter..."

    if ! download "$url" "$tmp" || [ ! -s "$tmp" ]; then
        echo "Warning: failed to download subconverter; continuing without it." >&2
        rm -f "$tmp"
        rm -rf "$stage"
        return 0
    fi
    if ! tar -xzf "$tmp" -C "$stage" || [ ! -x "$stage/subconverter/subconverter" ]; then
        echo "Warning: invalid subconverter archive; continuing without it." >&2
        rm -f "$tmp"
        rm -rf "$stage"
        return 0
    fi

    mkdir -p "$stage/subconverter/base"
    cat > "$stage/subconverter/base/pref.ini" <<'EOF'
[server]
listen=127.0.0.1
port=25500
EOF

    mkdir -p /usr/local
    rm -rf "$SUBCONVERTER_DIR"
    mv "$stage/subconverter" "$SUBCONVERTER_DIR"
    ln -sfn "$SUBCONVERTER_DIR/subconverter" "$SUBCONVERTER_BIN"
    rm -f "$tmp"
    rm -rf "$stage"
    echo "Subconverter installed on 127.0.0.1:25500."
}

write_config() {
    mkdir -p "$CONFIG_DIR"
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        chmod 0600 "$CONFIG_DIR/config.yaml" || true
        return
    fi

    jwt_secret="$(random_hex 32)"
    cred_key="$(random_hex 32)"

    cat > "$CONFIG_DIR/config.yaml" <<EOF
server:
  port: 8080
  mode: release

database:
  path: "$DATA_DIR/3m-ui.db"

jwt:
  secret: "$jwt_secret"

security:
  credential_key: "$cred_key"

mihomo:
  binary: "/usr/local/bin/mihomo"
  config: "$DATA_DIR/mihomo/config.yaml"
EOF
    chmod 0600 "$CONFIG_DIR/config.yaml"
}

install_systemd() {
    cat > "/etc/systemd/system/$SERVICE_NAME.service" <<EOF
[Unit]
Description=3m-ui server panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Environment=THREE_M_UI_CONFIG=$CONFIG_DIR/config.yaml
WorkingDirectory=$DATA_DIR
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable --now "$SERVICE_NAME"

    if [ -x "$SUBCONVERTER_BIN" ]; then
        cat > /etc/systemd/system/subconverter.service <<EOF
[Unit]
Description=Subconverter subscription conversion utility
After=network.target

[Service]
Type=simple
ExecStart=$SUBCONVERTER_BIN
WorkingDirectory=$SUBCONVERTER_DIR
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
ReadWritePaths=$SUBCONVERTER_DIR

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable --now subconverter
    fi
}

install_openrc() {
    cat > "/etc/init.d/$SERVICE_NAME" <<EOF
#!/sbin/openrc-run
name="3m-ui"
description="3m-ui server panel"
command="$BIN_PATH"
command_background="yes"
pidfile="/run/$SERVICE_NAME.pid"
directory="$DATA_DIR"
export THREE_M_UI_CONFIG="$CONFIG_DIR/config.yaml"
depend() { need net; }
EOF
    chmod +x "/etc/init.d/$SERVICE_NAME"
    rc-update add "$SERVICE_NAME" default
    rc-service "$SERVICE_NAME" restart

    if [ -x "$SUBCONVERTER_BIN" ]; then
        cat > /etc/init.d/subconverter <<EOF
#!/sbin/openrc-run
name="subconverter"
description="Subconverter subscription conversion utility"
command="$SUBCONVERTER_BIN"
command_background="yes"
pidfile="/run/subconverter.pid"
directory="$SUBCONVERTER_DIR"
depend() { need net; }
EOF
        chmod +x /etc/init.d/subconverter
        rc-update add subconverter default
        rc-service subconverter restart
    fi
}

need_root
echo "Detected Linux distribution: $(os_id)"
install_prerequisites
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$DATA_DIR/mihomo" "$LOG_DIR"

arch="$(arch_name)"
tmp="$(mktemp)"
panel_url="https://github.com/$REPO/releases/latest/download/3m-ui-linux-${arch}$(asset_suffix)"
echo "Downloading 3m-ui..."
if ! download "$panel_url" "$tmp" || [ ! -s "$tmp" ]; then
    echo "Failed to download 3m-ui: $panel_url" >&2
    rm -f "$tmp"
    exit 1
fi
chmod 0755 "$tmp"
install -m 0755 "$tmp" "$BIN_PATH"
rm -f "$tmp"

write_config
install_mihomo
install_subconverter

case "$(init_system)" in
    systemd) install_systemd ;;
    openrc) install_openrc ;;
esac

echo ""
echo "3m-ui installed successfully"
echo "Open http://SERVER_IP:8080/"
if [ -f "$DATA_DIR/.initial_admin_password" ]; then
    echo "Initial admin username: admin"
    echo "Initial admin password file: $DATA_DIR/.initial_admin_password"
fi
echo "Static mode: ${THREE_M_UI_STATIC:-auto}"
