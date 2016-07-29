#!/bin/bash

:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package arch-repo testing testing-db x86_64 package_one

expected='Success = true
Error = ""

[Data]
  packages = ["package_one"]'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"
tests:assert-success

:remove-package arch-repo testing testing-db x86_64 package_one

expected='Success = true
Error = ""

[Data]
  packages = []'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"
tests:assert-success
