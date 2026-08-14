#!/bin/sh
# 3m-ui unified management logic
# The command entrypoint is /usr/local/bin/3m-ui.
# The actual Go application is /usr/local/lib/3m-ui/3m-ui-bin.

set -eu

APP_NAME="3m-ui"
BASE="/usr/local/lib/3m-ui"
APP_BIN="$BASE/3m-ui-bin"
VERSION_FILE="$BASE/VERSION"
DATA_DIR="/var/lib/3m-ui"
CONFIG_DIR="/etc/3m-ui"
SERVICE="3m-ui"

say() { printf '%s\n' "$*"; }
err() { say "Error: $*" >&2; exit 1; }

need_root() {
  [ "$(id -u)" = 0 ] || err "Please run as root."
}

init_system() {
  if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
    printf '%s' systemd
  elif command -v rc-service >/dev/null 2>&1; then
    printf '%s' openrc
  else
    printf '%s' unsupported
  fi
}

service_action() {
  action="$1"
  case "$(init_system)" in
    systemd) systemctl "$action" "$SERVICE" ;;
    openrc) rc-service "$SERVICE" "$action" ;;
    *) err "No supported init system found." ;;
  esac
}

status() {
  case "$(init_system)" in
    systemd) systemctl --no-pager --full status "$SERVICE" || true ;;
    openrc) rc-service "$SERVICE" status || true ;;
    *) say "init: unsupported" ;;
  esac
  say ""
  if [ -x "$APP_BIN" ]; then
    say "Application: $APP_BIN"
    if ! "$APP_BIN" --version 2>/dev/null; then
      if [ -s "$VERSION_FILE" ]; then
        say "Version: $(cat "$VERSION_FILE")"
      else
        say "Version: unavailable (installed binary does not expose --version)"
      fi
    fi
  else
    say "Application binary not found: $APP_BIN"
  fi
}

logs() {
  case "$(init_system)" in
    systemd) journalctl -u "$SERVICE" -n 100 --no-pager ;;
    openrc) tail -n 100 "$DATA_DIR/logs/3m-ui.log" 2>/dev/null || tail -n 100 "/var/log/3m-ui/3m-ui.log" 2>/dev/null || say "No log file found." ;;
    *) say "No supported log backend found." ;;
  esac
}

version() {
  [ -x "$APP_BIN" ] || err "3m-ui application is not installed."
  if "$APP_BIN" --version 2>/dev/null; then
    return 0
  fi
  if [ -s "$VERSION_FILE" ]; then
    say "3m-ui $(cat "$VERSION_FILE")"
    return 0
  fi
  err "Unable to determine installed version."
}

uninstall() {
  service_action stop 2>/dev/null || true
  case "$(init_system)" in
    systemd)
      systemctl disable "$SERVICE" 2>/dev/null || true
      rm -f "/etc/systemd/system/$SERVICE.service"
      systemctl daemon-reload
      ;;
    openrc)
      rc-update del "$SERVICE" default 2>/dev/null || true
      rm -f "/etc/init.d/$SERVICE"
      ;;
  esac
  rm -f /usr/local/bin/3m-ui
  rm -rf "$BASE"
  say "3m-ui application and management command removed."
  say "Data preserved at: $DATA_DIR"
  say "Configuration preserved at: $CONFIG_DIR"
  say "Mihomo was not removed."
}

usage() {
  cat <<EOF
Usage: 3m-ui <command>

Commands:
  status       Show service and application status
  start        Start 3m-ui
  restart      Restart 3m-ui
  stop         Stop 3m-ui
  logs         Show recent service logs
  version      Show installed 3m-ui version
  uninstall    Remove 3m-ui but preserve data/config
  help         Show this help

Run '3m-ui' without arguments to open the interactive management menu.
EOF
}

main() {
  need_root
  cmd=${1:-}
  case "$cmd" in
    status) status ;;
    start) service_action start ;;
    restart) service_action restart ;;
    stop) service_action stop ;;
    logs) logs ;;
    version) version ;;
    uninstall) uninstall ;;
    help|-h|--help) usage ;;
    '')
      if [ -x /usr/local/bin/3m-ui ] && [ "$(readlink -f /usr/local/bin/3m-ui 2>/dev/null || true)" = "$APP_BIN" ]; then
        err "Invalid installation: management entrypoint points to the application binary. Re-run the latest installer to migrate it."
      fi
      exec /usr/local/bin/3m-ui
      ;;
    *) err "Unknown command: $cmd" ;;
  esac
}

main "$@"
