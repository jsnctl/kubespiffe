build:
	go build -o kubespiffed

run: build
	./kubespiffed

test:
	go test ./...


docker: build
	docker build -t kubespiffed .
