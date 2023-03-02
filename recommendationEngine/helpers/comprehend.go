package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"recommendation-system/miscelaneous"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/zhenghaoz/gorse/client"
)

type ErrorSchema struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GorseItemSchema struct {
	ItemId     string   `json:"itemId"`
	IsHidden   bool     `json:"isHidden"`
	Categories []string `json:"categories"`
	Timestamp  string   `json:"timeStamp"`
	Labels     []string `json:"labels"`
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
	miscelaneous.ErrorLogger(payloaderr)

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

func GorseAddItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	bodyByte, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		log.Fatal(bodyErr)
		errbyte, _ := json.Marshal(&ErrorSchema{"error", "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errbyte)
		return
	}

	var reqBody GorseItemSchema

	reqerr := json.Unmarshal(bodyByte, &reqBody)

	miscelaneous.ErrorLogger(reqerr)

	if (reqBody.ItemId == "") || reqBody.Comment == "" || len(reqBody.Categories) == 0 || len(reqBody.Labels) == 0 || reqBody.Timestamp == "" {
		errBody := &ErrorSchema{"error", "item id, comment, category, labels, ishidden, and timestamp are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorse:8088", "")

	affected, _ := gorse.InsertItem(context.TODO(), client.Item{
		ItemId:     reqBody.ItemId,
		IsHidden:   reqBody.IsHidden,
		Categories: reqBody.Categories,
		Timestamp:  reqBody.Timestamp,
		Labels:     reqBody.Labels,
		Comment:    reqBody.Comment,
	})

	if affected.RowAffected < 1 {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Error saving item"}
		jsonStr, _ := json.Marshal(jsonResp)

		w.Write(jsonStr)
		return
	}

	type SuccessSchema struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		ItemId  string `json:"itemid"`
	}

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", "Item inserted successfully", reqBody.ItemId}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}

func GorseAddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	type GorseUserSchema struct {
		UserId   string `json:"userId"`
		Platform string `json:"platform"`
	}

	bodyByte, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		log.Fatal(bodyErr)
		errbyte, _ := json.Marshal(&ErrorSchema{"error", "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errbyte)
		return
	}

	var reqBody GorseUserSchema

	reqerr := json.Unmarshal(bodyByte, &reqBody)

	miscelaneous.ErrorLogger(reqerr)

	if (reqBody.UserId == "") || reqBody.Platform == "" {
		errBody := &ErrorSchema{"error", "user id, and platform are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorse:8088", "")

	affected, _ := gorse.InsertUser(context.TODO(), client.User{
		Labels: []string{reqBody.Platform},
		UserId: reqBody.UserId,
	})

	if affected.RowAffected < 1 {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Error saving user"}
		jsonStr, _ := json.Marshal(jsonResp)

		w.Write(jsonStr)
		return
	}

	type SuccessSchema struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		UserId  string `json:"userid"`
	}

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", "User inserted successfully", reqBody.UserId}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}

func GorseAddFeedback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	type GorseFeedbackSchema struct {
		FeedbackType string `json:"feedbackType"`
		UserId       string `json:"userId"`
		ItemId       string `json:"itemId"`
		Timestamp    string `json:"timeStamp"`
	}

	bodyByte, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		log.Fatal(bodyErr)
		errbyte, _ := json.Marshal(&ErrorSchema{"error", "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errbyte)
		return
	}

	var reqBody GorseFeedbackSchema

	reqerr := json.Unmarshal(bodyByte, &reqBody)

	miscelaneous.ErrorLogger(reqerr)

	if (reqBody.UserId == "") || reqBody.FeedbackType == "" || reqBody.ItemId == "" || reqBody.Timestamp == "" {
		errBody := &ErrorSchema{"error", "user id, feedback type, item id, and timestamp are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorse:8088", "")

	affected, _ := gorse.InsertFeedback(context.TODO(), []client.Feedback{
		{
			FeedbackType: reqBody.FeedbackType, UserId: reqBody.UserId, ItemId: reqBody.ItemId, Timestamp: reqBody.Timestamp,
		},
	})

	if affected.RowAffected < 1 {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Error saving feedback"}
		jsonStr, _ := json.Marshal(jsonResp)

		w.Write(jsonStr)
		return
	}

	type SuccessSchema struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		UserId  string `json:"userid"`
	}

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", "Feedback inserted successfully", reqBody.UserId}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}

func GorseRecommend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	type GorseRecommendSchema struct {
		UserId   string `json:"userId"`
		Category string `json:"category"`
		Length   int    `json:"length"`
	}

	bodyByte, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		log.Fatal(bodyErr)
		errbyte, _ := json.Marshal(&ErrorSchema{"error", "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errbyte)
		return
	}

	var reqBody GorseRecommendSchema

	reqerr := json.Unmarshal(bodyByte, &reqBody)

	miscelaneous.ErrorLogger(reqerr)

	if (reqBody.UserId == "") || reqBody.Category == "" || reqBody.Length == 0 {
		errBody := &ErrorSchema{"error", "user id, category, and recommendation length are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorse:8088", "")

	recommendations, recErr := gorse.GetRecommend(context.TODO(), reqBody.UserId, reqBody.Category, reqBody.Length)

	miscelaneous.ErrorLogger(recErr)

	type SuccessSchema struct {
		Status          string   `json:"status"`
		Recommendations []string `json:"recommendations"`
		UserId          string   `json:"userid"`
	}

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", recommendations, reqBody.UserId}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}

func GorseApiRecommend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	type GorseRecommendSchema struct {
		UserId        string `json:"userId"`
		Category      string `json:"category"`
		WriteBackType string `json:"writeBackType"`
		Platform      string `json:"platform"`
		DelayMins     int    `json:"delayMins"`
		Length        int    `json:"length"`
	}

	bodyByte, bodyErr := io.ReadAll(r.Body)

	if bodyErr != nil {
		log.Fatal(bodyErr)
		errbyte, _ := json.Marshal(&ErrorSchema{"error", "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errbyte)
		return
	}

	var reqBody GorseRecommendSchema

	reqerr := json.Unmarshal(bodyByte, &reqBody)

	miscelaneous.ErrorLogger(reqerr)

	if (reqBody.UserId == "") || reqBody.Category == "" || reqBody.Platform == "" || reqBody.DelayMins == 0 || reqBody.WriteBackType == "" || reqBody.Length == 0 {
		errBody := &ErrorSchema{"errorsss", "user id, category, write-back-type, platform, delay-minutes and recommendation length are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	userApiUrl := "http://gorse:8088/api/user/" + reqBody.UserId

	var newUser bool

	_, userStatusCode := miscelaneous.GetRequest(userApiUrl)

	if userStatusCode != 200 {
		type userSchema struct {
			UserId    string
			Comment   string
			Labels    []string
			Subscribe []string
		}

		platform := []string{reqBody.Platform}
		createUserUrl := "http://gorse:8088/api/user"

		data := &userSchema{UserId: reqBody.UserId, Labels: platform}
		dataByte, _ := json.Marshal(data)
		_, resStatusCode := miscelaneous.PostRequest(createUserUrl, dataByte)

		if resStatusCode != 200 {
			w.WriteHeader(http.StatusOK)
			jsonResp := &ErrorSchema{"error", "Error adding user. User does not exist"}
			jsonStr, _ := json.Marshal(jsonResp)

			w.Write(jsonStr)
			return
		}

		newUser = true
	} else {
		newUser = false
	}

	apiUrl := "http://gorse:8088/api/recommend/" + reqBody.UserId + "/" + reqBody.Category + "?write-back-type=" + reqBody.WriteBackType + "&write-back-delay=" + strconv.Itoa(reqBody.DelayMins) + "m&n=" + strconv.Itoa(reqBody.Length)

	resp, statusCode := miscelaneous.GetRequest(apiUrl)

	if statusCode != 200 {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Error fetching recommendation"}
		jsonStr, _ := json.Marshal(jsonResp)

		w.Write(jsonStr)
	}

	type SuccessSchema struct {
		Status          string   `json:"status"`
		Recommendations []string `json:"recommendations"`
		UserId          string   `json:"userid"`
		NewUser         bool     `json:"newuser"`
	}

	var recommendations []string

	recErr := json.Unmarshal(resp, &recommendations)

	miscelaneous.ErrorLogger(recErr)

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", recommendations, reqBody.UserId, newUser}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}
