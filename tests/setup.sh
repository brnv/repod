:curl() {
    /bin/curl -s -u user:valid ${@}
}

tests:clone repod.test bin/
tests:clone tests/mocks/gpg bin/gpg
tests:clone tests/utils/PKGBUILD PKGBUILD
tests:clone tests/mocks/nucleus-server .

_repod="127.0.0.1:64444"
_nucleus="127.0.0.1:64777"

api_url="$_repod/v1"

:nucleus() {
    tests:make-tmp-dir nucleus

    local pid=""
    tests:value pid blankd \
        -d $(tests:get-tmp-dir)/nucleus/ \
        -e $(tests:get-tmp-dir)/nucleus-server \
        -l "$_nucleus" \
        --tls

    tests:put-string blankd.pid "$pid"
}

:run-daemon() {
    :nucleus

    tests:eval go-test:run \
        repod.test \
        --listen="$_repod" \
        --root=$(tests:get-tmp-dir)/repositories/ \
        --nucleus "$_nucleus" \
        --tls-cert $(tests:get-tmp-dir)/nucleus/tls.crt
}

:run-local() {
    tests:eval go-test:run \
        repod.test \
        --root=$(tests:get-tmp-dir)/repositories/ "${@}"
}

:bootstrap-repository() {
    local repo="$1"
    local epoch="${2:-}"
    local database="${3:-}"
    local architecture="${4:-}"

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    mkdir -p $dir
}

:list-repositories() {
    local run_method="$1"
    shift 1

    if [[ $run_method == "local" ]]; then
        :run-local --list
    else
        :curl $api_url/
    fi
}

:list-epoches() {
    local run_method="$1"
    local repo="$2"
    shift 2

    if [[ $run_method == "local" ]]; then
        :run-local --list $repo
    else
        :curl $api_url/$repo
    fi
}

:list-packages() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    shift 5

    if [[ $run_method == "local" ]]; then
        :run-local --list $repo $epoch $database $architecture
    else
        :curl $api_url/$repo/$epoch/$database/$architecture
    fi
}

:add-package() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    shift 5

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    for package in $packages; do
        tests:ensure cp $testdir/PKGBUILD $dir/

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        if [[ $run_method == "local" ]]; then
            :run-local --add $repo $epoch $database $architecture $package \
                --file=$dir/$package-1-1-$architecture.pkg.tar.xz
        else
            :curl -F \
                package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz \
                -XPOST $api_url/$repo/$epoch/$database/$architecture/$package
        fi
    done
}

# TODO fix this
:add-package-fail() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    shift 5

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    for package in $packages; do
        cp $testdir/PKGBUILD $dir/

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        if [[ $run_method == "local" ]]; then
            :run-local \
                --add unknown_repo $epoch $database $architecture $package \
                --file=$dir/$package-1-1-$architecture.pkg.tar.xz
        else
            :curl -F \
                package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz \
                -XPOST \
                $api_url/unknown_repo/$epoch/$database/$architecture/$package
        fi
    done
}

:stat-package() {
    local repo="$1"
    local epoch="$2"
    local database="$3"
    local architecture="$4"
    local package="$5"
    local repodir="$(tests:get-tmp-dir)/repositories"

    stat $repodir/$repo/$epoch/$database/$architecture/$package
}

:remove-package() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    local package="$6"
    shift 6

    if [[ $run_method == "local" ]]; then
        :run-local --remove $repo $epoch $database $architecture $package
    else
        :curl -XDELETE $api_url/$repo/$epoch/$database/$architecture/$package
    fi
}

:describe-package() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    local package="$6"
    shift 6

    if [[ $run_method == "local" ]]; then
        :run-local --show $repo $epoch $database $architecture $package
    else
        :curl $api_url/$repo/$epoch/$database/$architecture/$package
    fi
}

:edit-package-description() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    local package="$6"
    local description="$7"
    shift 7

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$repo/$epoch/$database/$architecture

    cp $testdir/PKGBUILD $dir/

    PKGDESC=$description PKGDEST=$dir PKGNAME=$package \
        makepkg -p $testdir/PKGBUILD --clean --force

    if [[ $run_method == "local" ]]; then
        :run-local --edit $repo $epoch $database $architecture $package \
            --file $dir/$package-1-1-$architecture.pkg.tar.xz
    else
        :curl -F \
            package_file=@$dir/$package-1-1-$architecture.pkg.tar.xz -XPATCH \
            $api_url/$repo/$epoch/$database/$architecture/$package
    fi
}

:copy-package-to-epoch() {
    local run_method="$1"
    local repo="$2"
    local epoch="$3"
    local database="$4"
    local architecture="$5"
    local package="$6"
    local new_epoch="$7"
    shift 7

    if [[ $run_method == "local" ]]; then
        :run-local --edit $repo $epoch $database $architecture $package \
            --change-epoch $new_epoch
    else
        :curl -d "epoch_new=$new_epoch" -XPATCH \
            $api_url/$repo/$epoch/$database/$architecture/$package
    fi
}
