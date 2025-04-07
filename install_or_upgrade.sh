#!/bin/bash

set -euo pipefail

APP_NAME="file-move-trigger"
BIN_PATH="/usr/local/sbin/${APP_NAME}"
CONFIG_DIR="/etc/${APP_NAME}"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
TIMER_FILE="/etc/systemd/system/${APP_NAME}.timer"

echo "📦 Building Go binary..."
go build -o "${APP_NAME}.tmp" main.go || {
  echo "❌ Failed to build the binary. Exiting."
  exit 1
}

sudo install -m 755 -o root -g root "${APP_NAME}.tmp" "${BIN_PATH}" || {
  echo "❌ Failed to install the binary. Exiting."
  exit 1
}

echo "✅ Installed binary to ${BIN_PATH}"

echo "📁 Ensuring config directory exists..."
sudo install -d -m 755 -o root -g root "${CONFIG_DIR}" || {
  echo "❌ Failed to create config directory at ${CONFIG_DIR}. Exiting."
  exit 1
}

if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "📝 Creating default config at ${CONFIG_FILE}..."
    sudo install -m 755 -o root -g root config.yaml "${CONFIG_FILE}" || {
        echo "❌ Failed to copy config.yaml to ${CONFIG_FILE}. Exiting."
    exit 1
    }
else
  echo "🛠 Config already exists at ${CONFIG_FILE} — leaving it untouched."
fi

echo "🖇 Installing systemd unit files..."

sudo install -m 755 -o root -g root "systemd/${APP_NAME}.service" "${SERVICE_FILE}" || {
  echo "❌ Failed to copy service file to ${SERVICE_FILE}. Exiting."
  exit 1
}

sudo install -m 755 -o root -g root "systemd/${APP_NAME}.timer" "${TIMER_FILE}" || {
  echo "❌ Failed to copy timer file to ${TIMER_FILE}. Exiting."
  exit 1
}

echo "🔄 Reloading systemd daemon..."
systemctl daemon-reload

echo "✅ Enabling and starting ${APP_NAME}.timer..."
systemctl enable --now "${APP_NAME}.timer"

echo "🎉 Install/upgrade complete!"
