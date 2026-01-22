package model

type User struct {
	Id      int
	Name    string
	City    string
	State   string
	Country string
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
