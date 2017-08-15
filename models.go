package main

// {"distance": 9, "city": "Новоомск", "place": "Ресторан", "id": 1, "country": "Венесуэлла"}
type Locaton struct {
	ID       int    `json:"id"`
	Distance int    `json:"distance"`
	City     string `json:"city"`
	Place    string `json:"place"`
	Country  string `json:"country"`
}

// {"first_name": "Злата", "last_name": "Кисатович", "birth_date": -627350400, "gender": "f", "id": 1, "email": "coorzaty@me.com"}
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate int    `json:"birth_date"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
}

// {"user": 42, "location": 13, "visited_at": 1123175509, "id": 1, "mark": 4}
type Visit struct {
	ID        int  `json:"id"`
	User      int  `json:"user"`
	Location  int  `json:"location"`
	VisitedAt int  `json:"visited_at"`
	Mark      int8 `json:"mark"`
}
