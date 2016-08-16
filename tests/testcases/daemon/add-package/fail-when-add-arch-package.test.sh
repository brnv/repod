:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = []'
actual=$(:list-packages curl arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"

expected='Success = true
Error = ""
Status = 200

[Data]'

actual=$(:add-package curl arch-repo testing testing-db x86_64 package_one)

tests:assert-equals "$actual" "$expected"

actual=$(:add-package curl arch-repo testing testing-db x86_64 package_one)
tests:not tests:assert-equals "$actual" "$expected"