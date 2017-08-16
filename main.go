package main

import (
	"archive/zip"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi"
)

type safeUsers struct {
	v   map[int]User
	mux sync.Mutex
}
type safeLocations struct {
	v   map[int]Location
	mux sync.Mutex
}
type safeVisits struct {
	v   map[int]Visit
	mux sync.Mutex
}
var users = safeUsers{v: make(map[int]User)}
var locations = safeLocations{v: make(map[int]Location)}
var visits = safeVisits{v: make(map[int]Visit)}

var successUpdate = []byte("{}")

type UsersFile struct {
	Data []User `json:"users"`
}
type LocationsFile struct {
	Data []Location `json:"locations"`
}
type VisitsFile struct {
	Data []Visit `json:"visits"`
}

type Avg struct {
	Avg float32 `json:"avg"`
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
				users.v[element.ID] = element
			}
		case "locations":
			data := LocationsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				locations.v[element.ID] = element
			}
		case "visits":
			data := VisitsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				visits.v[element.ID] = element
			}
		}
		rc.Close()
	}
/*
	for _, visit := range visits {
		u, ok := users[visit.User]
		if !ok {
			log.Fatal("not found user for visit")
		}
		u.Visits = append(u.Visits, &visit)

		l, ok := locations[visit.Location]
		if !ok {
			log.Fatal("not found location for visit")
		}
		l.Visits = append(l.Visits, &visit)
	}
*/
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
		obj, ok := users.v[id]
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
		obj, ok := locations.v[id]
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
		obj, ok := visits.v[id]
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
		obj, ok := users.v[id]
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
		if blank.Gender != "m" && blank.Gender != "f" && blank.Gender != "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		} else if blank.Gender == "" {
			blank.Gender = obj.Gender
		}
		if blank.BirthDate == 0 {
			blank.BirthDate = obj.BirthDate
		}
		if blank.Email == "" {
			blank.Email = obj.Email
		}
		if blank.FirstName == "" {
			blank.FirstName = obj.FirstName
		}
		if blank.LastName == "" {
			blank.LastName = obj.LastName
		}

		users.mux.Lock()
		users.v[id] = blank
		users.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := Location{}
		obj, ok := locations.v[id]
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
		if blank.Place == "" {
			blank.Place = obj.Place
		}
		if blank.Country == "" {
			blank.Country = obj.Country
		}
		if blank.City == "" {
			blank.City = obj.City
		}
		if blank.Distance == 0 {
			blank.Distance = obj.Distance
		}
		locations.mux.Lock()
		locations.v[id] = blank
		locations.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/visits/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := Visit{}
		obj, ok := visits.v[id]
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
		if blank.Location == 0 {
			blank.Location = obj.Location
		}
		if blank.User == 0 {
			blank.User = obj.User
		}
		if blank.Mark == 0 {
			blank.Mark = obj.Mark
		}
		if blank.VisitedAt == 0 {
			blank.VisitedAt = obj.VisitedAt
		}
		visits.mux.Lock()
		visits.v[id] = blank
		visits.mux.Unlock()
		w.Write(successUpdate)
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := User{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !isValidUser(&blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		users.mux.Lock()
		users.v[blank.ID] = blank
		users.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/locations/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Location{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !isValidLocation(&blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		locations.mux.Lock()
		locations.v[blank.ID] = blank
		locations.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/visits/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Visit{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !isValidVisit(&blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		visits.mux.Lock()
		visits.v[blank.ID] = blank
		visits.mux.Unlock()
		w.Write(successUpdate)
	})

	// GET /users/<id>/visits
	r.Get("/users/{id}/visits", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		_, ok := users.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		obj := VisitsFile{}
		// TODO
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// GET /locations/<id>/avg
	r.Get("/locations/{id}/avg", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		_, ok := locations.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		obj := Avg{}
		// TODO
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":80", r)
}

func isValidUser(u *User) bool {
	if u.ID == 0 {
		return false
	}
	if u.Gender != "m" && u.Gender != "f" {
		return false
	}

	return true
}

func isValidLocation(l *Location) bool {
	if l.ID == 0 {
		return false
	}
	if l.City == "" {
		return false
	}
	if l.Country == "" {
		return false
	}
	if l.Place == "" {
		return false
	}

	return true
}

func isValidVisit(v *Visit) bool {
	if v.ID == 0 {
		return false
	}
	return true
}
