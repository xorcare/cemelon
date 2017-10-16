# Cemelon

**The tool collects address used in the bitcoin blockchain!**

## Usage of cemelon

~~~
Usage of ./cemelon:
  -a string
        The name of the file which will be used to record all addresses in the block (default "all-cemelon-addresses.txt")
  -e int
        The block number on which program finished collecting the addresses including this number (default -1)
  -f string
        The name of the file which will be written to the first address in the block (default "frs-cemelon-addresses.txt")
  -m int
        The number of addresses stored in the card to prevent re-entry of addresses (default 1000000)
  -n int
        The number of threads downloading data (default 1)
  -r    Not to collect the first address in the block
  -s int
        The block number at which to start collecting addresses (default -1)
  -z    Not to collect all addresses
~~~


## License

The MIT License (MIT)

**Copyright 2017 Vasilyuk Vasiliy <vasilyukvasiliy@gmail.com>**