#!/bin/bash

set -euo pipefail

:run() {
    PATH=$PATH:$(tests:get-tmp-dir)/bin repod \
        --listen-address=":6333" \
        --repositories-dir=$(tests:get-tmp-dir)/repositories/
}

:bootstrap-repository() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture
    mkdir -p $dir
}

curl() {
    /bin/curl -s ${@}
}

:list-repositories() {
    curl http://localhost:6333/v1/
}

:list-epoches() {
    local repo="$1"
    curl http://localhost:6333/v1/$repo/
}

:list-packages() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"

    curl http://localhost:6333/v1/$repo/$epoch/$database/$architecture
}

:add-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    shift 4

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    for package in $packages; do
        cp $testdir/PKGBUILD $dir/
        PKGDEST=$dir PKGNAME=$package makepkg -p $testdir/PKGBUILD -c -f

        curl -F \
            package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz -XPUT \
            http://localhost:6333/v1/$repo/$epoch/$database/$architecture/$package
    done
}

:remove-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"

    curl -XDELETE \
        http://localhost:6333/v1/$repo/$epoch/$database/$architecture/$package
}
