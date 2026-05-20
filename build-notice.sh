#!/bin/sh

# ------------------------------------
# Purpose:
# - Build notice file for all FOSS modules.
#
# Releases:
# - v1.0.0 - 2025-11-13: initial release
# - v2.0.0 - 2026-05-20: 'lichen' instead of 'go-licenses'
#
# Meta Aspects:
# - 'lichen' analyzes the respective platform-specific binary.
# - Consequently, plattform-specific NOTICE files are generated.
# - For 'lichen' to work, a binary must be generated first.
# - If a component changes (new, deleted, version change):
#   1. Build applications
#   2. Generate NOTICE files
#   3. Build applications again
#
# Remarks:
# - Requirements: lichen (go install github.com/uw-labs/lichen@latest)
# ------------------------------------

# set -o verbose
set -o errexit

# 1. Define 'lichen' template internally
LICHEN_TEMPLATE=$(cat << 'EOF'
Third-Party Licenses:
=====================

{{ range .Modules }}
------------------------------------------------------------
- Name    : {{ .Module.Path }}
- Version : {{ .Module.Version }}
- License : {{ range $i, $lic := .Module.Licenses }}{{ if $i }}, {{ end }}{{ $lic.Name }}{{ end }}
------------------------------------------------------------

{{ range $lic := .Module.Licenses }}
{{ $lic.Content }}
{{ end }}
{{ end }}
EOF
)

# 2. Build licenses
echo "Building licenses..."

# Darwin (macOS)
lichen --template "$LICHEN_TEMPLATE" ./binaries/darwin-amd64/gem-pro > NOTICE_DARWIN-AMD64.txt
lichen --template "$LICHEN_TEMPLATE" ./binaries/darwin-arm64/gem-pro > NOTICE_DARWIN-ARM64.txt

# Linux
lichen --template "$LICHEN_TEMPLATE" ./binaries/linux-amd64/gem-pro > NOTICE_LINUX-AMD64.txt
lichen --template "$LICHEN_TEMPLATE" ./binaries/linux-arm64/gem-pro > NOTICE_LINUX-ARM64.txt

# Windows
lichen --template "$LICHEN_TEMPLATE" ./binaries/windows-amd64/gem-pro.exe > NOTICE_WINDOWS-AMD64.txt
lichen --template "$LICHEN_TEMPLATE" ./binaries/windows-arm64/gem-pro.exe > NOTICE_WINDOWS-ARM64.txt

echo "All NOTICE files successfully created."
