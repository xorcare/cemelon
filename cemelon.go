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
		e                 error
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
			jsonDataString, e = GetDataStringByUrl(urlString)
			if e != nil {
				jsonDataString = ""
				time.Sleep(time.Second)
				fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", e)
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
				e = write2file(FirstAddressesInBlockFileName, num[1])
				if e != nil {
					fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", e)
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
				e = write2file(AllAddressesInBlockFileName, num[1])
				if e != nil {
					fmt.Fprintln(os.Stderr, nowTime(), "|", "Block index: ", blockIndexInt, "[Error]", e)
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

func write2file(filename string, dataString string) (e error) {
	f, e := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if e != nil {
		return
	}
	defer f.Close()

	_, e = fmt.Fprintln(f, dataString)
	if e != nil {
		return
	}
	return
}
