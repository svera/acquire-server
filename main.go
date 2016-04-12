package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/svera/tbg-server/bridges"
	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/config"
	"github.com/svera/tbg-server/hub"
	//"net/url"
	"fmt"
	"os"
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

	c, err := client.NewHuman(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	hubs[id].Register <- c

	go c.WritePump()
	c.ReadPump(hubs[id].Messages, hubs[id].Unregister)
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

	id := generateID()

	r.ParseForm()
	if bridge, err := bridges.Create(r.FormValue("game")); err != nil {
		panic("Game bridge not found")
	} else {
		h := hub.New(bridge)
		hubs[id] = h

		go hubs[id].Run()

		http.Redirect(w, r, "/"+id, 302)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./public/index.html")
	t.Execute(w, nil)
}

func main() {
	f, err := os.Open("./config.yml")
	if err != nil {
		fmt.Println("Couldn't load configuration file. Check that config.yml exists and that it can be read. Exiting...")
		return
	}
	if cfg, err := config.Load(f); err != nil {
		fmt.Println(err.Error())
	} else {
		r := mux.NewRouter()
		hubs = make(map[string]*hub.Hub)
		r.HandleFunc("/", index)
		r.HandleFunc("/create", create)
		r.HandleFunc("/{id:[a-zA-Z]+}/join", join)
		r.HandleFunc("/{id:[a-zA-Z]+}", room)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
		http.Handle("/", r)
		log.Fatal(http.ListenAndServe(cfg.Port, r))
	}
}

// TODO Implement proper random string generator
func generateID() string {
	return "a"
}
