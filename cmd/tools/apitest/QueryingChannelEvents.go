package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

/*
You can query for events tied to a specific channel by making a GET request to the event endpoint of its address.
GET /api/<version>/events/channels/<channel_registry_address>
*/
func QueryingChannelEvents(url string, Channel string, Block int64) (Events []interface{}, Status string, err error) {
	var resp *http.Response
	var count int
	var Blocks string
	if Block == 0 {
		Blocks = ""
	} else {
		Blocks = "?from_block=" + strconv.FormatInt(Block, 10)
	}
	for count = 0; count < MaxTry; count = count + 1 {
		resp, err = http.Get(url + "/api/1//events/channels/" + Channel + Blocks)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if count >= MaxTry {
		Status = "504 TimeOut"
	} else {
		Status = resp.Status
	}
	if resp == nil {
		Events = nil
	} else {
		if resp.Status == "200 OK" {
			p, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(p, &Events)
		} else {
			Events = nil
		}
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	return
}

func QueryingChannelEventsTest(url string) {
	start := time.Now()
	ShowTime()
	log.Println("Start Querying Channel Events")
	Channels, Status, err := QueryingNodeAllChannels(url)

	//test the channel which doesn't exist.
	_, Status, err = QueryingChannelEvents(url, "0xffffffffffffffffffffffffffffffffffffffff", 0)
	ShowError(err)
	//display the details of the error
	ShowQueryingChannelEventsMsgDetail(Status)
	switch Status {
	case "404 Not Found":
		log.Println("Test pass:Querying nonexistent Channel")
	default:
		log.Println("Test failed:Querying  nonexistent Channel:", Status)
		if HalfLife {
			log.Fatal("HalfLife,exit")
		}
	}

	for i := 0; i < len(Channels); i++ {
		_, Status, err = QueryingChannelEvents(url, Channels[i].ChannelAddress, 0)
		ShowError(err)
		//display the details of the error
		ShowQueryingChannelEventsMsgDetail(Status)
		switch Status {
		case "200 OK":
			log.Println("Test pass:QueryingChannelEvents:", Channels[i].ChannelAddress)
		default:
			fmt.Printf("Test failed:QueryingChannelEvents:", Channels[i].ChannelAddress, "  ", Status)
			if HalfLife {
				log.Fatal("HalfLife,exit")
			}
		}
	}
	duration := time.Since(start)
	ShowTime()
	log.Println("time used:", duration.Nanoseconds()/1000000, " ms")
}

//display the details of the error
func ShowQueryingChannelEventsMsgDetail(Status string) {
	switch Status {
	case "200 OK":
		log.Println("Successful query")
	case "400 Bad Request":
		log.Println("The provided query string is malformed")
	case "404 Not Found":
		log.Println("The channel does not exist")
	case "500 Server Error":
		log.Println("Internal Raiden node error")
	case "504 TimeOut":
		log.Println("No response,timeout")
	default:
		log.Println("Unknown error,QueryingChannelEvents Failure:", Status)
	}
}
