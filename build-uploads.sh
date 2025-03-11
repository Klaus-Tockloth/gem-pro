#!/bin/sh

# ------------------------------------
# Purpose:
# - Builds uploads (tar.gz or zip) for Github project repository (assets in release section).
#
# Releases:
# - v1.0.0 - 2025/03/11: initial release
# ------------------------------------

# set -o xtrace
set -o verbose

# recreate directory
rm -r ./uploads
mkdir ./uploads

# uploads 'darwin'
tar -cvzf ./uploads/macos-amd64_gem-pro.tar.gz ./binaries/darwin-amd64/gem-pro
tar -cvzf ./uploads/macos-arm64_gem-pro.tar.gz ./binaries/darwin-arm64/gem-pro

# uploads 'freebsd'
tar -cvzf ./uploads/freebsd-amd64_gem-pro.tar.gz ./binaries/freebsd-amd64/gem-pro
tar -cvzf ./uploads/freebsd-arm64_gem-pro.tar.gz ./binaries/freebsd-arm64/gem-pro

# uploads 'linux'
tar -cvzf ./uploads/linux-amd64_gem-pro.tar.gz ./binaries/linux-amd64/gem-pro
tar -cvzf ./uploads/linux-arm64_gem-pro.tar.gz ./binaries/linux-arm64/gem-pro

# uploads 'netbsd'
tar -cvzf ./uploads/netbsd-amd64_gem-pro.tar.gz ./binaries/netbsd-amd64/gem-pro
tar -cvzf ./uploads/netbsd-arm64_gem-pro.tar.gz ./binaries/netbsd-arm64/gem-pro

# uploads 'openbsd'
tar -cvzf ./uploads/openbsd-amd64_gem-pro.tar.gz ./binaries/openbsd-amd64/gem-pro
tar -cvzf ./uploads/openbsd-arm64_gem-pro.tar.gz ./binaries/openbsd-arm64/gem-pro

# uploads 'windows'
zip ./uploads/windows-amd64_gem-pro.zip ./binaries/windows-amd64/gem-pro.exe
zip ./uploads/windows-arm_gem-pro.zip ./binaries/windows-arm/gem-pro.exe
