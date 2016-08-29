:bootstrap-repository arch-repo testing arch-repo x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package curl arch-repo/testing/arch-repo/x86_64 package_one

expected='Success = true
Error = ""
Data = ["arch-repo-testing-arch-repo-x86_64 package_one 1-1"]
Status = 200'

actual=$(:list-packages curl arch-repo/testing/arch-repo/x86_64)
tests:assert-equals "$actual" "$expected"

:bootstrap-repository arch-repo stable arch-repo x86_64

:copy-package-to-new-root \
    curl arch-repo/testing/arch-repo/x86_64 package_one \
    arch-repo/stable/arch-repo/x86_64

expected='Success = true
Error = ""
Data = ["arch-repo-stable-arch-repo-x86_64 package_one 1-1"]
Status = 200'
actual=$(:list-packages curl arch-repo/stable/arch-repo/x86_64)
tests:assert-equals "$actual" "$expected"
