:bootstrap-repository arch-repo testing testing-db x86_64
:bootstrap-repository ubuntu-repo testing testing-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""
Data = ["arch-repo", "ubuntu-repo"]
Status = 200'

tests:assert-equals "$(:list-repositories curl)" "$expected"
