#!/bin/bash
set -e

CONSUL_ADDR="${CONSUL_ADDR:-localhost:8500}"

wait_for_consul() {
    echo "Waiting for Consul to be ready..."
    until curl -s "http://${CONSUL_ADDR}/v1/status/leader" | grep -q '"'; do
        sleep 1
    done
    echo "Consul is ready"
}

put_config() {
    local key=$1
    local file=$2
    
    if [ -f "$file" ]; then
        echo "Uploading config: $key"
        curl -s -X PUT "http://${CONSUL_ADDR}/v1/kv/${key}" \
            --data-binary @"$file" > /dev/null
    else
        echo "Config file not found: $file"
    fi
}

wait_for_consul

put_config "config/auth/config.yaml" "/configs/auth/config.yaml"
put_config "config/user/config.yaml" "/configs/user/config.yaml"
put_config "config/asset/config.yaml" "/configs/asset/config.yaml"
put_config "config/gateway/config.yaml" "/configs/gateway/config.yaml"
put_config "config/worker/config.yaml" "/configs/worker/config.yaml"

echo "All configurations uploaded to Consul"
