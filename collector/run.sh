#!/bin/bash

go build -buildmode=plugin -o ./plugins/web.so ../plugins/web 
go build -buildmode=plugin -o ./plugins/ping.so ../plugins/ping 
sudo /usr/local/go/bin/go run main.go