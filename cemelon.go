package main

// The MIT License (MIT)
// Copyright 2017 Vasilyuk Vasiliy <vasilyuk.vasiliy@gmail.com>

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	StartBlockIndex               int    = -1
	EndBlockIndex                 int    = -1
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

func main() {
	var (
		err               error
		blockIndexInt     int             = 0
		blockIndexStr     string          = ""
		prevBlockIndexInt int             = -1
		jsonDataString    string          = ""
		isWritten         map[string]bool = map[string]bool{}
		isDone            bool            = true
	)

	if EndBlockIndex < 0 || StartBlockIndex < 0 {
		flag.Usage()
		os.Exit(0)
	}

	blockIndexInt = StartBlockIndex
	for blockIndexInt <= EndBlockIndex {
		blockIndexStr = strconv.Itoa(blockIndexInt)
		fmt.Fprintln(os.Stdout, nowTime(), "|", "Block index: ", blockIndexInt)

		if prevBlockIndexInt != blockIndexInt {
			isWritten = map[string]bool{}
			jsonDataString = ""
		}

		if jsonDataString == "" {
			urlString := "https://blockchain.info/block-height/" + blockIndexStr + "?format=json"
			jsonDataString, err = FetchUrlByte(urlString)
			if err != nil {
				time.Sleep(time.Second)
				fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", err)
				continue
			}
		}

		regex := regexp.MustCompile("\"addr\":\"([13][a-km-zA-HJ-NP-Z1-9]{25,34})\"")
		addresses := regex.FindAllStringSubmatch(jsonDataString, -1)

		if len(addresses) < 1 {
			fmt.Fprintln(os.Stderr, nowTime(), "|", "Not found address, block index:", blockIndexInt)
			jsonDataString = ""
			time.Sleep(time.Second * 8)
			continue
		}

		isDone = true
		for j, num := range addresses {
			if j == 0 && !isWritten[blockIndexStr+"frs"] && !NotCollectFirstAddresses {
				err := write2file(FirstAddressesInBlockFileName, num[1])
				if err != nil {
					fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", err)
					isDone = false
					break
				} else {
					isWritten[blockIndexStr+"frs"] = true
				}
			}

			if NotCollectAllAddresses {
				isDone = true
				break
			}

			if !isWritten[blockIndexStr+"all"] {
				err := write2file(AllAddressesInBlockFileName, num[1])
				if err != nil {
					fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", err)
					isDone = false
					break
				} else {
					isWritten[blockIndexStr+"frs"] = true
				}
			}
		}

		if isDone {
			blockIndexInt++
		}
	}
}

func FetchUrlByte(urlString string) (responseBodyString string, err error) {
	httpClient := &http.Client{Transport: &http.Transport{}}
	response, err := httpClient.Get(urlString)
	defer response.Body.Close()
	if err != nil {
		return
	}

	if response.StatusCode < http.StatusOK || response.StatusCode > http.StatusIMUsed {
		return responseBodyString, errors.New("FetchUrlByte: Bad response status code!")
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	responseBodyString = string(responseBody)
	return
}

func nowTime() string {
	return time.Now().UTC().Format(time.RFC1123)
}

func write2file(filename string, dataString string) (err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		return
	}

	_, err = fmt.Fprintln(f, dataString)
	if err != nil {
		return
	}
	return
}
