package main

import (
	"github.com/gorilla/mux"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire-server/hub"
	"html/template"
	"log"
	"net/http"
)

var hubs map[string]*hub.Hub

func join(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if _, ok := hubs[id]; !ok {
		http.Error(w, "Game doesn't exist", 404)
		return
	}

	c, err := client.New(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	hubs[id].Register <- c

	go c.WritePump()
	c.ReadPump(hubs[id].Broadcast, hubs[id].Unregister)
}

func create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	}

	id := generateId()
	h := hub.New()
	hubs[id] = h

	go hubs[id].Run()
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./public/index.html")
	t.Execute(w, nil)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/{id:[a-zA-Z]+}/join", join)
	r.HandleFunc("/create", create)

	log.Fatal(http.ListenAndServe(":8000", r))
}

// TODO Implement proper random string generator
func generateId() string {
	return "a"
}
