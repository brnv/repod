:bootstrap-repository arch-repo testing arch-repo x86_64

tests:run-background bg_repod :run-daemon
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package curl arch-repo testing arch-repo x86_64 package_one

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = ["arch-repo-testing package_one 1-1"]'
actual=$(:list-packages curl arch-repo testing arch-repo x86_64)
tests:assert-equals "$actual" "$expected"

:bootstrap-repository arch-repo stable arch-repo x86_64

:copy-package-to-epoch \
    curl arch-repo testing arch-repo x86_64 package_one stable

expected='Success = true
Error = ""
Status = 200

[Data]
  packages = ["arch-repo-stable package_one 1-1"]'
actual=$(:list-packages curl arch-repo stable arch-repo x86_64)
tests:assert-equals "$actual" "$expected"
