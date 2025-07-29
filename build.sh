#!/bin/bash

# 项目名称
PROJECT_NAME="gossh"

# 检查是否传入版本号参数
if [ -z "$1" ]; then
  echo "Error: Please pass the version number as the first argument."
  echo "Usage: $0 v1.0.0"
  exit 1
fi

VERSION=$1  # 从参数获取版本号

# 编译输出目录
OUTPUT_DIR="dist"

# 支持的平台及架构
PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64" "darwin/amd64")

# 清理旧构建文件夹
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

for PLATFORM in "${PLATFORMS[@]}"; do
  OS=$(echo $PLATFORM | cut -d'/' -f1)
  ARCH=$(echo $PLATFORM | cut -d'/' -f2)

  # 简洁的二进制文件名，不含版本号
  BINARY_NAME="${PROJECT_NAME}_${OS}_${ARCH}"
  if [ "$OS" == "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
  fi

  echo "Building for $OS/$ARCH ..."
  GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 go build -o "$OUTPUT_DIR/$BINARY_NAME"

  if [ $? -ne 0 ]; then
    echo "Build failed for $OS/$ARCH"
    exit 1
  fi

  # 打包压缩，压缩包文件名中体现版本号
  cd $OUTPUT_DIR
  PACKAGE_NAME="${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}"
  if [ "$OS" == "windows" ]; then
    zip -r "${PACKAGE_NAME}.zip" "$BINARY_NAME"
    rm "$BINARY_NAME"
  else
    tar czf "${PACKAGE_NAME}.tar.gz" "$BINARY_NAME"
    rm "$BINARY_NAME"
  fi
  cd -
done

echo "Build and packaging completed. Packages are in the $OUTPUT_DIR folder."

