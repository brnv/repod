:bootstrap-repository yet-unknown-repo/testing/testing-db/x86_64

:list-packages yet-unknown-repo/testing/testing-db/x86_64 yet-unknown-repo
    tests:eval echo "$(tests:get-stdout) $(tests:get-stderr)"
    tests:assert-stdout "can't obtain repository system"
