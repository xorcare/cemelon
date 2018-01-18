#!/bin/sh

patter="([13][a-km-zA-HJ-NP-Z1-9]{25,34})"

TM=$(date +%Y-%m-%d_%H-%M-%S)
LRT="largest-recent-transactions-"${TM}

rm *.tmp.txt

curl https://blockchain.info/largest-recent-transactions | egrep -o -e "${patter}" | sort -u >> ${LRT}.tmp.txt
cat ${LRT}.tmp.txt | sort -u > ${LRT}.txt

rm *.tmp.txt