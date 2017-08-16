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
	users = make(map[int]User)
	locations = make(map[int]Location)
	visits = make(map[int]Visit)

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
				users[element.ID] = element
			}
		case "locations":
			data := LocationsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				locations[element.ID] = element
			}
		case "visits":
			data := VisitsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				visits[element.ID] = element
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

		users[id] = blank
		w.Write(successUpdate)
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
		locations[id] = blank
		w.Write(successUpdate)
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
		visits[id] = blank
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
		users[blank.ID] = blank
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
		locations[blank.ID] = blank
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
		visits[blank.ID] = blank
		w.Write(successUpdate)
	})

	// GET /users/<id>/visits
	r.Get("/users/{id}/visits", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		_, ok := users[id]
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
		_, ok := locations[id]
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
