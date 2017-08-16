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
	v   map[int]*User
	mux sync.Mutex
}
type safeLocations struct {
	v   map[int]*Location
	mux sync.Mutex
}
type safeVisits struct {
	v   map[int]*Visit
	mux sync.Mutex
}
var users = safeUsers{v: make(map[int]*User)}
var locations = safeLocations{v: make(map[int]*Location)}
var visits = safeVisits{v: make(map[int]*Visit)}

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
				users.v[element.ID] = &element
			}
		case "locations":
			data := LocationsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				locations.v[element.ID] = &element
			}
		case "visits":
			data := VisitsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				visits.v[element.ID] = &element
			}
		}
		rc.Close()
	}

	for _, visit := range visits.v {
		u, ok := users.v[visit.User]
		if !ok {
			log.Fatal("not found user for visit")
		}
		u.Visits = append(u.Visits, visit)

		l, ok := locations.v[visit.Location]
		if !ok {
			log.Fatal("not found location for visit")
		}
		l.Visits = append(l.Visits, visit)
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
		blank := UserRequest{}
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
		switch "null" {
		case string(blank.BirthDate):
		case string(blank.FirstName):
		case string(blank.LastName):
		case string(blank.Email):
		case string(blank.Gender):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		gender, err := jsonRawToString(&blank.Gender)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		birthDate, err := jsonRawToInt(&blank.BirthDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		email, err := jsonRawToString(&blank.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		firstName, err := jsonRawToString(&blank.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		lastName, err := jsonRawToString(&blank.LastName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if gender != "" {
			obj.Gender = gender
		}
		if birthDate != 0 {
			obj.BirthDate = birthDate
		}
		if email != "" {
			obj.Email = email
		}
		if firstName != "" {
			obj.FirstName = firstName
		}
		if lastName != "" {
			obj.LastName = lastName
		}

		if !isValidUser(obj) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		users.mux.Lock()
		users.v[id] = obj
		users.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/locations/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := LocationRequest{}
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
		switch "null" {
		case string(blank.Place):
		case string(blank.Country):
		case string(blank.City):
		case string(blank.Distance):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		place, err := jsonRawToString(&blank.Place)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		country, err := jsonRawToString(&blank.Country)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		city, err := jsonRawToString(&blank.City)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		distance, err := jsonRawToInt(&blank.Distance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if place != "" {
			obj.Place = place
		}
		if city != "" {
			obj.City = city
		}
		if country != "" {
			obj.Country = country
		}
		if distance != 0 {
			obj.Distance = distance
		}

		if !isValidLocation(obj) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		locations.mux.Lock()
		locations.v[id] = obj
		locations.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/visits/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := VisitRequest{}
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
		switch "null" {
		case string(blank.Location):
		case string(blank.Mark):
		case string(blank.VisitedAt):
		case string(blank.User):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		location, err := jsonRawToInt(&blank.Location)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user, err := jsonRawToInt(&blank.User)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mark, err := jsonRawToInt(&blank.Mark)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		visitedAt, err := jsonRawToInt(&blank.VisitedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if location != 0 {
			obj.Location = location
		}
		if user != 0 {
			obj.User = user
		}
		if mark != 0 {
			obj.Mark = mark
		}
		if visitedAt == 0 {
			obj.VisitedAt = visitedAt
		}

		visits.mux.Lock()
		visits.v[id] = obj
		visits.mux.Unlock()
		w.Write(successUpdate)
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := &User{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !isValidUser(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		users.mux.Lock()
		users.v[blank.ID] = blank
		users.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/locations/new", func(w http.ResponseWriter, r *http.Request) {
		blank := &Location{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !isValidLocation(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		locations.mux.Lock()
		locations.v[blank.ID] = blank
		locations.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/visits/new", func(w http.ResponseWriter, r *http.Request) {
		blank := &Visit{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !isValidVisit(blank) {
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
	if u.FirstName == "" {
		return false
	}
	if u.LastName == "" {
		return false
	}
	if u.Email == "" {
		return false
	}
	if u.BirthDate == 0 {
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
	if l.Distance == 0 {
		return false
	}

	return true
}

func isValidVisit(v *Visit) bool {
	if v.ID == 0 {
		return false
	}
	if v.User == 0 {
		return false
	}
	if v.Mark == 0 {
		return false
	}
	if v.Location == 0 {
		return false
	}
	if v.VisitedAt == 0 {
		return false
	}
	return true
}

func jsonRawToString(r *json.RawMessage) (string, error) {
	if len(*r) == 0 {
		return "" , nil
	}
	var result *string
	err := json.Unmarshal(*r, &result)
	return *result, err
}
func jsonRawToInt(r *json.RawMessage) (int, error) {
	if len(*r) == 0 {
		return 0, nil
	}
	var result *int
	err := json.Unmarshal(*r, &result)
	return *result, err
}
