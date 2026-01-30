package model

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

func NewUser(Id int, Name, City, State, Country string) *User {
	return &User{
		Id:      Id,
		Name:    Name,
		City:    City,
		State:   State,
		Country: Country,
	}
}
