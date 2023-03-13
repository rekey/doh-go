go mod tidy
mkdir -p ./dist
CGO_ENABLED=0 
GOOS=linux
GOARCH=amd64
go build -o ./docker/web/web cmd/web.go
go build -o ./docker/tls/tls cmd/tls.go
docker build -t rekey/doh docker/web
docker build -t rekey/doh:tls docker/tls