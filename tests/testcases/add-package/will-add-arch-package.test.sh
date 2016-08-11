:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = []'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"

expected='Success = true
Error = ""
Status = 200

[Data]'

actual=$(:add-package arch-repo testing testing-db x86_64 package_one)

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = ["testing-db-testing package_one 1-1"]'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"
