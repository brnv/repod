:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo/testing/testing-db/x86_64 package

:list-packages arch-repo/testing/testing-db/x86_64
    tests:assert-stdout-re "arch-repo-testing-testing-db-x86_64 package"

:add-package arch-repo/testing/testing-db/x86_64 package-one
:add-package arch-repo/testing/testing-db/x86_64 package-two-2

tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package

:remove-package arch-repo/testing/testing-db/x86_64 package

:list-packages arch-repo/testing/testing-db/x86_64

tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package-one

tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package-two-2

tests:not tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package
