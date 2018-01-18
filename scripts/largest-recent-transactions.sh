#!/bin/sh

RegExp="[13][a-km-zA-HJ-NP-Z1-9]{25,34}"
RegExpFilter="[13][a-km-zA-HJ-NP-Z1-9]{1,}[A-HJ-NP-Z]{1,}[a-km-zA-HJ-NP-Z1-9]{1,}"
ServiceAddress="https://blockchain.info/largest-recent-transactions"

TM=$(date +%Y-%m-%d_%H-%M-%S)
LRT="largest-recent-transactions-"${TM}

curl ${ServiceAddress} | egrep -o -e "${RegExp}" | egrep -e "${RegExpFilter}" | sort -u >> ${LRT}.tmp.txt
cat ${LRT}.tmp.txt | sort -u > ${LRT}.txt

rm ${LRT}.tmp.txt