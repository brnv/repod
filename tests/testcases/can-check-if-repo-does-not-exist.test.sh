:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="Success = false
Error = \"given repository doesn't exist\""

tests:assert-equals "$(:list-packages repo-arch testing testing-db x86_64)" "$expected"
tests:assert-success
