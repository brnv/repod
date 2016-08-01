:bootstrap-repository yet-unknown-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="Success = false
Error = \"can't detect repository from request\""

tests:assert-equals "$(:list-packages yet-unknown-repo testing testing-db x86_64)" "$expected"
tests:assert-success
