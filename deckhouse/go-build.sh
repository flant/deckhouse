#!/bin/sh

set -e

export GO111MODULE=on

shellOpVer=$(go list -m all | grep shell-operator | cut -d' ' -f 2-)
addonOpVer=$(go list -m all | grep addon-operator | cut -d' ' -f 2-)
deckhouseVer=$(git ls-tree @ -- deckhouse | sed -r 's/^[0-9]{6}\s+tree\s+([0-9a-f]{40})\s+.*$/\1/')

CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'main.DeckhouseVersion=$deckhouseVer' -X 'main.AddonOperatorVersion=$addonOpVer' -X 'main.ShellOperatorVersion=$shellOpVer'" -o deckhouse ./cmd/deckhouse