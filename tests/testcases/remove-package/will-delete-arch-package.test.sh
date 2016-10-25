:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo/testing/testing-db/x86_64 package

:list-packages arch-repo/testing/testing-db/x86_64
tests:assert-stdout-re "arch-repo-testing-testing-db-x86_64 package"

:add-package arch-repo/testing/testing-db/x86_64 package-one
:add-package arch-repo/testing/testing-db/x86_64 package-two-2

:list-packages arch-repo/testing/testing-db/x86_64
tests:assert-stdout-re "package "
tests:assert-stdout-re "package-one"
tests:assert-stdout-re "package-two-2"

:remove-package arch-repo/testing/testing-db/x86_64 package

:list-packages arch-repo/testing/testing-db/x86_64
tests:not tests:assert-stdout-re "package "
