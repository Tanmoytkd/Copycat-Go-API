package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)



func main() {
	token := struct {
		Email, Password string
	}{"tkd@gmail.com", "12345"}

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
