#!/bin/bash
#From https://github.com/oneclickvirt/memorytest
#2024.06.24

rm -rf /usr/bin/memorytest
rm -rf memorytest
os=$(uname -s)
arch=$(uname -m)

case $os in
  Linux)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-linux-amd64
        ;;
      "i386" | "i686")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-linux-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-linux-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-darwin-amd64
        ;;
      "i386" | "i686")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-darwin-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-darwin-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  FreeBSD)
    case $arch in
      amd64)
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-freebsd-amd64
        ;;
      "i386" | "i686")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-freebsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-freebsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  OpenBSD)
    case $arch in
      amd64)
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-openbsd-amd64
        ;;
      "i386" | "i686")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-openbsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memorytest https://github.com/oneclickvirt/memorytest/releases/download/output/memorytest-openbsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $os"
    exit 1
    ;;
esac

chmod 777 memorytest
cp memorytest /usr/bin/memorytest
