:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package curl arch-repo/testing/testing-db/x86_64 package_one

tests:eval :describe-package \
    curl arch-repo/testing/testing-db/x86_64 package_one

tests:assert-re stdout "This is test package description"

:edit-package-description \
    curl arch-repo/testing/testing-db/x86_64 package_one "New description"

tests:eval :describe-package \
    curl arch-repo/testing/testing-db/x86_64 package_one

tests:assert-re stdout "New description"
