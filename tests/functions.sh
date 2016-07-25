#!/bin/bash

set -euo pipefail

:run() {
    PATH=$PATH:$(tests:get-tmp-dir)/bin repod \
        --listen-address=":6333" \
        --repositories-dir=$(tests:get-tmp-dir)/repositories/
}

:bootstrap-repositories() {
    local repos=${@}
    for repo in $repos; do
        mkdir -p $(tests:get-tmp-dir)/repositories/$repo
    done
}

:bootstrap-epoches() {
    local repo="$1"
    shift

    local epoches=${@}

    for epoch in $epoches; do
        mkdir -p $(tests:get-tmp-dir)/repositories/$repo/$epoch
    done
}

curl() {
    /bin/curl -s $1
}

:curl-repositories-list() {
    curl http://localhost:6333/v1/
}

:curl-epoches-list() {
    local repo="$1"
    curl http://localhost:6333/v1/$repo/
}
