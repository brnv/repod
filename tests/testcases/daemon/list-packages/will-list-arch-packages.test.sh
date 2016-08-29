:bootstrap-repository arch-repo testing test-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package curl arch-repo/testing/test-db/x86_64 package1
:add-package curl arch-repo/testing/test-db/x86_64 package2

expected='Success = true
Error = ""
Data = ["arch-repo-testing-test-db-x86_64 package1 1-1", "arch-repo-testing-test-db-x86_64 package2 1-1"]
Status = 200'

actual=$(:list-packages curl arch-repo/testing/test-db/x86_64)

tests:assert-equals "$actual" "$expected"
