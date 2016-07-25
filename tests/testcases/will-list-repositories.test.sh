#!/bin/bash

:bootstrap-repositories repo1 repo2 repo3

tests:run-background bg_repod :run

tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""

[Data]
  repositories = ["repo1", "repo2", "repo3"]'

tests:assert-equals "$(:curl-repositories-list)" "$expected"

tests:assert-success
