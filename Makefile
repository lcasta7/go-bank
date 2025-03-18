build:
	@go build -o bin/gobank

run: build
	@./bin/gobank

seed: build
	@./bin/gobank  --seed

createAdmin: build
	@./bin/gobank --create-admin --password="gobank" --first_name="chuck" --last_name="norris"

test:
	@go test -v ./...
