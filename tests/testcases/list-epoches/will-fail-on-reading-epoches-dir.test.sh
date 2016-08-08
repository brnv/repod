:bootstrap-repository arch-repo

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="no epoches found for repo"
tests:ensure :list-epoches arch-repo
tests:assert-stdout "$expected"
