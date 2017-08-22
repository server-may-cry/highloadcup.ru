package main

import (
	"archive/zip"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"sort"
	"fmt"
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

const defaultMapSize = 100000
var users = safeUsers{v: make(map[int]User, defaultMapSize)}
var locations = safeLocations{v: make(map[int]Location, defaultMapSize)}
var visits = safeVisits{v: make(map[int]Visit, defaultMapSize)}

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

type VisitInUser struct {
	Mark      int    `json:"mark"`
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`
}

type VisitsResponse struct {
	Data []*VisitInUser `json:"visits"`
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
				cp := element
				users.v[element.ID] = cp
			}
		case "locations":
			data := LocationsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				cp := element
				locations.v[element.ID] = cp
			}
		case "visits":
			data := VisitsFile{}
			err = decoder.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				cp := element
				visits.v[element.ID] = cp
			}
		}
		rc.Close()
	}

	for _, visit := range visits.v {
		u, ok := users.v[visit.User]
		if !ok {
			log.Fatal("not found user for visit")
		}
		u.Visits = append(u.Visits, &visit)

		l, ok := locations.v[visit.Location]
		if !ok {
			log.Fatal("not found location for visit")
		}
		l.Visits = append(l.Visits, &visit)
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
			http.Error(w, "", http.StatusInternalServerError)
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
			http.Error(w, "", http.StatusInternalServerError)
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
			http.Error(w, "", http.StatusInternalServerError)
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
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		switch "null" {
		case string(blank.BirthDate):
			fallthrough
		case string(blank.FirstName):
			fallthrough
		case string(blank.LastName):
			fallthrough
		case string(blank.Email):
			fallthrough
		case string(blank.Gender):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		gender, err := jsonRawToString(&blank.Gender)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		birthDate, err := jsonRawToInt(&blank.BirthDate)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		email, err := jsonRawToString(&blank.Email)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		firstName, err := jsonRawToString(&blank.FirstName)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		lastName, err := jsonRawToString(&blank.LastName)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if gender != "" {
			obj.Gender = gender
		}
		if birthDate != 0 {
			obj.BirthDate = int64(birthDate)
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

		w.Write(successUpdate)
		users.mux.Lock()
		users.v[id] = obj
		users.mux.Unlock()
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
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		switch "null" {
		case string(blank.Place):
			fallthrough
		case string(blank.Country):
			fallthrough
		case string(blank.City):
			fallthrough
		case string(blank.Distance):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		place, err := jsonRawToString(&blank.Place)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		country, err := jsonRawToString(&blank.Country)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		city, err := jsonRawToString(&blank.City)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		distance, err := jsonRawToInt(&blank.Distance)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
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

		w.Write(successUpdate)
		locations.mux.Lock()
		locations.v[id] = obj
		locations.mux.Unlock()
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
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		switch "null" {
		case string(blank.Location):
			fallthrough
		case string(blank.Mark):
			fallthrough
		case string(blank.VisitedAt):
			fallthrough
		case string(blank.User):
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		location, err := jsonRawToInt(&blank.Location)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		user, err := jsonRawToInt(&blank.User)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		mark, err := jsonRawToInt(&blank.Mark)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		visitedAt, err := jsonRawToInt(&blank.VisitedAt)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
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
		if visitedAt != 0 {
			obj.VisitedAt = visitedAt
		}

		w.Write(successUpdate)
		visits.mux.Lock()
		visits.v[id] = obj
		visits.mux.Unlock()
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := User{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidUser(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		w.Write(successUpdate)
		users.mux.Lock()
		users.v[blank.ID] = blank
		users.mux.Unlock()
	})
	r.Post("/locations/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Location{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidLocation(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		w.Write(successUpdate)
		locations.mux.Lock()
		locations.v[blank.ID] = blank
		locations.mux.Unlock()
	})
	r.Post("/visits/new", func(w http.ResponseWriter, r *http.Request) {
		blank := Visit{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidVisit(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		w.Write(successUpdate)
		visits.mux.Lock()
		visits.v[blank.ID] = blank
		visits.mux.Unlock()
		u, _ := users.v[blank.User]
		u.Visits = append(u.Visits, &blank)
		users.mux.Lock()
		users.v[u.ID] = u
		users.mux.Unlock()
		l, _ := locations.v[blank.Location]
		l.Visits = append(l.Visits, &blank)
		locations.mux.Lock()
		locations.v[l.ID] = l
		locations.mux.Unlock()
	})

	// GET /users/<id>/visits
	r.Get("/users/{id}/visits", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		u, ok := users.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		fromDate := r.URL.Query().Get("fromDate")
		var fromDateInt int
		if fromDate != "" {
			fromDateInt, err = strconv.Atoi(fromDate)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		toDate := r.URL.Query().Get("toDate")
		var toDateInt int
		if toDate != "" {
			toDateInt, err = strconv.Atoi(toDate)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		country := r.URL.Query().Get("country")
		if country != "" && len(country) > 50 {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		toDistance := r.URL.Query().Get("toDistance")
		var toDistanceInt int
		if toDistance != "" {
			toDistanceInt, err = strconv.Atoi(toDistance)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			if toDistanceInt < 0 {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		obj := VisitsResponse{Data: make([]*VisitInUser, 0)}
		for _, v := range u.Visits {
			location, _ := locations.v[v.Location]
			if country != "" && country != location.Country {
				continue
			}
			if toDistance != "" && toDistanceInt <= location.Distance {
				continue
			}
			if fromDate != "" && fromDateInt > v.VisitedAt {
				continue
			}
			if toDate != "" && toDateInt < v.VisitedAt {
				continue
			}
			obj.Data = append(obj.Data, &VisitInUser{
				Mark: v.Mark,
				VisitedAt: v.VisitedAt,
				Place: location.Place,
			})
		}
		sort.Slice(obj.Data, func (i, j int) bool {
			return obj.Data[i].VisitedAt < obj.Data[j].VisitedAt
		})
		err = json.NewEncoder(w).Encode(obj)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
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
		l, ok := locations.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		fromDate := r.URL.Query().Get("fromDate")
		var fromDateInt int
		if fromDate != "" {
			fromDateInt, err = strconv.Atoi(fromDate)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		toDate := r.URL.Query().Get("toDate")
		var toDateInt int
		if toDate != "" {
			toDateInt, err = strconv.Atoi(toDate)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		fromAge := r.URL.Query().Get("fromAge")
		var fromAgeInt int
		if fromAge != "" {
			fromAgeInt, err = strconv.Atoi(fromAge)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		toAge := r.URL.Query().Get("toAge")
		var toAgeInt int
		if toAge != "" {
			toAgeInt, err = strconv.Atoi(toAge)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
		}
		gender := r.URL.Query().Get("gender")
		if gender != "" && gender != "f" && gender != "m" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		var marks []float64
		for _, v := range l.Visits {
			u, _ := users.v[v.User]
			var userAge int
			userAge = int(time.Now().Sub(time.Unix(u.BirthDate, 0)).Hours() / 24 / 365)
			if fromDate != "" && v.VisitedAt < fromDateInt {
				continue
			}
			if toDate != "" && v.VisitedAt > toDateInt {
				continue
			}
			if fromAge != "" && fromAgeInt > userAge {
				continue
			}
			if toAge != "" && toAgeInt < userAge {
				continue
			}
			if gender != "" && u.Gender != gender {
				continue
			}
			marks = append(marks, float64(v.Mark))
		}
		var total float64 = 0
		for _, value:= range marks {
			total += value
		}
		response := fmt.Sprintf(`{"avg":%s}`, strconv.FormatFloat(total / float64(len(marks)), 'f', 5, 64))
		w.Write([]byte(response))
	})

	http.ListenAndServe(":80", r)
}

func isValidUser(u User) bool {
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
	if len(u.Email) > 50 {
		return false
	}
	if u.BirthDate == 0 {
		return false
	}

	return true
}

func isValidLocation(l Location) bool {
	if l.ID == 0 {
		return false
	}
	if l.City == "" {
		return false
	}
	if len(l.City) > 50 {
		return false
	}
	if l.Country == "" {
		return false
	}
	if len(l.Country) > 50 {
		return false
	}
	if l.Place == "" {
		return false
	}
	if l.Distance <= 0 {
		return false
	}

	return true
}

func isValidVisit(v Visit) bool {
	if v.ID == 0 {
		return false
	}
	if v.User == 0 {
		return false
	}
	if v.Mark < 0 || v.Mark > 5 {
		return false
	}
	if v.Location == 0 {
		return false
	}

	return true
}

func jsonRawToString(r *json.RawMessage) (string, error) {
	if len(*r) == 0 {
		return "" , nil
	}
	var result string
	err := json.Unmarshal(*r, &result)
	return result, err
}
func jsonRawToInt(r *json.RawMessage) (int, error) {
	if len(*r) == 0 {
		return 0, nil
	}
	var result int
	err := json.Unmarshal(*r, &result)
	return result, err
}
