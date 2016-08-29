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
    local dir=$testdir/repositories/$repo/$epoch/$database/x86_64
    local packages=$testdir/packages/$repo/$epoch/$database/x86_64

    mkdir -p $dir
    mkdir -p $packages
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

:list-packages() {
    local run_method="$1"
    local path="$2"
    local system="${3:-archlinux}"
    shift 2

    if [[ $run_method == "local" ]]; then
        :run-local --list $path
    else
        :curl $api_url/list?path=$path\&system=$system
    fi
}

:add-package() {
    local run_method="$1"
    local path="$2"
    shift 2

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/packages/$path

    for package in $packages; do
        tests:ensure cp $testdir/PKGBUILD $dir/

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        pkgfile="$dir/$package-1-1-x86_64.pkg.tar.xz"

        if [[ $run_method == "local" ]]; then
            :run-local --add $path --file="$pkgfile"
        else
            :curl -F package_file=@$pkgfile -XPOST \
                $api_url/add?path=$path\&system=archlinux
        fi
    done
}

# TODO fix this
:add-package-fail() {
    local run_method="$1"
    local path="$2"
    shift 2

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/packages/$path

    for package in $packages; do
        tests:ensure cp $testdir/PKGBUILD $dir/

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        if [[ $run_method == "local" ]]; then
            :run-local --add unknown_repo \
                --file=$dir/$package-1-1-x86_64.pkg.tar.xz
        else
            :curl -F \
                package_file=@$dir/$package-1-1-x86_64.pkg.tar.xz \
                -XPOST $api_url/add?path=unknown_repo
        fi
    done
}

:stat-package() {
    local path="$1"
    local package="$2"
    shift 2

    local repodir="$(tests:get-tmp-dir)/repositories"

    stat $repodir/$path/$package*.tar.xz
}

:remove-package() {
    local run_method="$1"
    local path="$2"
    local package="$3"
    shift 3

    if [[ $run_method == "local" ]]; then
        :run-local --remove $path $package
    else
        :curl -XDELETE $api_url/package/$package?path=$path\&system=archlinux
    fi
}

:describe-package() {
    local run_method="$1"
    local path="$2"
    local package="$3"
    shift 3

    if [[ $run_method == "local" ]]; then
        :run-local --show $path $package
    else
        :curl $api_url/package/$package?path=$path\&system=archlinux
    fi
}

:edit-package-description() {
    local run_method="$1"
    local path="$2"
    local package="$3"
    local description="$4"
    shift 4

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/repositories/$path

    cp $testdir/PKGBUILD $dir/

    PKGDESC=$description PKGDEST=$dir PKGNAME=$package \
        makepkg -p $testdir/PKGBUILD --clean --force

    if [[ $run_method == "local" ]]; then
        :run-local --edit $path $package \
            --file $dir/$package-1-1-x86_64.pkg.tar.xz
    else
        :curl -F \
            package_file=@$dir/$package-1-1-x86_64.pkg.tar.xz -XPATCH \
            $api_url/package/$package?path=$path\&system=archlinux
    fi
}

:copy-package-to-new-root() {
    local run_method="$1"
    local path="$2"
    local package="$3"
    local new_root="$4"
    shift 4

    if [[ $run_method == "local" ]]; then
        :run-local --edit $path $package \
            --copy-to $new_root
    else
        :curl -d "copy-to=$new_root" -XPATCH \
            $api_url/package/$package?path=$path\&system=archlinux

    fi
}
