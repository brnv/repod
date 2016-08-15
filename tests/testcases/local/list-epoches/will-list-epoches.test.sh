:bootstrap-repository arch-repo testing testing-db x86_64
:list-epoches local arch-repo
tests:assert-stdout "testing"
