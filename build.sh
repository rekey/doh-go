go mod tidy
mkdir -p ./dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./dist/runner cmd/web.go
docker build -t rekey/doh .