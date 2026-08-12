#!/usr/bin/env sh
set -eu
umask 077

REPO="dzx941/3m-ui"
BIN_PATH="/usr/local/bin/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
SERVICE_NAME="3m-ui"
MIHOMO_BIN="/usr/local/bin/mihomo"
SUBCONVERTER_DIR="/usr/local/subconverter"
SUBCONVERTER_BIN="/usr/local/bin/subconverter"

[ "$(id -u)" -eq 0 ] || { echo "Error: Please run as root." >&2; exit 1; }
[ -x "$BIN_PATH" ] || { echo "Error: 3m-ui is not installed at $BIN_PATH" >&2; exit 1; }

command_exists(){ command -v "$1" >/dev/null 2>&1; }
arch_name(){
    case "$(uname -m)" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        armv7l|armv7*) echo armv7 ;;
        *) echo "Error: Unsupported architecture" >&2; exit 1 ;;
    esac
}
asset_suffix(){
    case "${THREE_M_UI_STATIC:-auto}" in
        1|true|yes) echo "-static" ;;
        0|false|no) echo "" ;;
        *)
            if [ -r /etc/os-release ] && grep -q '^ID=alpine' /etc/os-release; then
                echo "-static"
            elif [ -e /lib/ld-musl-x86_64.so.1 ] || [ -e /lib/ld-musl-aarch64.so.1 ] || [ -e /lib/ld-musl-armhf.so.1 ]; then
                echo "-static"
            else
                echo ""
            fi
            ;;
    esac
}
download(){
    if command_exists curl; then
        curl -fsSL --retry 3 --connect-timeout 10 "$1" -o "$2"
    elif command_exists wget; then
        wget -qO "$2" "$1"
    else
        echo "Error: curl or wget is required." >&2
        exit 1
    fi
}

latest_mihomo_tag(){
    tag=""
    if command_exists curl; then
        tag="$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/MetaCubeX/mihomo/releases/latest 2>/dev/null | sed 's#.*/##')" || true
    fi
    case "$tag" in
        v[0-9]*) echo "$tag" ;;
        *) echo "" ;;
    esac
}

mihomo_asset(){
    case "$(uname -m)" in
        x86_64|amd64) echo "mihomo-linux-amd64-compatible" ;;
        aarch64|arm64) echo "mihomo-linux-arm64" ;;
        armv7l|armv7*) echo "mihomo-linux-armv7" ;;
        *) return 1 ;;
    esac
}

MIHOMO_STAGE_BIN=""
SUBCONVERTER_STAGE_DIR=""
PANEL_STAGE_BIN=""

cleanup(){
    [ -n "${PANEL_STAGE_BIN:-}" ] && rm -f "$PANEL_STAGE_BIN" || true
    [ -n "${MIHOMO_STAGE_BIN:-}" ] && rm -f "$MIHOMO_STAGE_BIN" || true
    [ -n "${SUBCONVERTER_STAGE_DIR:-}" ] && rm -rf "$SUBCONVERTER_STAGE_DIR" || true
}
trap cleanup EXIT INT TERM

stage_mihomo(){
    asset="$(mihomo_asset)"
    tag="$(latest_mihomo_tag)"
    if [ -z "$tag" ]; then
        echo "Warning: unable to determine latest Mihomo release; retaining current core." >&2
        return 0
    fi

    tmp="$(mktemp)"
    stage="$(mktemp)"
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag}/${asset}-${tag}.gz"
    echo "Checking Mihomo core update ($tag)..."

    if ! download "$url" "$tmp" || [ ! -s "$tmp" ]; then
        echo "Warning: Mihomo download failed; retaining current core." >&2
        rm -f "$tmp" "$stage"
        return 0
    fi
    if ! gzip -dc "$tmp" > "$stage" || [ ! -s "$stage" ]; then
        echo "Warning: Mihomo extraction failed; retaining current core." >&2
        rm -f "$tmp" "$stage"
        return 0
    fi
    chmod 0755 "$stage"
    if ! "$stage" -v >/dev/null 2>&1; then
        echo "Warning: Mihomo executable validation failed; retaining current core." >&2
        rm -f "$tmp" "$stage"
        return 0
    fi
    MIHOMO_STAGE_BIN="$stage"
    rm -f "$tmp"
    echo "Mihomo validation passed."
}

stage_subconverter(){
    case "$(uname -m)" in
        x86_64|amd64) asset="linux64" ;;
        aarch64|arm64) asset="aarch64" ;;
        armv7l|armv7*) asset="armv7" ;;
        *) return 0 ;;
    esac

    if [ ! -x "$SUBCONVERTER_BIN" ]; then
        echo "Subconverter is not installed; skipping subconverter update."
        return 0
    fi

    tmp="$(mktemp)"
    stage_dir="$(mktemp -d)"
    url="https://github.com/asdlokj1qpi233/subconverter/releases/download/v0.9.9/subconverter_${asset}.tar.gz"
    echo "Checking subconverter update..."

    if ! download "$url" "$tmp" || [ ! -s "$tmp" ]; then
        echo "Warning: subconverter download failed; retaining current installation." >&2
        rm -f "$tmp"
        rm -rf "$stage_dir"
        return 0
    fi
    if ! tar -xzf "$tmp" -C "$stage_dir" || [ ! -x "$stage_dir/subconverter/subconverter" ]; then
        echo "Warning: subconverter archive validation failed; retaining current installation." >&2
        rm -f "$tmp"
        rm -rf "$stage_dir"
        return 0
    fi

    mkdir -p "$stage_dir/subconverter/base"
    cat > "$stage_dir/subconverter/base/pref.ini" <<'EOF'
[server]
listen=127.0.0.1
port=25500
EOF

    SUBCONVERTER_STAGE_DIR="$stage_dir"
    rm -f "$tmp"
    echo "Subconverter validation passed."
}

stop_services(){
    if command_exists systemctl; then
        [ -f "/etc/systemd/system/$SERVICE_NAME.service" ] && systemctl stop "$SERVICE_NAME" || true
        [ -f "/etc/systemd/system/subconverter.service" ] && systemctl stop subconverter || true
    elif command_exists rc-service; then
        [ -f "/etc/init.d/$SERVICE_NAME" ] && rc-service "$SERVICE_NAME" stop || true
        [ -f /etc/init.d/subconverter ] && rc-service subconverter stop || true
    fi
}

start_services(){
    if command_exists systemctl; then
        [ -f "/etc/systemd/system/$SERVICE_NAME.service" ] && systemctl start "$SERVICE_NAME" || true
        [ -f /etc/systemd/system/subconverter.service ] && systemctl start subconverter || true
    elif command_exists rc-service; then
        [ -f "/etc/init.d/$SERVICE_NAME" ] && rc-service "$SERVICE_NAME" start || true
        [ -f /etc/init.d/subconverter ] && rc-service subconverter start || true
    fi
}

backup_dir="$DATA_DIR/backups/$(date +%Y%m%d%H%M%S)"
mkdir -p "$backup_dir"
[ -d "$CONFIG_DIR" ] && cp -a "$CONFIG_DIR" "$backup_dir/config"

# Stage and validate everything before stopping a running service.
tmp="$(mktemp)"
panel_url="https://github.com/$REPO/releases/latest/download/3m-ui-linux-$(arch_name)$(asset_suffix)"
echo "Checking 3m-ui panel update..."
if ! download "$panel_url" "$tmp" || [ ! -s "$tmp" ]; then
    echo "Error: failed to download 3m-ui. Current service was not touched." >&2
    rm -f "$tmp"
    exit 1
fi
chmod 0755 "$tmp"
if ! "$tmp" --version >/dev/null 2>&1; then
    echo "Error: downloaded 3m-ui binary failed executable validation. Current service was not touched." >&2
    rm -f "$tmp"
    exit 1
fi
PANEL_STAGE_BIN="$tmp"
echo "3m-ui validation passed."

stage_mihomo
stage_subconverter

stop_services

# Atomic-ish replacement: install/move only after all staged downloads validated.
install -m 0755 "$PANEL_STAGE_BIN" "$BIN_PATH"
rm -f "$PANEL_STAGE_BIN"
PANEL_STAGE_BIN=""

if [ -n "$MIHOMO_STAGE_BIN" ] && [ -s "$MIHOMO_STAGE_BIN" ]; then
    install -m 0755 "$MIHOMO_STAGE_BIN" "$MIHOMO_BIN"
    rm -f "$MIHOMO_STAGE_BIN"
    MIHOMO_STAGE_BIN=""
fi

if [ -n "$SUBCONVERTER_STAGE_DIR" ] && [ -x "$SUBCONVERTER_STAGE_DIR/subconverter/subconverter" ]; then
    rm -rf "$SUBCONVERTER_DIR"
    mv "$SUBCONVERTER_STAGE_DIR/subconverter" "$SUBCONVERTER_DIR"
    ln -sfn "$SUBCONVERTER_DIR/subconverter" "$SUBCONVERTER_BIN"
    SUBCONVERTER_STAGE_DIR=""
fi

start_services

echo "3m-ui updated successfully."
echo "Config backup: $backup_dir"
