:bootstrap-repository arch-repo testing testing-db x86_64
:add-package local arch-repo/testing/testing-db/x86_64 package_one
:describe-package local arch-repo/testing/testing-db/x86_64 package_one
tests:assert-stdout "This is test package description"

:edit-package-description \
    local arch-repo/testing/testing-db/x86_64 package_one "New description"

:describe-package local arch-repo/testing/testing-db/x86_64 package_one
tests:assert-stdout "New description"
