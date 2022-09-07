package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type userType struct {
	Name      string `json:"name" valid:"required,type(string),stringlength(3|30),matches(^[a-zA-Z]+(?:[ ]+[a-zA-Z]+)*$)"`
	Email     string `json:"email" valid:"required,type(string),email,stringlength(1|40)"`
	Gender    string `json:"gender" valid:"required,type(string),stringlength(4|6),matches(^(Female|female|Male|male))"`
	BirthYear int    `json:"birthyear" valid:"required,type(int),range(1920|2060)"`
	AccessKey string `json:"accesskey"`
	Type      string `json:"type"`
}

// Map form for Map validation for Patch
var userMapRules = map[string]interface{}{
	"name":      "type(string),stringlength(3|30),matches(^[a-zA-Z]+(?:[ ]+[a-zA-Z]+)*$)",
	"email":     "type(string),email,stringlength(1|40)",
	"gender":    "type(string),stringlength(4|6),matches(^(Female|female|Male|male))",
	"birthyear": "type(float64),range(1920|2060)",
}

// Note no space on validate struct tag allowed
// type userValid struct {
// 	Name  string `json:"name" validate:"required,min=3,max=30"`
// 	Email string `json:"email" validate:"required,email,max=40"`
// }

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

// users is the handler for "/api/v1/users" resource
func users(w http.ResponseWriter, r *http.Request) {

	v := r.URL.Query()
	//fmt.Printf("v :%+v", v)
	key, ok := v["key"]
	if !ok {
		http.Error(w, "400 - Missing key in URL", http.StatusBadRequest)
		return
	}

	// vakidate key for parameter key-value
	if !userIsAdmin(key[0]) {
		// w.WriteHeader(http.StatusUnauthorized)
		// w.Write([]byte("401 - Unauthorized Access"))
		http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
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

	bufferMap, err := userGetRecordsDB(db)

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
			http.Error(w, "400 - Missing key in URL", http.StatusBadRequest)
			return
		}

		fmt.Println(key[0])

		// validate key is registered
		if _, ok := userIsRegistered(key[0]); !ok {
			http.Error(w, "401 - Unauthorised Access Key", http.StatusUnauthorized)
			return
		}

		// Need to check DB has a row for this record with the ID
		// count returns the number of record matches
		count, err = userGetRowCountDB(db, params["uid"])
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
		if !userIsAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
			http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
			return
		}

		fmt.Println("uid =", params["uid"])
		user, err := userGetOneRecordDB(db, params["uid"])
		if err != nil {
			http.Error(w, "404 - User Id not found", http.StatusNotFound)
			return
		}
		fmt.Printf("User : %+v", user)

		err = json.NewEncoder(w).Encode(user) //key:value
		if err != nil {
			fmt.Println("error marshalling")
			http.Error(w, "500 - Unable to marshal json", http.StatusInternalServerError)
		}
	}

	// Delete may have a body but not encouraged, safest not to use
	if r.Method == "DELETE" {
		// count == 1
		if userIsAdmin(key[0]) {
			userDeleteRecordDB(db, params["uid"])
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - User deleted: " + params["uid"]))
			fmt.Println("UserMap ", userMap)
			return
		}

		// validate key with user id
		if !verifiedUser(key[0], params["uid"]) {
			http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
			return
		}

		userDeleteRecordDB(db, params["uid"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 - User deleted: " + params["uid"]))
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

				// struct validaion with struct tag
				if ok, err := govalidator.ValidateStruct(newUser); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("400 - Data Validation Failed: " + err.Error()))
					return
				}
				// trims leading and trailing spaces
				//newUser.Name = govalidator.Trim(newUser.Name, "")

				fmt.Printf("New User : %+v\n", newUser)

				// check if there is a row for this record with the New User email ID
				count, err := userGetRowCountDB(db, newUser.Email)

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

					userInsertRecordDB(db, &newUser)
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
				w.WriteHeader(http.StatusBadRequest) // error
				w.Write([]byte("400 - Please supply user Information in JSON format"))
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
				// validator
				// struct validaion with struct tag
				if ok, err := govalidator.ValidateMap(newUserReq, userMapRules); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("400 - Data Validation Failed: " + err.Error()))
					return
				}

				if userIsAdmin(key[0]) {
					// Edit row if row exist
					//if err := userUpdateRecord(db, newUser, params["uid"]); err != nil {
					if err := userUpdateRecordDB(db, newUserReq, params["uid"]); err != nil {
						if err == errDuplicateID {
							w.WriteHeader(http.StatusConflict)
							w.Write([]byte("409 - Error - New Email is used"))
							return
						}
						fmt.Println("Error", err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("500 - Internal Server Error"))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("200 - User item Patched: " + params["uid"]))
					return
				}

				fmt.Println("key :", key[0])
				fmt.Println("user ID :", params["uid"])

				// check if the user to be patched is a valid user
				if !verifiedUser(key[0], params["uid"]) {
					http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
					return
				}

				if err := userUpdateRecordDB(db, newUserReq, params["uid"]); err != nil {
					if err == errDuplicateID {
						w.WriteHeader(http.StatusConflict)
						w.Write([]byte("409 - Error - New Email is used"))
						return
					}
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("200 - User item Patch: " + params["uid"]))

			} else {
				// Problem with the body from response
				w.WriteHeader(http.StatusBadRequest) // error
				w.Write([]byte("400 - Please supply user Information in JSON format"))
			}
		}
	}
}
