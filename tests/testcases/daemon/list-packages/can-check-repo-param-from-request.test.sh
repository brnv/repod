:bootstrap-repository yet-unknown-repo testing testing-db x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

tests:ensure :list-packages curl ubuntu/testing/testing-db/x86_64 not-implemented
tests:assert-stdout "can't start work with repo"
tests:assert-stdout "400"
