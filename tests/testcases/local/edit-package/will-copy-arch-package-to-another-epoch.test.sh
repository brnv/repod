:bootstrap-repository arch-repo testing arch-repo x86_64
:add-package local arch-repo/testing/arch-repo/x86_64 package_one
:list-packages local arch-repo/testing/arch-repo/x86_64
tests:assert-stdout "arch-repo-testing-arch-repo-x86_64 package_one 1-1"

:bootstrap-repository arch-repo stable arch-repo x86_64

:copy-package-to-new-root \
    local arch-repo/testing/arch-repo/x86_64 package_one arch-repo/stable/arch-repo/x86_64

:list-packages local arch-repo/stable/arch-repo/x86_64
tests:assert-stdout "arch-repo-stable-arch-repo-x86_64 package_one 1-1"
