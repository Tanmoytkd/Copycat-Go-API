package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/shomali11/util/xhashes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	// We need a flag for setting the listening address.
	// We set the default to port 8080, which is a common HTTP port
	// for servers with local-only access.
	addr = flag.String("addr", ":8080", "http service address")

	db *sql.DB
)

//this function checks for error and throws a runtime
//error if an error is found
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}



// ## main
func main() {
	var currentTime time.Time = time.Now()
	fmt.Println(currentTime)


	// The main function starts by parsing the commandline.
	flag.Parse()

	db = connectDatabase()
	defer db.Close()

	// Now we can create a new `Router` instance...
	r := mux.NewRouter()

	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("POST")

	r.HandleFunc("/sessions/create", createSession).Methods("POST")

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintln(writer, "Hello there")
	})

	fmt.Println("Listening on port", *addr)

	// Finally, we just have to start the http Server. We pass the listening address
	// as well as our router instance.
	err := http.ListenAndServe(*addr, r)

	// For this demo, let's keep error handling simple.
	// `log.Fatal` prints out an error message and exits the process.
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func isValidToken(token string) bool  {
	accountChecker, err := db.Prepare("SELECT EXISTS (SELECT `id` FROM `accounts` WHERE `token`=?)");
	checkError(err)

	alreadyRegistered := false
	err = accountChecker.QueryRow(token).Scan(&alreadyRegistered)
	checkError(err)
	return alreadyRegistered
}

func renewToken(oldToken string) string {
	newToken, err := GenerateRandomStringURLSafe(256)
	checkError(err)

	_, err = db.Exec("UPDATE `accounts` SET `token`=? WHERE `token`=?", newToken, oldToken)
	checkError(err)
	return newToken
}

// ## The handler functions

// Let's implement the show function now. Typically, handler functions receive two parameters:
//
// * A Response Writer, and
// * a Request object.
//

func register(writer http.ResponseWriter, request *http.Request) {
	type AccountInfo struct {
		Name, Email, Password string
	}

	data, _ := ioutil.ReadAll(request.Body)

	fmt.Println("input data: ", string(data))

	var account AccountInfo
	json.Unmarshal(data, &account)

	account.Password = xhashes.SHA256(account.Password)

	token, err := GenerateRandomStringURLSafe(256)
	checkError(err)

	accountChecker, err := db.Prepare("SELECT EXISTS (SELECT `id` FROM `accounts` WHERE `email`=?)");
	checkError(err)

	alreadyRegistered := false
	err = accountChecker.QueryRow(account.Email).Scan(&alreadyRegistered)
	checkError(err)
	if alreadyRegistered {
		writer.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(writer, "An account has already been registered from "+account.Email)
		return
	}

	stmt, err := db.Prepare("INSERT INTO `accounts` (`id`, `name`, `email`, `password`, `token`) VALUES (NULL, ?, ?, ?, ?);")
	checkError(err)
	_, err = stmt.Exec(account.Name, account.Email, account.Password, token)
	checkError(err)

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "{\"token\":\""+token+"\"}")

}

func logout(writer http.ResponseWriter, request *http.Request) {
	type AccountInfo struct {
		Token string
	}

	data, _ := ioutil.ReadAll(request.Body)

	fmt.Println("input data: ", string(data))

	var account AccountInfo
	json.Unmarshal(data, &account)

	renewToken(account.Token)


	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "Logged out successfully")
}

func login(writer http.ResponseWriter, request *http.Request) {
	type AccountInfo struct {
		Email, Password string
	}

	data, _ := ioutil.ReadAll(request.Body)

	fmt.Println("input data: ", string(data))

	var account AccountInfo
	json.Unmarshal(data, &account)

	account.Password = xhashes.SHA256(account.Password)

	accountChecker, err := db.Prepare("SELECT EXISTS (SELECT `id` FROM `accounts` WHERE `email`=?)");
	checkError(err)

	alreadyRegistered := false
	err = accountChecker.QueryRow(account.Email).Scan(&alreadyRegistered)
	checkError(err)
	if !alreadyRegistered {
		writer.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(writer, "Please create an account first.")
		return
	}

	passChecker, err := db.Prepare("SELECT EXISTS (SELECT `id` FROM `accounts` WHERE `email`=? AND `password`=?)");
	checkError(err)

	passwordCorrect := false
	err = passChecker.QueryRow(account.Email, account.Password).Scan(&passwordCorrect)
	checkError(err)

	if !passwordCorrect {
		writer.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(writer, "Email or Password incorrect.")
		return
	}

	var id int

	stmt, err := db.Prepare("SELECT `id` FROM `accounts` WHERE `email`=? AND `password`=?")
	checkError(err)
	err = stmt.QueryRow(account.Email, account.Password).Scan(&id)
	checkError(err)

	token, err := GenerateRandomStringURLSafe(256)
	checkError(err)

	_, err = db.Exec("UPDATE `accounts` SET `token`=? WHERE `id`=?", token, id)
	checkError(err)

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "{\"token\":\""+token+"\"}")
}

func createSession(writer http.ResponseWriter, request *http.Request) {
	type SessionInfo struct {
		Token string
		Name string
		Code string
		Password string
		Data string
	}

	data, _ := ioutil.ReadAll(request.Body)

	fmt.Println("input data: ", string(data))

	var account SessionInfo
	json.Unmarshal(data, &account)

	if !isValidToken(account.Token) {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(writer, "Please login to create Sessions.")
		return
	}

	account.Password = xhashes.SHA256(account.Password)
	hash := xhashes.SHA256(account.Code+"$"+account.Password)


	_, err := db.Exec("INSERT INTO `collab_sessions`(`name`, `code`, `password`, `data`, `hash`) VALUES (?, ?, ?, ?, ?)",
		account.Name, account.Code, account.Password, account.Data, hash)
	checkError(err)

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "{\"message\": \"Session \""+account.Name+"\" created successfully\", " +
		"\"hash\":\""+hash+"\"")
}



/*
After saving, we can run the code locally by calling
```
go run rest.go
```
TEST USING CURL:
curl -X POST -H "Content-Type: application/json" -d '{"Name": "Tanmoy Krishna Das", "Email":"tanmoykrishnadas@gmail.com", "Password":"12345678"}' 40.117.123.41:8080/register
curl -X POST -H "Content-Type: application/json" -d '{"Email":"tanmoykrishnadas@gmail.com", "Password":"12345678"}' 40.117.123.41:8080/login
*/
