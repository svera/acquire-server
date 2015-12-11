package main

import (
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire-server/hub"
	"log"
	"net/http"
)

var h = hub.New()

func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	c, err := client.New(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	h.Register <- c

	go c.WritePump()
	c.ReadPump(h.Broadcast, h.Unregister)
}

func main() {
	go h.Run()
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", serveWs)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
