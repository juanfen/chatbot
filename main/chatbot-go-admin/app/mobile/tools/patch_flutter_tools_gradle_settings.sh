#!/usr/bin/env bash
# 将本目录下的 flutter_tools_gradle_settings.chatbot.gradle.kts 覆盖到当前 Flutter SDK
# （packages/flutter_tools/gradle/settings.gradle.kts），避免 :gradle:compileKotlin 去连 repo.maven.apache.org。
# 用法：在 mobile 目录执行: bash tools/patch_flutter_tools_gradle_settings.sh

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MOBILE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROP="$MOBILE_ROOT/android/local.properties"
SRC="$SCRIPT_DIR/flutter_tools_gradle_settings.chatbot.gradle.kts"

if [[ ! -f "$PROP" ]]; then
  echo "未找到 android/local.properties，请先在 mobile 执行 fvm flutter pub get。" >&2
  exit 1
fi
if [[ ! -f "$SRC" ]]; then
  echo "未找到补丁源文件: $SRC" >&2
  exit 1
fi

RAW="$(grep '^flutter.sdk=' "$PROP" | cut -d= -f2- | tr -d '\r')"
SDK="${RAW//\\//}"
TARGET="$SDK/packages/flutter_tools/gradle/settings.gradle.kts"

if [[ ! -d "$(dirname "$TARGET")" ]]; then
  echo "Flutter SDK 路径异常: $TARGET" >&2
  exit 1
fi

cp "$SRC" "$TARGET"
echo "已覆盖: $TARGET"
