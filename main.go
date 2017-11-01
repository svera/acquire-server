package main

import (
	"log"
	"net/http"

	"github.com/svera/sackson-server/drivers"

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
		drivers.Load()
		go hb.Run()

		r.HandleFunc("/", newClient)
		fmt.Printf("Sackson server listening on port %s\n", cfg.Port)
		fmt.Printf("Git commit hash: %s\n", gitHash)

		if cfg.Secure {
			log.Fatal(http.ListenAndServeTLS(cfg.Port, cfg.SecureCertFileName, cfg.SecureKeyFileName, r))
		} else {
			log.Fatal(http.ListenAndServe(cfg.Port, r))
		}
	}
}

func newClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameDriverName := r.FormValue("g")
	if !drivers.Exist(gameDriverName) {
		if cfg.Debug {
			log.Printf("Tried connection to non-existent game driver: %s\n", gameDriverName)
		}
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	if c, err := client.NewHuman(w, r, cfg); err == nil {
		c.SetGame(gameDriverName)
		hb.Register <- c
		go c.WritePump()
		c.ReadPump(hb.Messages, hb.Unregister)
	} else {
		if cfg.Debug {
			log.Println(err.Error())
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
