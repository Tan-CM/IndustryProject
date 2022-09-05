// Lab 3 This is the server implementation for REST API
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var m2 = sync.Mutex{}

var errDuplicateID = errors.New("Duplicate ID")

//var errJSONMapKeys = errors.New("Errors in JSON Map Keys")

// Use this as a fast cache to verify the user key
// Map only has the user type information
type userT struct {
	email      string
	accessType string
}

// Maps Access Key to user data
var userMap = map[string]userT{}

// validate only admin users
func userIsUser(key string) bool {
	v, ok := userMap[key]
	if ok && v.accessType == "user" {
		return true
	}
	return false
}

// validate only non-admin uses
func userIsAdmin(key string) bool {
	v, ok := userMap[key]
	if ok && v.accessType == "admin" {
		return true
	}
	return false
}

// // validate all registered user
// func userIsRegisteredx(key string) bool {

// 	fmt.Println("Key :", key)
// 	_, ok := userMap[key]
// 	return ok
// }

// validate all registered user, and get user data
func userIsRegistered(key string) (userT, bool) {

	v, ok := userMap[key]
	return v, ok
}

// Check user or admin identity match with access key
func verifiedUser(key string, Id string) bool {

	fmt.Println("Key :", key)
	if v, ok := userMap[key]; ok {
		return v.email == Id
	}
	return false // non existent key
}

func userCacheInit() {
	db, err := sql.Open("mysql", cfgUser.FormatDSN())
	// handle error
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
	err = userGetRecordsInitDB(db)
	if err != nil {
		panic(err.Error()) // panic because server cannot function
	}
}

func userGetRecordsInitDB(db *sql.DB) error {

	m2.Lock()
	defer m2.Unlock()

	// query to get all rows of table (persons) of my_db
	rows, err := db.Query("Select * FROM Users")
	if err != nil {
		fmt.Println("Error initial reading from SQL User Table")
		//panic(err.Error())
		return err
	}
	defer rows.Close()

	//var id string

	// extract row by row to create
	for rows.Next() {
		// map this type to the record in the table
		var user userType
		err = rows.Scan(&user.Email, &user.Name, &user.AccessKey, &user.Type)

		if err != nil {
			fmt.Println("Error reading rows from SQL User!!!")
			//panic(err.Error())
			return err
		}
		//fmt.Printf("User : %+v\n", user)

		// initialise map
		var userX userT
		userX.email = user.Email
		userX.accessType = user.Type

		userMap[user.AccessKey] = userX
	}

	fmt.Printf("UserMap : %+v\n", userMap)
	return nil
}

// // GetRecords gets all the rows of the current table and return as a slice of map
func userGetRecordsDB(db *sql.DB) (*usersType, error) {
	m2.Lock()
	defer m2.Unlock()

	// query to get all rows of table (persons) of my_db
	rows, err := db.Query("Select * FROM Users")
	if err != nil {
		fmt.Println("Error initial reading from SQL User Table")
		//panic(err.Error())
		return nil, err
	}
	defer rows.Close()

	var usersData usersType
	var userList = []userType{}

	// extract row by row to create
	for rows.Next() {
		// map this type to the record in the table
		var user userType
		err = rows.Scan(&user.Email, &user.Name, &user.AccessKey, &user.Type)

		if err != nil {
			fmt.Println("Error reading rows from SQL User!!!")
			//panic(err.Error())
			return nil, err
		}
		//fmt.Printf("User : %+v\n", user)
		userList = append(userList, user)
	}
	usersData.Count = len(userList)
	usersData.Users = userList

	//fmt.Printf("UserDataMap : %+v\n", usersData)

	return &usersData, nil
}

// GetOneRecord checks if there is a existence of a record based on the ID primary key
// If there is a record, it returns a map of the record key:title pair
// error = nil, there is a record
// error = emptyRow, there is no record
// func userGetOneRecordxx(ID string) (*userT, error) {

// 	if len(ID) == 0 {
// 		return nil, errIllegalID
// 	}

// 	// check for validity of ID
// 	if v, ok := userMap[ID]; ok {
// 		return &v, nil
// 	} else {
// 		return &v, fmt.Errorf("invalid ID (%v)", ID)
// 	}

// }

func userGetOneRecordDB(db *sql.DB, ID string) (*userType, error) {

	fmt.Println("ID :", ID)
	row, err := db.Query("Select * FROM USERS where EMAIL=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	var user userType
	// extract row by row to create slice of foodType
	for row.Next() {
		// map this type to the record in the table

		err = row.Scan(&user.Email, &user.Name, &user.AccessKey, &user.Type)

		if err != nil {
			fmt.Println("Error reading rows from user DB!!!")
			//panic(err.Error())
			return &user, err
		}
	}
	return &user, nil
}

// Returns the number of rows that match the ID
func userGetRowCountDB(db *sql.DB, ID string) (int, error) {

	fmt.Println("ID :", ID)
	row, err := db.Query("Select count(*) FROM USERS where EMAIL=?", ID)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	if row.Next() {
		var count int
		err = row.Scan(&count)
		if err != nil {
			fmt.Println("Error reading rows from user DB!!!")
			panic(err.Error())
		}
		return count, nil
	} else {
		return -1, errEmptyRow
	}
}

// DeleteRecord deletes a record from the current table using the ID primary key
func userDeleteRecordDB(db *sql.DB, ID string) {
	m2.Lock()
	defer m2.Unlock()

	// create the sql query to delete with primary key
	// Note deleting a non-existent record is considered as deleted, so will always passed

	//query := fmt.Sprintf("DELETE FROM foods WHERE ID='%s'", ID)
	//row, err := db.Query(query)
	user, err := userGetOneRecordDB(db, ID)
	if err != nil {
		panic(err.Error())
	}

	row, err := db.Query("DELETE FROM USERS WHERE EMAIL=?", ID)

	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Delete User Key from Read Cache
	delete(userMap, user.AccessKey)

	fmt.Println("Delete Successful")
}

// InsertRecord instead a row record into the current table based on the primary key and title
func userInsertRecordDB(db *sql.DB, p *userType) {
	m2.Lock()
	defer m2.Unlock()

	// create the sql query to insert record
	// note the quote for string
	// query := fmt.Sprintf("INSERT INTO foods VALUES ('%s', '%s')", ID, title)
	// _, err := db.Query(query)
	// Id is auto generated by SQL
	row, err := db.Query("INSERT INTO USERS VALUES (?, ?, ?, ?)",
		p.Email, p.Name, p.AccessKey, p.Type)
	if err != nil {
		panic(err.Error())
	}
	defer row.Close()

	// Add Read Cache with New record
	var userX userT
	userX.email = p.Email
	userX.accessType = p.Type

	userMap[p.AccessKey] = userX

	fmt.Println("UserMap ", userMap)
	fmt.Println("Insert Successful")
}

func userUpdateRecordDB(db *sql.DB, user mapInterface, ID string) error {
	m2.Lock()
	defer m2.Unlock()

	// Read the old record first
	userTemp, err := userGetOneRecordDB(db, ID)
	if err != nil {
		panic(err.Error())
	}

	// if email is to be changed, then, need to check that new email is unique before making change to db
	if v, ok := user["email"]; ok {
		// Need to check if the new email is unique
		if v.(string) != ID {
			count, err := userGetRowCountDB(db, v.(string))
			if err != nil {
				panic(err.Error())
			}
			if count != 0 {
				return errDuplicateID // Duplicated ID
			}
		}
		userTemp.Email = v.(string)
		if err != nil {
			panic(err.Error())
		}
	}

	if v, ok := user["name"]; ok {
		userTemp.Name = v.(string)
		if err != nil {
			panic(err.Error())
		}
	}

	// Update email ID in DB
	row, err := db.Query("Update USERS SET Email=?, Name=? where Email=?", userTemp.Email, userTemp.Name, ID)
	if err != nil {
		panic(err.Error())
	}

	// Update User cache with new email, using the old Access key
	newUser := userMap[userTemp.AccessKey] // copy old data first
	newUser.email = userTemp.Email         // change email

	// Update map only for email change
	userMap[userTemp.AccessKey] = newUser
	//}

	defer row.Close()

	fmt.Println("Update Successful")
	return nil
}
