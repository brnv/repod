#!/bin/bash

:bootstrap-repository arch-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

:add-package arch-repo testing testing-db x86_64 package1
:add-package arch-repo testing testing-db x86_64 package2

expected='Success = true
Error = ""

[Data]
  packages = ["package1", "package2"]'
actual=$(:list-packages arch-repo testing testing-db x86_64)

tests:assert-equals "$actual" "$expected"
tests:assert-success
