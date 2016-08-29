:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package curl arch-repo/testing/testing-db/x86_64 package_one

expected='Success = true
Error = ""
Data = ["arch-repo-testing-testing-db-x86_64 package_one 1-1"]
Status = 200'
actual=$(:list-packages curl arch-repo/testing/testing-db/x86_64)

tests:assert-equals "$actual" "$expected"

tests:ensure :stat-package arch-repo/testing/testing-db/x86_64 package_one

:remove-package curl arch-repo/testing/testing-db/x86_64 package_one

expected='Success = true
Error = ""
Data = []
Status = 200'
actual=$(:list-packages curl arch-repo/testing/testing-db/x86_64)

tests:assert-equals "$actual" "$expected"

tests:not tests:ensure \
    :stat-package arch-repo testing testing-db x86_64 package_one
