# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.20-bullseye AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 cd ./cmd && go build -buildvcs=false -o /goph_keeper_server

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base AS build-release-stage

WORKDIR /

COPY --from=build-stage /goph_keeper_server /goph_keeper_server
COPY --from=build-stage ./app/build/key.pem /key.pem
COPY --from=build-stage ./app/build/cert.pem /cert.pem

EXPOSE $GOPHKEEPER_GRPC_PORT

CMD ["/goph_keeper_server"]