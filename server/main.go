// Lab 3 This is the server implementation for REST API
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var urlKey string
var hostPort string

//var sqlDBConnection string
var cfgFood mysql.Config     // configuration for DSN
var cfgDietProf mysql.Config // configuration for DSN
var cfgUser mysql.Config     // configuration for DSN

// init() initialises the system
// Set up the environment
// Note:  For this exercise, both client and server uses the same .env
// In actual deployment, the .env file will be not be shared
func init() {

	// set path for the env file
	envFile := path.Join("..", "config", ".env")

	//err := godotenv.Load(".env")
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalln("Error loading .env file: ", err)
	}

	// getting env variables SITE_TITLE and DB_HOST
	siteTitle := os.Getenv("SERVER_TITLE")
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	urlKey = os.Getenv("SERVER_URLKEY")

	// Create Host Port from environment variable
	hostPort = fmt.Sprintf("%s:%s", serverHost, serverPort)

	fmt.Printf("Site Title = %s\n", siteTitle)
	fmt.Printf("Use http:// %s\n", hostPort)

	// SQL DB Data Source Name config
	cfgFood = mysql.Config{
		User:   os.Getenv("SQL_USER"),
		Passwd: os.Getenv("SQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("SQL_ADDR"),
		DBName: os.Getenv("SQL_foodDB"),
		//DBName: "foodDB",
	}

	// SQL DB Data Source Name config
	cfgDietProf = mysql.Config{
		User:   os.Getenv("SQL_USER"),
		Passwd: os.Getenv("SQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("SQL_ADDR"),
		DBName: os.Getenv("SQL_dietProfileDB"),
		//DBName: "dietProfileDB",
	}

	// SQL DB Data Source Name config
	cfgUser = mysql.Config{
		User:   os.Getenv("SQL_USER"),
		Passwd: os.Getenv("SQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("SQL_ADDR"),
		DBName: os.Getenv("SQL_userDB"),
		//DBName: "userDB",
	}

	//	govalidator.SetFieldsRequiredByDefault(true)
	// to differential nil and zero value
	govalidator.SetNilPtrAllowedByRequired(true)
}

// main() main function to start the http multiplexer
// maps URI resource to the handler
func main() {
	// Initialise Food Read Cache
	readCacheInit()

	router := mux.NewRouter()
	// can use to restrict to certain host
	//router.Host("www.example.com")

	// Note URL path is case sensitive, so exact path is needed to get it to work
	// register URL paths to handlers
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	//router.HandleFunc("/api/v1/", home)
	//router.HandleFunc("/api/v1/foods", allfoods)\
	subrouter.HandleFunc("/", home).Methods("GET")
	subrouter.HandleFunc("/foods", allfoods).Methods("GET")
	subrouter.HandleFunc("/food/{fid}", food).Methods("GET", "PUT", "POST", "DELETE", "PATCH")
	// if .Method is not defined, all methods are allowed
	// note more than one key can be used, so mux.Vars contains the key-value pairs

	// User food Preference
	subrouter.HandleFunc("/dietProfile/{uid}", dietUserProfile).Methods("GET", "POST", "PATCH", "DELETE")
	//	subrouter.HandleFunc("/dietProfile/{uid}", dietSelectUserProfile).Methods("GET", "POST", "PATCH", "DELETE")

	// FoodIntake
	// {select = {Metric, Value}}
	subrouter.HandleFunc("/foodIntake/{select}", foodTotal).Methods("GET")

	// Users
	subrouter.HandleFunc("/users", users).Methods("GET")
	subrouter.HandleFunc("/user/{uid}", user).Methods("GET", "PATCH", "POST", "DELETE")

	fmt.Printf("Listening at %s\n", hostPort)
	// log.Fatal(http.ListenAndServe(":5000", router))
	//log.Fatal(http.ListenAndServe(hostPort, router))
	log.Fatal(http.ListenAndServe(hostPort, subrouter))
}

// initialise Data in cache
func readCacheInit() {
	userCacheInit()
	foodCacheInit()
	dietProfCacheInit()
}
