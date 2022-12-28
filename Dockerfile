# Builder container
FROM docker.io/library/golang:1.19-alpine AS builder
WORKDIR /go/src/app
COPY go.mod ./
COPY go.sum ./
COPY *.go ./
RUN go get -d -v ./...
RUN go build -o pulp-migrator

# App container
FROM alpine
COPY --from=builder /go/src/app/pulp-migrator /bin/
ENTRYPOINT ["/bin/pulp-migrator"]
