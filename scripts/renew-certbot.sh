#!/usr/bin/env bash

set -Eeuo pipefail

docker run --rm \
  -v /etc/letsencrypt:/etc/letsencrypt \
  -v timexeed_certbot-webroot:/var/www/certbot \
  certbot/certbot:latest \
  renew \
  --webroot \
  --webroot-path=/var/www/certbot \
  --quiet

docker exec timexeed-nginx nginx -t
docker exec timexeed-nginx nginx -s reload

echo "[SUCCESS] Certificate renewal check completed."
