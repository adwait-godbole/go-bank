# Go Bank

In the Makefile, there is no need of doing "go clean -testcache"
before running "go test -v -cover -short ./...". In order to achieve the same functionality,
we can just do "go test -v -cover -short -count=1 ./..." where "-count=1" simply
disables test caching altogether.