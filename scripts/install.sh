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
    [ "$(id -u)" -eq 0 ] || {
        echo "Please run as root."
        exit 1
    }
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
            echo "Unsupported architecture: $(uname -m)" >&2
            exit 1
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


    echo "Unsupported init system"
    exit 1
}


latest_url() {

    echo "https://github.com/$REPO/releases/latest/download/3m-ui-linux-$1"

}


download() {

    if command_exists curl; then

        curl -fsSL "$1" -o "$2"

    elif command_exists wget; then

        wget -qO "$2" "$1"

    else

        echo "curl or wget required"
        exit 1

    fi

}


write_config() {

    [ -f "$CONFIG_DIR/config.yaml" ] && return


    jwt_secret="$(date +%s | sha256sum | awk '{print $1}')"

    cred_key="$(date +%s%N | sha256sum | awk '{print $1}')"


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

Restart=always
RestartSec=5

Environment=THREE_M_UI_CONFIG=$CONFIG_DIR/config.yaml

WorkingDirectory=$DATA_DIR


[Install]
WantedBy=multi-user.target

EOF


systemctl daemon-reload

systemctl enable --now $SERVICE_NAME

}



install_openrc() {


cat >/etc/init.d/$SERVICE_NAME <<EOF

#!/sbin/openrc-run


name="3m-ui"

description="3m-ui server panel"


command="$BIN_PATH"

command_background=yes


pidfile="/run/3m-ui.pid"


directory="$DATA_DIR"


export THREE_M_UI_CONFIG="$CONFIG_DIR/config.yaml"


depend(){
    need net
}

EOF


chmod +x /etc/init.d/$SERVICE_NAME


rc-update add $SERVICE_NAME default

rc-service $SERVICE_NAME start


}



need_root


mkdir -p \
"$INSTALL_DIR" \
"$CONFIG_DIR" \
"$DATA_DIR/mihomo" \
"$LOG_DIR"



arch="$(arch_name)"

tmp="$(mktemp)"


download \
"$(latest_url "$arch")" \
"$tmp"



install -m755 "$tmp" "$BIN_PATH"


rm -f "$tmp"



write_config



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
