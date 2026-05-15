#!/bin/sh
set -e

REPO="tb-suzukiTsukasa/backlog-cli"
BINARY="bk"
INSTALL_DIR="/usr/local/bin"

# OS・アーキテクチャを検出
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)
    echo "未対応のOS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "未対応のアーキテクチャ: $ARCH" >&2
    exit 1
    ;;
esac

# 最新バージョンを取得
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  | grep '"tag_name"' | sed 's/.*"tag_name": "\(.*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "バージョンの取得に失敗しました。GITHUB_TOKEN が設定されているか確認してください。" >&2
  exit 1
fi

ARCHIVE="backlog-cli_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "bk ${VERSION} をインストールします (${OS}/${ARCH})..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL -H "Authorization: token ${GITHUB_TOKEN}" "$URL" -o "${TMP}/${ARCHIVE}"
tar xzf "${TMP}/${ARCHIVE}" -C "$TMP"

if [ ! -w "$INSTALL_DIR" ]; then
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"
echo "インストール完了: ${INSTALL_DIR}/${BINARY}"
bk --version
