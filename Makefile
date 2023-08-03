build:
	go build cmd/main.go
unpack:
	go run cmd/main.go unpack -logo $(logo)
repack:
	go run cmd/main.go repack -logo $(logo) -output $(output)
