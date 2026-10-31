#!/usr/bin/env bash
# Server-side bootstrap for the Travel Guangdong website.
# Run on the target host from the extracted payload directory:
#   DASHSCOPE_API_KEY=... TENCENT_MAP_API_KEY=... bash bootstrap.sh
#
# Idempotent for files/services. The database seed is imported only when the
# users table is empty, so re-running will not duplicate demo data.
set -euo pipefail

APP_DIR=/opt/travel-gd
DB_NAME=travel_gd
DB_USER=travelgd_app
PUBLIC_PORT=8091
BACKEND_PORT=18080
PAYLOAD_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# Optional provider keys and model come from the environment. Empty is allowed.
: "${DASHSCOPE_API_KEY:=}"
: "${TENCENT_MAP_API_KEY:=}"
: "${AI_MODEL:=qwen-plus}"
: "${AMAP_WEB_KEY:=}"
: "${AMAP_ROUTE_URL:=https://restapi.amap.com/v3/direction}"
: "${AMAP_TRANSIT_CITY:=广州}"

echo "==> App user and directories"
if ! id travelgd >/dev/null 2>&1; then
    sudo -n useradd --system --home-dir "$APP_DIR" --shell /usr/sbin/nologin travelgd
fi
sudo -n mkdir -p "$APP_DIR"/{bin,web,etc}
sudo -n chown -R travelgd:travelgd "$APP_DIR"
sudo -n chmod 755 "$APP_DIR"

echo "==> Isolated MySQL schema and user"
DB_PW=$(openssl rand -hex 24)
JWT_SECRET=$(openssl rand -hex 32)
sudo -n mysql <<SQL
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\`
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PW}';
ALTER USER '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PW}';
CREATE USER IF NOT EXISTS '${DB_USER}'@'127.0.0.1' IDENTIFIED BY '${DB_PW}';
ALTER USER '${DB_USER}'@'127.0.0.1' IDENTIFIED BY '${DB_PW}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'127.0.0.1';
FLUSH PRIVILEGES;
SQL

sudo -n mysql < "$PAYLOAD_DIR/schema.sql"
USER_COUNT=$(sudo -n mysql -N -e "SELECT COUNT(*) FROM \`${DB_NAME}\`.users;")
if [ "$USER_COUNT" = "0" ]; then
    echo "    importing seed data"
    sudo -n mysql < "$PAYLOAD_DIR/seed.sql"
else
    echo "    seed data already present (users=${USER_COUNT}), skipping"
fi

echo "==> Backend binary and static site"
sudo -n install -o travelgd -g travelgd -m 0755 \
    "$PAYLOAD_DIR/travel-gd-server" "$APP_DIR/bin/travel-gd-server"
sudo -n find "$APP_DIR/web" -mindepth 1 -delete
sudo -n cp -r "$PAYLOAD_DIR/dist/." "$APP_DIR/web/"
sudo -n chown -R travelgd:travelgd "$APP_DIR/web"
sudo -n find "$APP_DIR/web" -type d -exec chmod 755 {} +
sudo -n find "$APP_DIR/web" -type f -exec chmod 644 {} +

echo "==> Environment file (0600 root)"
sudo -n tee "$APP_DIR/etc/app.env" >/dev/null <<EOF
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=${DB_NAME}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PW}
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRE_HOURS=72
DASHSCOPE_API_KEY=${DASHSCOPE_API_KEY}
TENCENT_MAP_API_KEY=${TENCENT_MAP_API_KEY}
AI_MODEL=${AI_MODEL}
AMAP_WEB_KEY=${AMAP_WEB_KEY}
AMAP_ROUTE_URL=${AMAP_ROUTE_URL}
AMAP_TRANSIT_CITY=${AMAP_TRANSIT_CITY}
PORT=${BACKEND_PORT}
HOST=127.0.0.1
GIN_MODE=release
EOF
sudo -n chown root:root "$APP_DIR/etc/app.env"
sudo -n chmod 600 "$APP_DIR/etc/app.env"

echo "==> systemd unit"
sudo -n install -m 0644 "$PAYLOAD_DIR/travel-gd.service" /etc/systemd/system/travel-gd.service
sudo -n systemctl daemon-reload
sudo -n systemctl enable travel-gd.service
sudo -n systemctl restart travel-gd.service

echo "==> nginx site on :${PUBLIC_PORT}"
sudo -n install -m 0644 "$PAYLOAD_DIR/nginx-travel-gd.conf" /etc/nginx/sites-available/travel-gd.conf
sudo -n ln -sfn /etc/nginx/sites-available/travel-gd.conf /etc/nginx/sites-enabled/travel-gd.conf
sudo -n nginx -t
sudo -n systemctl reload nginx

echo "==> Firewall"
sudo -n ufw allow "${PUBLIC_PORT}/tcp" >/dev/null

echo "==> Done"
echo "    backend : http://127.0.0.1:${BACKEND_PORT} (loopback only)"
echo "    website : http://<server-ip>:${PUBLIC_PORT}"
echo "    Remember to allow TCP ${PUBLIC_PORT} in the Azure NSG."
