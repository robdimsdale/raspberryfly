package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/stianeikeland/go-rpio"
)

var (
	currentState bool

	p props

	decoder *schema.Decoder
)

type props struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
	URL      string `schema:"url"`
	Pipeline string `schema:"pipeline"`
	Job      string `schema:"job"`
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	usernameFlag := flags.String("username", "", "Username for concourse")
	passwordFlag := flags.String("password", "", "Password for concourse")
	urlFlag := flags.String("url", "", "base url for concourse")
	pipelineFlag := flags.String("pipeline", "", "pipeline for concourse")
	jobFlag := flags.String("job", "", "job for concourse")
	flags.Parse(os.Args[1:])

	p = props{
		Username: *usernameFlag,
		Password: *passwordFlag,
		URL:      *urlFlag,
		Pipeline: *pipelineFlag,
		Job:      *jobFlag,
	}

	fmt.Println("raspberryfly airbourne")

	decoder = schema.NewDecoder()

	r := mux.NewRouter()

	r.HandleFunc("/", getHomeHandler).Methods("GET")

	r.HandleFunc("/", postHomeHandler).Methods("POST")

	http.Handle("/", r)

	fmt.Println("Listening...")
	// http.ListenAndServe(":3000", nil)
	go http.ListenAndServe(":3000", nil)

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
			trigger(p)
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

func trigger(p props) {
	endpoint := p.URL + "/pipelines/" + p.Pipeline + "/jobs/" + p.Job + "/builds"
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}

	req.SetBasicAuth(p.Username, p.Password)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("resp: %v\n", resp)
}

func getHomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Here")
	t, _ := template.ParseFiles("web/index.html")
	t.Execute(w, p)
}

func postHomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POSTED")
	//parse input
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(&p, r.PostForm)
	if err != nil {
		panic(err)
	}

	fmt.Printf("props %#v\n", p)

	http.Redirect(w, r, "/", 302)
}
