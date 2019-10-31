#!/bin/bash
addresses=( "localhost:1050" "localhost:1051" "localhost:1052" )
files=( "prova1.txt" "prova2.txt" "gpl-3.0.txt" )

./master/master -files ${files[0]},${files[1]},${files[2]} \
-ports ${addresses[0]},${addresses[1]},${addresses[2]}
