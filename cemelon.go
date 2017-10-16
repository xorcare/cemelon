package main

// The MIT License (MIT)
// Copyright 2017 Vasilyuk Vasiliy <vasilyukvasiliy@gmail.com>

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	StartBlockIndex               int    = -1
	EndBlockIndex                 int    = -1
	countStreams                  int    = 1
	NotCollectFirstAddresses      bool   = false
	NotCollectAllAddresses        bool   = false
	FirstAddressesInBlockFileName string = "frs-cemelon-addresses.txt"
	AllAddressesInBlockFileName   string = "all-cemelon-addresses.txt"
)

func init() {
	flag.IntVar(&StartBlockIndex, "s", StartBlockIndex,
		"The block number at which to start collecting addresses")

	flag.IntVar(&EndBlockIndex, "e", EndBlockIndex,
		"The block number on which program finished collecting the addresses including this number")

	flag.IntVar(&countStreams, "n", countStreams,
		"The number of threads downloading data")

	flag.StringVar(&FirstAddressesInBlockFileName, "f", FirstAddressesInBlockFileName,
		"The name of the file which will be written to the first address in the block")

	flag.StringVar(&AllAddressesInBlockFileName, "a", AllAddressesInBlockFileName,
		"The name of the file which will be used to record all addresses in the block")

	flag.BoolVar(&NotCollectFirstAddresses, "r", NotCollectFirstAddresses,
		"Not to collect the first address in the block")

	flag.BoolVar(&NotCollectAllAddresses, "z", NotCollectAllAddresses,
		"Not to collect all addresses")

	flag.Parse()
}

type InformationRecord struct {
	Filename   string
	Message    string
	BlockIndex int
}

func Write2FileFromChan(cn <-chan InformationRecord, wg *sync.WaitGroup) {
	var files map[string]*os.File = map[string]*os.File{}
	var err error = nil
	for {
		dan := <-cn
		wg.Add(1)
		for counter := 0; counter <= 64; counter++ {
			if files[dan.Filename] == nil {
				files[dan.Filename], err =
					os.OpenFile(dan.Filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					files[dan.Filename] = nil
					fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
					continue
				}
				defer files[dan.Filename].Close()
			}
			_, err = fmt.Fprintln(files[dan.Filename], dan.Message)
			if err != nil {
				fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
			} else {
				break
			}
			if counter == 64 {
				log.Fatalln(nowTime(), "|", "Block index: ", dan.BlockIndex, "[Error]", err)
			}
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}
}

func main() {
	if EndBlockIndex < 0 || StartBlockIndex < 0 || (EndBlockIndex-StartBlockIndex) < 0 {
		flag.Usage()
		os.Exit(0)
	}

	count := EndBlockIndex - StartBlockIndex
	step := int(count / countStreams)
	var wg sync.WaitGroup
	var writer chan InformationRecord = make(chan InformationRecord, 32*countStreams)

	go Write2FileFromChan(writer, &wg)

	if count > 0 && step > 1 {
		for i := StartBlockIndex; i <= EndBlockIndex; i += step {
			end := i + step - 1
			start := i
			if end > EndBlockIndex || (EndBlockIndex-end) < step {
				end = EndBlockIndex
				i = EndBlockIndex
			}
			go worker(&wg, writer, start, end)
		}
		time.Sleep(time.Second)

	} else {
		go worker(&wg, writer, StartBlockIndex, EndBlockIndex)
	}

	for {
		time.Sleep(time.Millisecond)
		wg.Wait()

		if len(writer) == 0 {
			break
		}
	}
}

func worker(wg *sync.WaitGroup, cn chan<- InformationRecord, startIndex, endIndex int) {
	wg.Add(1)
	defer wg.Done()

	var (
		err               error
		blockIndexInt     int             = 0
		blockIndexStr     string          = ""
		prevBlockIndexInt int             = -1
		jsonDataString    string          = ""
		isWritten         map[string]bool = map[string]bool{}
		isDone            bool            = true
	)

	blockIndexInt = startIndex
	for blockIndexInt <= endIndex {
		blockIndexStr = strconv.Itoa(blockIndexInt)
		fmt.Fprintln(os.Stdout, nowTime(), "|", "Block index: ", blockIndexInt)

		if prevBlockIndexInt != blockIndexInt {
			isWritten = map[string]bool{}
			jsonDataString = ""
		}

		if jsonDataString == "" {
			urlString := "https://blockchain.info/block-height/" + blockIndexStr + "?format=json"
			jsonDataString, err = GetDataStringByUrl(urlString)
			if err != nil {
				fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", err)
				jsonDataString = ""
				time.Sleep(time.Second * 16)
				continue
			}
		}

		regex := regexp.MustCompile("\"addr\":\"([13][a-km-zA-HJ-NP-Z1-9]{25,34})\"")
		addresses := regex.FindAllStringSubmatch(jsonDataString, -1)

		if len(addresses) < 1 {
			fmt.Fprintln(os.Stderr, nowTime(), "|", "Not found address, block index:", blockIndexInt)
			jsonDataString = ""
			time.Sleep(time.Second * 16)
			continue
		}

		isDone = true
		for j, num := range addresses {
			if j == 0 && !isWritten[blockIndexStr+"frs"] && !NotCollectFirstAddresses {
				cn <- InformationRecord{
					Filename:   FirstAddressesInBlockFileName,
					Message:    num[1],
					BlockIndex: blockIndexInt,
				}
				isWritten[blockIndexStr+"frs"] = true
			}

			if NotCollectAllAddresses {
				isDone = true
				break
			}

			if !isWritten[num[1]+"all"] {
				cn <- InformationRecord{
					Filename:   AllAddressesInBlockFileName,
					Message:    num[1],
					BlockIndex: blockIndexInt,
				}
				isWritten[num[1]+"all"] = true
			}
		}

		if isDone {
			blockIndexInt++
			if len(isWritten) > 1000 {
				isWritten = map[string]bool{}
			}
		}
	}
}

func GetDataStringByUrl(url string) (s string, e error) {
	s = ""
	c := &http.Client{Transport: &http.Transport{}}
	response, e := c.Get(url)
	if e != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode > http.StatusIMUsed {
		return s, errors.New("GetDataStringByUrl: Bad response status code!")
	}

	responseBody, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return
	}

	s = string(responseBody)
	return
}

func nowTime() string {
	return time.Now().UTC().Format(time.RFC1123)
}
