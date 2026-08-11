#!/usr/bin/env sh
set -eu
REPO="dzx941/3m-ui"; BIN_PATH="/usr/local/bin/3m-ui"; CONFIG_DIR="/etc/3m-ui"; DATA_DIR="/var/lib/3m-ui"; SERVICE_NAME="3m-ui"
[ "$(id -u)" -eq 0 ] || { echo "Please run as root." >&2; exit 1; }
[ -x "$BIN_PATH" ] || { echo "3m-ui is not installed at $BIN_PATH" >&2; exit 1; }
command_exists(){ command -v "$1" >/dev/null 2>&1; }
arch_name(){ case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; armv7l|armv7*) echo armv7;; *) echo "Unsupported architecture" >&2; exit 1;; esac; }
asset_suffix(){ case "${THREE_M_UI_STATIC:-0}" in 1|true|yes) echo "-static";; *) echo "";; esac; }
download(){ if command_exists curl; then curl -fsSL "$1" -o "$2"; elif command_exists wget; then wget -qO "$2" "$1"; else echo "curl or wget is required." >&2; exit 1; fi; }
stop_service(){ if command_exists systemctl && [ -f /etc/systemd/system/$SERVICE_NAME.service ]; then systemctl stop $SERVICE_NAME; elif command_exists rc-service && [ -f /etc/init.d/$SERVICE_NAME ]; then rc-service $SERVICE_NAME stop; fi; }
start_service(){ if command_exists systemctl && [ -f /etc/systemd/system/$SERVICE_NAME.service ]; then systemctl start $SERVICE_NAME; elif command_exists rc-service && [ -f /etc/init.d/$SERVICE_NAME ]; then rc-service $SERVICE_NAME start; fi; }
backup_dir="$DATA_DIR/backups/$(date +%Y%m%d%H%M%S)"; mkdir -p "$backup_dir"; [ -d "$CONFIG_DIR" ] && cp -a "$CONFIG_DIR" "$backup_dir/config"
tmp="$(mktemp)"; download "https://github.com/$REPO/releases/latest/download/3m-ui-linux-$(arch_name)$(asset_suffix)" "$tmp"
stop_service
install -m 0755 "$tmp" "$BIN_PATH"; rm -f "$tmp"
start_service
echo "3m-ui updated. Config backup: $backup_dir"
