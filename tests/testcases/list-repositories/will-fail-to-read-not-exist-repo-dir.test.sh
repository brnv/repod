tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="can't read repo dir"
tests:ensure :list-repositories
tests:assert-stdout "$expected"
