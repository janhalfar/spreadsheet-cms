SHELL = /bin/bash

build:
	go build -o /usr/local/bin/spreadsheet-cms main.go
build-example:
	echo "hello"
	go run main.go -template example/template.html -out example -asset-dir example/assets -csv example/data.csv