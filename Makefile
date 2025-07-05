mockgen:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./... -cover	

migrate:
	goose -dir ./migrations postgres "postgres://user:password@localhost:5432/db?sslmode=disable" up

docker-run:
	docker run --name metrics-postgres \
		-e POSTGRES_USER=user \
		-e POSTGRES_PASSWORD=password \
		-e POSTGRES_DB=db \
		-p 5432:5432 \
		-d postgres:15