:bootstrap-repository arch-repo testing arch-repo x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected="can't start work with repo"
tests:ensure :list-packages arch-rep testing arch-repo x86_64
tests:assert-stdout "$expected"
