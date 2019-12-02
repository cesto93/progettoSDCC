#!/bin/bash

sudo ./zookeeper/bin/zkServer.sh start
cd ./go/src/progettoSDCC/bin
./agent -aws
