.PHONY: build run clean

build:
	GOOS=linux GOARCH=amd64 go build -o policy-poc .

run: build
	func start

clean:
	rm -f policy-poc
