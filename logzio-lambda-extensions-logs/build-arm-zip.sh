env GOOS=linux GOARCH=arm64 go build -o bin/extensions/logzio-lambda-extensions-logs main.go
chmod +x bin/extensions/logzio-lambda-extensions-logs
cd bin
zip -r extension.zip extensions/