#!/usr/bin/env sh
set -eu

REPO="dzx941/3m-ui"

INSTALL_DIR="/usr/local/bin"
BIN_PATH="$INSTALL_DIR/3m-ui"

CONFIG_DIR="/etc/3m-ui"
DATA_DIR="/var/lib/3m-ui"
LOG_DIR="/var/log/3m-ui"

SERVICE_NAME="3m-ui"


need_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "Please run as root."
        exit 1
    fi
}


command_exists() {
    command -v "$1" >/dev/null 2>&1
}


arch_name() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo amd64
            ;;
        aarch64|arm64)
            echo arm64
            ;;
        armv7l|armv7*)
            echo armv7
            ;;
        *)
            echo "Unsupported architecture"
            exit 1
            ;;
    esac
}


os_id() {
    if [ -r /etc/os-release ]; then
        . /etc/os-release
        echo "${ID:-unknown}"
    else
        echo unknown
    fi
}

install_prerequisites() {
    # Best-effort support for common Linux distributions. Static release
    # binaries do not require sqlite CGO runtime libraries, but these tools are
    # still useful for installation and service management.
    if command_exists apt-get; then
        apt-get update
        apt-get install -y ca-certificates curl
    elif command_exists dnf; then
        dnf install -y ca-certificates curl
    elif command_exists yum; then
        yum install -y ca-certificates curl
    elif command_exists apk; then
        apk add --no-cache ca-certificates curl
    elif command_exists pacman; then
        pacman -Sy --noconfirm ca-certificates curl
    elif command_exists zypper; then
        zypper --non-interactive install ca-certificates curl
    fi
}

asset_suffix() {
    case "${THREE_M_UI_STATIC:-auto}" in
        1|true|yes)
            echo "-static"
            ;;
        0|false|no)
            echo ""
            ;;
        *)
            # Alpine and other musl-based systems should use the static
            # modernc.org/sqlite build automatically.
            if [ -r /etc/os-release ] && grep -q '^ID=alpine' /etc/os-release; then
                echo "-static"
            elif [ -e /lib/ld-musl-x86_64.so.1 ] ||
                 [ -e /lib/ld-musl-aarch64.so.1 ] ||
                 [ -e /lib/ld-musl-armhf.so.1 ]; then
                echo "-static"
            else
                echo ""
            fi
            ;;
    esac
}


init_system() {

    if command_exists systemctl &&
       [ -d /run/systemd/system ]; then

        echo systemd
        return

    fi


    if command_exists rc-service; then

        echo openrc
        return

    fi


    echo "No supported init system"
    exit 1
}



download() {

    url="$1"
    output="$2"


    if command_exists curl; then

        curl -fsSL "$url" -o "$output"

    elif command_exists wget; then

        wget -qO "$output" "$url"

    else

        echo "curl/wget required"
        exit 1

    fi
}




install_mihomo() {
    MIHOMO_DIR="$DATA_DIR/mihomo"
    MIHOMO_BIN="/usr/local/bin/mihomo"

    if [ -x "$MIHOMO_BIN" ]; then
        return
    fi

    mkdir -p "$MIHOMO_DIR"

    # Dynamic lookup of latest MetaCubeX/mihomo release version tag (No API hit)
    tag_name=$(curl -sI https://github.com/MetaCubeX/mihomo/releases/latest | grep -i "location" | awk -F'/' '{print $NF}' | tr -d '\r\n ')
    if [ -z "$tag_name" ]; then
        tag_name="v1.19.29"
    fi

    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) filename="mihomo-linux-amd64-compatible-${tag_name}.gz" ;;
        aarch64|arm64) filename="mihomo-linux-arm64-${tag_name}.gz" ;;
        armv7l|armv7*) filename="mihomo-linux-armv7-${tag_name}.gz" ;;
        *) echo "Unsupported mihomo architecture: $arch"; return ;;
    esac

    echo "Downloading Mihomo $tag_name for $arch..."
    tmp="$(mktemp)"
    url="https://github.com/MetaCubeX/mihomo/releases/download/${tag_name}/${filename}"

    if download "$url" "$tmp" && [ -s "$tmp" ]; then
        gzip -dc "$tmp" > "$MIHOMO_BIN" 2>/dev/null || true
        if [ -s "$MIHOMO_BIN" ]; then
            chmod 755 "$MIHOMO_BIN" || true
            echo "Mihomo installed successfully to $MIHOMO_BIN"
        else
            echo "Error: Unpacked Mihomo binary is empty. Exiting." >&2
            rm -f "$tmp" "$MIHOMO_BIN"
            exit 1
        fi
    else
        echo "Error: Failed to download Mihomo from $url. Exiting." >&2
        rm -f "$tmp"
        exit 1
    fi
    rm -f "$tmp"
}

install_subconverter() {
    SUBCONVERTER_DIR="/usr/local/subconverter"
    SUBCONVERTER_BIN="/usr/local/bin/subconverter"

    if [ -x "$SUBCONVERTER_BIN" ]; then
        echo "Subconverter is already installed."
        return
    fi

    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) sub_arch="linux64" ;;
        aarch64|arm64) sub_arch="aarch64" ;;
        armv7l|armv7*) sub_arch="armv7" ;;
        *)
            echo "Warning: Unsupported subconverter architecture: $arch. Skipping subconverter installation."
            return
            ;;
    esac

    echo "Installing subconverter for $arch..."
    tmp="$(mktemp)"
    url="https://github.com/asdlokj1qpi233/subconverter/releases/download/v0.9.9/subconverter_${sub_arch}.tar.gz"

    if download "$url" "$tmp" && [ -s "$tmp" ]; then
        mkdir -p "/usr/local"
        if tar -xzf "$tmp" -C "/usr/local" 2>/dev/null; then
            if [ -x "$SUBCONVERTER_DIR/subconverter" ]; then
                ln -sf "$SUBCONVERTER_DIR/subconverter" "$SUBCONVERTER_BIN" || true
                echo "Subconverter installed successfully to $SUBCONVERTER_BIN"
            else
                echo "Warning: Subconverter executable not found after unpack. Skipping subconverter."
            fi
        else
            echo "Warning: Subconverter tar unpack failed. Skipping subconverter."
        fi
    else
        echo "Warning: Failed to download subconverter from $url. Skipping subconverter."
    fi
    rm -f "$tmp"
}

write_config() {

    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        return
    fi


    jwt_secret=$(date +%s | sha256sum | awk '{print $1}')

    cred_key=$(date +%s%N | sha256sum | awk '{print $1}')


cat > "$CONFIG_DIR/config.yaml" <<EOF
server:
  port: 8080
  mode: release

database:
  path: "$DATA_DIR/3m-ui.db"

jwt:
  secret: "$jwt_secret"

security:
  credential_key: "$cred_key"

mihomo:
  binary: "/usr/local/bin/mihomo"
  config: "$DATA_DIR/mihomo/config.yaml"
EOF

}



install_systemd() {


cat >/etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=3m-ui server panel
After=network-online.target
Wants=network-online.target


[Service]
Type=simple

ExecStart=$BIN_PATH

Environment=THREE_M_UI_CONFIG=$CONFIG_DIR/config.yaml

WorkingDirectory=$DATA_DIR

Restart=always
RestartSec=5


[Install]
WantedBy=multi-user.target
EOF



systemctl daemon-reload

systemctl enable --now $SERVICE_NAME

install_subconverter_systemd

}

install_subconverter_systemd() {
    if [ ! -x "/usr/local/bin/subconverter" ]; then
        return
    fi

cat >/etc/systemd/system/subconverter.service <<EOF
[Unit]
Description=Subconverter subscription conversion utility
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/subconverter
WorkingDirectory=/usr/local/subconverter
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable --now subconverter
}



install_openrc() {


cat > /etc/init.d/$SERVICE_NAME <<EOF
#!/sbin/openrc-run

name="3m-ui"
description="3m-ui server panel"

command="$BIN_PATH"

command_background="yes"

pidfile="/run/$SERVICE_NAME.pid"

directory="$DATA_DIR"


export THREE_M_UI_CONFIG="$CONFIG_DIR/config.yaml"


depend() {
    need net
}

EOF


chmod +x /etc/init.d/$SERVICE_NAME


rc-update add $SERVICE_NAME default


rc-service $SERVICE_NAME restart

install_subconverter_openrc

}

install_subconverter_openrc() {
    if [ ! -x "/usr/local/bin/subconverter" ]; then
        return
    fi

cat > /etc/init.d/subconverter <<EOF
#!/sbin/openrc-run

name="subconverter"
description="Subconverter subscription conversion utility"

command="/usr/local/bin/subconverter"
command_background="yes"
pidfile="/run/subconverter.pid"
directory="/usr/local/subconverter"

depend() {
    need net
}
EOF

    chmod +x /etc/init.d/subconverter
    rc-update add subconverter default
    rc-service subconverter restart
}



need_root


echo "Detected Linux distribution: $(os_id)"
install_prerequisites


mkdir -p \
"$INSTALL_DIR" \
"$CONFIG_DIR" \
"$DATA_DIR/mihomo" \
"$LOG_DIR"



arch=$(arch_name)


tmp=$(mktemp)


download \
"https://github.com/$REPO/releases/latest/download/3m-ui-linux-$arch$(asset_suffix)" \
"$tmp"

if [ ! -s "$tmp" ]; then
    echo "Error: Downloaded 3m-ui binary is empty. Exiting." >&2
    rm -f "$tmp"
    exit 1
fi

install -m 755 "$tmp" "$BIN_PATH"


rm -f "$tmp"


write_config


install_mihomo


install_subconverter


case "$(init_system)" in

systemd)
    install_systemd
    ;;

openrc)
    install_openrc
    ;;

esac



echo ""
echo "3m-ui installed successfully"
echo "Open http://SERVER_IP:8080/"
if [ -f "$DATA_DIR/.initial_admin_password" ]; then
    echo "Initial admin username: admin"
    echo "Initial admin password file: $DATA_DIR/.initial_admin_password"
fi
echo "Static build selection: ${THREE_M_UI_STATIC:-auto} (Alpine/musl automatically uses -static)"
