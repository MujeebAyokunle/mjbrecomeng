package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"recommendation-system/miscelaneous"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
)

// var baseUrl string = os.Getenv("baseUrl")

type ErrorSchema struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GorseItemSchema struct {
	ItemId     string   `json:"itemId"`
	IsHidden   bool     `json:"isHidden"`
	Categories []string `json:"categories"`
	Timestamp  string   `json:"timeStamp"`
	UserId     string   `json:"userId"`
	Labels     []string `json:"labels"`
	Platform   string   `json:"platform"`
	Comment    string   `json:"comment"`
}

func getS3Client() *comprehend.Comprehend {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:                        aws.String("us-east-1"),
		MaxRetries:                    aws.Int(2),
		CredentialsChainVerboseErrors: aws.Bool(true),

		// https://github.com/aws/aws-sdk-go/issues/2914
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}))

	ec2RoleProvider := &ec2rolecreds.EC2RoleProvider{
		Client: ec2metadata.New(sess, &aws.Config{
			HTTPClient: &http.Client{Timeout: 10 * time.Second},
		}),
		ExpiryWindow: 0,
	}

	creds := credentials.NewChainCredentials([]credentials.Provider{ec2RoleProvider})
	return comprehend.New(sess, &aws.Config{Credentials: creds})
	// return s3.New(sess, &aws.Config{Credentials: creds})
}

func ComprehendFunc(w http.ResponseWriter, r *http.Request) {

	type TextSchema struct {
		Goal string `json:"goal"`
	}

	var payload TextSchema

	body, err := io.ReadAll(r.Body)
	miscelaneous.ErrorLogger(err)

	payloaderr := json.Unmarshal(body, &payload)
	if payloaderr != nil {
		log.Fatal(payloaderr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if payload.Goal == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Goal is required"}
		jsonStr, _ := json.Marshal(jsonResp)

		w.Write(jsonStr)
		return
	}

	// Create a Session with a custom region
	// sess := session.Must(session.NewSession(&aws.Config{
	// 	Region:                        aws.String("us-east-1"),
	// 	MaxRetries:                    aws.Int(2),
	// 	CredentialsChainVerboseErrors: aws.Bool(true),
	// 	// HTTP client is required to fetch EC2 metadata values
	// 	// having zero timeout on the default HTTP client sometimes makes
	// 	// it fail with Credential error
	// 	// https://github.com/aws/aws-sdk-go/issues/2914
	// 	HTTPClient: &http.Client{Timeout: 30 * time.Second},
	// }))

	// Create a Comprehend client from just a session.
	// client := comprehend.New(sess)
	client := getS3Client()

	// Create a Comprehend client with additional configuration
	// client := comprehend.New(sess, aws.NewConfig().WithRegion("us-west-2"))

	var text string = payload.Goal

	// for _, text := range listTexts {
	fmt.Println(text)

	// myChan := make(chan int, 3)
	mut := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	wg.Add(3)
	go func(mut *sync.Mutex, wg *sync.WaitGroup) {
		defer wg.Done()
		mut.Lock()
		params := comprehend.DetectSentimentInput{}
		params.SetLanguageCode("en")
		params.SetText(text)

		req, resp := client.DetectSentimentRequest(&params)

		err := req.Send()
		if err == nil { // resp is now filled
			// fmt.Println(*resp.Sentiment)
			fmt.Println(*resp)
		} else {
			fmt.Println(err)
		}
		mut.Unlock()
	}(mut, wg)

	go func(mut *sync.Mutex, wg *sync.WaitGroup) {
		defer wg.Done()
		mut.Lock()
		params := comprehend.DetectEntitiesInput{}
		params.SetLanguageCode("en")
		params.SetText(text)

		req, resp := client.DetectEntitiesRequest(&params)

		err := req.Send()
		if err == nil { // resp is now filled
			// fmt.Println(*resp.Sentiment)
			fmt.Println(*resp)
		} else {
			fmt.Println(err)
		}
		mut.Unlock()
	}(mut, wg)

	go func(mut *sync.Mutex, wg *sync.WaitGroup) {
		defer wg.Done()
		mut.Lock()
		params := comprehend.DetectKeyPhrasesInput{}
		params.SetLanguageCode("en")
		params.SetText(text)

		req, resp := client.DetectKeyPhrasesRequest(&params)

		err := req.Send()
		if err == nil { // resp is now filled
			// fmt.Println(*resp.Sentiment)
			fmt.Println(*resp)
		} else {
			fmt.Println(err)
		}
		mut.Unlock()
	}(mut, wg)

	wg.Wait()

	w.Write([]byte("Finished"))
}
