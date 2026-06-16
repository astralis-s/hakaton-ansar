#!/usr/bin/env bash
# One-shot deploy for Amana (web + Telegram bot) on a fresh Ubuntu/Debian server.
# Installs Docker if needed, fetches the code, generates secrets, and starts the
# stack on port 80. Idempotent — safe to re-run to update.
#
# Usage (as root):
#   TELEGRAM_BOT_TOKEN='123:ABC' bash deploy/server-setup.sh
# or remotely:
#   curl -fsSL https://raw.githubusercontent.com/astralis-s/hakaton-ansar/<sha>/deploy/server-setup.sh | TELEGRAM_BOT_TOKEN='123:ABC' bash
set -euo pipefail

REPO="https://github.com/astralis-s/hakaton-ansar"
BRANCH="claude/confident-gates-64hts3"
DIR="/opt/amana"

say() { echo "==> $*"; }

rand() { openssl rand -hex "${1:-32}" 2>/dev/null || head -c "${1:-32}" /dev/urandom | od -An -tx1 | tr -d ' \n'; }

say "Amana deploy starting"
if [ -z "${TELEGRAM_BOT_TOKEN:-}" ]; then
  echo "   NOTE: TELEGRAM_BOT_TOKEN not set — the bot will be disabled (web still works)." >&2
fi

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

# 2) Free port 80 (stop common web servers that may already hold it)
for svc in apache2 nginx httpd lighttpd; do
  systemctl stop "$svc" 2>/dev/null || true
  systemctl disable "$svc" 2>/dev/null || true
done

# 3) Code
if ! command -v git >/dev/null 2>&1; then
  apt-get update -y && apt-get install -y git || true
fi
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
grep -q '^JWT_SECRET=' .env       || echo "JWT_SECRET=$(rand 32)"     >> .env
grep -q '^POSTGRES_PASSWORD=' .env || echo "POSTGRES_PASSWORD=$(rand 16)" >> .env
grep -v '^TELEGRAM_BOT_TOKEN=' .env > .env.new 2>/dev/null || true
mv .env.new .env
echo "TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}" >> .env
chmod 600 .env

# 5) Build & run
say "Building and starting the stack (first build downloads images + deps; a few minutes)"
docker compose -f docker-compose.prod.yml up -d --build

# 6) Wait for health and report
say "Waiting for the app to become healthy"
ok=0
for _ in $(seq 1 60); do
  if curl -fsS http://localhost/health >/dev/null 2>&1; then ok=1; break; fi
  sleep 3
done
echo
docker compose -f docker-compose.prod.yml ps || true
IP="$(curl -fsS https://api.ipify.org 2>/dev/null || echo '<server-ip>')"
echo
echo "============================================================"
if [ "$ok" = 1 ]; then
  echo " ✅ Amana is UP:  http://$IP/"
else
  echo " ⚠  App did not report healthy yet. Check logs:"
  echo "    docker compose -f $DIR/docker-compose.prod.yml logs -f app"
fi
echo "    Login (demo):  owner@amana.ru / owner12345"
echo "    Swagger:       http://$IP/swagger/"
echo "    Telegram bot:  $([ -n "${TELEGRAM_BOT_TOKEN:-}" ] && echo 'ENABLED' || echo 'disabled (no token)')"
echo "============================================================"
