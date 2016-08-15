:bootstrap-repository arch-repo testing testing-db x86_64
:list-packages local arch-repo testing testing-db x86_64
tests:assert-stdout-empty
:add-package local arch-repo testing testing-db x86_64 package_one

:add-package local arch-repo testing testing-db x86_64 package_one
tests:assert-fail
