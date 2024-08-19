package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Reservation struct{}

func (res *Reservation) Create(w http.ResponseWriter, r *http.Request) {
	type newReservation struct {
		User            string
		Shop            string
		ReservationType string
		Date            string
	}

	var resrv newReservation

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&resrv)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println(resrv)
}
