#!/bin/bash
#USER=$1
#PASS=$2
cd ./go/src/progettoSDCC
git pull

go build -o ./source/application/word_counter/worker/worker ./source/application/word_counter/worker/worker.go
go build -o ./source/application/word_counter/master/master ./source/application/word_counter/master/master.go

