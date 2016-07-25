#!/bin/bash

:bootstrap-repositories repo1

:bootstrap-epoches repo1 epoch1

tests:run-background bg_repod :run

tests:wait-file-matches $(tests:get-background-stderr $bg_repod) "serving" 1 2

expected='Success = true
Error = ""

[Data]
  epoches = ["epoch1"]'

tests:assert-equals "$(:curl-epoches-list repo1)" "$expected"

tests:assert-success
