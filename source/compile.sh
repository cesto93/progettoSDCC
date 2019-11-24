#!/bin/bash
go build -o ./application/word_counter/worker/worker ./application/word_counter/worker/worker.go
go build -o ./application/word_counter/master/master ./application/word_counter/master/master.go
go build -o ./application/word_counter/client/client ./application/word_counter/client/client.go
go build -o ./monitoring/agent ./monitoring/agent.go
