package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	username := flags.String("username", "", "Username for concourse")
	password := flags.String("password", "", "Password for concourse")
	url := flags.String("url", "", "base url for concourse")
	pipeline := flags.String("pipeline", "", "pipeline for concourse")
	job := flags.String("job", "", "job for concourse")
	flags.Parse(os.Args[1:])

	fmt.Println("raspberryfly airbourne")

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
