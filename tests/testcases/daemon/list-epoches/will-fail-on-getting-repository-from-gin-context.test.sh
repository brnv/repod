:bootstrap-repository arch-repo

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

tests:ensure :list-epoches curl not-exists
tests:assert-stdout "can't start work with repo"
tests:assert-stdout "400"
