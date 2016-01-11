package main

import (
	"github.com/gorilla/mux"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire-server/hub"
	"github.com/svera/acquire/player"
	"html/template"
	"log"
	"net/http"
	//"net/url"
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

	newPlayer := player.New(vars["playerName"])

	c, err := client.New(w, r, newPlayer)
	if err != nil {
		log.Println(err)
		return
	}

	hubs[id].Register <- c

	go c.WritePump()
	c.ReadPump(hubs[id].Broadcast, hubs[id].Unregister)
}

func room(w http.ResponseWriter, r *http.Request) {
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

	t, _ := template.ParseFiles("./public/game.html")
	t.Execute(w, nil)
}

func create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	}

	id := generateId()
	h := hub.New()
	hubs[id] = h

	go hubs[id].Run()

	http.Redirect(w, r, "/"+id+"?playerName="+r.FormValue("playerName"), 302)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./public/index.html")
	t.Execute(w, nil)
}

func main() {
	r := mux.NewRouter()
	hubs = make(map[string]*hub.Hub)
	r.HandleFunc("/", index)
	r.HandleFunc("/create", create)
	r.HandleFunc("/{id:[a-zA-Z]+}/join", join)
	r.HandleFunc("/{id:[a-zA-Z]+}", room)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8000", r))
}

// TODO Implement proper random string generator
func generateId() string {
	return "a"
}
