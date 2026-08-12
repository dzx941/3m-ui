#!/usr/bin/env sh
set -eu
REPO="dzx941/3m-ui"
BIN_PATH="/usr/local/bin/3m-ui"
CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
SERVICE_NAME="3m-ui"

[ "$(id -u)" -eq 0 ] || { echo "Error: Please run as root." >&2; exit 1; }
[ -x "$BIN_PATH" ] || { echo "Error: 3m-ui is not installed at $BIN_PATH" >&2; exit 1; }

command_exists(){ command -v "$1" >/dev/null 2>&1; }
arch_name(){ case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; armv7l|armv7*) echo armv7;; *) echo "Error: Unsupported architecture" >&2; exit 1;; esac; }
asset_suffix(){ 
    case "${THREE_M_UI_STATIC:-auto}" in
      1|true|yes) echo "-static";;
      0|false|no) echo "";;
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
download(){ if command_exists curl; then curl -fsSL "$1" -o "$2"; elif command_exists wget; then wget -qO "$2" "$1"; else echo "Error: curl or wget is required." >&2; exit 1; fi; }

# Initialize validation variables
MIHOMO_STAGE_BIN=""
SUBCONVERTER_STAGE_DIR=""
PANEL_STAGE_BIN=""

install_mihomo_update() {
    # Dynamic latest version lookup
    tag_name=$(curl -sI https://github.com/MetaCubeX/mihomo/releases/latest | grep -i "location" | awk -F'/' '{print $NF}' | tr -d '\r\n ')
    if [ -z "$tag_name" ]; then
        tag_name="v1.19.29"
    fi

    case "$(uname -m)" in
      x86_64|amd64) filename="mihomo-linux-amd64-compatible-${tag_name}.gz";;
      aarch64|arm64) filename="mihomo-linux-arm64-${tag_name}.gz";;
      armv7l|armv7*) filename="mihomo-linux-armv7-${tag_name}.gz";;
      *) return;;
    esac

    echo "Checking Mihomo core update ($tag_name)..."
    tmp="$(mktemp)"
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag_name}/${filename}"
    if download "$url" "$tmp" && [ -s "$tmp" ]; then
       stage_bin=$(mktemp)
       if gzip -dc "$tmp" > "$stage_bin" 2>/dev/null && [ -s "$stage_bin" ]; then
          chmod 755 "$stage_bin"
          MIHOMO_STAGE_BIN="$stage_bin"
          echo "Mihomo download and validation passed."
       else
          echo "Warning: Mihomo unpack failed. Retaining current installation." >&2
          rm -f "$stage_bin"
       fi
       rm -f "$tmp"
    else
       echo "Warning: Mihomo download failed from $url. Retaining current installation." >&2
       rm -f "$tmp"
    fi
}

install_subconverter_update() {
    case "$(uname -m)" in
      x86_64|amd64) asset="linux64";;
      aarch64|arm64) asset="aarch64";;
      armv7l|armv7*) asset="armv7";;
      *) return;;
    esac

    echo "Checking subconverter update..."
    tmp="$(mktemp)"
    url="https://github.com/asdlokj1qpi233/subconverter/releases/download/v0.9.9/subconverter_${asset}.tar.gz"
    if download "$url" "$tmp" && [ -s "$tmp" ]; then
       stage_dir=$(mktemp -d)
       if tar -xzf "$tmp" -C "$stage_dir" 2>/dev/null && [ -x "$stage_dir/subconverter/subconverter" ]; then
          SUBCONVERTER_STAGE_DIR="$stage_dir"
          echo "Subconverter download and validation passed."
       else
          echo "Warning: Subconverter unpack failed. Retaining current installation." >&2
          rm -rf "$stage_dir"
       fi
       rm -f "$tmp"
    else
       echo "Warning: Subconverter download failed from $url. Retaining current installation." >&2
       rm -f "$tmp"
    fi
}

stop_services(){
    if command_exists systemctl; then
        [ -f /etc/systemd/system/$SERVICE_NAME.service ] && systemctl stop $SERVICE_NAME || true
        [ -f /etc/systemd/system/subconverter.service ] && systemctl stop subconverter || true
    elif command_exists rc-service; then
        [ -f /etc/init.d/$SERVICE_NAME ] && rc-service $SERVICE_NAME stop || true
        [ -f /etc/init.d/subconverter ] && rc-service subconverter stop || true
    fi
}

start_services(){
    if command_exists systemctl; then
        [ -f /etc/systemd/system/$SERVICE_NAME.service ] && systemctl start $SERVICE_NAME || true
        [ -f /etc/systemd/system/subconverter.service ] && systemctl start subconverter || true
    elif command_exists rc-service; then
        [ -f /etc/init.d/$SERVICE_NAME ] && rc-service $SERVICE_NAME start || true
        [ -f /etc/init.d/subconverter ] && rc-service subconverter start || true
    fi
}

# 1. Download and validate new 3m-ui binary FIRST
backup_dir="$DATA_DIR/backups/$(date +%Y%m%d%H%M%S)"
mkdir -p "$backup_dir"
[ -d "$CONFIG_DIR" ] && cp -a "$CONFIG_DIR" "$backup_dir/config"

echo "Checking 3m-ui panel update..."
tmp="$(mktemp)"
url="https://github.com/$REPO/releases/latest/download/3m-ui-linux-$(arch_name)$(asset_suffix)"

if download "$url" "$tmp" && [ -s "$tmp" ]; then
    PANEL_STAGE_BIN="$tmp"
    echo "3m-ui download and validation passed."
else
    echo "Error: Failed to download 3m-ui from $url. Aborting update to protect current installation." >&2
    rm -f "$tmp"
    exit 1
fi

# 2. Download and validate other components
install_mihomo_update
install_subconverter_update

# 3. Stop Services ONLY after all components are successfully validated
stop_services

# 4. Perform Safe In-Place Replacements
# Replace 3m-ui panel
if [ -n "$PANEL_STAGE_BIN" ] && [ -f "$PANEL_STAGE_BIN" ]; then
    install -m 0755 "$PANEL_STAGE_BIN" "$BIN_PATH"
    rm -f "$PANEL_STAGE_BIN"
fi

# Replace Mihomo Core (if downloaded and validated)
if [ -n "$MIHOMO_STAGE_BIN" ] && [ -f "$MIHOMO_STAGE_BIN" ]; then
    mv -f "$MIHOMO_STAGE_BIN" "/usr/local/bin/mihomo"
    chmod 755 "/usr/local/bin/mihomo"
fi

# Replace Subconverter (if downloaded and validated)
if [ -n "$SUBCONVERTER_STAGE_DIR" ] && [ -d "$SUBCONVERTER_STAGE_DIR/subconverter" ]; then
    rm -rf /usr/local/subconverter
    mv -f "$SUBCONVERTER_STAGE_DIR/subconverter" "/usr/local/subconverter"
    ln -sf /usr/local/subconverter/subconverter "/usr/local/bin/subconverter"
    rm -rf "$SUBCONVERTER_STAGE_DIR"
fi

# 5. Restart Services
start_services

echo "3m-ui updated successfully. Config backup: $backup_dir"
