:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""
Status = 200

[Data]
  epoches = ["testing"]'

tests:assert-equals "$(:list-epoches arch-repo)" "$expected"
tests:assert-success
