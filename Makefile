.PHONY: build
build: clean build-client build-server

build-client:
	cd client/cmd/ && go build -o client

build-server:
	cd server/cmd/ && go build -o server

.PHONY: clean
clean:
	rm -f client/cmd/client
	rm -f server/cmd/server

up-server: gen-keys
	echo -n "GOPHKEEPER_SERVICE_TOKEN_SECRET_KEY=" > server/build/dev_secret_key.env
	openssl rand -hex 20 >> server/build/dev_secret_key.env
	docker compose -f "server/build/docker-compose.yml" up -d --build

down-server:
	docker compose -f "server/build/docker-compose.yml" down -v

start-server:
	docker compose -f "server/build/docker-compose.yml" start

stop-server:
	docker compose -f "server/build/docker-compose.yml" stop

run-client:
	go run -tags debug ./client/cmd/

.PHONY: lint
lint:
	golangci-lint run ./client/...
	golangci-lint run ./server/...
	golangci-lint run ./pkg/...
	golangci-lint run ./common/...

gen-keys:
	go run /usr/local/go/src/crypto/tls/generate_cert.go -duration=168h -ca=true -host='localhost' $(date +"%b %d %H:%M:%S %Y")
	cp cert.pem server/build/cert.pem
	cp key.pem server/build/key.pem
	cp cert.pem client/build/cert.pem
	rm cert.pem && rm key.pem

gen-grpc:
	protoc --go_out=. --go_opt=paths=import   --go-grpc_out=. --go-grpc_opt=paths=import   common/api/keeper.proto
