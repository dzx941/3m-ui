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
REQUESTED_VERSION=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
PLAIN='\033[0m'

msg() { printf '%b\n' "$*"; }
ok() { msg "${GREEN}[OK]${PLAIN} $*"; }
warn() { msg "${YELLOW}[WARN]${PLAIN} $*"; }
err() { msg "${RED}[ERROR]${PLAIN} $*" >&2; }

usage() {
    cat <<EOF
3m-ui installer

Usage:
  $0 [VERSION] [options]

Examples:
  $0
  $0 v1.2.3
  $0 --yes
  $0 v1.2.3 --no-mihomo

Options:
  -y, --yes          Non-interactive installation
      --no-mihomo    Do not install Mihomo when it is missing
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
        v[0-9]*)
            [ -z "$REQUESTED_VERSION" ] || { err "Only one version may be specified."; exit 2; }
            REQUESTED_VERSION="$arg" ;;
        *) err "Unknown option: $arg"; usage >&2; exit 2 ;;
    esac
done

need_root() {
    [ "$(id -u)" -eq 0 ] || { err "Please run this script as root."; exit 1; }
}
command_exists() { command -v "$1" >/dev/null 2>&1; }

arch_name() {
    case "$(uname -m)" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        armv7l|armv7*) echo armv7 ;;
        *) err "Unsupported CPU architecture: $(uname -m)"; exit 1 ;;
    esac
}

os_id() {
    if [ -r /etc/os-release ]; then
        . /etc/os-release
        printf '%s' "${ID:-unknown}"
    elif [ -r /usr/lib/os-release ]; then
        . /usr/lib/os-release
        printf '%s' "${ID:-unknown}"
    else
        printf '%s' unknown
    fi
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
    for tool in curl ca-certificates gzip tar install sed; do
        command_exists "$tool" || missing="$missing $tool"
    done
    [ -z "$missing" ] && return 0

    msg "${YELLOW}Installing required packages:$missing${PLAIN}"
    case "$(os_id)" in
        alpine)
            apk add --no-cache ca-certificates curl gzip tar sed
            ;;
        debian|ubuntu|linuxmint|raspbian)
            apt-get update
            apt-get install -y ca-certificates curl gzip tar sed
            ;;
        fedora)
            dnf install -y ca-certificates curl gzip tar sed
            ;;
        centos|rhel|rocky|almalinux|oracle)
            if command_exists dnf; then dnf install -y ca-certificates curl gzip tar sed; else yum install -y ca-certificates curl gzip tar sed; fi
            ;;
        arch|manjaro)
            pacman -Sy --noconfirm ca-certificates curl gzip tar sed
            ;;
        opensuse*|sles)
            zypper --non-interactive install ca-certificates curl gzip tar sed
            ;;
        *)
            err "Cannot install missing dependencies automatically on $(os_id):$missing"
            exit 1
            ;;
    esac
}

download() {
    url="$1"
    output="$2"
    if command_exists curl; then
        curl -fL --retry 3 --retry-delay 1 --connect-timeout 10 --max-time 300 "$url" -o "$output"
    elif command_exists wget; then
        wget -qO "$output" "$url"
    else
        err "curl or wget is required."
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
        v[0-9]*) printf '%s' "$tag"; return 0 ;;
    esac
    printf '%s' ""
}

asset_suffix() {
    case "$STATIC_MODE" in
        1) echo "-static" ;;
        0) echo "" ;;
        *)
            if [ "$(os_id)" = alpine ] || [ -e /lib/ld-musl-x86_64.so.1 ] || [ -e /lib/ld-musl-aarch64.so.1 ] || [ -e /lib/ld-musl-armhf.so.1 ]; then
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
    mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
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

panel_version() {
    if [ -n "$REQUESTED_VERSION" ]; then
        printf '%s' "$REQUESTED_VERSION"
        return 0
    fi
    latest_tag "https://github.com/$REPO"
}

mihomo_version() {
    latest_tag "https://github.com/MetaCubeX/mihomo"
}

install_mihomo() {
    if [ -x "$MIHOMO_BIN" ] && "$MIHOMO_BIN" -v >/dev/null 2>&1; then
        ok "Existing Mihomo detected: $MIHOMO_BIN"
        return 0
    fi
    [ "$INSTALL_MIHOMO" -eq 1 ] || { warn "Mihomo installation disabled."; return 0; }

    asset="$(mihomo_asset)"
    tag="$(mihomo_version)"
    [ -n "$tag" ] || { err "Unable to determine the latest Mihomo release."; exit 1; }
    tmp="$(mktemp)"
    stage="$(mktemp)"
    trap 'rm -f "$tmp" "$stage"' EXIT INT TERM
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag}/${asset}-${tag}.gz"
    msg "${GREEN}Downloading Mihomo $tag...${PLAIN}"
    download "$url" "$tmp"
    gzip -dc "$tmp" > "$stage"
    chmod 0755 "$stage"
    "$stage" -v >/dev/null 2>&1 || { err "Downloaded Mihomo failed validation."; exit 1; }
    install -m 0755 "$stage" "$MIHOMO_BIN"
    rm -f "$tmp" "$stage"
    trap - EXIT INT TERM
    ok "Mihomo installed: $MIHOMO_BIN"
}

install_panel() {
    arch="$(arch_name)"
    tag="$(panel_version)"
    [ -n "$tag" ] || { err "Unable to determine the latest 3m-ui release. GitHub may be unavailable."; exit 1; }
    tmp="$(mktemp)"
    trap 'rm -f "$tmp"' EXIT INT TERM
    url="https://github.com/$REPO/releases/download/${tag}/3m-ui-linux-${arch}$(asset_suffix)"
    msg "${GREEN}Installing 3m-ui $tag for $arch...${PLAIN}"
    download "$url" "$tmp"
    chmod 0755 "$tmp"
    "$tmp" --version >/dev/null 2>&1 || { err "Downloaded 3m-ui failed executable validation."; exit 1; }
    install -m 0755 "$tmp" "$BIN_PATH"
    rm -f "$tmp"
    trap - EXIT INT TERM
    ok "3m-ui installed: $BIN_PATH"
}

stop_service() {
    case "$(init_system)" in
        systemd) systemctl stop "$SERVICE_NAME" 2>/dev/null || true ;;
        openrc) rc-service "$SERVICE_NAME" stop 2>/dev/null || true ;;
    esac
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
    systemctl enable "$SERVICE_NAME" >/dev/null
    systemctl restart "$SERVICE_NAME"
}

install_openrc() {
    cat > "/etc/init.d/$SERVICE_NAME" <<EOF
#!/sbin/openrc-run
description="3m-ui server panel and Mihomo Core"
command="$BIN_PATH"
command_background="yes"
pidfile="/run/$SERVICE_NAME.pid"
directory="$DATA_DIR"
export THREE_M_UI_CONFIG="$CONFIG_DIR/config.yaml"

output_log="$LOG_DIR/$SERVICE_NAME.log"
error_log="$LOG_DIR/$SERVICE_NAME.log"
respawn_delay=5
supervisor=supervise-daemon

depend() {
    need net
    after firewall
}
EOF
    chmod 0755 "/etc/init.d/$SERVICE_NAME"
    rc-update add "$SERVICE_NAME" default >/dev/null 2>&1 || true
    rc-service "$SERVICE_NAME" restart
}

install_service() {
    case "$(init_system)" in
        systemd) install_systemd ;;
        openrc) install_openrc ;;
        *) err "Unsupported init system. Supported: systemd and OpenRC."; exit 1 ;;
    esac
}

configure_interactive() {
    [ "$YES" -eq 1 ] && return 0
    [ -t 0 ] || return 0

    msg ""
    msg "${GREEN}╭────────────────────────────────────╮${PLAIN}"
    msg "${GREEN}│          3m-ui Installer           │${PLAIN}"
    msg "${GREEN}│    Mihomo Core Management UI       │${PLAIN}"
    msg "${GREEN}╰────────────────────────────────────╯${PLAIN}"
    msg ""
    msg "OS: $(os_id)"
    msg "Architecture: $(arch_name)"
    msg "Init: $(init_system)"

    if [ -z "$REQUESTED_VERSION" ]; then
        latest="$(latest_tag "https://github.com/$REPO")"
        [ -n "$latest" ] && msg "Latest 3m-ui: $latest"
    else
        msg "Requested 3m-ui: $REQUESTED_VERSION"
    fi

    printf "Install Mihomo Core too? [Y/n] "
    read -r answer || true
    case "$answer" in n|N|no|NO) INSTALL_MIHOMO=0 ;; esac
}

show_status() {
    sleep 1
    case "$(init_system)" in
        systemd) systemctl --no-pager --full status "$SERVICE_NAME" || true ;;
        openrc) rc-service "$SERVICE_NAME" status || true ;;
    esac
    msg ""
    ok "3m-ui installation finished."
    msg "Panel: http://SERVER_IP:8080/"
    msg "The default administrator credentials are unchanged; complete the required password change on first login."
}

main() {
    need_root
    install_prerequisites
    configure_interactive

    mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$DATA_DIR/mihomo" "$LOG_DIR"
    write_config

    if [ -f "$BIN_PATH" ]; then
        msg "${YELLOW}Existing 3m-ui installation detected; configuration and data will be preserved.${PLAIN}"
        stop_service
    fi

    install_panel
    install_mihomo
    install_service
    show_status
}

main "$@"
