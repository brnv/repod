#!/bin/bash

:bootstrap-repository arch-repo testing testing-db x86_64
:bootstrap-repository ubuntu-repo testing testing-db x86_64

tests:run-background bg_repod :run
tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""

[Data]
  repositories = ["arch-repo", "ubuntu-repo"]'

tests:assert-equals "$(:list-repositories)" "$expected"

tests:assert-success
