:bootstrap-repository arch-repo/testing/testing-db/x86_64

:add-package arch-repo/testing/testing-db/x86_64 package_one

:add-package arch-repo/testing/testing-db/x86_64 package_one
    tests:eval echo "$(tests:get-stdout) $(tests:get-stderr)"
    tests:assert-re stdout "can't add package"
