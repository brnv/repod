:bootstrap-repository arch-repo testing testing-db x86_64
:add-package local arch-repo testing testing-db x86_64 package_one
:list-packages local arch-repo testing testing-db x86_64
tests:assert-stdout "testing-db-testing package_one 1-1"
