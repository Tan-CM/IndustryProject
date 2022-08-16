package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type nutriValue struct {
	Energy      float32 `json:"energy"`
	Protein     float32 `json:"protein"`
	FatTotal    float32 `json:"fatTotal"`
	FatSat      float32 `json:"fatSat"`
	Fibre       float32 `json:"fibre"`
	Carb        float32 `json:"carb"`
	Cholesterol float32 `json:"cholesterol"`
	Sodium      float32 `json:"sodium"`
}

type dailyNutrMetric struct {
	Energy      float32 `json:"energy"`
	Protein     float32 `json:"protein"`
	FatTotal    float32 `json:"fatTotal"`
	FatSat      float32 `json:"fatSat"`
	Fibre       float32 `json:"fibre"`
	Carb        float32 `json:"carb"`
	Cholesterol float32 `json:"cholesterol"`
	Sodium      float32 `json:"sodium"`
}

// Recommended Max Daily Intake for Singapore (www.healthhub.sg)
var foodDailyLimit = map[string]nutriValue{
	"MALE": {
		Energy:      2500, //g
		Protein:     56,   //g
		FatTotal:    70,   //g
		FatSat:      21,   //g
		Fibre:       26,   //g
		Carb:        300,  //g
		Cholesterol: 300,  //mg
		Sodium:      2000, //mg
	},
	"FEMALE": {
		Energy:      2000, //g
		Protein:     46,   //g
		FatTotal:    56,   //g
		FatSat:      17,   //g
		Fibre:       20,   //g
		Carb:        240,  //g
		Cholesterol: 240,  //mg
		Sodium:      2000, //mg
	},
}

// food() is the hanlder for "/api/v1/foods/{fid}" resource
func foodTotal(w http.ResponseWriter, r *http.Request) {

	// vakidate key for parameter key-value
	if !validKey(r) {
		// w.WriteHeader(http.StatusNotFound)
		// w.Write([]byte("401 - Invalid key"))
		http.Error(w, "401 - Invalid key", http.StatusNotFound)
		return
	}

	// // Use mysql as driverName and a valid DSN as dataSourceName:
	// // user set up password that can access this db connection
	// //db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:58710)/foodDB")
	// //db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:58710)/foodDB")
	// db, err := sql.Open("mysql", cfg.FormatDSN())

	// // handle error
	// if err != nil {
	// 	panic(err.Error())
	// }
	// // defer the close till after the main function has finished executing
	// defer db.Close()
	// //	fmt.Println("Database opened")

	// mux.Vars(r) is the variable immediately after the URL
	// it can have more than one parameters
	params := mux.Vars(r)
	fmt.Printf("Parameter = %+v\n", params)
	fmt.Println("Select ", params["select"])

	// URL parameter is where we can get the parameter
	vMap := r.URL.Query()
	fmt.Printf("URL Query : %+v\n", vMap)
	fmt.Printf("Id :%+v\n", vMap["Id"])
	count := len(vMap["Id"])
	fmt.Println("Number of ID parameters :", count)

	// use make or directly initialise an empty map
	foodlist := map[string]float32{}

	// Processing the URL stream to ensure it is in the expected format
	// convert parameter to parameter map
	for _, v := range vMap["Id"] {
		s := strings.Split(v, ",")
		// There should be 2 values per parameter
		if len(s) != 2 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Missing weight parameter with '&Id=FoodID,weight(gram)' for every food item"))
			return
		}

		// weight in S[1]
		fmt.Println("S:", s)
		if fvalue, err := strconv.ParseFloat(s[1], 32); err == nil {
			foodlist[s[0]] = (float32)(fvalue)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Error in weight parameter, please supply parameter with '&Id=FoodID,weight(gram)' for every food item"))
			return
		}

	}

	fmt.Printf("foodlist : %#v\n", foodlist)

	var err error

	db, err := sql.Open("mysql", cfg.FormatDSN())

	// handle error
	if err != nil {
		panic(err.Error())
	}

	var food productType
	// if Id pattern is not found, then use the ID directly
	bufferMap := &map[string]productType{}

	// number of ID matches the number of valid ID and values
	if count != len(foodlist) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("422 - Error in parameter, please supply parameter with '&Id=FoodID,weight(gram)' for every food item"))
		return
	}

	// initialise valueTotal to zero
	valueTotal := nutriValue{}

	// End point determines the return content
	switch strings.ToUpper(params["select"]) {
	case "VALUE":

		for k, v := range foodlist {
			bufferMap, err = GetOneRecord(db, k) // get food data
			food = (*bufferMap)[k]

			fmt.Printf("food : %#v, v: %+v\n", food, v)

			if err != nil {
				http.Error(w, "404 - Invalid food id in parameter", http.StatusNotFound)
			}

			valueTotal.Energy += food.Energy * v / food.Weight
			valueTotal.Protein += food.Protein * v / food.Weight
			valueTotal.FatTotal += food.FatTotal * v / food.Weight
			valueTotal.FatSat += food.FatSat * v / food.Weight
			valueTotal.Fibre += food.Fibre * v / food.Weight
			valueTotal.Carb += food.Carb * v / food.Weight
			valueTotal.Cholesterol += food.Cholesterol * v / food.Weight
			valueTotal.Sodium += food.Sodium * v / food.Weight
		}

		fmt.Printf("valueTotal = %+v\n", valueTotal)
		json.NewEncoder(w).Encode(&valueTotal) //key:value

	case "METRIC":

		// check gender parameter exist
		value, ok := vMap["gender"]
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Missing Gender parameter, use &gender=male or &gender=female"))
			return
		}

		fmt.Printf("Gender : %+v\n", value)
		gender := strings.ToUpper(value[0])

		// validate gender parameter is ok
		dailyLimit, ok := foodDailyLimit[gender]
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Wrong Gender parameter value, use &gender=male or &gender=female"))
			return
		}

		// compute total nutrient value
		for k, v := range foodlist {
			bufferMap, err = GetOneRecord(db, k) // get food data
			food = (*bufferMap)[k]

			fmt.Printf("food : %#v, v: %+v\n", food, v)

			if err != nil {
				http.Error(w, "404 - Invalid food id in parameter", http.StatusNotFound)
			}

			valueTotal.Energy += food.Energy * v / food.Weight
			valueTotal.Protein += food.Protein * v / food.Weight
			valueTotal.FatTotal += food.FatTotal * v / food.Weight
			valueTotal.FatSat += food.FatSat * v / food.Weight
			valueTotal.Fibre += food.Fibre * v / food.Weight
			valueTotal.Carb += food.Carb * v / food.Weight
			valueTotal.Cholesterol += food.Cholesterol * v / food.Weight
			valueTotal.Sodium += food.Sodium * v / food.Weight
		}

		// Compute Metric for each value against daily limit
		// initialise metricTotal to zero
		metricTotal := valueTotal

		metricTotal.Energy /= dailyLimit.Energy
		metricTotal.Protein /= dailyLimit.Protein
		metricTotal.FatTotal /= dailyLimit.FatTotal
		metricTotal.FatSat /= dailyLimit.FatSat
		metricTotal.Fibre /= dailyLimit.Fibre
		metricTotal.Carb /= dailyLimit.Carb
		metricTotal.Cholesterol /= dailyLimit.Cholesterol
		metricTotal.Sodium /= dailyLimit.Sodium
		fmt.Printf("MetricTotal = %+v\n", metricTotal)
		json.NewEncoder(w).Encode(&metricTotal) //key:value
	}

	//	fmt.Printf("Daily Limit = %+v", foodDailyLimit)
}
