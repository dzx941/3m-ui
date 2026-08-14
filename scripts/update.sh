#!/usr/bin/env sh
set -eu
umask 077

REPO="dzx941/3m-ui"
BIN_PATH="/usr/local/bin/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
SERVICE_NAME="3m-ui"
MIHOMO_BIN="/usr/local/bin/mihomo"
UPDATE_MIHOMO=1
STATIC_MODE="auto"

usage() {
    cat <<EOF
3m-ui updater

Usage: $0 [options]

Options:
  -y, --yes          Non-interactive update
      --no-mihomo    Update only 3m-ui
      --static       Prefer the static 3m-ui release asset
      --dynamic      Prefer the dynamic 3m-ui release asset
  -h, --help         Show this help
EOF
}
for arg in "$@"; do
    case "$arg" in
        -y|--yes) : ;;
        --no-mihomo) UPDATE_MIHOMO=0 ;;
        --static) STATIC_MODE=1 ;;
        --dynamic) STATIC_MODE=0 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $arg" >&2; usage >&2; exit 2 ;;
    esac
done

[ "$(id -u)" -eq 0 ] || { echo "Error: please run as root." >&2; exit 1; }
[ -x "$BIN_PATH" ] || { echo "Error: 3m-ui is not installed at $BIN_PATH" >&2; exit 1; }
command_exists() { command -v "$1" >/dev/null 2>&1; }
arch_name(){ case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; armv7l|armv7*) echo armv7;; *) echo "Error: unsupported architecture" >&2; exit 1;; esac; }
is_alpine(){ [ -r /etc/os-release ] && . /etc/os-release && [ "${ID:-}" = alpine ]; }
asset_suffix(){ case "$STATIC_MODE" in 1) echo -static;; 0) echo "";; *) if is_alpine || [ -e /lib/ld-musl-x86_64.so.1 ] || [ -e /lib/ld-musl-aarch64.so.1 ] || [ -e /lib/ld-musl-armhf.so.1 ]; then echo -static; else echo ""; fi;; esac; }
init_system(){ if [ -d /run/systemd/system ] && command_exists systemctl; then echo systemd; elif command_exists rc-service && [ -d /etc/init.d ]; then echo openrc; else echo unsupported; fi; }
download(){ if command_exists curl; then curl -fL --retry 3 --retry-delay 1 --connect-timeout 10 --max-time 300 "$1" -o "$2"; elif command_exists wget; then wget -qO "$2" "$1"; else echo "Error: curl or wget is required." >&2; exit 1; fi; }
latest_tag(){ tag="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$1/releases/latest" 2>/dev/null | sed 's#.*/##')" || true; case "$tag" in v[0-9]*) echo "$tag";; *) echo "";; esac; }
mihomo_asset(){ case "$(uname -m)" in x86_64|amd64) echo mihomo-linux-amd64-compatible;; aarch64|arm64) echo mihomo-linux-arm64;; armv7l|armv7*) echo mihomo-linux-armv7;; *) return 1;; esac; }

stop_service(){
    case "$(init_system)" in
        systemd) systemctl stop "$SERVICE_NAME" 2>/dev/null || true ;;
        openrc) rc-service "$SERVICE_NAME" stop 2>/dev/null || true ;;
    esac
}
start_service(){
    case "$(init_system)" in
        systemd) systemctl start "$SERVICE_NAME" ;;
        openrc) rc-service "$SERVICE_NAME" start ;;
        *) echo "Error: unsupported init system." >&2; exit 1 ;;
    esac
}
service_ok(){
    case "$(init_system)" in
        systemd) systemctl is-active --quiet "$SERVICE_NAME" ;;
        openrc) rc-service "$SERVICE_NAME" status >/dev/null 2>&1 ;;
        *) return 1 ;;
    esac
}

backup_dir="$DATA_DIR/backups/$(date +%Y%m%d%H%M%S)"
mkdir -p "$backup_dir"
[ -d "$CONFIG_DIR" ] && cp -a "$CONFIG_DIR" "$backup_dir/config"
[ -f "$DATA_DIR/3m-ui.db" ] && cp -p "$DATA_DIR/3m-ui.db" "$backup_dir/3m-ui.db"

panel_tmp="$(mktemp)"
mihomo_tmp=""
cleanup(){ rm -f "$panel_tmp" ${mihomo_tmp:+"$mihomo_tmp"} || true; }
trap cleanup EXIT INT TERM

panel_tag="$(latest_tag "https://github.com/$REPO")"
[ -n "$panel_tag" ] || { echo "Error: unable to determine latest 3m-ui release." >&2; exit 1; }
panel_url="https://github.com/$REPO/releases/download/${panel_tag}/3m-ui-linux-$(arch_name)$(asset_suffix)"
echo "[1/5] Downloading 3m-ui $panel_tag..."
download "$panel_url" "$panel_tmp"
chmod 0755 "$panel_tmp"
"$panel_tmp" --version >/dev/null 2>&1 || { echo "Error: downloaded 3m-ui failed validation; service was not touched." >&2; exit 1; }
echo "[OK] 3m-ui binary validated."

if [ "$UPDATE_MIHOMO" -eq 1 ] && [ -x "$MIHOMO_BIN" ]; then
    mihomo_tag="$(latest_tag https://github.com/MetaCubeX/mihomo)"
    if [ -n "$mihomo_tag" ]; then
        mihomo_tmp="$(mktemp)"
        url="https://github.com/MetaCubeX/mihomo/releases/download/${mihomo_tag}/$(mihomo_asset)-${mihomo_tag}.gz"
        echo "[2/5] Downloading Mihomo $mihomo_tag..."
        if download "$url" "$mihomo_tmp.gz" && gzip -dc "$mihomo_tmp.gz" > "$mihomo_tmp"; then
            chmod 0755 "$mihomo_tmp"
            if "$mihomo_tmp" -v >/dev/null 2>&1; then
                echo "[OK] Mihomo binary validated."
            else
                echo "[WARN] Mihomo validation failed; keeping current core."
                rm -f "$mihomo_tmp" "$mihomo_tmp.gz"; mihomo_tmp=""
            fi
        else
            echo "[WARN] Mihomo download/extraction failed; keeping current core."
            rm -f "$mihomo_tmp" "$mihomo_tmp.gz"; mihomo_tmp=""
        fi
    else
        echo "[WARN] Cannot determine latest Mihomo release; keeping current core."
    fi
else
    echo "[2/5] Mihomo update skipped."
fi

# Everything is downloaded and validated before stopping the running service.
echo "[3/5] Stopping 3m-ui..."
stop_service
old_panel="$BIN_PATH.update-old"
cp -p "$BIN_PATH" "$old_panel"
install -m 0755 "$panel_tmp" "$BIN_PATH"
if [ -n "${mihomo_tmp:-}" ] && [ -s "$mihomo_tmp" ]; then
    cp -p "$MIHOMO_BIN" "$MIHOMO_BIN.update-old" 2>/dev/null || true
    install -m 0755 "$mihomo_tmp" "$MIHOMO_BIN"
fi

echo "[4/5] Starting 3m-ui..."
if ! start_service || ! service_ok; then
    echo "[ERROR] New version failed to start. Rolling back..." >&2
    install -m 0755 "$old_panel" "$BIN_PATH"
    if [ -f "$MIHOMO_BIN.update-old" ]; then install -m 0755 "$MIHOMO_BIN.update-old" "$MIHOMO_BIN"; fi
    start_service || true
    rm -f "$old_panel" "$MIHOMO_BIN.update-old"
    echo "Rollback complete. Config backup: $backup_dir" >&2
    exit 1
fi

rm -f "$old_panel" "$MIHOMO_BIN.update-old"
echo "[5/5] Update completed successfully."
echo "Config/data backup: $backup_dir"
