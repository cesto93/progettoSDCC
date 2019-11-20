#!/bin/bash
addresses=( "localhost:1050" "localhost:1051" "localhost:1052" )
files=( "../../../words_to_count/prova1.txt" "../../../words_to_count/prova2.txt" "../../../words_to_count/gpl-3.0.txt" )
master=1049

./master/master -workerAddr ${addresses[0]},${addresses[1]},${addresses[2]} -masterPort $master
