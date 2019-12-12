#!/bin/bash
go build -o ./bin/worker ./source/application/word_counter/worker/worker.go
go build -o ./bin/master ./source/application/word_counter/master/master.go
go build -o ./bin/client ./source/application/word_counter/client/client.go
go build -o ./bin/agent ./source/monitoring/agent.go
