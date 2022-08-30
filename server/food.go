package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type productType struct {
	//	Id          string  `json:"id"`
	Category    string  `json:"category" valid:"required,stringlength(3|20),matches(^[a-zA-Z]+$)"`
	Name        string  `json:"name" valid:"required,stringlength(3|60),matches(^[a-zA-Z]+(?:[ ]+[a-zA-Z]+)*$)"`
	Weight      float32 `json:"weight" valid:"required,range(0|1000)"`
	Energy      float32 `json:"energy" valid:"required,range(0|1000)"`
	Protein     float32 `json:"protein" valid:"required,range(0|100)"`
	FatTotal    float32 `json:"fatTotal" valid:"required,range(0|100)"`
	FatSat      float32 `json:"fatSat" valid:"required,range(0|100)"`
	Fibre       float32 `json:"fibre" valid:"required,range(0|100)"`
	Carb        float32 `json:"carb" valid:"required,range(0|500)"`
	Cholesterol float32 `json:"cholesterol" valid:"required,range(0|1000)"`
	Sodium      float32 `json:"sodium" valid:"required,range(0|5000)"`
}

type foodListType struct {
	Count   int                     `json:"count"`
	FoodMap *map[string]productType `json:"foodMap"`
}

// Define JSON keys and type for JSON validation for interface{} is float64
var foodKeyTypeRules = map[string]string{
	"Id":          "string",
	"Category":    "string",
	"Name":        "string",
	"Weight":      "float64",
	"Energy":      "float64",
	"Protein":     "float64",
	"FatTotal":    "float64",
	"FatSat":      "float64",
	"Fibre":       "float64",
	"Carb":        "float64",
	"Cholesterol": "float64",
	"Sodium":      "float64",
}

// Map form for Map validation (govalidator)
// map is for POST, PUT and PATCH so required has to be removed
var foodMapRules = map[string]interface{}{
	"Id":          "required,matches(^[a-zA-Z]{3}[0-9]{4}$)",
	"Category":    "required,stringlength(3|20),matches(^[a-zA-Z]+$)",
	"Name":        "required,stringlength(3|60),matches(^[a-zA-Z]+(?:[ ]+[a-zA-Z]+)*$)",
	"Weight":      "required,range(0|1000)",
	"Energy":      "required,range(0|1000)",
	"Protein":     "range(0|100)", // required removed to accept zero value
	"FatTotal":    "range(0|100)",
	"FatSat":      "range(0|100)",
	"Fibre":       "range(0|100)",
	"Carb":        "range(0|500)",
	"Cholesterol": "range(0|1000)",
	"Sodium":      "range(0|5000)",
}

// Map form for Map validation (govalidator)
// map is for POST, PUT and PATCH so required has to be removed
var foodNoIdMapRules = map[string]interface{}{
	"Category":    "required,type(string),stringlength(3|20),matches(^[a-zA-Z]+$)",
	"Name":        "required,type(string),stringlength(3|60),matches(^[a-zA-Z]+(?:[ ]+[a-zA-Z]+)*$)",
	"Weight":      "required,type(float64),range(0|1000)",
	"Energy":      "required,type(float64),range(0|1000)",
	"Protein":     "type(float64),range(0|100)", // required removed to accept zero value
	"FatTotal":    "type(float64),range(0|100)",
	"FatSat":      "type(float64),range(0|100)",
	"Fibre":       "type(float64),range(0|100)",
	"Carb":        "type(float64),range(0|500)",
	"Cholesterol": "type(float64),range(0|1000)",
	"Sodium":      "type(float64),range(0|5000)",
}

func foodCacheInit() {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = getProductRecordsInit(db)
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
		http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
		return
	}

	var foodList foodListType
	var err error

	foodList.FoodMap, err = getProductRecords()

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
			http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
			return
		}

		fmt.Println("fid =", params["fid"])

		// find string with {PrefixId*} for group ID search, using backtick `*` for raw character instead of \\*
		pattern := regexp.MustCompile("[a-zA-Z]+[0-9]*\\*")
		IdFound := pattern.FindString(params["fid"])

		fmt.Printf("a: %+v", IdFound)

		// if Id pattern is not found, then use the ID directly
		bufferMap := &map[string]productType{}
		var err error

		if len(IdFound) == 0 {
			// check if there is a row for this record with the ID
			bufferMap, err = getOneRecord(params["fid"])
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
			foodList.FoodMap, err = getPrefixedRecords(prefixID)
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
			http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
			return
		}

		count, err := getRowCount(db, params["fid"])
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
		deleteRecord(db, params["fid"])
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
			var newFood mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)
				fmt.Printf("UnMarshall :%+v\n", newFood)

				// validate the JSON Num of keys-value in body are correct
				if len(newFood) != len(foodNoIdMapRules) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Incompatible or Incomplete JSON Data Error"))
					return
				}

				// validate the JSON keys and value type in body are correct
				if ok, err := validateKeysValueTypes(newFood, foodKeyTypeRules); !ok {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
					return
				}

				// struct value validaion with struct tag
				if ok, err := govalidator.ValidateMap(newFood, foodNoIdMapRules); !ok {
					//if ok, err := govalidator.ValidateStruct(newFood); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
					return
				}

				// check if there is a row for this record with the ID
				count, err := getRowCount(db, params["fid"])
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}
				switch {
				case count == 0:
					var food productType
					_, err := updateFoodMapToStruct(&food, newFood, foodMapRules)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
					}

					err = insertRecord(db, food, params["fid"])
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
						return
					}
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Food item added: " + params["fid"] + " Category: " + newFood["Category"].(string) +
						" Name: " + newFood["Name"].(string)))

					fmt.Println("Food :", newFood)

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
				http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
				return
			}

			//var newFood productType
			var newFood mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)

				fmt.Printf("New Food JSON : %+v\n", newFood)

				// validate the JSON Num of keys-value in body are correct
				if len(newFood) != len(foodMapRules) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Incompatible or Incomplete JSON Data Error"))
					return
				}

				// validate the JSON keys and value type in body are correct
				if ok, err := validateKeysValueTypes(newFood, foodKeyTypeRules); !ok {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
					return
				}

				// struct value validaion with Map interface{} values
				if ok, err := govalidator.ValidateMap(newFood, foodMapRules); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
					return
				}

				// check if there is a row for this record with the ID
				count, err := getRowCount(db, params["fid"])
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
					var food productType
					_, err := updateFoodMapToStruct(&food, newFood, foodMapRules)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
					}

					fmt.Println("new Food :", newFood)
					err = editRecord(db, newFood["Id"].(string), food, params["fid"])
					if err != nil {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - JSON Body Error: " + err.Error()))
						return
					}
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
		//---PATCH is for patching selective data ---
		if r.Method == "PATCH" {
			// vakidate key for parameter key-value
			if !validAdmin(key[0]) {
				// w.WriteHeader(http.StatusNotFound)
				// w.Write([]byte("401 - Invalid key"))
				http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
				return
			}

			var newFood mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newFood)
				// validate the JSON keys and value type in body are correct
				if ok, err := validateKeysValueTypes(newFood, foodKeyTypeRules); !ok {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
					return
				}

				// build rules dynamically base on interface{} because govalidator requires complete rules
				buildRules, err := buildVMapTemplate(newFood, foodMapRules)
				if err != nil {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Validation Build Rule Failed, " + err.Error()))
					return
				}
				fmt.Printf("Template : %+v\n", buildRules)

				// struct value validaion with Map interface{} values
				if ok, err := govalidator.ValidateMap(newFood, *buildRules); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
					return
				}

				fmt.Printf("newFood : %+v\n", newFood)
				fmt.Printf("newFood Size: %+v\n", len(newFood))

				// check if there is a row for this record with the ID
				count, err := getRowCount(db, params["fid"])
				if err != nil {
					fmt.Println("Error", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				switch {
				case count == 0:
					http.Error(w, "404 - Food Id not found", http.StatusNotFound)

				case count == 1:

					fmt.Printf("New Food : %+v, %+v\n", newFood, params["fid"])

					// Edit row if row exist
					err := updateRecord(db, newFood, *buildRules, params["fid"])
					if err != nil {
						w.WriteHeader(http.StatusUnprocessableEntity)
						w.Write([]byte("422 - JSON Body Error, " + err.Error()))
						return
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

// updateFoodMapToStruct extract the data in the Map and convert to struct type
// It checks for the validity of the key
// return (number of key found, err)
// Update p based fields of food interface, using foodFieldMap as key reference check
func updateFoodMapToStruct(p *productType, food mapInterface, foodFieldMap mapInterface) (int, error) {

	var count int

	// Note "Id" key if present is also checked by not pu into p
	for k, v := range food {
		if _, ok := foodFieldMap[k]; !ok {
			return count, fmt.Errorf("invalid Key Used : (%v)", k)
		}
		count++
		fmt.Println("Key :", k)
		switch k {
		case "Category":
			p.Category = v.(string)
		case "Name":
			p.Name = v.(string)
		case "Weight":
			p.Weight = (float32)(v.(float64))
		case "Energy":
			p.Energy = (float32)(v.(float64))
		case "Protein":
			p.Protein = (float32)(v.(float64))
		case "FatTotal":
			p.FatTotal = (float32)(v.(float64))
		case "FatSat":
			p.FatSat = (float32)(v.(float64))
		case "Fibre":
			p.Fibre = (float32)(v.(float64))
		case "Carb":
			p.Carb = (float32)(v.(float64))
		case "Cholesterol":
			p.Cholesterol = (float32)(v.(float64))
		case "Sodium":
			p.Sodium = (float32)(v.(float64))
			// default:
			// 	return count, fmt.Errorf("invalid Key Used : (%v)", k)
		}

	}
	return count, nil
}
