#!/usr/bin/env bash
# DMXAPI 配置工具一键安装脚本 (Linux / macOS)
# 用法: curl -fsSL https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.sh | bash

set -euo pipefail

REPO="dmxapi/opencode_dmxapi"
RELEASES_API="https://cnb.cool/${REPO}/-/releases"
BIN_PREFIX="opencode-dmxapi"

tmp_dir=""
trap '[ -n "$tmp_dir" ] && rm -rf "$tmp_dir"' EXIT

USE_COLOR=1
if [ -n "${NO_COLOR-}" ] || [ "${TERM-}" = "dumb" ] || [ ! -t 1 ]; then
  USE_COLOR=0
fi

color() {
  if [ "$USE_COLOR" = "1" ]; then
    printf '\033[%sm%s\033[0m\n' "$1" "$2"
  else
    printf '%s\n' "$2"
  fi
}
info()  { color "1;36" "==> $*"; }
warn()  { color "1;33" "!! $*"; }
err()   { color "1;31" "✖ $*" >&2; }

# --- 检测 OS / ARCH ---
detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  case "$os" in
    Linux)  os="linux" ;;
    Darwin) os="macos" ;;
    *) err "不支持的操作系统: $os"; exit 1 ;;
  esac
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) err "不支持的架构: $arch"; exit 1 ;;
  esac
  echo "${os}-${arch}"
}

# --- 获取最新 tag（过滤 prerelease，按 semver 取最大） ---
fetch_latest_tag() {
  local json tags stable tag
  if ! json="$(curl -fsSL -H 'Accept: application/json' "$RELEASES_API")"; then
    err "无法访问 $RELEASES_API"
    exit 1
  fi
  tags="$(printf '%s' "$json" | grep -oE '"tag_name"[[:space:]]*:[[:space:]]*"[^"]+"' | sed -E 's/.*"([^"]+)"$/\1/' || true)"
  if [ -z "$tags" ]; then
    err "无法解析最新版本号"
    exit 1
  fi
  stable="$(printf '%s\n' "$tags" | grep -v -- '-' || true)"
  if [ -n "$stable" ]; then
    tag="$(printf '%s\n' "$stable" | sort -V | tail -1)"
  else
    tag="$(printf '%s\n' "$tags" | sort -V | tail -1)"
    warn "未找到稳定版本，使用最新预发布版本: $tag"
  fi
  echo "$tag"
}

main() {
  command -v curl >/dev/null 2>&1 || { err "需要 curl"; exit 1; }

  info "检测平台..."
  local platform tag asset url bin_path
  platform="$(detect_platform)"
  info "平台: $platform"

  info "获取最新版本..."
  tag="$(fetch_latest_tag)"
  info "版本: $tag"

  asset="${BIN_PREFIX}-${tag}-${platform}"
  url="https://cnb.cool/${REPO}/-/releases/download/${tag}/${asset}"

  tmp_dir="$(mktemp -d -t dmxapi-XXXXXX)"
  bin_path="${tmp_dir}/${asset}"

  info "下载 $asset ..."
  if ! curl -fL --progress-bar -o "$bin_path" "$url"; then
    err "下载失败: $url"
    err "请确认 release 资产已发布，或访问 $RELEASES_API 手动下载"
    exit 1
  fi

  # 可选 SHA256 校验：老 release 可能不带 .sha256 文件，下载失败则跳过
  local sha_url="${url}.sha256"
  local sha_path="${bin_path}.sha256"
  if curl -fsSL -o "$sha_path" "$sha_url" 2>/dev/null; then
    local expected actual
    expected="$(awk '{print $1}' "$sha_path")"
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$bin_path" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
      actual="$(shasum -a 256 "$bin_path" | awk '{print $1}')"
    else
      warn "未找到 sha256sum/shasum，跳过校验"
      actual="$expected"
    fi
    if [ "$expected" != "$actual" ]; then
      err "SHA256 校验失败！期望 $expected，实际 $actual"
      exit 1
    fi
    info "SHA256 校验通过"
  else
    warn "未找到 SHA256 校验文件，跳过完整性校验（向后兼容旧 release）"
  fi

  chmod +x "$bin_path"
  if [ "$(uname -s)" = "Darwin" ]; then
    xattr -dr com.apple.quarantine "$bin_path" 2>/dev/null || true
  fi

  info "启动配置..."
  echo
  if [ -r /dev/tty ]; then
    "$bin_path" < /dev/tty
  else
    "$bin_path"
  fi
}

main "$@"
