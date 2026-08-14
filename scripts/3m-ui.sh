#!/bin/sh
# 3m-ui standalone management script
# Install / Update / Uninstall / Status / Restart / Logs
# Compatible with systemd and OpenRC.

set -eu

APP_NAME="3m-ui"
BIN="/usr/local/bin/3m-ui"
DATA_DIR="/var/lib/3m-ui"
CONFIG_DIR="/etc/3m-ui"
SERVICE="3m-ui"
REPO="dzx941/3m-ui"

say() { printf '%s\n' "$*"; }
err() { say "Error: $*" >&2; exit 1; }

need_root() {
  [ "$(id -u)" = 0 ] || err "Please run as root."
}

init_system() {
  if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
    INIT=systemd
  elif command -v rc-service >/dev/null 2>&1; then
    INIT=openrc
  else
    INIT=none
  fi
}

service_exists() {
  [ "$INIT" = systemd ] && systemctl list-unit-files "${SERVICE}.service" >/dev/null 2>&1 && return 0
  [ "$INIT" = openrc ] && [ -f "/etc/init.d/$SERVICE" ] && return 0
  return 1
}

service_stop() {
  case "$INIT" in
    systemd) systemctl stop "$SERVICE" 2>/dev/null || true ;;
    openrc) rc-service "$SERVICE" stop 2>/dev/null || true ;;
  esac
}

service_start() {
  case "$INIT" in
    systemd) systemctl daemon-reload; systemctl enable --now "$SERVICE" ;;
    openrc) rc-update add "$SERVICE" default >/dev/null 2>&1 || true; rc-service "$SERVICE" restart ;;
    *) err "No supported init system found." ;;
  esac
}

service_restart() {
  case "$INIT" in
    systemd) systemctl restart "$SERVICE" ;;
    openrc) rc-service "$SERVICE" restart ;;
    *) err "No supported init system found." ;;
  esac
}

status() {
  case "$INIT" in
    systemd) systemctl status "$SERVICE" --no-pager || true ;;
    openrc) rc-service "$SERVICE" status || true ;;
    *) say "init: unsupported" ;;
  esac
  if [ -x "$BIN" ]; then
    say "binary: $BIN"
    "$BIN" --version 2>/dev/null || true
  fi
}

logs() {
  case "$INIT" in
    systemd) journalctl -u "$SERVICE" -n 100 --no-pager ;;
    openrc) tail -n 100 "/var/log/$SERVICE.log" 2>/dev/null || say "No log file found." ;;
    *) say "No supported log backend found." ;;
  esac
}

uninstall() {
  need_root
  init_system
  service_stop
  case "$INIT" in
    systemd) systemctl disable "$SERVICE" 2>/dev/null || true; rm -f "/etc/systemd/system/$SERVICE.service"; systemctl daemon-reload ;;
    openrc) rc-update del "$SERVICE" default 2>/dev/null || true; rm -f "/etc/init.d/$SERVICE" ;;
  esac
  rm -f "$BIN"
  say "3m-ui binary and service removed."
  say "Data preserved at: $DATA_DIR"
  say "Configuration preserved at: $CONFIG_DIR"
  say "Mihomo was not removed."
}

usage() {
  cat <<EOF
Usage: $0 <command>

Commands:
  install      Install or repair 3m-ui using scripts/install.sh
  update       Update 3m-ui using scripts/update.sh
  uninstall    Remove 3m-ui but preserve data/config
  status       Show service and binary status
  restart      Restart 3m-ui
  stop         Stop 3m-ui
  start        Start 3m-ui
  logs         Show recent service logs
  version      Show installed version
  help         Show this help

Examples:
  $0 install
  $0 update
  $0 status
  $0 logs
  $0 uninstall
EOF
}

main() {
  need_root
  init_system
  cmd=${1:-help}
  case "$cmd" in
    install)
      SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
      [ -x "$SCRIPT_DIR/install.sh" ] || err "scripts/install.sh not found or not executable."
      exec "$SCRIPT_DIR/install.sh" "${2:-}"
      ;;
    update)
      SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
      [ -x "$SCRIPT_DIR/update.sh" ] || err "scripts/update.sh not found or not executable."
      exec "$SCRIPT_DIR/update.sh" "${2:-}"
      ;;
    uninstall) uninstall ;;
    status) status ;;
    restart) service_restart ;;
    stop) service_stop ;;
    start) service_start ;;
    logs) logs ;;
    version)
      [ -x "$BIN" ] || err "3m-ui is not installed."
      "$BIN" --version
      ;;
    help|-h|--help) usage ;;
    *) usage; exit 2 ;;
  esac
}

main "$@"
