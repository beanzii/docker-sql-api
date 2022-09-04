package models

import (
	"encoding/json"
	"net/http"
)

type Device struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

func (d *Device) Load(r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(&d)
}
