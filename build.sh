go mod tidy
mkdir -p ./dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./dist/web cmd/web.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./dist/tls cmd/tls.go
docker build -t rekey/doh .