package controller

import (
	"fmt"
	"net/http"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connection from ", r.RemoteAddr)
	w.Write([]byte("Testing"))
}

func PostUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connection from ", r.RemoteAddr)
	w.Write([]byte("Testing"))
}

func PatchUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connection from ", r.RemoteAddr)
	w.Write([]byte("Testing"))
}

func PutUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connection from ", r.RemoteAddr)
	w.Write([]byte("Testing"))
}
