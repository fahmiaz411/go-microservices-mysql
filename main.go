package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// go get -u github.com/go-sql-driver/mysql if package is missing!
// If you encounter problems like I did with a newer version of Go. Run the following:
// GO111MODULE="off" go get github.com/go-sql-driver/mysql
// Restart IDE

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getUsers() []*User {
	// Open up our database connection.
	db, err := sql.Open("mysql", "root:mypassword@tcp(db:3306)/testdb")

	// if there is an error opening the connection, handle it
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	// Execute the query
	results, err := db.Query("SELECT * FROM users")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	var users []*User
	for results.Next() {
		var u User
		// for each row, scan the result into our tag composite object
		err = results.Scan(&u.ID, &u.Name)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		users = append(users, &u)
	}

	return users
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func userPage(w http.ResponseWriter, r *http.Request) {
	users := getUsers()

	fmt.Println("Endpoint Hit: usersPage")
	json.NewEncoder(w).Encode(users)
}

func main() {
	// Connect to the database with the name of the database container and it's login details.
	fmt.Println("Connecting to db")
	var conn, err = sql.Open("mysql", "root:mypassword@tcp(db:3306)/testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// MySQL server isn't fully active yet.
	// Block until connection is accepted. This is a docker problem with v3 & container doesn't start
	// up in time.
	for conn.Ping() != nil {
		fmt.Println("Attempting connection to db")
		time.Sleep(5 * time.Second)
	}
	fmt.Println("Connected")

	// Optional: Below is a quick demo.
	// drop table if it already exists
	fmt.Println("Dropping table")
	_, err = conn.Exec(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		panic(err)
	}

	// create a new table
	fmt.Println("Creating table")
	_, err = conn.Exec(`
	CREATE TABLE users (
		id int auto_increment primary key,
		name varchar(255)
	);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert into the new table
	fmt.Println("Inserting person")
	_, err = conn.Exec("INSERT INTO users (name) VALUES ('fahmi')")
	if err != nil {
		log.Fatal(err)
	}

	// Create struct to store data assuming non NULL values for testing purposes.
	var person struct {
		ID        int
		LastName  string
		FirstName string
		Address   string
		City      string
	}

	// Get all the users
	fmt.Println("Getting person")
	result, err := conn.Query("SELECT * FROM People;")
	if err != nil {
		log.Fatal(err)
	}
	defer result.Close()

	// Get the results and store them in person.
	if result.Next() {
		err = result.Scan(
			&person.ID,
			&person.LastName,
			&person.FirstName,
			&person.Address,
			&person.City,
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v", person)
	}

	// For testing only - sleep to keep container alive.
	// time.Sleep(1 * time.Minute)

	http.HandleFunc("/", homePage)
	http.HandleFunc("/users", userPage)

	log.Fatal(http.ListenAndServe(":8080", nil))
}