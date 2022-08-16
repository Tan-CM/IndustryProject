// Lab 3 This is the server implementation for REST API
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var urlKey string
var hostPort string

//var sqlDBConnection string
var cfg mysql.Config // configuration for DSN

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
	cfg = mysql.Config{
		User:   os.Getenv("SQL_USER"),
		Passwd: os.Getenv("SQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("SQL_ADDR"),
		DBName: os.Getenv("SQL_DB"),
	}
}

// main() main function to start the http multiplexer
// maps URI resource to the handler
func main() {
	// Initialise Food Read Cache
	foodCacheInit()

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
	// passing a variable into a path as a value in {} to the next slash, use curly braces {for fid}
	// This var in {fid} is the key defined map of mux.Var(requester) by Gorilla
	//.Methods limit the allow methods
	// can use regex to filter variable, instead of allowing any fid to pass
	//router.HandleFunc("/api/v1/foods/{fid:IOT\\d{3}}", course).Methods("GET", "PUT", "POST", "DELETE")
	//router.HandleFunc("/api/v1/foods/{fid}", course).Methods("GET", "PUT", "POST", "DELETE")
	subrouter.HandleFunc("/foods/{fid}", food).Methods("GET", "PUT", "POST", "DELETE")
	// if .Method is not defined, all methods are allowed
	// note more than one key can be used, so mux.Vars contains the key-value pairs

	// Nutrition
	// Max 5 variable for IDs allowed
	subrouter.HandleFunc("/foodIntake/{select}", foodTotal).Methods("GET")

	fmt.Printf("Listening at %s\n", hostPort)
	// log.Fatal(http.ListenAndServe(":5000", router))
	//log.Fatal(http.ListenAndServe(hostPort, router))
	log.Fatal(http.ListenAndServe(hostPort, subrouter))
}
