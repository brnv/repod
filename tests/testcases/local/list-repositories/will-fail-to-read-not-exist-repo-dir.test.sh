tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2
tests:ensure :list-repositories local
tests:assert-stdout-empty
