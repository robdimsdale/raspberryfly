package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio"
)

var currentState bool

func main() {

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	username := flags.String("username", "", "Username for concourse")
	password := flags.String("password", "", "Password for concourse")
	url := flags.String("url", "", "base url for concourse")
	pipeline := flags.String("pipeline", "", "pipeline for concourse")
	job := flags.String("job", "", "job for concourse")
	flags.Parse(os.Args[1:])

	fmt.Println("raspberryfly airbourne")

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pin14 := rpio.Pin(14)
	pin15 := rpio.Pin(15)
	pin18 := rpio.Pin(18)

	// Unmap gpio memory when done
	defer rpio.Close()

	pin14.Input()
	pin14.PullDown()

	pin15.Input()
	pin15.PullDown()

	pin18.Input()
	pin18.PullDown()

	for {
		res14 := pin14.Read()
		res15 := pin15.Read()
		res18 := pin18.Read()
		// fmt.Printf("pin 14: %v, pin 15: %v, pin 18:%v\n", res14, res15, res18)

		state := res14 == 1 && res15 == 1 && res18 == 1
		if shouldTrigger(state) {
			trigger(url, username, password, pipeline, job)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func shouldTrigger(newState bool) bool {
	if currentState {
		currentState = newState
		return false
	}

	currentState = newState
	return newState
}

func trigger(url, username, password, pipeline, job *string) {
	endpoint := *url + "/pipelines/" + *pipeline + "/jobs/" + *job + "/builds"
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}

	req.SetBasicAuth(*username, *password)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("resp: %v\n", resp)
}
