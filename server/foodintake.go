package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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

// Note JSON field needs to be exported to encoding/json to enable Encoding/Decoding, so it has to be in CAPITAL
type nutriMetricVsProfile struct {
	Energy      string `json:"energy"`
	Protein     string `json:"protein"`
	FatTotal    string `json:"fatTotal"`
	FatSat      string `json:"fatSat"`
	Fibre       string `json:"fibre"`
	Carb        string `json:"carb"`
	Cholesterol string `json:"cholesterol"`
	Sodium      string `json:"sodium"`
}

// JSON reply for MyProfile in Map of replyType
type replyType struct {
	Actual float32 `json:"actual"`
	Target float32 `json:"target"`
	Status string  `json:"status"`
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
	v := r.URL.Query()
	//fmt.Printf("v :%+v", v)
	key, ok := v["key"]
	if !ok {
		http.Error(w, "401 - Missing key in URL", http.StatusNotFound)
		return
	}

	// vakidate key for parameter key-value
	userX, ok := userIsRegistered(key[0])
	if !ok {
		// w.WriteHeader(http.StatusNotFound)
		// w.Write([]byte("401 - Invalid key"))
		http.Error(w, "401 - Invalid key", http.StatusNotFound)
		return
	}

	fmt.Printf("User = %+v\n", userX)

	//{{baseURL}}/foodIntake/{select} can be "Value" or "Metric"
	// get route variable of map[string]string
	params := mux.Vars(r)
	fmt.Printf("Parameter = %+v\n", params)
	fmt.Println("Select ", params["select"])

	// ?key={{urlKey}}&gender=male&food=CHN0001,132&food=CHN0002,18&food=CHN0003,16&food=CHN0004,22&food=CHN0005,18
	// URL parameter is where we can get all the parameters
	vMap := r.URL.Query()
	fmt.Printf("URL Query : %+v\n", vMap)
	fmt.Printf("food :%+v\n", vMap["food"])
	count := len(vMap["food"])
	fmt.Println("Number of food parameters :", count)

	// use make or directly initialise an empty map
	foodlist := map[string]float32{}

	// Processing the URL stream to ensure it is in the expected format
	// convert parameter to parameter map
	for _, v := range vMap["food"] {
		s := strings.Split(v, ",")
		// There should be 2 values per parameter
		if len(s) != 2 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Missing weight parameter with '&Id=FoodID,weight(gram)' for every food item"))
			return
		}

		// ID in s[0], weight in S[1]
		fmt.Println("S:", s)
		if fvalue, err := strconv.ParseFloat(s[1], 32); err == nil {
			foodlist[s[0]] = (float32)(fvalue)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Error in weight parameter, please supply parameter with '&Id=FoodID,weight(gram)' for every food item"))
			return
		}

	}

	//fmt.Printf("foodlist : %#v\n", foodlist)

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
			food, err := foodGetOneRecord(k) // get food data

			//fmt.Printf("food : %#v, v: %+v\n", food, v)

			if err != nil {
				http.Error(w, "404 - Invalid food id in parameter", http.StatusNotFound)
			}
			computeNutrientValue(&valueTotal, food, v)
		}

		fmt.Printf("valueTotal = %+v\n", valueTotal)
		json.NewEncoder(w).Encode(&valueTotal) //key:value

	case "METRIC":

		gender := strings.ToUpper(userX.gender)
		// validate gender parameter is ok
		dailyLimit, ok := foodDailyLimit[gender]
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Wrong Gender parameter value"))
			return
		}

		// compute total nutrient value
		for k, v := range foodlist {
			food, err := foodGetOneRecord(k) // get food data

			//fmt.Printf("food : %#v, v: %+v\n", food, v)

			if err != nil {
				http.Error(w, "404 - Invalid food id in parameter", http.StatusNotFound)
			}

			computeNutrientValue(&valueTotal, food, v)
		}

		// Compute Metric for each value against daily limit
		// initialise metricTotal to zero
		metricTotal := valueTotal
		computeNutrientMetric(&metricTotal, &dailyLimit)
		fmt.Printf("MetricTotal = %+v\n", metricTotal)

		json.NewEncoder(w).Encode(&metricTotal) //key:value

	case "MYPROFILE":

		gender := strings.ToUpper(userX.gender)
		// validate gender parameter is ok
		dailyLimit, ok := foodDailyLimit[gender]
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Wrong Gender parameter value"))
			return
		}

		// compute total nutrient value
		for k, v := range foodlist {
			food, err := foodGetOneRecord(k) // get food data

			//fmt.Printf("food : %#v, v: %+v\n", food, v)

			if err != nil {
				http.Error(w, "404 - Invalid food id in parameter", http.StatusNotFound)
			}

			computeNutrientValue(&valueTotal, food, v)
		}

		// Compute Metric for each value against daily limit
		// initialise metricTotal to zero
		metricTotal := valueTotal
		computeNutrientMetric(&metricTotal, &dailyLimit)
		fmt.Printf("MetricTotal = %+v\n", metricTotal)

		// Get user's diet profile
		profile, err := dietProfGetOneRecord(userX.email)

		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Error in Diet Profile ID, " + err.Error()))
			return
		}

		// create the reply message
		replyMap := make(map[string]replyType)

		// use reflect to get info about the struct fields
		metaValue1 := reflect.ValueOf(&metricTotal).Elem()
		metaValue2 := reflect.ValueOf(profile).Elem()

		fmt.Println("Meta Value 1:", metaValue1)
		fmt.Println("Meta Value 2:", metaValue2)

		// scan through the struct field list
		for i := 0; i < metaValue1.NumField(); i++ {
			field := replyType{}

			name := metaValue1.Type().Field(i).Name
			field.Actual = metaValue1.FieldByName(name).Interface().(float32)
			field.Target = metaValue2.FieldByName(name).Interface().(float32)
			if field.Actual > field.Target {
				field.Status = "Failed Target"
			} else {
				field.Status = "Met Target"
			}
			replyMap[name] = field
		}

		fmt.Println("Reply :", replyMap)
		json.NewEncoder(w).Encode(&replyMap) //key:value
	}
}

// Extrapolate the Nutrient Value based on the actual weight
func computeNutrientValue(valueTotal *nutriValue, food *foodType, actualWt float32) {
	valueTotal.Energy += food.Energy * actualWt / food.Weight
	valueTotal.Protein += food.Protein * actualWt / food.Weight
	valueTotal.FatTotal += food.FatTotal * actualWt / food.Weight
	valueTotal.FatSat += food.FatSat * actualWt / food.Weight
	valueTotal.Fibre += food.Fibre * actualWt / food.Weight
	valueTotal.Carb += food.Carb * actualWt / food.Weight
	valueTotal.Cholesterol += food.Cholesterol * actualWt / food.Weight
	valueTotal.Sodium += food.Sodium * actualWt / food.Weight
}

// Extrapolate the Nutrient Value based on the actual weight
func computeNutrientMetric(metricTotal *nutriValue, dailyLimit *nutriValue) {
	metricTotal.Energy /= dailyLimit.Energy
	metricTotal.Protein /= dailyLimit.Protein
	metricTotal.FatTotal /= dailyLimit.FatTotal
	metricTotal.FatSat /= dailyLimit.FatSat
	metricTotal.Fibre /= dailyLimit.Fibre
	metricTotal.Carb /= dailyLimit.Carb
	metricTotal.Cholesterol /= dailyLimit.Cholesterol
	metricTotal.Sodium /= dailyLimit.Sodium
}
