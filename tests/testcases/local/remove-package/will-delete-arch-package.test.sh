:bootstrap-repository arch-repo testing testing-db x86_64

:add-package local arch-repo/testing/testing-db/x86_64 package_one
:list-packages local arch-repo/testing/testing-db/x86_64
tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package_one 1-1"

tests:ensure :stat-package arch-repo/testing/testing-db/x86_64 package_one

:remove-package local arch-repo/testing/testing-db/x86_64 package_one
:list-packages local arch-repo/testing/testing-db/x86_64
tests:assert-stdout-empty

tests:not tests:ensure :stat-package \
    arch-repo/testing/testing-db/x86_64 package_one
