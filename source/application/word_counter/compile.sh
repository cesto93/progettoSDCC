#!/bin/bash
go build -o ./worker/worker ./worker/worker.go
go build -o ./master/master ./master/master.go
go build -o ./client/client ./client/client.go
