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

:bootstrap-packages-arch() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    shift 4

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    for package in $packages; do
        dir=$testdir/repositories/$repo/$epoch/$database/$architecture
        mkdir -p $dir
        touch $dir/$package
    done
}

curl() {
    /bin/curl -s ${@}
}

:curl-repositories-list() {
    curl http://localhost:6333/v1/
}

:curl-epoches-list() {
    local repo="$1"
    curl http://localhost:6333/v1/$repo/
}

:curl-list-packages() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"

    curl http://localhost:6333/v1/$repo/$epoch/$database/$architecture
}

:curl-add-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"
    local package_file="$6"

    curl -F package_file=@$package_file -XPUT \
        http://localhost:6333/v1/$repo/$epoch/$database/$architecture/$package
}
