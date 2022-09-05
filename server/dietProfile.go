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
type dietProfileType struct {
	Energy      float32 `json:"energy" valid:"required,range(0|1000)"`
	Protein     float32 `json:"protein" valid:"required,range(0|100)"`
	FatTotal    float32 `json:"fatTotal" valid:"required,range(0|100)"`
	FatSat      float32 `json:"fatSat" valid:"required,range(0|100)"`
	Fibre       float32 `json:"fibre" valid:"required,range(0|100)"`
	Carb        float32 `json:"carb" valid:"required,range(0|500)"`
	Cholesterol float32 `json:"cholesterol" valid:"required,range(0|1000)"`
	Sodium      float32 `json:"sodium" valid:"required,range(0|5000)"`
}

type dietProfileListType struct {
	Count       int                         `json:"count"`
	DietProfile *map[string]dietProfileType `json:"dietProfile"`
}

// Define JSON keys and type for JSON validation for interface{} is float64
var dietProfKeyTypeRules = map[string]string{
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
var dietProfMapRules = map[string]interface{}{
	"Energy":      "required,range(0|1000)",
	"Protein":     "range(0|100)", // required removed to accept zero value
	"FatTotal":    "range(0|100)",
	"FatSat":      "range(0|100)",
	"Fibre":       "range(0|100)",
	"Carb":        "range(0|500)",
	"Cholesterol": "range(0|1000)",
	"Sodium":      "range(0|5000)",
}

// dietProf() is the hanlder for "/api/v1/dietProf" resource
func dietUserProfile(w http.ResponseWriter, r *http.Request) {

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
	db, err := sql.Open("mysql", cfgDietProf.FormatDSN())

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
		fmt.Println("Diet Profile Get ")

		if strings.ToUpper(params["uid"]) == "ALL" {
			// if not admin key, then user authentication is needed
			if !userIsAdmin(key[0]) {
				http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
				return
			}
			var dietProfileList dietProfileListType
			var err error

			dietProfileList.DietProfile, err = dietProfGetRecords()

			if err != nil {
				http.Error(w, "SQL DB Read Error", http.StatusInternalServerError)
				return
			}

			dietProfileList.Count = len(*dietProfileList.DietProfile)
			//fmt.Printf("BufferMap :%+v\n", *bufferMap)

			// returns all the foods in JSON and send to IO Response writer
			err = json.NewEncoder(w).Encode(dietProfileList)

			if err != nil {
				fmt.Println("error marshalling")
				http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			}
			return
		}

		// find string with {PrefixId*} for group ID search, using backtick `*` for raw character instead of \\*
		pattern := regexp.MustCompile("[a-zA-Z0-9]+\\*")
		IdFound := pattern.FindString(params["uid"])

		fmt.Println("IdFound :", IdFound)
		if len(IdFound) == 0 {
			// if not admin key, then user authentication is needed
			if !userIsAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
				http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
				return
			}

			// using email as ID to get profile
			fp, err := dietProfGetOneRecord(params["uid"])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("404 - Diet Profile Not Setup ! "))
				return
			}
			fmt.Printf("Food Preference : %+v", *fp)
			err = json.NewEncoder(w).Encode(fp) //key:value
			if err != nil {
				fmt.Println("error marshalling")
				http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			}
			return
		} else {
			// remove * to get the ID prefix
			prefixID := strings.TrimSuffix(IdFound, "*")
			fmt.Println("PrefixID :", prefixID)

			var dietProfileList dietProfileListType

			//bufferMap, err = GetPrefixedRecords(prefixID)
			dietProfileList.DietProfile, err = dietProfGetPrefixedRecords(prefixID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("404 - Food id not found :" + err.Error()))
				return
			}

			// create a food map to be populated to match search
			dietProfileList.Count = len(*dietProfileList.DietProfile)
			fmt.Printf("foodList : %+v\n", dietProfileList)
			err = json.NewEncoder(w).Encode(&dietProfileList) //key:value
			if err != nil {
				fmt.Println("error marshalling")
				http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			}
			return
		}
	}

	// Delete may have a body but not encouraged, safest not to use
	if r.Method == "DELETE" {
		fmt.Println("Diet Profile Delete ")

		// if not admin key, then user authentication is needed
		if !userIsAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
			http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
			return
		}

		if _, err := dietProfGetOneRecord(params["uid"]); err != nil {
			fmt.Println("No Diet Profile for user")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Diet Profile Not Setup ! "))
			return
		}

		dietProfDeleteRecord(db, params["uid"])
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("202 - Diet Profile deleted: "))
	}

	// check for json application
	if r.Header.Get("Content-type") == "application/json" {
		// POST is for creating new food item
		if r.Method == "POST" { // check request method
			// if not admin key, then user authentication is needed
			if !userIsAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
				http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
				return
			}

			// Need to validate if user exists in user DB before profile can be created
			if userIsAdmin(key[0]) {

				dbUser, err := sql.Open("mysql", cfgUser.FormatDSN())

				// handle error
				if err != nil {
					panic(err.Error())
				}
				// defer the close till after the main function has finished executing
				defer dbUser.Close()

				count, err := userGetRowCountDB(dbUser, params["uid"])
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
					return
				}

				// valid count == 1
				switch {
				case count == 0:
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - User Id (" + params["uid"] + ") not created, please create User before Profile"))
					return

				case count > 1:
					// some sql data error if any such error
					fmt.Println("SQL database error")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("500 - Internal Server Error"))
				}
			}

			// read the string sent to the service
			var newProfile mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &newProfile)
				fmt.Printf("UnMarshall :%+v\n", newProfile)

				// validate the JSON Num of keys-value in body are correct
				if len(newProfile) != len(dietProfMapRules) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Incompatible or Incomplete JSON Data Error"))
					return
				}

				// validate the JSON keys and value type in body are correct
				if ok, err := validateKeysValueTypes(newProfile, dietProfKeyTypeRules); !ok {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
					return
				}

				// struct value validaion with struct tag
				if ok, err := govalidator.ValidateMap(newProfile, dietProfMapRules); !ok {
					//if ok, err := govalidator.ValidateStruct(newFood); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
					return
				}

				// validate if ID exist
				_, err = dietProfGetOneRecord(params["uid"])
				if err == nil {
					w.WriteHeader(http.StatusConflict) // id key already exist
					w.Write([]byte("409 - Duplicate Diet Profile ID"))
					return
				} else {
					if err == errIllegalID {
						http.Error(w, "500 - Diet Profile Illegal Id", http.StatusNotFound)
					}
				}

				var dietProfile dietProfileType
				_, err := updateDietProfMapToStruct(&dietProfile, newProfile, dietProfMapRules)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
				}

				err = dietProfInsertRecord(db, dietProfile, params["uid"])
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
					return
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("201 - Diet Profile created for : " + params["uid"]))

				fmt.Println("Diet Profile :", dietProfile)

			} else {
				// Problem with the body from response
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please supply diet Profile Body " + "in JSON format"))
			}
		}

		//---PUT is for creating or updating exiting food item---
		// if r.Method == "PUT" {
		// 	// vakidate key for parameter key-value
		// 	if !userIsAdmin(key[0]) {
		// 		// w.WriteHeader(http.StatusNotFound)
		// 		// w.Write([]byte("401 - Invalid key"))
		// 		http.Error(w, "401 - - Unauthorized Access", http.StatusNotFound)
		// 		return
		// 	}

		// 	//var newFood foodType
		// 	var newFood mapInterface
		// 	reqBody, err := ioutil.ReadAll(r.Body)
		// 	if err == nil {
		// 		// parse JSON to object data structure
		// 		json.Unmarshal(reqBody, &newFood)

		// 		fmt.Printf("New Food JSON : %+v\n", newFood)

		// 		// validate the JSON Num of keys-value in body are correct
		// 		if len(newFood) != len(foodMapRules) {
		// 			w.WriteHeader(http.StatusUnprocessableEntity)
		// 			w.Write([]byte("422 - Incompatible or Incomplete JSON Data Error"))
		// 			return
		// 		}

		// 		// validate the JSON keys and value type in body are correct
		// 		if ok, err := validateKeysValueTypes(newFood, foodKeyTypeRules); !ok {
		// 			w.WriteHeader(http.StatusUnprocessableEntity)
		// 			w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
		// 			return
		// 		}

		// 		// struct value validaion with Map interface{} values
		// 		if ok, err := govalidator.ValidateMap(newFood, foodMapRules); !ok {
		// 			w.WriteHeader(http.StatusBadRequest)
		// 			w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
		// 			return
		// 		}

		// 		// check if there is a row for this record with the ID
		// 		count, err := foodGetRowCount(db, params["fid"])
		// 		if err != nil {
		// 			fmt.Println("Error", err)
		// 			w.WriteHeader(http.StatusInternalServerError)
		// 			w.Write([]byte("500 - Internal Server Error"))
		// 			return
		// 		}

		// 		fmt.Println("Count :", count)
		// 		fmt.Println("New Food", newFood)

		// 		switch {
		// 		case count == 0:
		// 			// Add row if none exist
		// 			// InsertRecord(db, &newFood, params["fid"])
		// 			// w.WriteHeader(http.StatusCreated)
		// 			// w.Write([]byte("201 - Food item added: " + params["fid"]))
		// 			http.Error(w, "404 - Food Id not found", http.StatusNotFound)

		// 		case count == 1:
		// 			// Edit row if row exist
		// 			var food foodType
		// 			_, err := updateFoodMapToStruct(&food, newFood, foodMapRules)
		// 			if err != nil {
		// 				w.WriteHeader(http.StatusBadRequest)
		// 				w.Write([]byte("422 - JSON Map Structure Error, " + err.Error()))
		// 			}

		// 			fmt.Println("new Food :", newFood)
		// 			err = foodEditRecord(db, newFood["Id"].(string), food, params["fid"])
		// 			if err != nil {
		// 				w.WriteHeader(http.StatusUnprocessableEntity)
		// 				w.Write([]byte("422 - JSON Body Error: " + err.Error()))
		// 				return
		// 			}
		// 			w.WriteHeader(http.StatusAccepted)
		// 			// w.Write([]byte("202 - Food item Updated: " + params["fid"] +
		// 			// 	" Category: " + newFood.Category + " Name:" + newFood.Name))
		// 			w.Write([]byte("202 - Food item Updated: " + params["fid"]))

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
		// }
		//---PATCH is for patching selective data ---
		if r.Method == "PATCH" {
			// if not admin key, then user authentication is needed
			if !userIsAdmin(key[0]) && !verifiedUser(key[0], params["uid"]) {
				http.Error(w, "401 - Unauthorized Access", http.StatusUnauthorized)
				return
			}

			var dietProfile mapInterface
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// parse JSON to object data structure
				json.Unmarshal(reqBody, &dietProfile)
				// validate the JSON keys and value type in body are correct
				if ok, err := validateKeysValueTypes(dietProfile, dietProfKeyTypeRules); !ok {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Data Type Error, " + err.Error()))
					return
				}

				// build rules dynamically base on interface{} because govalidator requires complete rules
				buildRules, err := buildVMapTemplate(dietProfile, dietProfMapRules)
				if err != nil {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Validation Build Rule Failed, " + err.Error()))
					return
				}
				fmt.Printf("Template : %+v\n", buildRules)

				// struct value validaion with Map interface{} values
				if ok, err := govalidator.ValidateMap(dietProfile, *buildRules); !ok {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("422 - JSON Data Value Error, " + err.Error()))
					return
				}

				fmt.Printf("Diet Profile : %+v\n", dietProfile)
				fmt.Printf("Diet Profile Size: %+v\n", len(dietProfile))

				// validate if ID exist
				_, err = dietProfGetOneRecord(params["uid"])
				if err != nil {
					if err == errIllegalID {
						http.Error(w, "500 - Diet Profile Illegal Id", http.StatusNotFound)
						return
					}
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Diet Profile ID Error, " + err.Error()))
					return
				}

				fmt.Printf("Diet Profile : %+v, %+v\n", dietProfile, params["uid"])

				// Edit row if row exist
				err = dietProfUpdateRecord(db, dietProfile, *buildRules, params["uid"])
				if err != nil {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - JSON Body Error, " + err.Error()))
					return
				}
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("202 - Diet Profile item is Patched: " + params["uid"]))

			} else {
				w.WriteHeader(http.StatusUnprocessableEntity) // error
				w.Write([]byte("422 - Please supply diet Profile Body " + "in JSON format"))
			}
		}
	}
}

// updateFoodMapToStruct extract the data in the Map and convert to struct type
// It checks for the validity of the key
// return (number of key found, err)
// Update p based fields of food interface, using foodFieldMap as key reference check
func updateDietProfMapToStruct(p *dietProfileType, food mapInterface, foodFieldMap mapInterface) (int, error) {

	var count int

	// Note "Id" key if present is also checked by not pu into p
	for k, v := range food {
		if _, ok := foodFieldMap[k]; !ok {
			return count, fmt.Errorf("invalid Key Used : (%v)", k)
		}
		count++
		fmt.Println("Key :", k)
		switch k {
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
