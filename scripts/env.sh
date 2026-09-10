#!/usr/bin/env bash
# Portable build environment for uniTerm.
#
# On a fresh machine, source this BEFORE building so that Go resolves the exact
# pinned versions (go.sum) via a reachable module proxy, independent of the
# machine's global Go config:
#
#     source scripts/env.sh    # then:
#     wails3 build             #  or:  wails3 dev   /   task build
#
# Everything here has a sensible default and respects a value you already set,
# so it is safe to source unconditionally.
set -a

# Module proxy. Defaults to goproxy.cn, a fast mirror (China-friendly; the
# default proxy.golang.org is often unreachable from CN networks). Override by
# exporting GOPROXY before sourcing.
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

# Toolchain policy. `auto` makes Go use the exact toolchain required by
# go.mod (go 1.26.2) and download it on first use on a new machine, instead of
# silently building with whatever older local go is installed.
export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

# Resolve modules per go.sum. There is deliberately no vendor/ here (Wails
# embeds build-time DLLs that prevent `go mod vendor`), so keep -mod=mod (the
# normal mode) rather than letting Go auto-switch to a partial vendor dir.
export GOFLAGS="${GOFLAGS:--mod=mod}"

set +a

echo "GOPROXY=${GOPROXY}"
echo "GOTOOLCHAIN=${GOTOOLCHAIN}"
echo "GOFLAGS=${GOFLAGS}"
echo "Go module cache: $(command -v go >/dev/null 2>&1 && go env GOMODCACHE || echo '<go not on PATH — install Go 1.26+ first>')"