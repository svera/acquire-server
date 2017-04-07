package main

import (
	"log"
	"net/http"

	"fmt"
	"os"

	"github.com/gorilla/mux"
	"github.com/svera/sackson-server/client"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/hub"
	"github.com/svera/sackson-server/observer"
)

var (
	hb      *hub.Hub
	cfg     *config.Config
	gitHash = "No git hash provided"
)

func main() {
	f, err := os.Open("./sackson.yml")
	if err != nil {
		fmt.Println("Couldn't load configuration file. Check that sackson.yml exists and that it can be read. Exiting...")
		return
	}
	if cfg, err = config.Load(f); err != nil {
		fmt.Println(err.Error())
	} else {
		r := mux.NewRouter()
		obs := observer.New()
		hb = hub.New(cfg, obs)
		go hb.Run()

		r.HandleFunc("/", newClient)
		http.Handle("/", r)
		fmt.Printf("Sackson server listening on port %s\n", cfg.Port)
		fmt.Printf("Git commit hash: %s\n", gitHash)
		log.Fatal(http.ListenAndServe(cfg.Port, r))
	}
}

func newClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	if c, err := client.NewHuman(w, r, cfg); err == nil {
		hb.Register <- c
		go c.WritePump()
		c.ReadPump(hb.Messages, hb.Unregister)
	} else {
		log.Println(err)
		return
	}
}
