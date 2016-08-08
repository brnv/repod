:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="can't ensure repo paths"
tests:ensure :list-packages not-exist testing testing-db x86_64
tests:assert-stdout "$expected"
