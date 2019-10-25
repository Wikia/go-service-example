package handlers

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Text string
}

func helloWorldJSON() string {
	m := Message{"Hello World"}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err) // no, not really
	}

	return string(b)
}
func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	m := Message{"Hello World"}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err) // no, not really
	}
	w.Write(b)
}
