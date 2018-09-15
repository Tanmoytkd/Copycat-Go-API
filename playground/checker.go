package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func login() {
	reader := bufio.NewReader(os.Stdin)
	var email, password string
	fmt.Println("Please enter your email:")
	email, err := reader.ReadString('\n')
	email = strings.Replace(email, "\n", "", -1)
	fmt.Println("Please enter your password")
	password, err = reader.ReadString('\n')
	password = strings.Replace(password, "\n", "", -1)

	token := struct {
		Email, Password string
	}{email, password}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(token)

	x, err:= http.Post("http://localhost:8080/login", "application/json; charset=utf-8" , b)
	//NewRequest("GET", "http://localhost:8080/", reader)

	if err!=nil {
		log.Fatal(err)
		return
	}

	//fmt.Println(x)
	fmt.Println(x.Header.Get("Token" ))

	fmt.Println("\n")

	data, _ := ioutil.ReadAll(x.Body)
	fmt.Println(string(data))
	defer x.Body.Close()
}

func register()  {
	var email, password, name string
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your name:")
	name, err := reader.ReadString('\n')
	name = strings.Replace(name, "\n", "", -1)
	fmt.Println("Please enter your email:")
	email, err = reader.ReadString('\n')
	email = strings.Replace(email, "\n", "", -1)
	fmt.Println("Please enter your password")
	password, err = reader.ReadString('\n')
	password = strings.Replace(password, "\n", "", -1)

	token := struct {
		Email, Password, Name string
	}{email, password, name}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(token)

	x, err:= http.Post("http://localhost:8080/register", "application/json; charset=utf-8" , b)
	//NewRequest("GET", "http://localhost:8080/", reader)

	if err!=nil {
		log.Fatal(err)
		return
	}

	//fmt.Println(x)
	fmt.Println(x.Header.Get("Token" ))

	fmt.Println("\n")

	data, _ := ioutil.ReadAll(x.Body)
	fmt.Println(string(data))
	defer x.Body.Close()
}

func logout()  {

}


func main() {
	//register()
	login()
}
