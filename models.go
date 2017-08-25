package main

import (
	"encoding/json"
)

// {"distance": 9, "city": "Новоомск", "place": "Ресторан", "id": 1, "country": "Венесуэлла"}
type Location struct {
	ID       int    `json:"id"`
	Distance int    `json:"distance"`
	City     string `json:"city"`
	Place    string `json:"place"`
	Country  string `json:"country"`

	// cache
	Visits map[int]*Visit `json:"-"`
	JSON   []byte         `json:"-"`
}

// {"first_name": "Злата", "last_name": "Кисатович", "birth_date": -627350400, "gender": "f", "id": 1, "email": "coorzaty@me.com"}
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate int64  `json:"birth_date"`
	Age       int    `json:"-"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`

	// cache
	Visits map[int]*Visit `json:"-"`
	JSON   []byte         `json:"-"`
}

// {"user": 42, "location": 13, "visited_at": 1123175509, "id": 1, "mark": 4}
type Visit struct {
	ID        int `json:"id"`
	User      int `json:"user"`
	Location  int `json:"location"`
	VisitedAt int `json:"visited_at"`
	Mark      int `json:"mark"`

	// cache
	JSON []byte `json:"-"`
}

// if any field is `null` return status 400
// if there is no field in update then skip field
type LocationRequest struct {
	ID       json.RawMessage `json:"id"`
	Distance json.RawMessage `json:"distance"`
	City     json.RawMessage `json:"city"`
	Place    json.RawMessage `json:"place"`
	Country  json.RawMessage `json:"country"`
}
type UserRequest struct {
	ID        json.RawMessage `json:"id"`
	FirstName json.RawMessage `json:"first_name"`
	LastName  json.RawMessage `json:"last_name"`
	BirthDate json.RawMessage `json:"birth_date"`
	Gender    json.RawMessage `json:"gender"`
	Email     json.RawMessage `json:"email"`
}
type VisitRequest struct {
	ID        json.RawMessage `json:"id"`
	User      json.RawMessage `json:"user"`
	Location  json.RawMessage `json:"location"`
	VisitedAt json.RawMessage `json:"visited_at"`
	Mark      json.RawMessage `json:"mark"`
}
