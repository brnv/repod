:bootstrap-repository arch-repo testing testing-db x86_64

:add-package local arch-repo/testing/testing-db/x86_64 package1
:add-package local arch-repo/testing/testing-db/x86_64 package2

:list-packages local arch-repo/testing/testing-db/x86_64

tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package1 1-1"
tests:assert-stdout "arch-repo-testing-testing-db-x86_64 package2 1-1"
