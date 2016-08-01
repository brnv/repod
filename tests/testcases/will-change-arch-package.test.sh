:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package arch-repo testing testing-db x86_64 package_one
tests:eval :describe-package arch-repo testing testing-db x86_64 package_one
tests:assert-re stdout "This is test package description"
tests:assert-success

:edit-package-description arch-repo testing testing-db x86_64 package_one "New description"
tests:eval :describe-package arch-repo testing testing-db x86_64 package_one
tests:assert-re stdout "New description"
tests:assert-success