#!/bin/sh

# ------------------------------------
# Purpose:
# - Builds uploads (tar.gz or zip) for Github project repository (assets in release section).
#
# Releases:
# - v1.0.0 - 2025-03-11: initial release
# - v1.1.0 - 2026-02-19: Windows ARM-32 support removed, ARM-64 added
# - v1.2.0 - 2026-06-03: FreeBSD, OpenBSD and NetBSD support removed
# ------------------------------------

# set -o xtrace
set -o verbose

# recreate directory
rm -r ./uploads
mkdir ./uploads

# uploads 'darwin'
tar -cvzf ./uploads/macos-amd64_gem-pro.tar.gz ./binaries/darwin-amd64/gem-pro
tar -cvzf ./uploads/macos-arm64_gem-pro.tar.gz ./binaries/darwin-arm64/gem-pro

# uploads 'linux'
tar -cvzf ./uploads/linux-amd64_gem-pro.tar.gz ./binaries/linux-amd64/gem-pro
tar -cvzf ./uploads/linux-arm64_gem-pro.tar.gz ./binaries/linux-arm64/gem-pro

# uploads 'windows'
zip ./uploads/windows-amd64_gem-pro.zip ./binaries/windows-amd64/gem-pro.exe
zip ./uploads/windows-arm64_gem-pro.zip ./binaries/windows-arm64/gem-pro.exe
