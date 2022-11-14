package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"github.com/sethvargo/go-password/password"
)

type User struct{
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ReturnData struct{
	TimesVerified int 
	SuccessfulVerifications int 
	UnsuccessfulVerifications int 
	err error
}


var users []User
var Verifies ReturnData

func main(){
	
	//Data storage 
	usersBytes, _ := ioutil.ReadFile("users.json")
	json.Unmarshal(usersBytes, &users)

	//Initialize Router 
	r := mux.NewRouter()
	r.HandleFunc("/users", newUser).Methods("POST")
	r.HandleFunc("/users/login/{id}", verifyUser).Methods("POST")
	
	log.Fatal(http.ListenAndServe(":8001", r))
}

func newUser(w http.ResponseWriter, r *http.Request){
	var newUser User 
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		errNewUser := errors.New("invalid body requirements")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errNewUser)
		return
	}
	password, err := password.Generate(8, 3, 1, false, false)
	if err != nil {
	  log.Fatal(err)
	}
	newUser.Password = password
	users = append(users, newUser)
	usersMarshalled, err := json.Marshal(users)
	if err != nil {
		errNewUser := errors.New("unable to marshall item data")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errNewUser)
		return
	}
	_ = ioutil.WriteFile("users.json", usersMarshalled, 0644)
	log.Print("new user ", newUser.Username, " added")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newUser)
}

func verifyUser(w http.ResponseWriter, r *http.Request){
	var userInJSON User
	

	Verifies.TimesVerified = Verifies.TimesVerified + 1

	vars := mux.Vars(r)
	userId := vars["id"]
	dt := time.Now().Weekday()
	if dt.String() == "Sunday" {
		w.WriteHeader(http.StatusInternalServerError)
		errVerifyUser := errors.New("we are closed on sundays")
		json.NewEncoder(w).Encode(errVerifyUser)
	}
	err := json.NewDecoder(r.Body).Decode(&userInJSON)
	if err != nil {
		Verifies.err = errors.New("invalid body requirements")
		w.WriteHeader(http.StatusBadRequest)
		Verifies.UnsuccessfulVerifications = Verifies.UnsuccessfulVerifications + 1
		json.NewEncoder(w).Encode(Verifies)
		return 
	}
	for _, user := range users{
		if user.ID == userId{
			if user.Username == userInJSON.Username && user.Password == userInJSON.Password{
				w.WriteHeader(http.StatusOK)
				Verifies.SuccessfulVerifications = Verifies.SuccessfulVerifications + 1
				json.NewEncoder(w).Encode(Verifies)
				return 
			} else {
				Verifies.err = errors.New("incorrect username or password")
				w.WriteHeader(http.StatusBadRequest)
				Verifies.UnsuccessfulVerifications = Verifies.UnsuccessfulVerifications + 1
				json.NewEncoder(w).Encode(Verifies)
				return 
			}
		}
	}
	Verifies.err = errors.New("user not found")
	Verifies.UnsuccessfulVerifications = Verifies.UnsuccessfulVerifications + 1
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Verifies)
}