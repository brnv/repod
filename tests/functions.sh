#!/bin/bash

set -euo pipefail

:run() {
    PATH=$PATH:$(tests:get-tmp-dir)/bin repod \
        --listen-address=":6333" \
        --repository-dir=$(tests:get-tmp-dir)/repositories/
}

:bootstrap-repositories() {
    for repo in ${@}; do
        mkdir -p $(tests:get-tmp-dir)/repositories/$repo
    done
}

curl() {
    /bin/curl -s $1
}

:curl-repositories-list() {
    curl http://localhost:6333/v1/
}
