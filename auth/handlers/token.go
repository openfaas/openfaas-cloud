package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func generateToken() {

	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)

	data, err := ioutil.ReadFile("key")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, keyErr := jwt.ParseECPrivateKeyFromPEM(data)
	if keyErr != nil {
		log.Fatal("Load private key", keyErr)
	}

	pubData, pubErr := ioutil.ReadFile("key.pub")
	if pubErr != nil {
		log.Fatal(pubErr)
	}

	publicKey, pubKeyErr := jwt.ParseECPublicKeyFromPEM(pubData)
	if pubKeyErr != nil {
		log.Fatal("Load pub key ", pubKeyErr)
	}

	token := jwt.NewWithClaims(method, jwt.StandardClaims{
		Id:        "alex",
		Issuer:    "openfaas-cloud@github",
		ExpiresAt: time.Now().Add(48 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   "alexellis",
		Audience:  ".system.gw.io",
	})

	val, err := token.SignedString(privateKey)
	fmt.Println(val, token.Valid)

	if err != nil {
		log.Fatal(err)
	}

	parsed, parseErr := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	fmt.Println(parsed, parsed.Valid, parseErr)
}
