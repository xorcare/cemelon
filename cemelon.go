package main

// The MIT License (MIT)
// Copyright 2017-2018 Vasiliy Vasilyuk <vasilyukvasiliy@gmail.com>

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vasilyukvasiliy/blockchain"
)

var (
	startBlockIndex          = -1
	endBlockIndex            = -1
	countStreams             = 1
	whitenAddressSize        = 262144
	notCollectFirstAddresses = false
	notCollectAllAddresses   = false
	OutFileBaseName          = "cemelon.txt"
	isWhitenAddress          Map
)

func init() {
	fmt.Println(`
#######################################################
#                                                     #
#   #####                                             #
#  #     # ###### #    # ###### #       ####  #    #  #
#  #       #      ##  ## #      #      #    # ##   #  #
#  #       #####  # ## # #####  #      #    # # #  #  #
#  #       #      #    # #      #      #    # #  # #  #
#  #     # #      #    # #      #      #    # #   ##  #
#   #####  ###### #    # ###### ######  ####  #    #  #
#                                                     #
#  The MIT License (MIT)                              #
#  Copyright 2017-2018 Vasiliy Vasilyuk               #
#  Email: vasilyukvasiliy@gmail.com                   #
#######################################################
#            Github: https://git.io/vNIKR             #
#######################################################

`)

	fmt.Println("Runned:", strings.Join(os.Args, " "))

	flag.IntVar(&startBlockIndex, "s", startBlockIndex, "The block number at which to start collecting addresses")
	flag.IntVar(&endBlockIndex, "e", endBlockIndex, "The block number on which program finished collecting the addresses including this number")
	flag.IntVar(&whitenAddressSize, "m", whitenAddressSize, "The number of addresses stored in the card to prevent re-entry of addresses")
	flag.BoolVar(&notCollectFirstAddresses, "r", notCollectFirstAddresses, "Not to collect the first address in the block")
	flag.IntVar(&countStreams, "n", countStreams, "The number of threads downloading data")
	flag.StringVar(&OutFileBaseName, "o", OutFileBaseName, "Output data file base name")
	flag.BoolVar(&notCollectAllAddresses, "z", notCollectAllAddresses, "Not to collect all addresses")
	flag.Parse()

	isWhitenAddress = *NewMap()
}

type Map struct {
	mx sync.Mutex
	m  map[string]bool
}

func (c *Map) Count() int {
	c.mx.Lock()
	defer c.mx.Unlock()
	return len(c.m)
}

func (c *Map) Exist(key string) bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	_, ok := c.m[key]
	return ok
}

func (c *Map) Store(key string, value bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key] = value
}

func (c *Map) Clear() {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m = map[string]bool{}
}

func NewMap() *Map {
	return &Map{
		m: make(map[string]bool),
	}
}

type InformationRecord struct {
	Filename   string
	Message    string
	BlockIndex int
}

func Write2FileFromChan(cn <-chan InformationRecord, wg *sync.WaitGroup) {
	var files = map[string]*os.File{}
	var err error = nil

	pid := strconv.Itoa(os.Getpid())
	for {
		dan := <-cn
		dan.Filename = pid + "-" + dan.Filename
		for counter := 0; counter <= 64; counter++ {
			if files[dan.Filename] == nil {
				files[dan.Filename], err =
					os.OpenFile(dan.Filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					files[dan.Filename] = nil
					fmt.Fprintln(os.Stderr, nowTimeRFC1123(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
					continue
				}
				defer files[dan.Filename].Close()
			}
			_, err = fmt.Fprintln(files[dan.Filename], dan.Message)
			if err != nil {
				fmt.Fprintln(os.Stderr, nowTimeRFC1123(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
			} else {
				break
			}
			if counter == 64 {
				log.Fatalln(nowTimeRFC1123(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func main() {
	if endBlockIndex < 0 || startBlockIndex < 0 || (endBlockIndex-startBlockIndex) < 0 {
		flag.Usage()
		os.Exit(0)
	}

	count := endBlockIndex - startBlockIndex
	step := int(count / countStreams)
	var wg sync.WaitGroup
	chanInformationRecords := make(chan InformationRecord, 2*countStreams)

	go Write2FileFromChan(chanInformationRecords, &wg)

	if count > 0 && step > 1 {
		for i := startBlockIndex; i <= endBlockIndex; i += step {
			end := i + step - 1
			start := i
			if end > endBlockIndex || (endBlockIndex-end) < step {
				end = endBlockIndex
				i = endBlockIndex
			}
			go worker(&wg, chanInformationRecords, start, end)
		}
	} else {
		go worker(&wg, chanInformationRecords, startBlockIndex, endBlockIndex)
	}

	time.Sleep(time.Second)
	wg.Wait()

	for {
		time.Sleep(time.Second)
		if len(chanInformationRecords) == 0 {
			break
		}
	}

	time.Sleep(time.Second)
}

func worker(wg *sync.WaitGroup, cn chan<- InformationRecord, startIndex, endIndex int) {
	wg.Add(1)
	defer wg.Done()

	var block *blockchain.Block = nil
	var (
		blockIndexInt     = 0
		blockIndexStr     = ""
		prevBlockIndexInt = -1
		isDone            = true
	)

	blc := blockchain.New()
	blockIndexInt = startIndex
	for blockIndexInt <= endIndex {
		blockIndexStr = strconv.Itoa(blockIndexInt)
		fmt.Fprintln(os.Stdout, nowTimeRFC1123(), "|", "Block index: ", blockIndexInt)

		if prevBlockIndexInt != blockIndexInt {
			block = nil
		}

		if block == nil {
			res, err := blc.GetBlockHeight(blockIndexStr)
			if err != nil {
				fmt.Fprintln(os.Stderr, nowTimeRFC1123(), "|", "Block index: ", blockIndexInt, "[Error]", err)
				block = nil
				time.Sleep(time.Second * 16)
				continue
			}
			block = &res.Blocks[0]
		}

		if len(block.Tx) < 1 {
			fmt.Fprintln(os.Stderr, nowTimeRFC1123(), "|", "Not found address, block index:", blockIndexInt)
			block = nil
			time.Sleep(time.Second * 16)
			continue
		}

		addresses := make([]string, 0, 0)
		for _, tx := range block.Tx {
			for _, out := range tx.Out {
				addresses = append(addresses, out.Addr)
			}
		}

		if !isWhitenAddress.Exist(block.Hash) {
			cn <- InformationRecord{
				Filename:   "blk-" + OutFileBaseName,
				Message:    block.Hash,
				BlockIndex: blockIndexInt,
			}
			isWhitenAddress.Store(block.Hash, true)
		}

		isDone = true
		for j, address := range addresses {
			if j == 0 && !isWhitenAddress.Exist(blockIndexStr) && !notCollectFirstAddresses {
				cn <- InformationRecord{
					Filename:   "frs-" + OutFileBaseName,
					Message:    address,
					BlockIndex: blockIndexInt,
				}
				isWhitenAddress.Store(blockIndexStr, true)
			}

			if notCollectAllAddresses {
				isDone = true
				break
			}

			if !isWhitenAddress.Exist(address) {
				cn <- InformationRecord{
					Filename:   "all-" + OutFileBaseName,
					Message:    address,
					BlockIndex: blockIndexInt,
				}
				isWhitenAddress.Store(address, true)
			}
		}

		if isDone {
			blockIndexInt++
			if isWhitenAddress.Count() > whitenAddressSize {
				isWhitenAddress.Clear()
			}
		}
	}
}

func nowTimeRFC1123() string {
	return time.Now().UTC().Format(time.RFC1123)
}
