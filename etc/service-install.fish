#!/bin/env fish

make verify
 or exit 2

make build
 or exit 2


sudo cp etc/yoake-server.service /etc/systemd/system/yoake-server.service
    or exit 2

sudo systemctl daemon-reload
    or exit 2

sudo systemctl stop yoake-server.service

sudo find ~caddy/yoake -mindepth 1 -delete
    or exit 2

sudo -ucaddy mkdir -p ~caddy/yoake
    or exit 2

sudo -ucaddy make INSTALLDEST=(echo ~caddy/yoake) install
    or exit 2

cat config-prod.yml  | sudo -ucaddy tee ~caddy/yoake/config.yml > /dev/null
    or exit 2

sudo systemctl start yoake-server.service

systemctl status yoake-server.service