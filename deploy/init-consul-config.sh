#!/bin/sh
set -e

CONSUL_ADDR="${CONSUL_ADDR:-localhost:8500}"
UPLOAD_MARKER="${CONSUL_UPLOAD_MARKER:-/consul/data/.config-uploaded}"

wait_for_consul() {
    echo "Waiting for Consul to be ready..."
    while true; do
        if consul info 2>/dev/null | grep -q "leader = true"; then
            break
        fi
        sleep 1
    done
    echo "Consul is ready"
}

put_config() {
    key=$1
    file=$2
    
    if [ ! -f "$file" ]; then
        echo "Config file not found: $file"
        return
    fi

    echo "Uploading config: $key"
    attempt=0
    while true; do
        if consul kv put -http-addr="http://${CONSUL_ADDR}" "$key" "@$file" > /dev/null 2>&1; then
            break
        fi
        attempt=$((attempt + 1))
        if [ "$attempt" -ge 30 ]; then
            echo "Failed to upload config after retries: $key"
            exit 1
        fi
        sleep 1
    done
}

wait_for_consul

if [ -f "$UPLOAD_MARKER" ]; then
    echo "Consul config already uploaded. Skipping."
    exit 0
fi

put_config "config/auth/config.yaml" "/configs/auth/config.yaml"
put_config "config/user/config.yaml" "/configs/user/config.yaml"
put_config "config/gateway/config.yaml" "/configs/gateway/config.yaml"

touch "$UPLOAD_MARKER"
echo "All configurations uploaded to Consul"
