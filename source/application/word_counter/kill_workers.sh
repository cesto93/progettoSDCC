#!/bin/bash
ports=( "1050" "1051" "1052")

for (( i=0; i<${#ports[@]}; i++ ));
do
  lsof -i tcp:${ports[$i]} | awk 'NR!=1 {print $2}' | xargs kill
done 
