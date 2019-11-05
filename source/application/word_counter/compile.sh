#!/bin/bash
#cd ./worker
go build -o ./worker/worker ./worker/worker.go
#cd ../master
go build -o ./master/master ./master/master.go
