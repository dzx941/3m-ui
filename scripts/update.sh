#!/usr/bin/env sh
set -eu
umask 077

REPO="dzx941/3m-ui"
BASE="/usr/local/lib/3m-ui"
APP_BIN="$BASE/3m-ui-bin"
ENTRY="/usr/local/bin/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
SERVICE_NAME="3m-ui"
MIHOMO_BIN="/usr/local/bin/mihomo"
UPDATE_MIHOMO=1

for arg in "$@"; do
  case "$arg" in
    -y|--yes) :;;
    --no-mihomo) UPDATE_MIHOMO=0;;
    -h|--help) printf '%s\n' 'Usage: update.sh [--yes] [--no-mihomo]'; exit 0;;
    *) printf '%s\n' "Unknown option: $arg" >&2; exit 2;;
  esac
done

[ "$(id -u)" -eq 0 ] || { echo "Error: please run as root." >&2; exit 1; }
[ -x "$APP_BIN" ] || { echo "Error: 3m-ui application is not installed at $APP_BIN" >&2; exit 1; }
command_exists(){ command -v "$1" >/dev/null 2>&1; }
arch(){ case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; armv7l|armv7*) echo armv7;; *) echo "Error: unsupported architecture" >&2; exit 1;; esac; }
init_system(){ if [ -d /run/systemd/system ] && command_exists systemctl; then echo systemd; elif command_exists rc-service; then echo openrc; else echo unsupported; fi; }
download(){ if command_exists curl; then curl -fL --retry 3 --retry-delay 1 --connect-timeout 10 --max-time 300 "$1" -o "$2"; else wget -qO "$2" "$1"; fi; }
latest_tag(){ tag="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$1/releases/latest" 2>/dev/null | sed 's#.*/##')" || true; case "$tag" in v[0-9]*) echo "$tag";; *) echo "";; esac; }
mihomo_asset(){ case "$(uname -m)" in x86_64|amd64) echo mihomo-linux-amd64-compatible;; aarch64|arm64) echo mihomo-linux-arm64;; armv7l|armv7*) echo mihomo-linux-armv7;; *) return 1;; esac; }

stop(){ case "$(init_system)" in systemd) systemctl stop "$SERVICE_NAME" 2>/dev/null || true;; openrc) rc-service "$SERVICE_NAME" stop 2>/dev/null || true;; esac; }
start(){ case "$(init_system)" in systemd) systemctl start "$SERVICE_NAME";; openrc) rc-service "$SERVICE_NAME" start;; *) return 1;; esac; }
ok(){ case "$(init_system)" in systemd) systemctl is-active --quiet "$SERVICE_NAME";; openrc) rc-service "$SERVICE_NAME" status >/dev/null 2>&1;; *) return 1;; esac; }

backup="$DATA_DIR/backups/$(date +%Y%m%d%H%M%S)"; mkdir -p "$backup"
[ -d "$CONFIG_DIR" ] && cp -a "$CONFIG_DIR" "$backup/config"
[ -f "$DATA_DIR/3m-ui.db" ] && cp -p "$DATA_DIR/3m-ui.db" "$backup/3m-ui.db"

panel_tmp="$(mktemp)"; manager_tmp="$(mktemp)"; mihomo_tmp=""; old="$APP_BIN.update-old"
cleanup(){ rm -f "$panel_tmp" "$manager_tmp" "$mihomo_tmp" "$mihomo_tmp.gz" "$old" "$MIHOMO_BIN.update-old"; }
trap cleanup EXIT INT TERM

tag="$(latest_tag https://github.com/$REPO)"; [ -n "$tag" ] || { echo "Error: unable to determine latest 3m-ui release." >&2; exit 1; }
echo "[1/4] Downloading 3m-ui $tag..."
download "https://github.com/$REPO/releases/download/${tag}/3m-ui-linux-$(arch)" "$panel_tmp"
chmod 0755 "$panel_tmp"
"$panel_tmp" --version >/dev/null 2>&1 || { echo "Error: downloaded 3m-ui failed validation." >&2; exit 1; }

echo "[2/4] Refreshing management scripts..."
download "https://raw.githubusercontent.com/$REPO/main/scripts/3m-ui.sh" "$manager_tmp"
chmod 0755 "$manager_tmp"

if [ "$UPDATE_MIHOMO" -eq 1 ] && [ -x "$MIHOMO_BIN" ]; then
  mtag="$(latest_tag https://github.com/MetaCubeX/mihomo)" || true
  if [ -n "$mtag" ]; then
    mihomo_tmp="$(mktemp)"
    if download "https://github.com/MetaCubeX/mihomo/releases/download/${mtag}/$(mihomo_asset)-${mtag}.gz" "$mihomo_tmp.gz" && gzip -dc "$mihomo_tmp.gz" > "$mihomo_tmp"; then
      chmod 0755 "$mihomo_tmp"
      "$mihomo_tmp" -v >/dev/null 2>&1 || mihomo_tmp=""
    else mihomo_tmp=""; fi
  fi
fi

echo "[3/4] Installing validated files..."
stop
cp -p "$APP_BIN" "$old"
install -m 0755 "$panel_tmp" "$APP_BIN"
install -m 0755 "$manager_tmp" "$BASE/3m-ui.sh"
if [ -n "$mihomo_tmp" ] && [ -s "$mihomo_tmp" ]; then cp -p "$MIHOMO_BIN" "$MIHOMO_BIN.update-old" 2>/dev/null || true; install -m 0755 "$mihomo_tmp" "$MIHOMO_BIN"; fi

if ! start || ! ok; then
  echo "[ERROR] New version failed to start; rolling back." >&2
  install -m 0755 "$old" "$APP_BIN"
  [ -f "$MIHOMO_BIN.update-old" ] && install -m 0755 "$MIHOMO_BIN.update-old" "$MIHOMO_BIN" || true
  start || true
  exit 1
fi

rm -f "$old" "$MIHOMO_BIN.update-old"
echo "[4/4] Update completed successfully."
echo "Backup: $backup"
echo "Command: 3m-ui"
