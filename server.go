package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"os"
)

type Server struct {
	config *Config
	http   *http.Server
	router *mux.Router
}

type hash map[string]interface{}

func NewServer(c *Config) *Server {
	s := &Server{}

	s.config = c
	s.http = &http.Server{Handler: handlers.CombinedLoggingHandler(os.Stdout, s)}
	s.router = mux.NewRouter()

	s.router.NewRoute().Methods("POST").Path("/doors/{id}/mag/disengage").Handler(http.HandlerFunc(disengageDoorMag))
	s.router.NewRoute().Methods("POST").Path("/doors/{id}/mag/engage").Handler(http.HandlerFunc(engageDoorMag))
	s.router.NewRoute().Methods("POST").Path("/doors/{id}/open").Handler(http.HandlerFunc(unlockDoor))
	s.router.NewRoute().Methods("POST").Path("/doors/{id}/unlock").Handler(http.HandlerFunc(unlockDoor))
	s.router.NewRoute().Methods("GET").Path("/doors/{id}").Handler(http.HandlerFunc(showDoor))
	s.router.NewRoute().Methods("GET").Path("/doors").Handler(http.HandlerFunc(showDoors))
	s.router.NewRoute().Methods("GET").Path("/").Handler(http.HandlerFunc(root))
	s.router.NewRoute().Handler(http.HandlerFunc(notFound))

	return s
}

func (s *Server) Listen(addr string) (net.Listener, error) {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}

	go s.http.Serve(l)

	return l, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	context.Set(r, "server", s)
	s.router.ServeHTTP(w, r)
	context.Clear(r)
}

func disengageDoorMag(w http.ResponseWriter, r *http.Request) {
	var d *HTTPDoor

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

func engageDoorMag(w http.ResponseWriter, r *http.Request) {
	var d *HTTPDoor

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

func unlockDoor(w http.ResponseWriter, r *http.Request) {
	var d *HTTPDoor

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}
	if !checkDoorLock(w, r, d) {
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

func showDoor(w http.ResponseWriter, r *http.Request) {
	var d *HTTPDoor

	if !checkSecret(w, r) {
		return
	}
	if !checkDoor(w, r, &d) {
		return
	}

	response(w, 200, d)
}

func showDoors(w http.ResponseWriter, r *http.Request) {
	if !checkSecret(w, r) {
		return
	}

	s := context.Get(r, "server").(*Server)
	d := s.config.Doors()
	h := make([]*HTTPDoor, len(d))

	for i, o := range d {
		h[i] = &HTTPDoor{Door: *o}
	}

	response(w, 200, h)
}

func root(w http.ResponseWriter, r *http.Request) {
	response(w, 200, hash{"hi": true})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	response(w, 404, hash{"error": "not found"})
}

func checkDoor(w http.ResponseWriter, r *http.Request, o **HTTPDoor) bool {
	s := context.Get(r, "server").(*Server)
	v := mux.Vars(r)
	d := findDoor(s, v["id"])

	if d == nil {
		response(w, 404, hash{"error": "door not found"})
		return false
	}

	*o = &HTTPDoor{*d}
	return true
}

func checkDoorMag(w http.ResponseWriter, r *http.Request, d *HTTPDoor) bool {
	if d.Mag == nil {
		response(w, 422, hash{"error": "door does not support mag"})
		return false
	}

	return true
}

func checkDoorLock(w http.ResponseWriter, r *http.Request, d *HTTPDoor) bool {
	if d.Lock == nil {
		response(w, 422, hash{"error": "door does not support lock"})
		return false
	}

	return true
}

func checkSecret(w http.ResponseWriter, r *http.Request) bool {
	s := context.Get(r, "server").(*Server)

	if s.config.Secret != r.Form.Get("secret") {
		response(w, 403, hash{"error": "bad secret"})
		return false
	}

	return true
}

func findDoor(s *Server, id string) *Door {
	for _, d := range s.config.Doors() {
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
