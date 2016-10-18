package main

import (
	"log"
	"net/http"

	"fmt"
	"os"

	"github.com/gorilla/mux"
	"github.com/olebedev/emitter"
	"github.com/svera/sackson-server/client"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/hub"
)

var hb *hub.Hub
var gitHash = "No git hash provided"

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
		e := &emitter.Emitter{}
		e.Use("*", emitter.Skip)
		hb = hub.New(cfg, e, isDebugEnabled())
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

func isDebugEnabled() bool {
	debug := false
	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "debug" {
			debug = true
		}
	}
	return debug
}
