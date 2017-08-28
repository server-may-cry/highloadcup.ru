package dto


type UsersFile struct {
	Data []User `json:"users"`
}
type LocationsFile struct {
	Data []Location `json:"locations"`
}
type VisitsFile struct {
	Data []Visit `json:"visits"`
}

// {"distance": 9, "city": "Новоомск", "place": "Ресторан", "id": 1, "country": "Венесуэлла"}
type Location struct {
	ID       int    `json:"id"`
	Distance int    `json:"distance"`
	City     string `json:"city"`
	Place    string `json:"place"`
	Country  string `json:"country"`

	// links
	Visits map[int]struct{} `json:"-"`
}

// {"first_name": "Злата", "last_name": "Кисатович", "birth_date": -627350400, "gender": "f", "id": 1, "email": "coorzaty@me.com"}
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate int64  `json:"birth_date"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`

	// links
	Visits map[int]struct{} `json:"-"`
}

// {"user": 42, "location": 13, "visited_at": 1123175509, "id": 1, "mark": 4}
type Visit struct {
	ID        int `json:"id"`
	User      int `json:"user"`
	Location  int `json:"location"`
	VisitedAt int `json:"visited_at"`
	Mark      int `json:"mark"`
}

// if any field is `null` return status 400
// if there is no field in update then skip field
type LocationRequest struct {
	ID       int `json:"id"`
	Distance int `json:"distance"`
	City     string `json:"city"`
	Place    string `json:"place"`
	Country  string `json:"country"`
}

type UserRequest struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate int    `json:"birth_date"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
}

type VisitRequest struct {
	ID        int `json:"id"`
	User      int `json:"user"`
	Location  int `json:"location"`
	VisitedAt int `json:"visited_at"`
	Mark      int `json:"mark"`
}

type VisitInUser struct {
	Mark      int    `json:"mark"`
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`
}

type VisitsResponse struct {
	Data []VisitInUser `json:"visits"`
}
