package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"recommendation-system/router"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(filepath.Join("./", ".env"))

	if err != nil {
		log.Fatal(err)
	}

	var port = os.Getenv("port")

	fmt.Println("Welcome to GetZing")
	r := router.Router()
	fmt.Println("Running at port ", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}
