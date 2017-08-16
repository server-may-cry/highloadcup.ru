package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

var users map[int]User
var locations map[int]Location
var visits map[int]Visit

var usersMaxID, locationsMaxId, visitsMaxId int

func main() {
	r := chi.NewRouter()

	// GET /<entity>/<id>
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		obj, ok := users[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Get("/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		obj, ok := locations[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Get("/visits/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		obj, ok := visits[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// POST /<entity>/<id>
	r.Post("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := User{}
		obj, ok := users[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = obj.ID
		users[id] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Post("/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := Location{}
		obj, ok := locations[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = obj.ID
		locations[id] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Post("/visits/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := Visit{}
		obj, ok := visits[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		err = json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = obj.ID
		visits[id] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := User{}
		usersMaxID++
		newID := usersMaxID
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = newID
		users[newID] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Post("/locations/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Location{}
		locationsMaxId++
		newID := locationsMaxId
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = newID
		locations[newID] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Post("/visits/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Visit{}
		visitsMaxId++
		newID := visitsMaxId
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blank.ID = newID
		visits[newID] = blank
		err = json.NewEncoder(w).Encode(blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
