#!/bin/bash
#From https://github.com/oneclickvirt/memorytest
#2024.08.05

rm -rf /usr/bin/memorytest
rm -rf memorytest
os=$(uname -s)
arch=$(uname -m)

check_cdn() {
  local o_url=$1
  for cdn_url in "${cdn_urls[@]}"; do
    if curl -sL -k "$cdn_url$o_url" --max-time 6 | grep -q "success" >/dev/null 2>&1; then
      export cdn_success_url="$cdn_url"
      return
    fi
    sleep 0.5
  done
  export cdn_success_url=""
}

check_cdn_file() {
  check_cdn "https://raw.githubusercontent.com/spiritLHLS/ecs/main/back/test"
  if [ -n "$cdn_success_url" ]; then
    echo "CDN available, using CDN"
  else
    echo "No CDN available, no use CDN"
  fi
}

cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
check_cdn_file

download_file() {
    local url="$1"
    local output="$2"
    
    if ! wget -O "$output" "$url"; then
        echo "wget failed, trying curl..."
        if ! curl -L -o "$output" "$url"; then
            echo "Both wget and curl failed. Unable to download the file."
            return 1
        fi
    fi
    return 0
}

get_memorytest_url() {
    local os="$1"
    local arch="$2"

    case $os in
        Linux)
            case $arch in
                "x86_64" | "x86" | "amd64" | "x64") echo "memorytest-linux-amd64" ;;
                "i386" | "i686") echo "memorytest-linux-386" ;;
                "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64") echo "memorytest-linux-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        Darwin)
            case $arch in
                "x86_64" | "x86" | "amd64" | "x64") echo "memorytest-darwin-amd64" ;;
                "i386" | "i686") echo "memorytest-darwin-386" ;;
                "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64") echo "memorytest-darwin-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        FreeBSD)
            case $arch in
                amd64) echo "memorytest-freebsd-amd64" ;;
                "i386" | "i686") echo "memorytest-freebsd-386" ;;
                "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64") echo "memorytest-freebsd-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        OpenBSD)
            case $arch in
                amd64) echo "memorytest-openbsd-amd64" ;;
                "i386" | "i686") echo "memorytest-openbsd-386" ;;
                "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64") echo "memorytest-openbsd-arm64" ;;
                *) return 1 ;;
            esac
            ;;
        *) return 1 ;;
    esac
}

memorytest_filename=$(get_memorytest_url "$os" "$arch")
if [ -z "$memorytest_filename" ]; then
    echo "Unsupported operating system ($os) or architecture ($arch)"
    exit 1
fi
memorytest_url="${cdn_success_url}https://github.com/oneclickvirt/memorytest/releases/download/output/${memorytest_filename}"
if ! download_file "$memorytest_url" "memorytest"; then
    echo "Failed to download memorytest"
    exit 1
fi
chmod 777 memorytest
cp memorytest /usr/bin/memorytest
