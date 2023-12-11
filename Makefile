build:
	go build -o bin/ajudafortaleza cmd/ajudafortaleza/main.go

run: build
	./bin/ajudafortaleza