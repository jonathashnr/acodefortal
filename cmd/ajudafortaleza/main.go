package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Help Stronghold aka Ajuda Fortaleza :D")
	})
	addr := ":8080"
	fmt.Println("Servidor escutando em http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, mux))
}