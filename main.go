package main

import (
	"archive/zip"
	"encoding/json"
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
	"github.com/go-chi/chi/middleware"
	"github.com/server-may-cry/highloadcup.ru/dto"
)

type safeUsers struct {
	v   map[int]dto.User
	mux sync.Mutex
}
type safeLocations struct {
	v   map[int]dto.Location
	mux sync.Mutex
}
type safeVisits struct {
	v   map[int]*dto.Visit
	mux sync.Mutex
}

const debug = false

const defaultMapSize = 100000

var users = safeUsers{v: make(map[int]dto.User, defaultMapSize)}
var locations = safeLocations{v: make(map[int]dto.Location, defaultMapSize)}
var visits = safeVisits{v: make(map[int]*dto.Visit, defaultMapSize)}

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
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		parts := strings.Split(f.Name, "_")
		reader, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		}
		switch parts[0] {
		case "users":
			data := dto.UsersFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				cp := element
				cp.Visits = make(map[int]*dto.Visit)
				cp.Age = ageTo(time.Unix(cp.BirthDate, 0), timeStampStart)
				users.v[element.ID] = cp
			}
		case "locations":
			data := dto.LocationsFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				cp := element
				cp.Visits = make(map[int]*dto.Visit)
				locations.v[element.ID] = cp
			}
		case "visits":
			data := dto.VisitsFile{}
			err = data.UnmarshalJSON(content)
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range data.Data {
				cp := element
				visits.v[element.ID] = &cp
			}
		}
		rc.Close()
	}

	for _, visit := range visits.v {
		cpv := visit
		u, ok := users.v[visit.User]
		if !ok {
			log.Fatal("not found user for visit")
		}
		u.Visits[cpv.ID] = cpv
		users.v[u.ID] = u

		l, ok := locations.v[visit.Location]
		if !ok {
			log.Fatal("not found location for visit")
		}
		l.Visits[cpv.ID] = cpv
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
		bytes, _ := obj.MarshalJSON()
		w.Write(bytes)
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
		bytes, _ := obj.MarshalJSON()
		w.Write(bytes)
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
		bytes, _ := obj.MarshalJSON()
		w.Write(bytes)
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

		gender, err := jsonRawToString(blank.Gender)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		birthDate, err := jsonRawToInt(blank.BirthDate)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		email, err := jsonRawToString(blank.Email)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		firstName, err := jsonRawToString(blank.FirstName)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		lastName, err := jsonRawToString(blank.LastName)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		users.mux.Lock()
		obj, _ := users.v[id]
		if gender != "" {
			obj.Gender = gender
		}
		if birthDate != 0 || string(blank.BirthDate) == "0" {
			obj.BirthDate = int64(birthDate)
			obj.Age = ageTo(time.Unix(obj.BirthDate, 0), timeStampStart)
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

		place, err := jsonRawToString(blank.Place)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		country, err := jsonRawToString(blank.Country)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		city, err := jsonRawToString(blank.City)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		distance, err := jsonRawToInt(blank.Distance)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

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

		location, err := jsonRawToInt(blank.Location)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		user, err := jsonRawToInt(blank.User)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		mark, err := jsonRawToInt(blank.Mark)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		visitedAt, err := jsonRawToInt(blank.VisitedAt)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		visits.mux.Lock()
		obj, _ := visits.v[id]
		if location != 0 {
			if obj.Location != location {
				locations.mux.Lock()
				l, _ := locations.v[obj.Location]
				delete(l.Visits, obj.ID)
				locations.v[obj.Location] = l

				nl, _ := locations.v[location]
				nl.Visits[obj.ID] = obj
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
				nu.Visits[obj.ID] = obj
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

		//visits.v[id] = obj
		visits.mux.Unlock()
		w.Write(successUpdate)
	})

	// POST /<entity>/new
	r.Post("/users/new", func(w http.ResponseWriter, r *http.Request) {
		blank := dto.User{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidUser(blank) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		blank.Age = ageTo(time.Unix(blank.BirthDate, 0), timeStampStart)
		users.mux.Lock()
		users.v[blank.ID] = blank
		users.mux.Unlock()
		w.Write(successUpdate)
	})
	r.Post("/locations/new", func(w http.ResponseWriter, r *http.Request) {
		blank := dto.Location{}
		err := json.NewDecoder(r.Body).Decode(&blank)
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
		blank := &dto.Visit{}
		err := json.NewDecoder(r.Body).Decode(&blank)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !isValidVisit(*blank) {
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
		u.Visits[blank.ID] = blank
		users.v[blank.User] = u
		users.mux.Unlock()

		locations.mux.Lock()
		l, ok := locations.v[blank.Location]
		if !ok {
			http.Error(w, "", http.StatusBadRequest)
			locations.mux.Unlock()
			return
		}
		l.Visits[blank.ID] = blank
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
		for _, v := range u.Visits {
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
		bytes, _ := obj.MarshalJSON()
		w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
		w.Write(bytes)
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
			if fromAge != "" && fromAgeInt > u.Age {
				continue
			}
			if toAge != "" && toAgeInt <= u.Age {
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

	if debug {
		r.Mount("/debug", middleware.Profiler())
	}

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

func jsonRawToString(r json.RawMessage) (string, error) {
	if len(r) == 0 {
		return "", nil
	}
	var result string
	err := json.Unmarshal(r, &result)
	return result, err
}
func jsonRawToInt(r json.RawMessage) (int, error) {
	if len(r) == 0 {
		return 0, nil
	}
	var result int
	err := json.Unmarshal(r, &result)
	return result, err
}

func ageTo(born, to time.Time) int {
	years := to.Year() - born.Year()
	if born.Month() > to.Month() {
		years--
	} else if born.Month() == to.Month() && born.Day() < to.Day() {
		years--
	}

	return years
}
