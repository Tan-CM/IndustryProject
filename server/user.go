package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	uuid "github.com/satori/go.uuid"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type userType struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User this resource to Add user, Delete User, Validate user
// user() is the hanlder for "/api/v1/user/" resource
func user(w http.ResponseWriter, r *http.Request) {
	// Use mysql as driverName and a valid DSN as dataSourceName:
	// user set up password that can access this db connection
	//db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:58710)/foodDB")
	//db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:58710)/foodDB")
	db, err := sql.Open("mysql", cfg.FormatDSN())

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

	// Creating UUID Version 4
	// panic on error
	// u1 := uuid.Must(uuid.NewV4(), errors.New("UUID Error"))
	// fmt.Printf("UUIDv4: %s\n", u1)

	// or error handling
	u2 := uuid.NewV4()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	fmt.Printf("UUIDv4:1 %s\n", u2)

	u2 = uuid.NewV4()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	fmt.Printf("UUIDv4:2 %s\n", u2)

	// Parsing UUID from string input
	u2, err = uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	fmt.Printf("Successfully parsed:1 %s\n", u2)

	// Parsing UUID from string input
	u2, err = uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	fmt.Printf("Successfully parsed:2 %s\n", u2)

	// // Get does not have a body so only header
	// if r.Method == "GET" {
	// 	fmt.Println("fid =", params["fid"])

	// 	// find string with {PrefixId*} for group ID search
	// 	pattern := regexp.MustCompile("[a-zA-Z]+\\*")
	// 	IdFound := pattern.FindString(params["fid"])

	// 	fmt.Printf("a: %+v", IdFound)

	// 	// if Id pattern is not found, then use the ID directly
	// 	bufferMap := &map[string]productType{}
	// 	var err error

	// 	if len(IdFound) == 0 {
	// 		// check if there is a row for this record with the ID
	// 		bufferMap, err = GetOneRecord(params["fid"])

	// 	} else {
	// 		// remove * to get the ID prefix
	// 		prefixID := strings.TrimSuffix(IdFound, "*")
	// 		fmt.Println("PrefixID :", prefixID)

	// 		// create a food map to be populated to match search
	// 		bufferMap, err = GetPrefixedRecords(prefixID)

	// 	}
	// 	if err != nil {
	// 		http.Error(w, "404 - Food id not found", http.StatusNotFound)
	// 		return
	// 	}
	// 	fmt.Printf("bufferMap : %+v", bufferMap)
	// 	json.NewEncoder(w).Encode(bufferMap) //key:value
	// }

	// // Delete may have a body but not encouraged, safest not to use
	// if r.Method == "DELETE" {
	// 	count, err := GetRowCount(db, params["fid"])
	// 	if err != nil {
	// 		fmt.Println("Error", err)
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		w.Write([]byte("500 - Internal Server Error"))
	// 		return
	// 	}
	// 	if count == 0 {
	// 		// w.WriteHeader(http.StatusNotFound)
	// 		// w.Write([]byte("404 - Food id not found"))
	// 		// another way to send fixed error message with http.error()
	// 		http.Error(w, "404 - Food Id not found", http.StatusNotFound)
	// 		return
	// 	}
	// 	if count > 1 {
	// 		// some database error because there are more than one row with the same id
	// 		fmt.Println("Error - Duplicate IDs")
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		w.Write([]byte("500 - Internal Server Error"))
	// 		return
	// 	}
	// 	// count == 1
	// 	DeleteRecord(db, params["fid"])
	// 	w.WriteHeader(http.StatusAccepted)
	// 	w.Write([]byte("202 - Food item deleted: " + params["fid"]))
	// }

	// check for json application
	if r.Header.Get("Content-type") == "application/json" {
		// POST is for creating new food item
		if r.Method == "POST" { // check request method
			// read the string sent to the service
			var newUser userType
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newUser)
				if newUser.Name == "" || newUser.Email == "" { // empty name and email
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please supply user information in JSON format"))
					return
				} // check if food item exists; add only if food item does not exist
				fmt.Printf("New User : %+v\n", newUser)

				// check if there is a row for this record with the email ID
				//count, err := userGetRowCount(db, newUser.Email)
				count := 0
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				switch {
				case count == 0:
					// token := InsertRecord(db, &newUser)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - New User added: " + newUser.Name + " Email: " + newUser.Email +
						" Access Token: " /*+ token*/))

				case count == 1:
					w.WriteHeader(http.StatusConflict) // food id key already exist
					w.Write([]byte("409 - Duplicate food ID"))

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

		// //---PUT is for creating or updating exiting course---
		// if r.Method == "PUT" {
		// 	var newFood productType
		// 	reqBody, err := ioutil.ReadAll(r.Body)
		// 	if err == nil {
		// 		// parse JSON to object data structure
		// 		json.Unmarshal(reqBody, &newFood)
		// 		if newFood.Category == "" || newFood.Name == "" { // empty category or name in body
		// 			w.WriteHeader(http.StatusUnprocessableEntity)
		// 			w.Write([]byte("422 - Please supply food " + " information " + "in JSON format"))
		// 			return
		// 		} // check if food item exists; add only if does not exist

		// 		// check if there is a row for this record with the ID
		// 		count, err := GetRowCount(db, params["fid"])
		// 		if err != nil {
		// 			fmt.Println("Error", err)
		// 			w.WriteHeader(http.StatusInternalServerError)
		// 			w.Write([]byte("500 - Internal Server Error"))
		// 			return
		// 		}

		// 		fmt.Println("Count :", count)
		// 		fmt.Println("Product", newFood)

		// 		switch {
		// 		case count == 0:
		// 			// Add row if none exist
		// 			InsertRecord(db, &newFood, params["fid"])
		// 			w.WriteHeader(http.StatusCreated)
		// 			w.Write([]byte("201 - Food item added: " + params["fid"]))

		// 		case count == 1:
		// 			// Edit row if row exist
		// 			EditRecord(db, &newFood, params["fid"])
		// 			w.WriteHeader(http.StatusAccepted)
		// 			w.Write([]byte("202 - Food item updated: " + params["fid"] +
		// 				" Category: " + newFood.Category + " Name:" + newFood.Name))

		// 		case count > 1:
		// 			// some database error because there are more than one row with the same id
		// 			fmt.Println("Error - Duplicate IDs")
		// 			w.WriteHeader(http.StatusInternalServerError)
		// 			w.Write([]byte("500 - Internal Server Error"))
		// 		}

		// 	} else {
		// 		w.WriteHeader(http.StatusUnprocessableEntity) // error
		// 		w.Write([]byte("422 - Please supply " + "food information " + "in JSON format"))
		// 	}
		//}
	}
}