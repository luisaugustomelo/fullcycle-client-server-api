run-service:
	go run -tags=server server/server.go
	go run -tags=client client/client.go