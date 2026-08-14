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

for arg in "$@"; do
    case "$arg" in
        --purge) PURGE=1 ;;
        -y|--yes) YES=1 ;;
        *) echo "Unknown option: $arg" >&2; exit 2 ;;
    esac
done

[ "$(id -u)" -eq 0 ] || { echo "Please run as root." >&2; exit 1; }

if [ "$YES" -ne 1 ]; then
    if [ ! -t 0 ]; then
        echo "Non-interactive uninstall requires -y/--yes." >&2
        echo "Example: curl -fsSL https://github.com/dzx941/3m-ui/releases/latest/download/uninstall.sh | sh -s -- -y" >&2
        exit 1
    fi
    printf "Uninstall 3m-ui? This removes the service, binary, and panel configs. [y/N] "
    read -r ans
    case "$ans" in
        y|Y|yes|YES) ;;
        *) echo "Aborted."; exit 0 ;;
    esac
fi

if command -v systemctl >/dev/null 2>&1 && [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    systemctl disable --now "$SERVICE_NAME" || true
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload || true
fi
if command -v rc-service >/dev/null 2>&1 && [ -f "/etc/init.d/$SERVICE_NAME" ]; then
    rc-service "$SERVICE_NAME" stop || true
    rc-update del "$SERVICE_NAME" default || true
    rm -f "/etc/init.d/$SERVICE_NAME"
fi

rm -f "$BIN_PATH"
rm -rf "$CONFIG_DIR" "$LOG_DIR"

# Do not blindly delete /usr/local/bin/mihomo. Mihomo can be shared with other
# services and the installer intentionally reuses an existing installation.
# Keeping it avoids destructive uninstall side effects.

if [ "$PURGE" -eq 1 ]; then
    rm -rf "$DATA_DIR"
    echo "3m-ui uninstalled and all panel data purged."
else
    echo "3m-ui uninstalled. Persistent data kept at $DATA_DIR."
    echo "Mihomo was left untouched because it may be shared by other services."
    echo "Use --purge to remove the 3m-ui database and application data."
fi
