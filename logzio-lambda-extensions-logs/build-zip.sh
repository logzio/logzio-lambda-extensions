env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/extensions/logzio-lambda-extensions-logs main.go
chmod +x bin/extensions/logzio-lambda-extensions-logs
cd bin
zip -r extension.zip extensions/