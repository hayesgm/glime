#! /bin/bash

go build glime.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o glime.linux glime.go
