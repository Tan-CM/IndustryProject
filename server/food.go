package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type productType struct {
	//	Id          string  `json:"id"`
	Category    string  `json:"category"`
	Name        string  `json:"name"`
	Weight      float32 `json:"weight"`
	Energy      float32 `json:"energy"`
	Protein     float32 `json:"protein"`
	FatTotal    float32 `json:"fatTotal"`
	FatSat      float32 `json:"fatSat"`
	Fibre       float32 `json:"fibre"`
	Carb        float32 `json:"carb"`
	Cholesterol float32 `json:"cholesterol"`
	Sodium      float32 `json:"sodium"`
}

type foodListType struct {
	Count   int                     `json:"count"`
	FoodMap *map[string]productType `json:"foodMap"`
}

// Expected Key list for PATCH
// type keyValue struct {
// 	key string,
// 	type string,
// }
var keysFood = []string{
	"id",
	"category",
	"name",
	"weight",
	"energy",
	"protein",
	"fatTotal",
	"fatSat",
	"fibre",
	"carb",
	"cholesterol",
	"sodium",
}

func foodCacheInit() {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = GetProductRecordsInit(db)
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}

	db, err = sql.Open("mysql", cfgUser.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = GetUserRecordsInit(db)
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
}

// home is the handler for "/api/v1/" resource
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("testing")
	fmt.Fprintf(w, "Welcome to the REST FOOD API Server!")
}

// allfoods is the handler for "/api/v1/allfoods" resource
func allfoods(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	//fmt.Printf("v :%+v", v)
	key, ok := v["key"]
	if !ok {
		http.Error(w, "401 - Missing key in URL", http.StatusNotFound)
		return
	}

	if !validAdmin(key[0]) {
		// w.WriteHeader(http.StatusNotFound)
		// w.Write([]byte("401 - Invalid key"))
		http.Error(w, "401 - Invalid key", http.StatusNotFound)
		return
	}

	var foodList foodListType
	var err error

	foodList.FoodMap, err = GetProductRecords()

	if err != nil {
		http.Error(w, "SQL DB Read Error", http.StatusInternalServerError)
		return
	}

	foodList.Count = len(*foodList.FoodMap)
	//fmt.Printf("BufferMap :%+v\n", *bufferMap)

	// returns all the foods in JSON and send to IO Response writer
	//err = (* productTypeInterface)(productList).ToJSON(w)
	err = json.NewEncoder(w).Encode(foodList)

	if err != nil {
		fmt.Println("error marshalling")
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

// food() is the hanlder for "/api/v1/foods/{fid}" resource
func food(w http.ResponseWriter, r *http.Request) {

	v := r.URL.Query()
	//fmt.Printf("v :%+v", v)
	key, ok := v["key"]
	if !ok {
		http.Error(w, "401 - Missing key in URL", http.StatusNotFound)
		return
	}

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

	// Get does not have a body so only header
	if r.Method == "GET" {
		fmt.Println("userMap :", userMap)
		// vakidate key for parameter key-value
		if !validRegUser(key[0]) {
			// w.WriteHeader(http.StatusNotFound)
			// w.Write([]byte("401 - Invalid key"))
			http.Error(w, "401 - Invalid keyX", http.StatusNotFound)
			return
		}

		fmt.Println("fid =", params["fid"])

		// find string with {PrefixId*} for group ID search
		pattern := regexp.MustCompile("[a-zA-Z]+[0-9]*\\*")
		IdFound := pattern.FindString(params["fid"])

		fmt.Printf("a: %+v", IdFound)

		// if Id pattern is not found, then use the ID directly
		bufferMap := &map[string]productType{}
		var err error

		if len(IdFound) == 0 {
			// check if there is a row for this record with the ID
			bufferMap, err = GetOneRecord(params["fid"])
			if err != nil {
				http.Error(w, "404 - Food id not found", http.StatusNotFound)
				return
			}
			fmt.Printf("bufferMap : %+v", bufferMap)
			err = json.NewEncoder(w).Encode(bufferMap) //key:value
			if err != nil {
				fmt.Println("error marshalling")
				http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			}

		} else {
			// remove * to get the ID prefix
			prefixID := strings.TrimSuffix(IdFound, "*")
			fmt.Println("PrefixID :", prefixID)

			var foodList foodListType
			//bufferMap, err = GetPrefixedRecords(prefixID)
			foodList.FoodMap, err = GetPrefixedRecords(prefixID)
			if err != nil {
				http.Error(w, "404 - Food id not found", http.StatusNotFound)
				return
			}

			// create a food map to be populated to match search
			foodList.Count = len(*foodList.FoodMap)
			fmt.Printf("foodList : %+v\n", foodList)
			err = json.NewEncoder(w).Encode(&foodList) //key:value
			if err != nil {
				fmt.Println("error marshalling")
				http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			}
		}
	}

	// Delete may have a body but not encouraged, safest not to use
	if r.Method == "DELETE" {
		// vakidate key for parameter key-value
		if !validAdmin(key[0]) {
			// w.WriteHeader(http.StatusNotFound)
			// w.Write([]byte("401 - Invalid key"))
			http.Error(w, "401 - Invalid key", http.StatusNotFound)
			return
		}

		count, err := GetRowCount(db, params["fid"])
		if err != nil {
			fmt.Println("Error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Internal Server Error"))
			return
		}
		if count == 0 {
			// w.WriteHeader(http.StatusNotFound)
			// w.Write([]byte("404 - Food id not found"))
			// another way to send fixed error message with http.error()
			http.Error(w, "404 - Food Id not found", http.StatusNotFound)
			return
		}
		if count > 1 {
			// some database error because there are more than one row with the same id
			fmt.Println("Error - Duplicate IDs")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Internal Server Error"))
			return
		}
		// count == 1
		DeleteRecord(db, params["fid"])
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("202 - Food item deleted: " + params["fid"]))
	}

	// check for json application
	if r.Header.Get("Content-type") == "application/json" {
		// POST is for creating new food item
		if r.Method == "POST" { // check request method
			// vakidate key for parameter key-value
			if !validAdmin(key[0]) {
				// w.WriteHeader(http.StatusNotFound)
				// w.Write([]byte("401 - Invalid key"))
				http.Error(w, "401 - Unauthorized Access", http.StatusNotFound)
				return
			}
			// read the string sent to the service
			var newFood productType
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)
				if newFood.Category == "" || newFood.Name == "" { // empty title
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please supply food " + "information " + "in JSON format"))
					return
				} // check if food item exists; add only if food item does not exist
				fmt.Printf("Food : %+v\n", newFood)

				// check if there is a row for this record with the ID
				count, err := GetRowCount(db, params["fid"])
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				switch {
				case count == 0:
					InsertRecord(db, &newFood, params["fid"])
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Food item added: " + params["fid"] + " Category: " + newFood.Category +
						" Name: " + newFood.Name))

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
				w.Write([]byte("422 - Please supply food information " + "in JSON format"))
			}
		}

		//---PUT is for creating or updating exiting food item---
		if r.Method == "PUT" {
			// vakidate key for parameter key-value
			if !validAdmin(key[0]) {
				// w.WriteHeader(http.StatusNotFound)
				// w.Write([]byte("401 - Invalid key"))
				http.Error(w, "401 - Invalid key", http.StatusNotFound)
				return
			}

			//var newFood productType
			var newFood mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)

				fmt.Printf("New Food JSON : %+v\n", newFood)

				// if newFood.Category == "" || newFood.Name == "" { // empty category or name in body
				// 	w.WriteHeader(http.StatusUnprocessableEntity)
				// 	w.Write([]byte("422 - Please supply food " + " information " + "in JSON format"))
				// 	return
				// } // check if food item exists; add only if does not exist

				// validate the JSON keys and value type in body are correct
				if !foodValidKeysValues(newFood, keysFood) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Error in JSON body map --> key-value pairs"))
					return
				}
				// validate number of key-value pairs
				if len(newFood) != len(keysFood) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Error in JSON body --> Incomplete key-value pairs"))
					return
				}

				// check if there is a row for this record with the ID
				count, err := GetRowCount(db, params["fid"])
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				fmt.Println("Count :", count)
				fmt.Println("Product", newFood)

				switch {
				case count == 0:
					// Add row if none exist
					// InsertRecord(db, &newFood, params["fid"])
					// w.WriteHeader(http.StatusCreated)
					// w.Write([]byte("201 - Food item added: " + params["fid"]))
					http.Error(w, "404 - Food Id not found", http.StatusNotFound)

				case count == 1:
					// Edit row if row exist
					EditRecord(db, &newFood, params["fid"])
					w.WriteHeader(http.StatusAccepted)
					// w.Write([]byte("202 - Food item Updated: " + params["fid"] +
					// 	" Category: " + newFood.Category + " Name:" + newFood.Name))
					w.Write([]byte("202 - Food item Updated: " + params["fid"]))

				case count > 1:
					// some database error because there are more than one row with the same id
					fmt.Println("Error - Duplicate IDs")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
				}

			} else {
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please supply " + "food information " + "in JSON format"))
			}
		}
		//---PUT is for creating or updating exiting course---
		if r.Method == "PATCH" {
			// vakidate key for parameter key-value
			if !validAdmin(key[0]) {
				// w.WriteHeader(http.StatusNotFound)
				// w.Write([]byte("401 - Invalid key"))
				http.Error(w, "401 - Invalid key", http.StatusNotFound)
				return
			}

			var newFood mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)
				if len(newFood) == 0 { // empty body
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please supply food information body in JSON format"))
					return
				} // check if food item exists; add only if does not exist

				fmt.Printf("newFood : %+v\n", newFood)
				fmt.Printf("newFood Size: %+v\n", len(newFood))

				// check if there is a row for this record with the ID
				count, err := GetRowCount(db, params["fid"])
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				fmt.Println("Count :", count)
				fmt.Println("Product", newFood)

				switch {
				case count == 0:
					http.Error(w, "404 - Food Id not found", http.StatusNotFound)

				case count == 1:
					// validate the JSON keys in body are correct
					if !foodValidKeysValues(newFood, keysFood) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - Error in JSON body map key-value pairs"))
						return
					}

					// Edit row if row exist
					err := UpdateRecord(db, newFood, params["fid"])
					if err != nil {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - Please supply updated food information body in JSON format"))
					}
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - Food item is Patched: " + params["fid"]))

				case count > 1:
					// some database error because there are more than one row with the same id
					fmt.Println("Error - Duplicate IDs")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
				}

			} else {
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please supply " + "food information " + "in JSON format"))
			}
		}
	}
}

// count all the valid keys of JSON object
func foodValidKeysValues(mapX mapInterface, keys []string) bool {
	// validate JSON data is correct

	if len(mapX) == 0 {
		return false
	}

	var count int
	// validate keys
	for _, key := range keys {
		if _, ok := mapX[key]; ok {
			count++
		}
	}
	if len(mapX) != count {
		return false
	}

	//	fmt.Println("Key Count = ", count)

	// validate values
	count = 0
	for i := 0; i < len(keys); i++ {
		if i < 3 {
			if v, ok := mapX[keys[i]].(string); ok {
				count++
				fmt.Println(v)
			}
		} else {
			if v, ok := mapX[keys[i]].(float64); ok {
				count++
				fmt.Println(v)
			}
		}
	}

	//	fmt.Println("Value Type Match Count = ", count)

	if len(mapX) != count {
		return false
	}
	return true
}
