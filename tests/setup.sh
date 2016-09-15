:curl() {
    tests:eval /bin/curl -s -u user:valid ${@}
}

tests:clone repod.test bin/
tests:clone tests/mocks/gpg bin/gpg
tests:clone tests/utils/PKGBUILD PKGBUILD
tests:clone tests/mocks/nucleus-server .

_repod="127.0.0.1:64444"
_nucleus="127.0.0.1:64777"

export package_ver=$RANDOM
export package_rel=$RANDOM

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

if [[ $mode == "daemon" ]]; then
    :nucleus
fi

:run-daemon() {
    tests:eval go-test:run \
        repod.test \
        --listen="$_repod" \
        --root=$(tests:get-tmp-dir)/repositories/ \
        --nucleus "$_nucleus" \
        --tls-cert $(tests:get-tmp-dir)/nucleus/tls.crt \
        --debug
}

:run-local() {
    tests:eval go-test:run \
        repod.test \
        --root=$(tests:get-tmp-dir)/repositories/ "${@}" \
        --debug
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

    if [[ $mode == "daemon" ]]; then
        tests:run-background bg_repod :run-daemon
        tests:ensure tests:wait-file-matches \
            $(tests:get-background-stderr $bg_repod) "serving" 1 2
    fi
}

:list-repositories() {
    if [[ $mode == "cli" ]]; then
        :run-local --query
    else
        :curl $api_url/
    fi
}

:list-packages() {
    local path="$1"
    local system="${2:-archlinux}"

    if [[ $mode == "cli" ]]; then
        :run-local --query $path --system $system
    else
        :curl $api_url/list?path=$path\&system=$system
    fi
}

:add-package() {
    local path="$1"
    shift 1

    local packages=${@}

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/packages/$path

    for package in $packages; do
        tests:ensure cp $testdir/PKGBUILD $dir/

        PKGDEST=$dir PKGNAME=$package \
            makepkg -p $testdir/PKGBUILD --clean --force

        pkgfile="$dir/$package-$package_ver-$package_rel-x86_64.pkg.tar.xz"

        if [[ $mode == "cli" ]]; then
            :run-local --add $path --file="$pkgfile"
        else
            :curl -F package_file=@$pkgfile -XPOST \
                $api_url/add?path=$path\&system=archlinux
        fi
    done
}

:stat-package() {
    local path="$1"
    local package="$2"
    shift 2

    local repodir="$(tests:get-tmp-dir)/repositories"

    stat $repodir/$path/$package-[0-9]*-[0-9]*-*.tar.xz
}

:remove-package() {
    local path="$1"
    local package="$2"
    shift 2

    if [[ $mode == "cli" ]]; then
        :run-local --remove $path $package
    else
        :curl -XDELETE $api_url/package/$package?path=$path\&system=archlinux
    fi
}

:describe-package() {
    local path="$1"
    local package="$2"
    shift 2

    if [[ $mode == "cli" ]]; then
        :run-local --show $path $package
    else
        :curl $api_url/package/$package?path=$path\&system=archlinux
    fi
}

:edit-package-description() {
    local path="$1"
    local package="$2"
    local description="$3"
    shift 3

    local testdir=$(tests:get-tmp-dir)
    local dir=$testdir/packages/$path

    cp $testdir/PKGBUILD $dir/

    PKGDESC=$description PKGDEST=$dir PKGNAME=$package \
        makepkg -p $testdir/PKGBUILD --clean --force

    tree $(pwd)

    if [[ $mode == "cli" ]]; then
        :run-local --add $path --force \
            --file $dir/$package-$package_ver-$package_rel-x86_64.pkg.tar.xz
    else
        :curl -F \
            package_file=@$dir/$package-$package_ver-$package_rel-x86_64.pkg.tar.xz -XPOST \
            $api_url/add?path=$path\&system=archlinux\&force=1
    fi
}

:copy-package-to-new-root() {
    local path="$1"
    local package="$2"
    local new_root="$3"
    shift 3

    if [[ $mode == "cli" ]]; then
        :run-local --copy $path $package \
            --copy-to $new_root
    else
        :curl -XPOST \
            $api_url/package/$package?path=$path\&system=archlinux\&copy-to=$new_root
    fi
}
