package models

type User struct {
	FirstName string   `json:"firstname" bson:"firstname"`
	LastName  string   `json:"lastname" bson:"lastname"`
	Email     string   `json:"email" bson:"email"`
	Password  string   `json:"password" bson:"password"`
	Tasks     []string `json: "tasks" bson:"tasks"`
}
