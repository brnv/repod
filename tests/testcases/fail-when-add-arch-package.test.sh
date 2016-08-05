:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""

[Data]
  packages = []'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"

expected='Success = true
Error = ""

[Data]'

actual=$(:add-package arch-repo testing testing-db x86_64 package_one)

tests:assert-equals "$actual" "$expected"

actual=$(:add-package arch-repo testing testing-db x86_64 package_one)
tests:not tests:assert-equals "$actual" "$expected"
