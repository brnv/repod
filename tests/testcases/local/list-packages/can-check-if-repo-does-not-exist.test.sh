:bootstrap-repository arch-repo testing testing-db x86_64
:list-packages local not-exist testing testing-db x86_64
tests:assert-stdout-empty
