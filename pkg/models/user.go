package models

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) Load(r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(&u)
}
