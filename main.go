package main

import (
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

	if c, err := client.NewHuman(w, r); err == nil {
		c.SetName(fmt.Sprintf("Player %d", hubs[id].NumberClients()+1))
		hubs[id].Register <- c
		go c.WritePump()
		c.ReadPump(hubs[id].Messages, hubs[id].Unregister)
	} else {
		log.Println(err)
		return
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	}

	id := generateID()

	r.ParseForm()
	if bridge, err := bridges.Create(r.FormValue("game")); err != nil {
		http.Error(w, "Game bridge not found", 404)
	} else {
		hubs[id] = hub.New(bridge, func() { delete(hubs, id); fmt.Printf("Number of running games: %d\n", len(hubs)) })
		fmt.Printf("Number of running games: %d\n", len(hubs))

		go hubs[id].Run()
		fmt.Fprint(w, id)
	}
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
		r.HandleFunc("/create", create)
		r.HandleFunc("/{id:[a-zA-Z]+}/join", join)
		http.Handle("/", r)
		log.Printf("TBG Server listening on port %s\n", cfg.Port)
		log.Fatal(http.ListenAndServe(cfg.Port, r))
	}
}

// TODO Implement proper random string generator
func generateID() string {
	return "a"
}
