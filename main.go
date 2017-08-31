package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/server-may-cry/highloadcup.ru/dto"
	"github.com/server-may-cry/highloadcup.ru/helpers"
)

// "github.com/go-chi/chi/middleware"

var users = struct {
	v   map[int]dto.User
	mux sync.Mutex
}{
	v: make(map[int]dto.User),
}
var locations = struct {
	v   map[int]dto.Location
	mux sync.Mutex
}{
	v: make(map[int]dto.Location),
}
var visits = struct {
	v   map[int]dto.Visit
	mux sync.Mutex
}{
	v: make(map[int]dto.Visit),
}

var successUpdate = []byte("{}")
var timeStampStart time.Time

func init() {
	log.SetFlags(log.LstdFlags)
	timeStampStart = time.Now()

	r, err := zip.OpenReader("/tmp/data/data.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		parts := strings.Split(f.Name, "_")
		reader, err := f.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		content, err := ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			log.Println(err)
			continue
		}
		switch parts[0] {
		case "users":
			data := dto.UsersFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, element := range data.Data {
				cp := element
				cp.Visits = make(map[int]struct{})
				users.v[element.ID] = cp
			}
		case "locations":
			data := dto.LocationsFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, element := range data.Data {
				cp := element
				cp.Visits = make(map[int]struct{})
				locations.v[element.ID] = cp
			}
		case "visits":
			data := dto.VisitsFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, element := range data.Data {
				cp := element
				visits.v[element.ID] = cp
			}
		}
	}

	for _, visit := range visits.v {
		cpv := visit
		u, ok := users.v[visit.User]
		if !ok {
			log.Printf("not found user:%d for visit:%d\n", visit.User, visit.ID)
			continue
		}
		u.Visits[cpv.ID] = struct{}{}
		users.v[u.ID] = u

		l, ok := locations.v[visit.Location]
		if !ok {
			log.Printf("not found location:%d for visit:%d\n", visit.Location, visit.ID)
			continue
		}
		l.Visits[cpv.ID] = struct{}{}
		locations.v[l.ID] = l
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
		array_bytes, _ := obj.MarshalJSON()
		w.Write(array_bytes)
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
		array_bytes, _ := obj.MarshalJSON()
		w.Write(array_bytes)
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
		array_bytes, _ := obj.MarshalJSON()
		w.Write(array_bytes)
	})

	// POST /<entity>/<id>
	r.Post("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "", http.StatusNotFound) // StatusBadRequest
			return
		}
		blank := dto.UserRequest{}
		_, ok := users.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		gender := blank.Gender
		birthDate := blank.BirthDate
		email := blank.Email
		firstName := blank.FirstName
		lastName := blank.LastName

		users.mux.Lock()
		obj, _ := users.v[id]
		if gender != "" {
			obj.Gender = gender
		}
		if birthDate != 0 || string(blank.BirthDate) == "0" {
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
		blank := dto.LocationRequest{}
		_, ok := locations.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		place := blank.Place
		country := blank.Country
		city := blank.City
		distance := blank.Distance

		locations.mux.Lock()
		obj, _ := locations.v[id]
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
		blank := dto.VisitRequest{}
		_, ok := visits.v[id]
		if !ok {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		location := blank.Location
		user := blank.User
		mark := blank.Mark
		visitedAt := blank.VisitedAt

		visits.mux.Lock()
		obj, _ := visits.v[id]
		if location != 0 {
			if obj.Location != location {
				locations.mux.Lock()
				l, _ := locations.v[obj.Location]
				delete(l.Visits, obj.ID)
				locations.v[obj.Location] = l

				nl, _ := locations.v[location]
				nl.Visits[obj.ID] = struct{}{}
				locations.mux.Unlock()
			}
			obj.Location = location
		}
		if user != 0 {
			if obj.User != user {
				users.mux.Lock()
				u, _ := users.v[obj.User]
				delete(u.Visits, obj.ID)
				users.v[obj.User] = u

				nu, _ := users.v[user]
				nu.Visits[obj.ID] = struct{}{}
				users.mux.Unlock()
			}
			obj.User = user
		}
		if mark != 0 {
			obj.Mark = mark
		}
		if visitedAt != 0 {
			obj.VisitedAt = visitedAt
		}

		visits.v[id] = obj
		visits.mux.Unlock()
		w.Write(successUpdate)
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := dto.User{}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
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
		blank := dto.Location{}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
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
		blank := dto.Visit{}
		array_bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		json_string := string(array_bytes)
		if strings.Contains(json_string, "null") {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = blank.UnmarshalJSON(array_bytes)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidVisit(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		visits.mux.Lock()
		visits.v[blank.ID] = blank
		visits.mux.Unlock()

		users.mux.Lock()
		u, ok := users.v[blank.User]
		if !ok {
			http.Error(w, "", http.StatusBadRequest)
			users.mux.Unlock()
			return
		}
		u.Visits[blank.ID] = struct{}{}
		users.v[blank.User] = u
		users.mux.Unlock()

		locations.mux.Lock()
		l, ok := locations.v[blank.Location]
		if !ok {
			http.Error(w, "", http.StatusBadRequest)
			locations.mux.Unlock()
			return
		}
		l.Visits[blank.ID] = struct{}{}
		locations.v[blank.Location] = l
		locations.mux.Unlock()
		w.Write(successUpdate)
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
		obj := dto.VisitsResponse{Data: make([]dto.VisitInUser, 0)}
		for vID := range u.Visits {
			v, _ := visits.v[vID]
			location, ok := locations.v[v.Location]
			if !ok {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
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
			obj.Data = append(obj.Data, dto.VisitInUser{
				Mark:      v.Mark,
				VisitedAt: v.VisitedAt,
				Place:     location.Place,
			})
		}
		sort.Slice(obj.Data, func(i, j int) bool {
			return obj.Data[i].VisitedAt < obj.Data[j].VisitedAt
		})
		array_bytes, _ := obj.MarshalJSON()
		w.Header().Set("Content-Length", strconv.Itoa(len(array_bytes)))
		w.Write(array_bytes)
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
		for vID := range l.Visits {
			v, _ := visits.v[vID]
			u, ok := users.v[v.User]
			if !ok {
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			if fromDate != "" && v.VisitedAt < fromDateInt {
				continue
			}
			if toDate != "" && v.VisitedAt > toDateInt {
				continue
			}
			age := helpers.AgePass(time.Unix(u.BirthDate, 0), timeStampStart)
			if fromAge != "" && fromAgeInt > age {
				continue
			}
			if toAge != "" && toAgeInt <= age {
				continue
			}
			if gender != "" && u.Gender != gender {
				continue
			}
			marks = append(marks, float64(v.Mark))
		}
		var total float64
		for _, value := range marks {
			total += value
		}
		if len(marks) > 0 {
			total /= float64(len(marks))
		}
		response := fmt.Sprintf(`{"avg":%s}`, strconv.FormatFloat(total, 'f', 5, 64))
		w.Write([]byte(response))
	})

	// r.Mount("/debug", middleware.Profiler())

	fmt.Println("Ready")
	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal(err)
	}
}

func isValidUser(u dto.User) bool {
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

	return true
}

func isValidLocation(l dto.Location) bool {
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

func isValidVisit(v dto.Visit) bool {
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
