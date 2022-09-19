package main

import (
	"fmt"

	"encoding/json"

	"github.com/go-redis/redis"
)

type Author struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {

	// create a Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // no timeout
	})

	// Testing with ping
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	// we can call set with a `Key` and a `Value`.
	err = client.Set("name", "Elliot", 0).Err()
	// if there has been an error setting the value
	// handle the error
	if err != nil {
		fmt.Println(err)
	}

	// Get the value
	val, err := client.Get("name").Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val)

	// create a composite data -JSON data
	json, err := json.Marshal(Author{Name: "Elliot", Age: 25})
	if err != nil {
		fmt.Println(err)
	}

	// set it to Json data to Id
	err = client.Set("id1234", json, 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	// Get the composite data
	val, err = client.Get("id1234").Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val)
}
