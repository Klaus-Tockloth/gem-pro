#!/bin/sh

# ------------------------------------
# Purpose:
# - Build binaries for supported target systems.
#
# Releases:
# - v1.0.0 - 2025-03-11: initial release
# - v1.1.0 - 2025-05-19: gosec added
# - v1.2.0 - 2025-11-21: errexit added
# - v1.3.0 - 2025-12-10: revised
# - v1.4.0 - 2026-02-03: revised
# - v1.5.0 - 2026-02-19: Windows ARM-32 support removed, ARM-64 added
# - v1.6.0 - 2026-05-20: FreeBSD, OpenBSD and NetBSD support removed
# ------------------------------------

set -o errexit
set -v -o verbose

# recreate directory
rm -r ./binaries
mkdir ./binaries

# renew vendor content
go mod tidy
go mod vendor

# lint
golangci-lint run --no-config --enable gocritic
revive

# security
govulncheck ./...
gosec -exclude=G114,G115,G204,G302,G304 ./...

# show compiler version
go version

# compile 'darwin' (macOS)
env GOOS=darwin GOARCH=arm64 go build -v -o binaries/darwin-arm64/gem-pro
env GOOS=darwin GOARCH=amd64 go build -v -o binaries/darwin-amd64/gem-pro

# compile 'linux'
env GOOS=linux GOARCH=amd64 go build -v -o binaries/linux-amd64/gem-pro
env GOOS=linux GOARCH=arm64 go build -v -o binaries/linux-arm64/gem-pro

# compile 'windows'
env GOOS=windows GOARCH=amd64 go build -v -o binaries/windows-amd64/gem-pro.exe
env GOOS=windows GOARCH=arm64 go build -v -o binaries/windows-arm64/gem-pro.exe

