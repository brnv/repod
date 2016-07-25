#!/bin/bash

:bootstrap-repositories arch1
:bootstrap-epoches arch1 epoch1
:bootstrap-packages-arch arch1 epoch1 database1 architecture1 package1 package2

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""

[Data]
  packages = ["package1", "package2"]'
actual=$(:curl-list-packages arch1 epoch1 database1 architecture1)

tests:assert-equals "$actual" "$expected"
tests:assert-success
