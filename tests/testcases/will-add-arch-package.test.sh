#!/bin/bash

:bootstrap-repositories arch1
:bootstrap-epoches arch1 epoch1
:bootstrap-packages-arch arch1 epoch1 database1 architecture1 package1

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

tests:put-string new_package "binary package content"

:curl-add-package arch1 epoch1 database1 \
    architecture1 "new_package" $(tests:get-tmp-dir)/new_package

expected='Success = true
Error = ""

[Data]
  packages = ["new_package", "package1"]'
actual=$(:curl-list-packages arch1 epoch1 database1 architecture1)

tests:assert-equals "$actual" "$expected"
tests:assert-success
