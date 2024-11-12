// main.go
package main

import (
	"encoding/json"
	"example.com/vmwareHostsInfo/cmd/retrieve"
	"log"
	"net/http"
)

func handleGetData(w http.ResponseWriter, r *http.Request) {
	data, err := retrieve.GetData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the results to JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	http.HandleFunc("/data", handleGetData)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
