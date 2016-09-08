:bootstrap-repository arch-repo testing arch-repo x86_64
:bootstrap-repository arch-repo stable arch-repo x86_64

:add-package arch-repo/testing/arch-repo/x86_64 package_one

:list-packages arch-repo/testing/arch-repo/x86_64
    tests:assert-stdout "arch-repo-testing-arch-repo-x86_64 package_one 1-1"

:copy-package-to-new-root \
    arch-repo/testing/arch-repo/x86_64 \
    package_one arch-repo/stable/arch-repo/x86_64

:list-packages arch-repo/stable/arch-repo/x86_64
    tests:assert-stdout "arch-repo-stable-arch-repo-x86_64 package_one 1-1"
