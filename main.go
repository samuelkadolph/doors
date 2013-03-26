package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/samuelkadolph/go/phidgets"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Doors  []*Door
	Secret string
}

type hash map[string]interface{}

var config Config
var ifk *phidgets.InterfaceKit

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

	if err := d.MagDisengage(ifk); err != nil {
		response(w, 200, hash{}, hash{"success": false, "error": err.Error()})
	} else {
		response(w, 200, hash{}, hash{"success": true})
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

	if err := d.MagEngage(ifk); err != nil {
		response(w, 200, hash{}, hash{"success": false, "error": err.Error()})
	} else {
		response(w, 200, hash{}, hash{"success": true})
	}
}

func DoorOpen(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}

	if err := d.Open(ifk); err != nil {
		response(w, 200, hash{}, hash{"success": false, "error": err.Error()})
	} else {
		response(w, 200, hash{}, hash{"success": true})
	}
}

func ShowDoor(w http.ResponseWriter, r *http.Request) {
	var d *Door

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}

	response(w, 200, hash{}, d)
}

func Root(w http.ResponseWriter, r *http.Request) {
	response(w, 200, hash{}, hash{"hi": true})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	response(w, 404, hash{}, hash{"error": "not found"})
}

func checkDoor(w http.ResponseWriter, r *http.Request, o **Door) bool {
	v := mux.Vars(r)
	d := findDoor(v["door"])

	if d == nil {
		response(w, 404, hash{}, hash{"error": "door not found"})
		return false
	}
	if d.Lock == nil {
		response(w, 422, hash{}, hash{"error": "door does not support opening"})
		return false
	}

	*o = d
	return true
}

func checkDoorMag(w http.ResponseWriter, r *http.Request, d *Door) bool {
	if d.Mag == nil {
		response(w, 422, hash{}, hash{"error": "door does not support mag"})
		return false
	}

	return true
}

func checkSecret(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()

	if config.Secret != r.Form.Get("secret") {
		response(w, 403, hash{}, hash{"error": "bad secret"})
		return false
	}

	return true
}

func findDoor(name string) *Door {
	for _, d := range config.Doors {
		if d.Name == name {
			return d
		}
	}
	return nil
}

func response(w http.ResponseWriter, s int, h hash, b interface{}) {
	o, _ := json.Marshal(b)
	h["Content-Length"] = len(o)
	h["Content-Type"] = "application/json"
	for k, v := range h {
		w.Header().Set(k, fmt.Sprintf("%v", v))
	}
	w.WriteHeader(s)
	w.Write([]byte(o))
}

func init() {
	var err error
	var file []byte

	if file, err = ioutil.ReadFile("./config.json"); err != nil {
		log.Fatalf("Error while opening config.json - %s", err)
	}

	if err = json.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error while unmarshaling config - %s", err)
	}

	if ifk, err = phidgets.NewInterfaceKit(); err != nil {
		log.Fatalf("%s", err)
	}

	if err = ifk.Open(phidgets.Any); err != nil {
		log.Fatalf("%s", err)
	}

	if err = ifk.WaitForAttachment(2 * time.Second); err != nil {
		log.Fatalf("%s", err)
	}
}

func main() {
	r := mux.NewRouter()
	r.NewRoute().Methods("POST").Path("/doors/{door}/mag/disengage").Handler(http.HandlerFunc(DoorMagDisengage))
	r.NewRoute().Methods("POST").Path("/doors/{door}/mag/engage").Handler(http.HandlerFunc(DoorMagEngage))
	r.NewRoute().Methods("POST").Path("/doors/{door}/open").Handler(http.HandlerFunc(DoorOpen))
	r.NewRoute().Methods("GET").Path("/doors/{door}").Handler(http.HandlerFunc(ShowDoor))
	r.NewRoute().Methods("GET").Path("/").Handler(http.HandlerFunc(Root))
	r.NewRoute().Handler(http.HandlerFunc(NotFound))

	h := handlers.CombinedLoggingHandler(os.Stdout, r)

	log.Fatalf("Error while starting server %s", http.ListenAndServe(":4567", h))
}
