:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo/testing/testing-db/x86_64 package_one

:list-packages arch-repo/testing/testing-db/x86_64
    tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package_one 1-1"
