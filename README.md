# Go Bank

GoBank | Go, Gin, JWT, gRPC, Protocol Buffers, PostgreSQL, Asynq, Redis, Docker, GoMock | Github
• Built a banking system in Go using Gin framework and gRPC + Protocol Buffers for high-performance service
communication.
• Used pgx driver and golang-migrate for performing ACID transactions and zero-downtime migrations against
PostgreSQL.
• Integrated Asynq with Redis as a message broker for handling asynchronous email verifications and notifications.
• Achieved high test coverage using table driven tests and uber/mock testing framework.
• Built a minimal Docker image leveraging multi-stage builds and published to registries like Dockerhub and ECR.

In the Makefile, there is no need of doing "go clean -testcache"
before running "go test -v -cover -short ./...". In order to achieve the same functionality,
we can just do "go test -v -cover -short -count=1 ./..." where "-count=1" simply
disables test caching altogether.
