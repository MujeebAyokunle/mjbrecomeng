package miscelaneous

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func ErrorLogger(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Post struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserId int    `json:"userId"`
}

func PostRequest(url string, body []byte) ([]byte, int) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	ErrorLogger(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-Key", "getzingrecommendationengine23")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	post := &Post{}
	derr := json.NewDecoder(res.Body).Decode(post)
	if derr != nil {
		log.Fatal(derr)
	}

	resByte, _ := json.Marshal(post)

	return resByte, res.StatusCode

}

func GetRequest(url string) ([]byte, int) {

	req, err := http.NewRequest(http.MethodGet, url, nil)

	ErrorLogger(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-Key", "getzingrecommendationengine23")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	resBody, readErr := io.ReadAll(res.Body)

	if readErr != nil {
		log.Fatal(readErr)
	}

	return resBody, res.StatusCode

}
