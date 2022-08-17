package main

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type userType struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User this resource to Add user, Delete User, Validate user
// user() is the hanlder for "/api/v1/user/" resource
func user(w http.ResponseWriter, r *http.Request) {

}
