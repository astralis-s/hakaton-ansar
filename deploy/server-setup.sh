#!/usr/bin/env bash
# One-shot deploy for Amana (web + Telegram bot) on a fresh Ubuntu/Debian server.
# Installs Docker if needed, fetches the code, generates secrets, and starts the
# stack. Idempotent — safe to re-run to update.
#
# Plain HTTP on port 80:
#   TELEGRAM_BOT_TOKEN='123:ABC' bash deploy/server-setup.sh
# HTTPS for a domain (after its DNS A-record points to this server):
#   DOMAIN='amana-crm.ru' TELEGRAM_BOT_TOKEN='123:ABC' bash deploy/server-setup.sh
set -euo pipefail

REPO="https://github.com/astralis-s/hakaton-ansar"
BRANCH="claude/confident-gates-64hts3"
DIR="/opt/amana"

say() { echo "==> $*"; }
rand() { openssl rand -hex "${1:-32}" 2>/dev/null || head -c "${1:-32}" /dev/urandom | od -An -tx1 | tr -d ' \n'; }

say "Amana deploy starting"
[ -z "${TELEGRAM_BOT_TOKEN:-}" ] && echo "   NOTE: TELEGRAM_BOT_TOKEN not set — the bot will be disabled (web still works)." >&2

# 1) Docker (+ compose plugin)
if ! command -v docker >/dev/null 2>&1; then
  say "Installing Docker"
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker 2>/dev/null || true
if ! docker compose version >/dev/null 2>&1; then
  echo "ERROR: 'docker compose' plugin is unavailable. Install docker-compose-plugin and re-run." >&2
  exit 1
fi

# 2) Free ports 80/443 (stop common web servers that may already hold them)
for svc in apache2 nginx httpd lighttpd caddy; do
  systemctl stop "$svc" 2>/dev/null || true
  systemctl disable "$svc" 2>/dev/null || true
done

# 3) Code
command -v git >/dev/null 2>&1 || { apt-get update -y && apt-get install -y git; }
if [ -d "$DIR/.git" ]; then
  say "Updating existing checkout at $DIR"
  git -C "$DIR" fetch origin "$BRANCH"
  git -C "$DIR" checkout -B "$BRANCH" "origin/$BRANCH"
  git -C "$DIR" reset --hard "origin/$BRANCH"
else
  say "Cloning $REPO ($BRANCH)"
  rm -rf "$DIR"
  git clone -b "$BRANCH" "$REPO" "$DIR"
fi
cd "$DIR"

# 4) Secrets in .env (generated once, kept stable across redeploys; token refreshed)
touch .env
grep -q '^JWT_SECRET=' .env       || echo "JWT_SECRET=$(rand 32)"        >> .env
grep -q '^POSTGRES_PASSWORD=' .env || echo "POSTGRES_PASSWORD=$(rand 16)" >> .env
grep -v -E '^(TELEGRAM_BOT_TOKEN|SITE_ADDRESS)=' .env > .env.new 2>/dev/null || true
mv .env.new .env
echo "TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}" >> .env

# 5) Choose HTTP (port 80) or HTTPS (Caddy + Let's Encrypt) based on DOMAIN
if [ -n "${DOMAIN:-}" ]; then
  echo "SITE_ADDRESS=${DOMAIN}" >> .env
  COMPOSE="docker-compose.tls.yml"
  say "Deploying with HTTPS for ${DOMAIN} (Caddy will request a Let's Encrypt cert)."
  echo "   Make sure ${DOMAIN}'s DNS A-record already points to this server, or the cert will fail."
else
  COMPOSE="docker-compose.prod.yml"
  say "Deploying plain HTTP on port 80 (no DOMAIN set)."
fi
chmod 600 .env

# 6) Build & run
say "Building and starting the stack (first build downloads images + deps; a few minutes)"
docker compose -f "$COMPOSE" up -d --build

# 7) Wait for health and report
say "Waiting for the app to become healthy"
ok=0
for _ in $(seq 1 60); do
  if curl -fsS http://localhost/health >/dev/null 2>&1 || curl -fsSk https://localhost/health >/dev/null 2>&1; then ok=1; break; fi
  sleep 3
done
echo
docker compose -f "$COMPOSE" ps || true
IPS="$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -E '^[0-9]' | grep -vE '^(127\.|172\.1[7-9]\.|172\.2[0-9]\.|172\.3[0-1]\.|10\.)' | paste -sd', ' -)"
echo
echo "============================================================"
if [ "$ok" = 1 ]; then echo " ✅ Amana is UP"; else echo " ⚠  Not healthy yet — see: docker compose -f $DIR/$COMPOSE logs -f"; fi
if [ -n "${DOMAIN:-}" ]; then
  echo "    URL:   https://${DOMAIN}/   (works once DNS → this server has propagated)"
else
  echo "    URL:   http://<this-server-ip>/   (the IP you SSH'd into)"
fi
[ -n "$IPS" ] && echo "    This server's public IP(s): ${IPS}"
echo "    Login (demo):  owner@amana.ru / owner12345"
echo "    Swagger:       /swagger/"
echo "    Telegram bot:  $([ -n "${TELEGRAM_BOT_TOKEN:-}" ] && echo 'ENABLED' || echo 'disabled (no token)')"
echo "============================================================"
