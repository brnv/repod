#!/bin/bash

tests:involve tests/functions.sh

tests:clone repod bin/repod

tests:clone tests/mocks/gpg bin/gpg
tests:clone tests/mocks/repo-add bin/repo-add
