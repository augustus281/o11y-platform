run:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run cmd/main.go
docker-build:
	docker build -t user-service .
docker-run:
	docker run -d --name user-service \
		-p 8888:8888 \
		-e OTEL_EXPORTER_OTLP_ENDPOINT="http://host.docker.internal:4318" \
		user-service:latest
docker-restart:
	docker rm -f user-service && make docker-run