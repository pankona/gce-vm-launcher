package main

import (
	"log"
	"net/http"

	"github.com/pankona/gce-vm-launcher/cmd/cloudfuncs"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cloudfuncs.Launch(w, r)
	})
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
