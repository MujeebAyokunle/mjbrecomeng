package router

import (
	"recommendation-system/helpers"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/text/comprehend", helpers.ComprehendFunc).Methods("POST")
	router.HandleFunc("/gorse/additem", helpers.GorseAddItem).Methods("POST")
	router.HandleFunc("/gorse/adduser", helpers.GorseAddUser).Methods("POST")
	router.HandleFunc("/gorse/addfeedback", helpers.GorseAddFeedback).Methods("POST")
	router.HandleFunc("/gorse/recommend", helpers.GorseRecommend).Methods("POST")
	router.HandleFunc("/gorse/timedrecommend", helpers.GorseApiRecommend).Methods("POST")
	router.HandleFunc("/gorse/recommend/details", helpers.GorseFullRecommend).Methods("POST")

	return router
}
