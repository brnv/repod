:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package arch-repo testing testing-db x86_64 package1
:add-package arch-repo testing testing-db x86_64 package2

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = ["testing-db-testing package1 1-1", "testing-db-testing package2 1-1"]'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"
