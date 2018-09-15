package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/shomali11/util/xhashes"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// We need a data store. For our purposes, a simple map
// from string to string is completely sufficient.
type store struct {
	data map[string]string

	// Handlers run concurrently, and maps are not thread-safe.
	// This mutex is used to ensure that only one goroutine can update `data`.
	m sync.RWMutex
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}


var (
	// We need a flag for setting the listening address.
	// We set the default to port 8080, which is a common HTTP port
	// for servers with local-only access.
	addr = flag.String("addr", ":8080", "http service address")

	db *sql.DB

	// Now we create the data store.
	s = store{
		data: map[string]string{},
		m:    sync.RWMutex{},
	}
)

type tok struct {
	Token string
}

//this function checks for error and throws a runtime
//error if an error is found
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func connectDatabase() *sql.DB {
	const (
		host     = "localhost"
		database = "copycat"
		user     = "tanmoy"
		password = "jwjHr4RqGq0MOxpu@"
	)

	// Initialize connection string.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user, password, host, database)

	db, dbErr := sql.Open("mysql", connectionString);
	checkError(dbErr)
	err001 := db.Ping()
	checkError(err001)
	return db;
}

// ## main
func main() {
	// The main function starts by parsing the commandline.
	flag.Parse()

	db = connectDatabase()
	defer db.Close()

	// Now we can create a new `Router` instance...
	r := mux.NewRouter()

	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("POST")

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

/*
After saving, we can run the code locally by calling
```
go run rest.go
```
TEST USING CURL:
curl -X POST -H "Content-Type: application/json" -d '{"Name": "Tanmoy Krishna Das", "Email":"tanmoykrishnadas@gmail.com", "Password":"12345678"}' 40.117.123.41:8080/register
curl -X POST -H "Content-Type: application/json" -d '{"Email":"tanmoykrishnadas@gmail.com", "Password":"12345678"}' 40.117.123.41:8080/login
*/
