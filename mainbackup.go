package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"time"
	"todoappgo/configs"
	"todoappgo/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = []byte(configs.EnvSecretPassword())
var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

var client *mongo.Client

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func GenerateJWT() (string, error) {

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(1516239022, 0)),
		Issuer:    "Jezbot",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}
	return tokenString, nil
}

func ParseJWT(tokenString string) {

	type MyCustomClaims struct {
		Foo string `json:"foo"`
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("AllYourBase"), nil
	})

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		fmt.Printf("%v %v", claims.Foo, claims.RegisteredClaims.Issuer)
	} else {
		fmt.Println(err)
	}
}

func userSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)
	user.Password = getHash([]byte(user.Password))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			result, _ := userCollection.InsertOne(ctx, user)
			json.NewEncoder(response).Encode(result)
		} else {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message":"` + err.Error() + `"}`))
			return
		}
	} else {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"this user already exists!"}`))
		return
	}
}

func userLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		log.Println(err.Error())
		return
	}
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}
	jwtToken, err := GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	response.Write([]byte(`{"token":"` + jwtToken + `"}`))

}

func returnTasks(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	//foo := "foo"
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		log.Println(err.Error())
		return
	}

	taskString := "{\"tasks\": ["
	for i, s := range dbUser.Tasks {
		if i+1 < len(dbUser.Tasks) {
			taskString += "\"" + s + "\", "
		} else {
			taskString += "\"" + s + "\""
		}
	}
	taskString += "]}"
	response.Write([]byte(taskString))

}

func userUpdate(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	//var dbUser User
	json.NewDecoder(request.Body).Decode(&user)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	filter := bson.D{{"email", user.Email}}
	update := bson.D{{"$set", bson.D{{"tasks", user.Tasks}}}}
	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	fmt.Printf("Documents updated: %v\n", result.ModifiedCount)
	response.Write([]byte(`{"message":"` + "user updated successfully" + `"}`))
}

func main() {
	log.Println("Starting the application")

	router := mux.NewRouter()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	log.Println("/todo/login (POST)")
	log.Println("/todo/signup (POST)")
	log.Println("/todo/tasks (GET)")
	log.Println("/todo/update (PUT)")
	router.HandleFunc("/todo/login", userLogin).Methods("POST", "OPTIONS")
	router.HandleFunc("/todo/signup", userSignup).Methods("POST", "OPTIONS")

	//below here to be protected by JWT
	router.HandleFunc("/todo/tasks", returnTasks).Methods("GET", "OPTIONS")
	router.HandleFunc("/todo/update", userUpdate).Methods("PUT", "OPTIONS")

	log.Fatal(http.ListenAndServe(":6001", router))

}
