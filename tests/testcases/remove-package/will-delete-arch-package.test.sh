:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo/testing/testing-db/x86_64 package_one

:list-packages arch-repo/testing/testing-db/x86_64
    tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package_one 1-1"

tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package_one

:remove-package arch-repo/testing/testing-db/x86_64 package_one

tests:not tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package_one
