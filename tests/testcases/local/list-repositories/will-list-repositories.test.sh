:bootstrap-repository arch-repo testing testing-db x86_64
:bootstrap-repository ubuntu-repo testing testing-db x86_64

:list-repositories local
#tests:assert-stdout "arch-repo"
#tests:assert-stdout "ubuntu-repo"
