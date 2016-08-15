tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

tests:ensure :list-repositories curl
tests:assert-stdout "can't read repo dir"
tests:assert-stdout "500"
