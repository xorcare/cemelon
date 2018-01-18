#!/bin/sh

BTC="bitcoin"
BTH="bitcoin%20cash"
RegExp="[13][a-km-zA-HJ-NP-Z1-9]{25,34}"
RegExpFilter="[13][a-km-zA-HJ-NP-Z1-9]{1,}[A-HJ-NP-Z]{1,}[a-km-zA-HJ-NP-Z1-9]{1,}"

curl https://bitinfocharts.com/ru/top-100-richest-${BTC}-addresses.html | egrep -o -e "${RegExp}" | egrep -o -e "${RegExpFilter}" | sort -u >> ${BTC}.tmp.txt
curl https://bitinfocharts.com/ru/top-100-richest-${BTH}-addresses.html | egrep -o -e "${RegExp}" | egrep -o -e "${RegExpFilter}" | sort -u >> ${BTH}.tmp.txt

for (( i=1; i <= 101; i++ ))
do
curl https://bitinfocharts.com/ru/top-100-richest-${BTC}-addresses-${i}.html | egrep -o -e "${RegExp}" |  egrep -o -e "${RegExpFilter}" |sort -u >> ${BTC}.tmp.txt
curl https://bitinfocharts.com/ru/top-100-richest-${BTH}-addresses-${i}.html | egrep -o -e "${RegExp}" |  egrep -o -e "${RegExpFilter}" |sort -u >> ${BTH}.tmp.txt
done

TM=$(date +%Y-%m-%d_%H-%M-%S)
cat ${BTC}.tmp.txt | sort -u > top-10000-richest-bitcoin-addresses-${TM}.txt
cat ${BTH}.tmp.txt | sort -u > top-10000-richest-bitcoin-cash-addresses-${TM}.txt

rm ${BTH}.tmp.txt ${BTC}.tmp.txt
