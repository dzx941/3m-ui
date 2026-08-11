#!/usr/bin/env sh
set -eu
BIN_PATH="/usr/local/bin/3m-ui"; CONFIG_DIR="/etc/3m-ui"; DATA_DIR="/var/lib/3m-ui"; LOG_DIR="/var/log/3m-ui"; SERVICE_NAME="3m-ui"; PURGE=0
[ "${1:-}" = "--purge" ] && PURGE=1
[ "$(id -u)" -eq 0 ] || { echo "Please run as root." >&2; exit 1; }
printf "Uninstall 3m-ui? This removes the service, binary, and configs. [y/N] "; read ans
case "$ans" in y|Y|yes|YES) ;; *) echo "Aborted."; exit 0;; esac
if command -v systemctl >/dev/null 2>&1 && [ -f /etc/systemd/system/$SERVICE_NAME.service ]; then systemctl disable --now $SERVICE_NAME || true; rm -f /etc/systemd/system/$SERVICE_NAME.service; systemctl daemon-reload || true; fi
if command -v rc-service >/dev/null 2>&1 && [ -f /etc/init.d/$SERVICE_NAME ]; then rc-service $SERVICE_NAME stop || true; rc-update del $SERVICE_NAME default || true; rm -f /etc/init.d/$SERVICE_NAME; fi
rm -f "$BIN_PATH"; rm -rf "$CONFIG_DIR" "$LOG_DIR"
if [ "$PURGE" -eq 1 ]; then rm -rf "$DATA_DIR"; echo "3m-ui uninstalled and data purged."; else echo "3m-ui uninstalled. Data kept at $DATA_DIR (use --purge to remove it)."; fi
