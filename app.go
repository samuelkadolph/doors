package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/samuelkadolph/go/phidgets"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Config struct {
	Doors           []*Door
	FeedbackTimeout int
	Secret          string
}

type Door struct {
	ID           string
	Lock         *int
	LockFeedback *int
	Mag          *int
	MagFeedback  *int
	Name         string

	lockCond  *sync.Cond
	lockMutex *sync.Mutex
	magCond   *sync.Cond
	magMutex  *sync.Mutex
}

type hash map[string]interface{}

var config *Config
var ifk *phidgets.InterfaceKit

var host = flag.String("host", "", "Host for the server to listen on")
var configPath = flag.String("config", "./config.json", "Path to config file")
var port = flag.Int("port", 4567, "Port for the server to listen on")

func DoorIndex(w http.ResponseWriter, r *http.Request) {
	response(w, 200, config.Doors)
}

func DoorMagDisengage(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}
	if !checkDoorMag(w, r, d) {
		return
	}

	if err := d.MagDisengage(); err != nil {
		response(w, 200, hash{"success": false, "error": err})
	} else {
		response(w, 200, hash{"success": true})
	}
}

func DoorMagEngage(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}
	if !checkDoorMag(w, r, d) {
		return
	}

	if err := d.MagEngage(); err != nil {
		response(w, 200, hash{"success": false, "error": err})
	} else {
		response(w, 200, hash{"success": true})
	}
}

func DoorUnlock(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}
	if d.Lock == nil {
		response(w, 422, hash{"error": "door does not support opening"})
		return
	}

	ch, err := d.Unlock()

	if err == nil {
		err = <-ch
	}

	if err != nil {
		response(w, 200, hash{"success": false, "error": err})
	} else {
		response(w, 200, hash{"success": true})
	}
}

func DoorShow(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}

	response(w, 200, d)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	response(w, 404, hash{"error": "not found"})
}

func Root(w http.ResponseWriter, r *http.Request) {
	response(w, 200, hash{"hi": true})
}

func (d *Door) LockStatus() string {
	var err error
	var s bool

	if d.LockFeedback != nil {
		s, err = ifk.Inputs[*d.LockFeedback].State()
	} else if d.Lock != nil {
		s, err = ifk.Outputs[*d.Lock].State()
	} else {
		return "unsupported"
	}

	if err != nil {
		return "error"
	} else if s {
		return "unlocked"
	}

	return "locked"
}

func (d *Door) MagDisengage() error {
	return ifk.Outputs[*d.Mag].SetState(false)
}

func (d *Door) MagEngage() error {
	return ifk.Outputs[*d.Mag].SetState(true)
}

func (d *Door) MagStatus() string {
	var err error
	var s bool

	if d.MagFeedback != nil {
		s, err = ifk.Inputs[*d.MagFeedback].State()
	} else if d.Mag != nil {
		s, err = ifk.Outputs[*d.Mag].State()
	} else {
		return "unsupported"
	}

	if err != nil {
		return "error"
	} else if s {
		return "engaged"
	}

	return "disengaged"
}

func (d *Door) MarshalJSON() ([]byte, error) {
	o := make(map[string]interface{})

	o["id"] = d.ID
	o["lock"] = d.LockStatus()
	o["mag"] = d.MagStatus()
	o["name"] = d.Name

	return json.Marshal(o)
}

func (d *Door) Unlock() (<-chan error, error) {
	var err error

	ch := make(chan error, 1)

	if err = ifk.Outputs[*d.Lock].SetState(true); err != nil {
		return nil, err
	}

	time.Sleep(200 * time.Millisecond)

	if err = ifk.Outputs[*d.Lock].SetState(false); err != nil {
		return nil, err
	}

	ch <- nil

	return ch, nil
}

func checkDoor(w http.ResponseWriter, r *http.Request, o **Door) bool {
	v := mux.Vars(r)
	d := findDoor(v["door"])

	if d == nil {
		response(w, 404, hash{"error": "door not found"})
		return false
	}

	*o = d
	return true
}

func checkDoorMag(w http.ResponseWriter, r *http.Request, d *Door) bool {
	if d.Mag == nil {
		response(w, 422, hash{"error": "door does not support mag"})
		return false
	}

	return true
}

func checkSecret(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()

	if config.Secret != r.Form.Get("secret") {
		response(w, 403, hash{"error": "bad secret"})
		return false
	}

	return true
}

func findDoor(id string) *Door {
	for _, d := range config.Doors {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func response(w http.ResponseWriter, s int, b interface{}) {
	o, _ := json.Marshal(b)

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(o)))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(s)
	w.Write(o)
}

func loadConfig(path string) (*Config, error) {
	var config Config
	var err error
	var file []byte

	if file, err = ioutil.ReadFile(path); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	if config.Doors == nil {
		config.Doors = make([]*Door, 0)
	}

	return &config, nil
}

func loadInterfaceKit() (*phidgets.InterfaceKit, error) {
	var err error
	var ifk *phidgets.InterfaceKit

	if ifk, err = phidgets.NewInterfaceKit(); err != nil {
		return nil, err
	}

	if err = ifk.Open(phidgets.Any); err != nil {
		return nil, err
	}

	// if err = ifk.WaitForAttachment(2 * time.Second); err != nil {
	// 	return nil, err
	// }

	return ifk, nil
}

func main() {
	var err error

	flag.Parse()

	if config, err = loadConfig(*configPath); err != nil {
		log.Fatalf("Error while loading config - %s", err)
	}

	if ifk, err = loadInterfaceKit(); err != nil {
		log.Fatalf("Error while loading interface kit - %s", err)
	}

	r := mux.NewRouter()
	r.NewRoute().Methods("POST").Path("/doors/{door}/mag/disengage").Handler(http.HandlerFunc(DoorMagDisengage))
	r.NewRoute().Methods("POST").Path("/doors/{door}/mag/engage").Handler(http.HandlerFunc(DoorMagEngage))
	r.NewRoute().Methods("POST").Path("/doors/{door}/open").Handler(http.HandlerFunc(DoorUnlock))
	r.NewRoute().Methods("POST").Path("/doors/{door}/unlock").Handler(http.HandlerFunc(DoorUnlock))
	r.NewRoute().Methods("GET").Path("/doors/{door}").Handler(http.HandlerFunc(DoorShow))
	r.NewRoute().Methods("GET").Path("/doors").Handler(http.HandlerFunc(DoorIndex))
	r.NewRoute().Methods("GET").Path("/").Handler(http.HandlerFunc(Root))
	r.NewRoute().Handler(http.HandlerFunc(NotFound))

	a := fmt.Sprintf("%s:%d", *host, *port)
	h := handlers.CombinedLoggingHandler(os.Stdout, r)

	log.Fatalf("Error while starting server %s", http.ListenAndServe(a, h))
}
