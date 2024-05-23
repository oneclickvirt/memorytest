#!/bin/bash
#From https://github.com/oneclickvirt/memoryTest
#2024.05.23

rm -rf /usr/bin/memoryTest
os=$(uname -s)
arch=$(uname -m)

case $os in
  Linux)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-linux-amd64
        ;;
      "i386" | "i686")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-linux-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-linux-arm64
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
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-darwin-amd64
        ;;
      "i386" | "i686")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-darwin-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-darwin-arm64
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
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-freebsd-amd64
        ;;
      "i386" | "i686")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-freebsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-freebsd-arm64
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
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-openbsd-amd64
        ;;
      "i386" | "i686")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-openbsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O memoryTest https://github.com/oneclickvirt/memoryTest/releases/download/output/memoryTest-openbsd-arm64
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

chmod 777 memoryTest
if [ ! -f /usr/bin/memoryTest ]; then
  mv memoryTest /usr/bin/
  memoryTest
else
  ./memoryTest
fi
