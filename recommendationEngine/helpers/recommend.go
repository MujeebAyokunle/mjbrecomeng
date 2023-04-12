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
	"strings"

	"github.com/zhenghaoz/gorse/client"
)

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

	if reqerr != nil {
		log.Fatal(reqerr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if (reqBody.ItemId == "") || reqBody.Comment == "" || len(reqBody.Categories) == 0 || len(reqBody.Labels) == 0 || reqBody.Timestamp == "" {
		errBody := &ErrorSchema{"error", "itemId, comment, categories, labels, isHidden, and timeStamp are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	var loweCaseLabels []string

	for _, data := range reqBody.Labels {
		loweCaseLabels = append(loweCaseLabels, strings.ToLower(data))
	}

	// Create a client
	// gorse := client.NewGorseClient("http://gorseengine:8088", "")

	// affected, _ := gorse.InsertItem(context.TODO(), client.Item{
	// 	ItemId:     reqBody.ItemId,
	// 	IsHidden:   reqBody.IsHidden,
	// 	Categories: reqBody.Categories,
	// 	Timestamp:  reqBody.Timestamp,
	// 	Labels:     loweCaseLabels,
	// 	Comment:    reqBody.Comment,
	// })

	// if affected.RowAffected < 1 {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsonResp := &ErrorSchema{"error", "Error saving item"}
	// 	jsonStr, _ := json.Marshal(jsonResp)

	// 	w.Write(jsonStr)
	// 	return
	// }

	type reqBodySchema struct {
		ItemId     string
		IsHidden   bool
		Categories []string
		Timestamp  string
		Labels     []string
		Comment    string
	}

	userApiUrl := "http://gorseengine:8088/api/item"

	data := &reqBodySchema{reqBody.ItemId, reqBody.IsHidden, reqBody.Categories, reqBody.Timestamp, loweCaseLabels, reqBody.Comment}

	dataByte, _ := json.Marshal(data)
	type responseStruct struct {
		RowAffected int
	}
	rows, _ := miscelaneous.PostRequest(userApiUrl, dataByte)

	var affectedRow responseStruct
	json.Unmarshal(rows, &affectedRow)
	fmt.Println("affectedRow ", affectedRow)
	if affectedRow.RowAffected < 1 {
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

	if reqerr != nil {
		log.Fatal(reqerr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if (reqBody.UserId == "") || reqBody.Platform == "" {
		errBody := &ErrorSchema{"error", "user id, and platform are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	// gorse := client.NewGorseClient("http://gorseengine:8088", "")

	// affected, _ := gorse.InsertUser(context.TODO(), client.User{
	// 	Labels: []string{reqBody.Platform},
	// 	UserId: reqBody.UserId,
	// })

	// if affected.RowAffected < 1 {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsonResp := &ErrorSchema{"error", "Error saving user"}
	// 	jsonStr, _ := json.Marshal(jsonResp)

	// 	w.Write(jsonStr)
	// 	return
	// }

	userApiUrl := "http://gorseengine:8088/api/user/" + reqBody.UserId

	_, userStatusCode := miscelaneous.GetRequest(userApiUrl)
	fmt.Println("user check status code", userStatusCode)
	if userStatusCode != 200 {
		type userSchema struct {
			UserId    string
			Comment   string
			Labels    []string
			Subscribe []string
		}

		platform := []string{reqBody.Platform}
		createUserUrl := "http://gorseengine:8088/api/user"

		data := &userSchema{UserId: reqBody.UserId, Labels: platform}

		dataByte, _ := json.Marshal(data)
		_, resStatusCode := miscelaneous.PostRequest(createUserUrl, dataByte)

		if resStatusCode != 200 {
			w.WriteHeader(http.StatusOK)
			jsonResp := &ErrorSchema{"error", "Error adding user."}
			jsonStr, _ := json.Marshal(jsonResp)

			w.Write(jsonStr)
			return
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp := &ErrorSchema{"error", "Error adding user. User already exists"}
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

	if reqerr != nil {
		log.Fatal(reqerr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if (reqBody.UserId == "") || reqBody.FeedbackType == "" || reqBody.ItemId == "" || reqBody.Timestamp == "" {
		errBody := &ErrorSchema{"error", "user id, feedback type, item id, and timestamp are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorseengine:8088", "getzingrecommendationengine23")

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

	if reqerr != nil {
		log.Fatal(reqerr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if (reqBody.UserId == "") || reqBody.Category == "" || reqBody.Length == 0 {
		errBody := &ErrorSchema{"error", "user id, category, and recommendation length are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorseengine:8088", "getzingrecommendationengine23")

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

	if reqerr != nil {
		log.Fatal(reqerr)
		errBody := &ErrorSchema{"error", "Data type error"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if (reqBody.UserId == "") || reqBody.Category == "" || reqBody.Platform == "" || reqBody.DelayMins == 0 || reqBody.WriteBackType == "" || reqBody.Length == 0 {
		errBody := &ErrorSchema{"errorsss", "user id, category, write-back-type, platform, delay-minutes and recommendation length are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	userApiUrl := "http://gorseengine:8088/api/user/" + reqBody.UserId

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
		createUserUrl := "http://gorseengine:8088/api/user"

		data := &userSchema{UserId: reqBody.UserId, Labels: platform}
		dataByte, _ := json.Marshal(data)
		_, resStatusCode := miscelaneous.PostRequest(createUserUrl, dataByte)

		if resStatusCode != 200 {
			w.WriteHeader(http.StatusOK)
			jsonResp := &ErrorSchema{"error", "Error adding user. User already exists"}
			jsonStr, _ := json.Marshal(jsonResp)

			w.Write(jsonStr)
			return
		}

		newUser = true
	} else {
		newUser = false
	}

	apiUrl := "http://gorseengine:8088/api/recommend/" + reqBody.UserId + "/" + reqBody.Category + "?write-back-type=" + reqBody.WriteBackType + "&write-back-delay=" + strconv.Itoa(reqBody.DelayMins) + "m&n=" + strconv.Itoa(reqBody.Length)

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

func GorseFullRecommend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Content-Allow-Methods", "POST")

	type GorseRecommendSchema struct {
		UserId   string `json:"userId"`
		Category string `json:"category"`
		Length   int    `json:"length"`
		Platform string `json:"platform"`
		Intent   string `json:"intent"`
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

	if reqerr != nil {
		errBody := &ErrorSchema{"error", "userId must be a string, category must be string, and length must be an integer"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	if reqBody.UserId == "" || reqBody.Intent == "" || reqBody.Category == "" || reqBody.Length == 0 {
		errBody := &ErrorSchema{"error", "userId, category, intent, platform and length are required"}
		w.WriteHeader(http.StatusBadRequest)
		errByte, _ := json.Marshal(errBody)
		w.Write(errByte)
		return
	}

	userApiUrl := "http://gorseengine:8088/api/user/" + reqBody.UserId

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
		createUserUrl := "http://gorseengine:8088/api/user"

		data := &userSchema{UserId: reqBody.UserId, Labels: platform}
		dataByte, _ := json.Marshal(data)
		_, resStatusCode := miscelaneous.PostRequest(createUserUrl, dataByte)

		if resStatusCode != 200 {
			w.WriteHeader(http.StatusOK)
			jsonResp := &ErrorSchema{"error", "Error adding user. User already exists"}
			jsonStr, _ := json.Marshal(jsonResp)

			w.Write(jsonStr)
			return
		}

		newUser = true
	} else {
		newUser = false
	}

	// Create a client
	gorse := client.NewGorseClient("http://gorseengine:8088", "getzingrecommendationengine23")

	recommendations, recErr := gorse.GetRecommend(context.TODO(), reqBody.UserId, reqBody.Category, reqBody.Length)

	type recommendationSchema struct {
		Categories []string `json:"categories"`
		Comment    string   `json:"comment"`
		IsHidden   bool     `json:"isHidden"`
		ItemId     string   `json:"itemId"`
		Labels     []string `json:"labels"`
		Timestamp  string   `json:"timestamp"`
	}

	var recomendationsDetails []recommendationSchema
	var recommendationIds []string

	for _, rec := range recommendations {

		apiUrl := "http://gorseengine:8088/api/item/" + rec

		resp, statusCode := miscelaneous.GetRequest(apiUrl)

		if statusCode != 200 {
			w.WriteHeader(http.StatusBadRequest)
			jsonResp := &ErrorSchema{"error", "Error fetching recommendation details"}
			jsonStr, _ := json.Marshal(jsonResp)

			w.Write(jsonStr)
		}

		var recomendationDetail recommendationSchema

		json.Unmarshal(resp, &recomendationDetail)

		if miscelaneous.StringInSlice(strings.ToLower(reqBody.Intent), recomendationDetail.Labels) {
			recomendationsDetails = append(recomendationsDetails, recomendationDetail)
			recommendationIds = append(recommendationIds, rec)
		}

	}

	miscelaneous.ErrorLogger(recErr)

	type SuccessSchema struct {
		Status                 string                 `json:"status"`
		Recommendations        []string               `json:"recommendations"`
		RecommendationsDetails []recommendationSchema `json:"recommendationsdetails"`
		UserId                 string                 `json:"userid"`
		NewUser                bool                   `json:"newUser"`
	}

	w.WriteHeader(http.StatusOK)
	jsonResp := &SuccessSchema{"success", recommendationIds, recomendationsDetails, reqBody.UserId, newUser}
	jsonStr, _ := json.Marshal(jsonResp)

	w.Write(jsonStr)

}
