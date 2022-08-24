package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type userType struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AccessKey string `json:"accesskey"`
	Type      string `json:"type"`
}

type usersType struct {
	Count int        `json:"count"`
	Users []userType `json:"users"`
}

// ENUM data object is a string (not integer)
const (
	EMPTY = "0" // Empty == 0
	ADMIN = "1" // ADMIN == 1
	USER  = "2" // USER == 2
)

type userNameId struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type mapInterface map[string]interface{}

// Expected Key list for PATCH
var keysUser = []string{
	"name",
	"email",
}

// users is the handler for "/api/v1/users" resource
func users(w http.ResponseWriter, r *http.Request) {

	v := r.URL.Query()
	//fmt.Printf("v :%+v", v)
	key, ok := v["key"]
	if !ok {
		http.Error(w, "401 - Missing key in URL", http.StatusNotFound)
		return
	}

	// vakidate key for parameter key-value
	if !validAdmin(key[0]) {
		// w.WriteHeader(http.StatusNotFound)
		// w.Write([]byte("401 - Invalid key"))
		http.Error(w, "401 - Invalid key", http.StatusNotFound)
		return
	}

	// Use mysql as driverName and a valid DSN as dataSourceName:
	// user set up password that can access this db connection
	//db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:58710)/foodDB")
	//db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:58710)/foodDB")
	db, err := sql.Open("mysql", cfgUser.FormatDSN())

	// handle error
	if err != nil {
		panic(err.Error())
	}
	// defer the close till after the main function has finished executing
	defer db.Close()

	bufferMap, err := GetUsersRecords(db)

	if err != nil {
		http.Error(w, "SQL DB Read Error", http.StatusInternalServerError)
		return
	}
	//fmt.Printf("BufferMap :%+v\n", *bufferMap)

	err = json.NewEncoder(w).Encode(bufferMap)

	if err != nil {
		fmt.Println("error marshalling")
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

// User this resource to Add user, Delete User, Validate user
// user() is the hanlder for "/api/v1/user/{id}" resource
func user(w http.ResponseWriter, r *http.Request) {
	// Use mysql as driverName and a valid DSN as dataSourceName:
	// user set up password that can access this db connection
	//db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:58710)/foodDB")
	//db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:58710)/foodDB")
	db, err := sql.Open("mysql", cfgUser.FormatDSN())

	// handle error
	if err != nil {
		panic(err.Error())
	}
	// defer the close till after the main function has finished executing
	defer db.Close()
	//	fmt.Println("Database opened")

	// mux.Vars(r) is the variable immediately after the URL
	// it can have more than one parameters
	params := mux.Vars(r)
	fmt.Println("parameter =", params)

	fmt.Printf("Parameter = %+v\n", params)
	fmt.Println("Select ", params["select"])

	// All methods requires a Access Key except "POST"
	var key []string
	var ok bool
	var count int

	// Common for GET, DELETE, PATCH
	if r.Method != "POST" {
		// get the Access Key from the URL parameter
		v := r.URL.Query()
		//fmt.Printf("v :%+v", v)
		key, ok = v["key"]
		if !ok {
			http.Error(w, "401 - Missing key in URL", http.StatusNotFound)
			return
		}

		fmt.Println(key[0])

		// validate key is registered
		if !validRegUser(key[0]) {
			http.Error(w, "401 - Unauthorised Access Key", http.StatusNotFound)
			return
		}

		// Need to check DB has a row for this record with the ID
		// count returns the number of record matches
		count, err = userGetRowCount(db, params["uid"])
		if err != nil {
			fmt.Println("Error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Internal Server Error"))
			return
		}

		// for GET, DELETE and PUT, there must only be count = 1
		// count = 0, mean ID cannot be found
		if count == 0 {
			http.Error(w, "404 - User Id not found", http.StatusNotFound)
			return
		}

		// count > 1, means there is a problem with DB
		if count > 1 {
			fmt.Println("Error - Duplicate IDs in DB")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Internal Server Error"))
		}

		fmt.Println("Count :", count)
	}

	// Get does not have a body so only header
	// Get is used to validate a user
	if r.Method == "GET" {
		// case count == 1
		// if not admin key, then user authentication is needed
		if !validAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
			http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
			return
		}

		fmt.Println("uid =", params["uid"])
		user, err := userGetOneRecord(db, params["uid"])
		if err != nil {
			http.Error(w, "404 - User Id not found", http.StatusNotFound)
			return
		}
		fmt.Printf("User : %+v", user)

		err = json.NewEncoder(w).Encode(user) //key:value
		if err != nil {
			fmt.Println("error marshalling")
			http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
		}
	}

	// Delete may have a body but not encouraged, safest not to use
	if r.Method == "DELETE" {
		// count == 1
		if validAdmin(key[0]) {
			userDeleteRecord(db, params["uid"])
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - User deleted: " + params["uid"]))
			fmt.Println("UserMap ", userMap)
			return
		}

		if !verifiedUser(key[0], params["uid"]) {
			http.Error(w, "401 - Mismatched Access key with Id", http.StatusNotFound)
			return
		}

		userDeleteRecord(db, params["uid"])
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("202 - User deleted: " + params["uid"]))
		fmt.Println("UserMap ", userMap)
	}

	// check for json application
	if r.Header.Get("Content-type") == "application/json" {
		// POST is for creating new food item
		if r.Method == "POST" && strings.ToUpper(params["uid"]) == "NEW" { // check request method
			// read the string sent to the service
			var newUser userType
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure, note newUser has either a name or email or both
				json.Unmarshal(reqBody, &newUser)
				if newUser.Name == "" || newUser.Email == "" { // empty name and email
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please supply user information in JSON format"))
					return
				} // check if food item exists; add only if food item does not exist
				fmt.Printf("New User : %+v\n", newUser)

				// check if there is a row for this record with the New User email ID
				count, err := userGetRowCount(db, newUser.Email)

				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				switch {
				case count == 0:
					newUser.AccessKey = (uuid.NewV4()).String()
					_, ok := userMap[newUser.AccessKey]

					// repeat this if the Access key is not unique
					for ok {
						newUser.AccessKey = (uuid.NewV4()).String()
						_, ok = userMap[newUser.AccessKey]
						fmt.Println("Rare case of non-unique UUID")
					}

					newUser.Type = USER

					userInsertRecord(db, &newUser)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - New User Registered: " + newUser.Name + " Email: " + newUser.Email +
						" Access Key: " + newUser.AccessKey))

				case count == 1:
					w.WriteHeader(http.StatusConflict) // food id key already exist
					w.Write([]byte("409 - Email ID is already used, please use a another Email ID"))

				case count > 1:
					// some sql data error if any such error
					fmt.Println("SQL database error")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
				}

			} else {
				// Problem with the body from response
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please user information in JSON format"))
			}
		}

		//---PUT is for updating email or name of user record
		if r.Method == "PATCH" {
			// case of count == 1
			var newUserReq mapInterface
			//var newUser userNameId
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newUserReq)

				fmt.Printf("New User Request: %+v\n", newUserReq)

				// validate JSON data is correct
				if !userValidKeysValues(newUserReq, keysUser) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Error in JSON body map key-value"))
					return
				}

				var newUser userNameId
				// assertion
				if v, ok := newUserReq["name"]; ok {
					newUser.Name = v.(string)
				}
				if v, ok := newUserReq["email"]; ok {
					newUser.Email = v.(string)
				}

				if validAdmin(key[0]) {
					// Edit row if row exist
					if err := userUpdateRecord(db, newUser, params["uid"]); err != nil {
						if err == errDuplicateID {
							w.WriteHeader(http.StatusUnprocessableEntity)
							w.Write([]byte("422 - Error - New Email is used"))
							return
						}
						fmt.Println("Error", err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("500 - Internal Server Error"))
						return
					}
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - User item Patched: " + params["uid"]))
					return
				}

				fmt.Println("key :", key[0])
				fmt.Println("user ID :", params["uid"])

				if !verifiedUser(key[0], params["uid"]) {
					http.Error(w, "401 - Mismatched Access key with Id", http.StatusNotFound)
					return
				}

				if err := userUpdateRecord(db, newUser, params["uid"]); err != nil {
					if err == errDuplicateID {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - Error - New Email is used"))
						return
					}
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("202 - User item Patch: " + params["uid"]))

			} else {
				// Problem with the body from response
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please user information in JSON format"))
			}
		}
	}
}

// count all the valid keys of JSON object
func userValidKeysValues(mapX mapInterface, keys []string) bool {
	// validate JSON data is correct
	//	fmt.Println("length = ", len(mapX))
	if len(mapX) == 0 {
		return false
	}

	// validate key
	var count int
	for _, key := range keys {
		if _, ok := mapX[key]; ok {
			count++
		}
	}
	//	fmt.Println("Key Count = ", count)

	if len(mapX) != count {
		return false
	}

	// validate values
	count = 0
	for i := 0; i < len(keys); i++ {
		if v, ok := mapX[keys[i]].(string); ok {
			count++
			fmt.Println(v)
		}
	}

	//	fmt.Println("Value Type Match Count = ", count)

	if len(mapX) != count {
		return false
	}
	return true
}
