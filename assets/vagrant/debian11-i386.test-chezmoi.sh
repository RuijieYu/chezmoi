#!/bin/bash

set -eufo pipefail

GO_VERSION=$(awk '/GO_VERSION:/ { print $2 }' /chezmoi/.github/workflows/main.yml)

go get "golang.org/dl/go${GO_VERSION}"
"${HOME}/go/bin/go${GO_VERSION}" download
export PATH="${HOME}/sdk/go${GO_VERSION}/bin:${PATH}"

cd /chezmoi

go test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
