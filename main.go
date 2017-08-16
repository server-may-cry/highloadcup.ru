package main

import (
	"archive/zip"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

var users map[int]User
var locations map[int]Location
var visits map[int]Visit

var usersMaxID, locationsMaxID, visitsMaxID int

type UsersFile struct {
	Data []User `json:"users"`
}
type LocationsFile struct {
	Data []Location `json:"locations"`
}
type VisitsFile struct {
	Data []Visit `json:"visits"`
}

func init() {
	r, err := zip.OpenReader("/tmp/data/data.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		parts := strings.Split(f.Name, "_")
		reader, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		decoder := json.NewDecoder(reader)
		switch parts[0] {
		case "users":
			data := UsersFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				users[data.ID] = data
				if data.ID > usersMaxID {
					usersMaxID = data.ID
				}
			}
		case "locations":
			data := LocationsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				locations[data.ID] = data
				if data.ID > locationsMaxID {
					locationsMaxID = data.ID
				}
			}
		case "visits":
			data := VisitsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				visits[data.ID] = data
				if data.ID > visitsMaxID {
					visitsMaxID = data.ID
				}
			}
		}
	}
}

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
		locationsMaxID++
		newID := locationsMaxID
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
		visitsMaxID++
		newID := visitsMaxID
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

	http.ListenAndServe(":80", r)
}
