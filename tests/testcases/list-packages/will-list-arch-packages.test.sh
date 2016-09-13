:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo/testing/testing-db/x86_64 package1
:add-package arch-repo/testing/testing-db/x86_64 package2

:list-packages arch-repo/testing/testing-db/x86_64
    tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package1"
    tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package2"
