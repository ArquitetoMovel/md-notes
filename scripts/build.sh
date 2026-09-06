#!/usr/bin/env bash
# ==============================================================================
# Cross-platform build script for md-notes (mdn)
# Platforms: macOS (Darwin), Linux, Windows (amd64, arm64)
# ==============================================================================

set -euo pipefail

# ANSI color codes
BOLD="\033[1m"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[0;33m"
RED="\033[0;31m"
RESET="\033[0m"

# Project metadata
BINARY_NAME="mdn"
PACKAGE_PATH="./cmd/mdn"
BIN_DIR="bin"
DIST_DIR="dist"

# Default build parameters
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}"
BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

TARGET_OS="all"
TARGET_ARCH="all"
DO_PACKAGE=false
DO_CLEAN=false

print_usage() {
    echo -e "${BOLD}md-notes Cross-Platform Build Script${RESET}"
    echo ""
    echo "Usage: $(basename "$0") [options]"
    echo ""
    echo "Options:"
    echo "  -v, --version <ver>    Override version string (default: git describe / dev)"
    echo "  -o, --os <os>          Target OS: darwin, linux, windows, or all (default: all)"
    echo "  -a, --arch <arch>      Target Arch: amd64, arm64, or all (default: all)"
    echo "  -p, --package          Package binaries into tar.gz/zip with checksums"
    echo "  -c, --clean            Clean output directories before building"
    echo "  -h, --help             Display this help message"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0")                     # Build all 3 platforms (amd64 & arm64)"
    echo "  $(basename "$0") -o linux -a amd64   # Build only Linux amd64"
    echo "  $(basename "$0") -v 1.0.0 --package  # Build v1.0.0 and package release archives"
}

# Parse command line options
while [[ $# -gt 0 ]]; do
    case "$1" in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--os)
            TARGET_OS="$2"
            shift 2
            ;;
        -a|--arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        -p|--package)
            DO_PACKAGE=true
            shift
            ;;
        -c|--clean)
            DO_CLEAN=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown argument: $1${RESET}" >&2
            print_usage
            exit 1
            ;;
    esac
done

if [ "$DO_CLEAN" = true ]; then
    echo -e "${YELLOW}==> Cleaning ${BIN_DIR} and ${DIST_DIR}...${RESET}"
    rm -rf "${BIN_DIR}" "${DIST_DIR}"
fi

ALL_PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

SELECTED_PLATFORMS=()
for plat in "${ALL_PLATFORMS[@]}"; do
    os="${plat%/*}"
    arch="${plat#*/}"

    if [ "$TARGET_OS" != "all" ] && [ "$TARGET_OS" != "$os" ]; then
        continue
    fi
    if [ "$TARGET_ARCH" != "all" ] && [ "$TARGET_ARCH" != "$arch" ]; then
        continue
    fi
    SELECTED_PLATFORMS+=("$plat")
done

if [ ${#SELECTED_PLATFORMS[@]} -eq 0 ]; then
    echo -e "${RED}Error: No platform matched criteria (os=${TARGET_OS}, arch=${TARGET_ARCH})${RESET}" >&2
    exit 1
fi

echo -e "${BOLD}${BLUE}======================================================${RESET}"
echo -e "${BOLD}${BLUE} Building md-notes (${BINARY_NAME})${RESET}"
echo -e "${BOLD}${BLUE} Version:    ${RESET}${GREEN}${VERSION}${RESET}"
echo -e "${BOLD}${BLUE} Git Commit: ${RESET}${GREEN}${GIT_COMMIT}${RESET}"
echo -e "${BOLD}${BLUE} Build Date: ${RESET}${GREEN}${BUILD_DATE}${RESET}"
echo -e "${BOLD}${BLUE} Targets:    ${RESET}${GREEN}${#SELECTED_PLATFORMS[@]} target(s)${RESET}"
echo -e "${BOLD}${BLUE}======================================================${RESET}"

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}"

for plat in "${SELECTED_PLATFORMS[@]}"; do
    os="${plat%/*}"
    arch="${plat#*/}"
    out_dir="${BIN_DIR}/${os}_${arch}"
    ext=""
    if [ "$os" = "windows" ]; then
        ext=".exe"
    fi
    target_file="${out_dir}/${BINARY_NAME}${ext}"

    echo -e "${BLUE}==> Building for ${BOLD}${os}/${arch}${RESET} -> ${target_file}"
    mkdir -p "${out_dir}"
    CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" go build \
        -trimpath \
        -ldflags="${LDFLAGS}" \
        -o "${target_file}" \
        "${PACKAGE_PATH}"
done

DARWIN_AMD64="${BIN_DIR}/darwin_amd64/${BINARY_NAME}"
DARWIN_ARM64="${BIN_DIR}/darwin_arm64/${BINARY_NAME}"
DARWIN_UNIVERSAL_DIR="${BIN_DIR}/darwin_universal"
DARWIN_UNIVERSAL="${DARWIN_UNIVERSAL_DIR}/${BINARY_NAME}"

if [ -f "$DARWIN_AMD64" ] && [ -f "$DARWIN_ARM64" ] && command -v lipo >/dev/null 2>&1; then
    echo -e "${BLUE}==> Creating macOS Universal Binary (lipo)...${RESET}"
    mkdir -p "$DARWIN_UNIVERSAL_DIR"
    lipo -create -output "$DARWIN_UNIVERSAL" "$DARWIN_AMD64" "$DARWIN_ARM64"
    echo -e "${GREEN}    Created: ${DARWIN_UNIVERSAL}${RESET}"
fi

echo -e "\n${BOLD}${GREEN}Build completed successfully!${RESET}"

if [ "$DO_PACKAGE" = true ]; then
    echo -e "\n${BOLD}${BLUE}======================================================${RESET}"
    echo -e "${BOLD}${BLUE} Packaging Release Archives${RESET}"
    echo -e "${BOLD}${BLUE}======================================================${RESET}"

    mkdir -p "${DIST_DIR}"

    for plat in "${SELECTED_PLATFORMS[@]}"; do
        os="${plat%/*}"
        arch="${plat#*/}"
        archive_name="${BINARY_NAME}-${VERSION}-${os}-${arch}"

        if [ "$os" = "windows" ]; then
            echo -e "${YELLOW}==> Packaging ${archive_name}.zip...${RESET}"
            (cd "${BIN_DIR}/${os}_${arch}" && zip -q -9 "../../${DIST_DIR}/${archive_name}.zip" "${BINARY_NAME}.exe")
        else
            echo -e "${YELLOW}==> Packaging ${archive_name}.tar.gz...${RESET}"
            tar -czf "${DIST_DIR}/${archive_name}.tar.gz" -C "${BIN_DIR}/${os}_${arch}" "${BINARY_NAME}"
        fi
    done

    if [ -f "$DARWIN_UNIVERSAL" ]; then
        univ_archive="${BINARY_NAME}-${VERSION}-darwin-universal.tar.gz"
        echo -e "${YELLOW}==> Packaging ${univ_archive}...${RESET}"
        tar -czf "${DIST_DIR}/${univ_archive}" -C "${DARWIN_UNIVERSAL_DIR}" "${BINARY_NAME}"
    fi

    echo -e "${BLUE}==> Generating SHA-256 checksums...${RESET}"
    (cd "${DIST_DIR}" && {
        if command -v sha256sum >/dev/null 2>&1; then
            sha256sum *.tar.gz *.zip > checksums.txt
        else
            shasum -a 256 *.tar.gz *.zip > checksums.txt
        fi
    })

    echo -e "\n${BOLD}${GREEN}Release archives generated in ${DIST_DIR}/:${RESET}"
    ls -lh "${DIST_DIR}"
fi
