go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/app /dist/runner
docker build -t rekey/doh .