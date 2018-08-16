# Cemelon

**The tool collects address used in the bitcoin blockchain!**

## Usage of cemelon
```text
Usage of cemelon:
  -b    Save only the addresses with a balance
  -c    To check the balance of addresses and to hash160
  -e int
        The block number on which program finished collecting the addresses including this number (default -1)
  -f string
        Output data format string (default "%34s, %s, %16d")
  -m int
        The number of addresses stored in the card to prevent re-entry of addresses (default 262144)
  -n int
        The number of threads downloading data (default 1)
  -o string
        Output data file base name (default "cemelon.txt")
  -r    Not to collect the first address in the block
  -s int
        The block number at which to start collecting addresses (default -1)
  -z    Not to collect all addresses
```

## Dependencies

Install go packages:

```bash
go get github.com/xorcare/blockchain
```

## Scripts

 * **top100bitcoin-richest.sh** - Allows you to collect top 10,000 richest Bitcoin and Bitcoin Cash addresses

## Example of the execution log

```log
Program: Cemelon
Author: Vasiliy Vasilyuk
Github: https://git.io/fNhcc
License: BSD 3-Clause "New" or "Revised" License

Runned: cemelon -s 0 -e 10 -n 2 -c -b
Thu, 16 Aug 2018 19:39:39 UTC | Block index:  5
Thu, 16 Aug 2018 19:39:39 UTC | Block index:  0
Thu, 16 Aug 2018 19:39:43 UTC | Block index:  6
Thu, 16 Aug 2018 19:39:43 UTC | Block index:  1
Thu, 16 Aug 2018 19:39:43 UTC | Block index:  7
Thu, 16 Aug 2018 19:39:43 UTC | Block index:  8
Thu, 16 Aug 2018 19:39:43 UTC | Block index:  2
Thu, 16 Aug 2018 19:39:44 UTC | Block index:  9
Thu, 16 Aug 2018 19:39:44 UTC | Block index:  3
Thu, 16 Aug 2018 19:39:44 UTC | Block index:  4
Thu, 16 Aug 2018 19:39:44 UTC | Block index:  10

Process finished with exit code 0
```