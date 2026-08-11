#!/usr/bin/env sh
set -eu

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

[ "$(id -u)" -eq 0 ] || {
    echo "Please run as root." >&2
    exit 1
}

if [ "$YES" -ne 1 ]; then
    if [ ! -t 0 ]; then
        echo "Non-interactive uninstall requires -y/--yes." >&2
        echo "Example: curl -fsSL https://raw.githubusercontent.com/dzx941/3m-ui/main/scripts/uninstall.sh | sh -s -- -y" >&2
        exit 1
    fi
    printf "Uninstall 3m-ui? This removes the service, binary, and configs. [y/N] "
    read -r ans
    case "$ans" in
        y|Y|yes|YES) ;;
        *) echo "Aborted."; exit 0 ;;
    esac
fi

# 1. Disable and Stop 3m-ui
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

# 2. Disable and Stop subconverter
if command -v systemctl >/dev/null 2>&1 && [ -f "/etc/systemd/system/subconverter.service" ]; then
    systemctl disable --now subconverter || true
    rm -f "/etc/systemd/system/subconverter.service"
    systemctl daemon-reload || true
fi

if command -v rc-service >/dev/null 2>&1 && [ -f "/etc/init.d/subconverter" ]; then
    rc-service subconverter stop || true
    rc-update del subconverter default || true
    rm -f "/etc/init.d/subconverter"
fi

# 3. Clean Binaries and Directories
rm -f "$BIN_PATH"
rm -f "/usr/local/bin/mihomo"
rm -f "/usr/local/bin/subconverter"
rm -rf "/usr/local/subconverter"
rm -rf "$CONFIG_DIR" "$LOG_DIR"

if [ "$PURGE" -eq 1 ]; then
    rm -rf "$DATA_DIR"
    echo "3m-ui and subconverter uninstalled and all data purged."
else
    echo "3m-ui and subconverter uninstalled. Data kept at $DATA_DIR."
    echo "Use --purge to remove the database and all application data."
fi
