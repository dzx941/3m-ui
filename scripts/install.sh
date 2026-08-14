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
SERVICE_NAME="3m-ui"
YES=0
INSTALL_MIHOMO=1
STATIC_MODE="auto"

usage() {
    cat <<EOF
3m-ui installer

Usage: $0 [options]

Options:
  -y, --yes          Non-interactive installation
      --no-mihomo    Do not install Mihomo when it is not already present
      --static       Prefer the static 3m-ui release asset
      --dynamic      Prefer the dynamic 3m-ui release asset
  -h, --help         Show this help
EOF
}

for arg in "$@"; do
    case "$arg" in
        -y|--yes) YES=1 ;;
        --no-mihomo) INSTALL_MIHOMO=0 ;;
        --static) STATIC_MODE=1 ;;
        --dynamic) STATIC_MODE=0 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $arg" >&2; usage >&2; exit 2 ;;
    esac
done

need_root() {
    [ "$(id -u)" -eq 0 ] || { echo "Error: please run as root." >&2; exit 1; }
}
command_exists() { command -v "$1" >/dev/null 2>&1; }

arch_name() {
    case "$(uname -m)" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        armv7l|armv7*) echo armv7 ;;
        *) echo "Error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
    esac
}

is_alpine() {
    [ -r /etc/os-release ] && . /etc/os-release && [ "${ID:-}" = "alpine" ]
}

init_system() {
    if [ -d /run/systemd/system ] && command_exists systemctl; then
        echo systemd
    elif command_exists rc-service && [ -d /etc/init.d ]; then
        echo openrc
    else
        echo unsupported
    fi
}

install_prerequisites() {
    missing=""
    for tool in curl ca-certificates gzip tar install; do
        command_exists "$tool" || missing="$missing $tool"
    done
    [ -z "$missing" ] && return 0

    echo "Installing required tools:$missing"
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
    else
        echo "Error: cannot install missing dependencies automatically:$missing" >&2
        exit 1
    fi
}

download() {
    url="$1"
    output="$2"
    if command_exists curl; then
        curl -fL --retry 3 --retry-delay 1 --connect-timeout 10 --max-time 300 "$url" -o "$output"
    elif command_exists wget; then
        wget -qO "$output" "$url"
    else
        echo "Error: curl or wget is required." >&2
        exit 1
    fi
}

latest_tag() {
    base="$1"
    tag=""
    if command_exists curl; then
        tag="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "${base}/releases/latest" 2>/dev/null | sed 's#.*/##')" || true
    fi
    case "$tag" in
        v[0-9]*) echo "$tag"; return 0 ;;
    esac
    echo ""
}

asset_suffix() {
    case "$STATIC_MODE" in
        1) echo "-static" ;;
        0) echo "" ;;
        *)
            if is_alpine || [ -e /lib/ld-musl-x86_64.so.1 ] || [ -e /lib/ld-musl-aarch64.so.1 ] || [ -e /lib/ld-musl-armhf.so.1 ]; then
                echo "-static"
            else
                echo ""
            fi
            ;;
    esac
}

mihomo_asset() {
    case "$(uname -m)" in
        x86_64|amd64) echo mihomo-linux-amd64-compatible ;;
        aarch64|arm64) echo mihomo-linux-arm64 ;;
        armv7l|armv7*) echo mihomo-linux-armv7 ;;
        *) return 1 ;;
    esac
}

random_hex() {
    dd if=/dev/urandom bs=1 count="${1:-32}" 2>/dev/null | od -An -tx1 | tr -d ' \n'
}

write_config() {
    mkdir -p "$CONFIG_DIR"
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        chmod 0600 "$CONFIG_DIR/config.yaml"
        return 0
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
  binary: "$MIHOMO_BIN"
  config: "$DATA_DIR/mihomo/config.yaml"
EOF
    chmod 0600 "$CONFIG_DIR/config.yaml"
}

install_mihomo() {
    if [ -x "$MIHOMO_BIN" ] && "$MIHOMO_BIN" -v >/dev/null 2>&1; then
        echo "[OK] Existing Mihomo detected: $MIHOMO_BIN"
        return 0
    fi
    [ "$INSTALL_MIHOMO" -eq 1 ] || { echo "[SKIP] Mihomo installation disabled."; return 0; }

    asset="$(mihomo_asset)"
    tag="$(latest_tag https://github.com/MetaCubeX/mihomo)"
    [ -n "$tag" ] || { echo "Error: unable to determine the latest Mihomo release." >&2; exit 1; }
    tmp="$(mktemp)"
    stage="$(mktemp)"
    trap 'rm -f "$tmp" "$stage"' EXIT INT TERM
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag}/${asset}-${tag}.gz"
    echo "[1/3] Downloading Mihomo $tag..."
    download "$url" "$tmp"
    echo "[2/3] Extracting and validating Mihomo..."
    gzip -dc "$tmp" > "$stage"
    chmod 0755 "$stage"
    "$stage" -v >/dev/null 2>&1 || { echo "Error: downloaded Mihomo failed validation." >&2; exit 1; }
    mkdir -p "$(dirname "$MIHOMO_BIN")"
    install -m 0755 "$stage" "$MIHOMO_BIN"
    rm -f "$tmp" "$stage"
    trap - EXIT INT TERM
    echo "[OK] Mihomo installed: $MIHOMO_BIN"
}

install_panel() {
    arch="$(arch_name)"
    tag="$(latest_tag "https://github.com/$REPO")"
    [ -n "$tag" ] || { echo "Error: unable to determine the latest 3m-ui release." >&2; exit 1; }
    tmp="$(mktemp)"
    trap 'rm -f "$tmp"' EXIT INT TERM
    url="https://github.com/$REPO/releases/download/${tag}/3m-ui-linux-${arch}$(asset_suffix)"
    echo "Downloading 3m-ui $tag ($arch)..."
    download "$url" "$tmp"
    chmod 0755 "$tmp"
    "$tmp" --version >/dev/null 2>&1 || { echo "Error: downloaded 3m-ui failed executable validation." >&2; exit 1; }
    install -m 0755 "$tmp" "$BIN_PATH"
    rm -f "$tmp"
    trap - EXIT INT TERM
    echo "[OK] 3m-ui installed: $BIN_PATH"
}

install_systemd() {
    cat > "/etc/systemd/system/$SERVICE_NAME.service" <<EOF
[Unit]
Description=3m-ui server panel and Mihomo Core
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Environment=THREE_M_UI_CONFIG=$CONFIG_DIR/config.yaml
WorkingDirectory=$DATA_DIR
Restart=always
RestartSec=5
KillMode=control-group

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable --now "$SERVICE_NAME"
}

install_openrc() {
    cat > "/etc/init.d/$SERVICE_NAME" <<EOF
#!/sbin/openrc-run
name="3m-ui"
description="3m-ui server panel and Mihomo Core"
command="$BIN_PATH"
command_background="yes"
pidfile="/run/$SERVICE_NAME.pid"
directory="$DATA_DIR"
export THREE_M_UI_CONFIG="$CONFIG_DIR/config.yaml"

depend() {
    need net
}
EOF
    chmod 0755 "/etc/init.d/$SERVICE_NAME"
    rc-update add "$SERVICE_NAME" default >/dev/null 2>&1 || true
    rc-service "$SERVICE_NAME" restart
}

install_service() {
    init="$(init_system)"
    case "$init" in
        systemd) install_systemd ;;
        openrc) install_openrc ;;
        *) echo "Error: unsupported init system. Supported: systemd, OpenRC." >&2; exit 1 ;;
    esac
}

show_status() {
    sleep 1
    case "$(init_system)" in
        systemd) systemctl --no-pager --full status "$SERVICE_NAME" || true ;;
        openrc) rc-service "$SERVICE_NAME" status || true ;;
    esac
    echo
    echo "3m-ui installed successfully."
    echo "Open: http://SERVER_IP:8080/"
}

need_root
install_prerequisites
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$DATA_DIR/mihomo" "$LOG_DIR"

if [ "$YES" -ne 1 ] && [ -t 0 ]; then
    echo ""
    echo "╭────────────────────────────────────╮"
    echo "│       3m-ui Installer               │"
    echo "│  Mihomo Core Management Console     │"
    echo "╰────────────────────────────────────╯"
    echo ""
    printf "Install/upgrade Mihomo Core too? [Y/n] "
    read -r answer
    case "$answer" in n|N|no|NO) INSTALL_MIHOMO=0 ;; esac
fi

install_panel
write_config
install_mihomo
install_service
show_status
