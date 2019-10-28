#!/bin/bash
ports=( "1050" "1051" "1052" )

for (( i=0; i<${#ports[@]}; i++ ));
do
  ./worker/worker ${ports[$i]} &
done
