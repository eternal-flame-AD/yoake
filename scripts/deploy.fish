#!/usr/bin/env fish

echo "Building UI..."
cd ui && yarn build && cd ..  \
    || exit 1

echo "Building server..."
cargo build --release \
    || exit 1

echo "Copying files..."
scp target/release/yoake_server config-prod.yaml yoake: \
    || exit 1

echo "Deploying..."
ssh yoake "
    echo 'Stopping server...'
    sudo systemctl stop yoake-server
    sudo mv yoake_server /var/lib/caddy/yoake/
    sudo mv config-prod.yaml /var/lib/caddy/yoake/config.yaml
    sudo chown -R caddy:caddy /var/lib/caddy/yoake/
    echo 'Starting server...'
    sudo systemctl start yoake-server
    " \
    || exit 1