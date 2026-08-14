#!/usr/bin/env sh
set -eu
umask 077

BIN_PATH="/usr/local/bin/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
LOG_DIR="/var/log/3m-ui"
SERVICE_NAME="3m-ui"
PURGE=0
YES=0

usage() {
    cat <<EOF
3m-ui uninstaller

Usage: $0 [options]

Options:
  -y, --yes          Skip confirmation
      --purge        Also remove /var/lib/3m-ui and all persistent data
  -h, --help         Show this help
EOF
}

for arg in "$@"; do
    case "$arg" in
        -y|--yes) YES=1 ;;
        --purge) PURGE=1 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $arg" >&2; usage >&2; exit 2 ;;
    esac
done

[ "$(id -u)" -eq 0 ] || { echo "Error: please run as root." >&2; exit 1; }

init_system() {
    if [ -d /run/systemd/system ] && command -v systemctl >/dev/null 2>&1; then
        echo systemd
    elif command -v rc-service >/dev/null 2>&1 && [ -d /etc/init.d ]; then
        echo openrc
    else
        echo unsupported
    fi
}

if [ "$YES" -ne 1 ]; then
    if [ ! -t 0 ]; then
        echo "Non-interactive uninstall requires -y/--yes." >&2
        exit 1
    fi
    echo ""
    echo "3m-ui uninstall"
    echo "  Service: $SERVICE_NAME"
    echo "  Binary:  $BIN_PATH"
    echo "  Config:  $CONFIG_DIR"
    if [ "$PURGE" -eq 1 ]; then
        echo "  DATA:    $DATA_DIR  [WILL BE DELETED]"
    else
        echo "  Data:    $DATA_DIR  [KEPT]"
    fi
    printf "Continue? [y/N] "
    read -r answer
    case "$answer" in y|Y|yes|YES) ;; *) echo "Aborted."; exit 0 ;; esac
fi

case "$(init_system)" in
    systemd)
        if [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
            systemctl disable --now "$SERVICE_NAME" >/dev/null 2>&1 || true
            rm -f "/etc/systemd/system/$SERVICE_NAME.service"
            systemctl daemon-reload >/dev/null 2>&1 || true
        fi
        ;;
    openrc)
        if [ -f "/etc/init.d/$SERVICE_NAME" ]; then
            rc-service "$SERVICE_NAME" stop >/dev/null 2>&1 || true
            rc-update del "$SERVICE_NAME" default >/dev/null 2>&1 || true
            rm -f "/etc/init.d/$SERVICE_NAME"
        fi
        ;;
esac

rm -f "$BIN_PATH"
rm -rf "$CONFIG_DIR" "$LOG_DIR"

# Mihomo is intentionally not removed. It can be shared with another service,
# and 3m-ui must not destroy a user's independent Mihomo installation.
if [ "$PURGE" -eq 1 ]; then
    rm -rf "$DATA_DIR"
    echo "3m-ui uninstalled and all application data purged."
else
    echo "3m-ui uninstalled. Persistent data kept at: $DATA_DIR"
fi

echo "Mihomo was left untouched: $MIHOMO_BIN"
echo "Done."
