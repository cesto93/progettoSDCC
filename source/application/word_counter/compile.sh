#!/bin/bash
cd ./worker
go build worker.go
cd ../master
go build master.go
