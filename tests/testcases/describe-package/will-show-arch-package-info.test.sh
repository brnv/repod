:bootstrap-repository arch-repo testing testing-db x86_64

:add-package arch-repo testing testing-db x86_64 package_one

:describe-package arch-repo testing testing-db x86_64 package_one
tests:assert-stdout "This is test package description"
