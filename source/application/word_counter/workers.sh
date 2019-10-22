#!/bin/bash
ports=( "$@" )

for (( i=0; i<${#ports[@]}; i++ ));
do
  ./worker/worker ${ports[$i]} &
done
