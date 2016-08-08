#!/bin/bash

api_url="http://localhost:6333/v1"

:curl() {
    /bin/curl -s ${@}
}

tests:clone tests/repod bin/repod
tests:clone tests/mocks/gpg bin/gpg
tests:clone tests/utils/PKGBUILD PKGBUILD

:run() {
    PATH=$PATH:$(tests:get-tmp-dir)/bin repod \
        --listen=":6333" \
        --repos-dir=$(tests:get-tmp-dir)/repositories/
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

:list-repositories() {
    :curl $api_url/
}

:list-epoches() {
    local repo="$1"

    :curl $api_url/$repo
}

:list-packages() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"

    :curl $api_url/$repo/$epoch/$database/$architecture
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

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        :curl -F \
            package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz -XPOST \
            $api_url/$repo/$epoch/$database/$architecture/$package
    done
}

:remove-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"

    :curl -XDELETE \
        $api_url/$repo/$epoch/$database/$architecture/$package
}

:describe-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"

    :curl $api_url/$repo/$epoch/$database/$architecture/$package
}

:edit-package-description() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"
    local description="$6"

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    cp $testdir/PKGBUILD $dir/

    PKGDESC=$description PKGDEST=$dir PKGNAME=$package \
        makepkg -p $testdir/PKGBUILD --clean --force

    :curl -F \
        package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz -XPATCH \
        $api_url/$repo/$epoch/$database/$architecture/$package
}

:copy-package-to-epoch() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"
    local new_epoch="$6"

    :curl -d "epoch_new=$new_epoch" -XPATCH \
        $api_url/$repo/$epoch/$database/$architecture/$package
}
