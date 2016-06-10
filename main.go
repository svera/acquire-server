package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/config"
	"github.com/svera/tbg-server/hub"
	//"net/url"
	"fmt"
	"os"
)

var hb *hub.Hub

func newClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	if c, err := client.NewHuman(w, r); err == nil {
		c.SetName(fmt.Sprintf("Player %d", hb.NumberClients()+1))
		hb.Register <- c
		go c.WritePump()
		c.ReadPump(hb.Messages, hb.Unregister)
	} else {
		log.Println(err)
		return
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
		hb = hub.New(cfg)
		go hb.Run()

		r.HandleFunc("/", newClient)
		http.Handle("/", r)
		log.Printf("TBG Server listening on port %s\n", cfg.Port)
		log.Fatal(http.ListenAndServe(cfg.Port, r))
	}
}
