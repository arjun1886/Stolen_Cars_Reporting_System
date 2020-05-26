package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//The structure representing the stolen car case data
type Case struct {
	CaseID         string `bson:"case_id,omitempty" json:"case_id,omitempty"`
	ReportedBy     string `bson:"reported_by,omitempty" json:"reported_by,omitempty"`
	HandledBy      string `bson:"handled_by,omitempty" json:"handled_by,omitempty"`
	Status         string `bson:"status,omitempty" json:"status,omitempty"`
	Image          string `bson:"image,omitempty" json:"image,omitempty"`
	Model          string `bson:"model,omitempty" json:"model,omitempty"`
	Brand          string `bson:"brand,omitempty" json:"brand,omitempty"`
	DateRecovered  string `bson:"date_recovered,omitempty" json:"date_recovered,omitempty"`
	DateStolen     string `bson:"date_stolen,omitempty" json:"date_stolen,omitempty"`
	LocationStolen string `bson:"location_stolen,omitempty" json:"location_stolen,omitempty"`
	NumberPlate    string `bson:"number_plate,omitempty" json:"number_plate,omitempty"`
	TimeStamp      int64  `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
}

//The structure representing the Cop data
type Cop struct {
	Name                  string `bson:"name,omitempty" json:"name,omitempty"`
	LastOccupiedTimestamp int64  `bson:"last_occupied_timestamp,omitempty" json:"last_occupied_timestamp,omitempty"`
	Status                string `bson:"status,omitempty" json:"status,omitempty"`
}

//Custom error message in case of invalid requests
type CustomMessage struct {
	CustomString string
}

//Connections to the MongoDB Case and Cop collections are stored in this structure
type Collections struct {
	CopsCollection  *mongo.Collection
	CasesCollection *mongo.Collection
}

func (collections *Collections) reportStolen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	now := time.Now()
	timestamp := now.Unix()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		var custMessage CustomMessage
		custMessage.CustomString = "Please attach a valid image and a number plate"
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(custMessage)
		return
	}
	var newReportStolen Case
	json.Unmarshal(reqBody, &newReportStolen)
	newReportStolen.Status = "Unassigned"
	newReportStolen.TimeStamp = timestamp
	min := 1000
	max := 9999
	newReportStolen.CaseID = newReportStolen.NumberPlate + "_" + strconv.Itoa(rand.Intn(max-min)+min)
	newReportStolen.HandledBy = "None"
	newReportStolen.DateRecovered = "N/A"
	_, err = collections.CasesCollection.InsertOne(context.TODO(), newReportStolen)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	url := "http://localhost:8080/assign_case"
	client := &http.Client{}
	r, _ = http.NewRequest("POST", url, nil)
	_, err = client.Do(r)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(newReportStolen)
}

func (collections *Collections) addCop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		var custMessage CustomMessage
		custMessage.CustomString = "Please enter valid cop details"
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(custMessage)
		return
	}
	var newCop Cop
	json.Unmarshal(reqBody, &newCop)
	newCop.Status = "Unoccupied"
	newCop.LastOccupiedTimestamp = 0
	_, err = collections.CopsCollection.InsertOne(context.TODO(), newCop)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	url := "http://localhost:8080/assign_case"
	client := &http.Client{}
	r, _ = http.NewRequest("POST", url, nil)
	_, err = client.Do(r)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(newCop)
}

func (collections *Collections) getFreeCop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET,OPTIONS")
	if r.Method != http.MethodGet {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	// Array in which you can store the decoded documents
	var copResults []*Cop
	filter := bson.D{primitive.E{Key: "status", Value: "Unoccupied"}}
	cur, err := collections.CopsCollection.Find(context.TODO(), filter)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {
		var potentialCop Cop
		err := cur.Decode(&potentialCop)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		copResults = append(copResults, &potentialCop)
	}
	if err := cur.Err(); err != nil {
		w.WriteHeader(500)
		return
	}
	cur.Close(context.TODO())
	if len(copResults) == 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(Cop{})
		return
	}
	chosenCop := Cop{}

	//Choosing least  recently occupied cop
	minlastOccupiedTimestamp := copResults[0].LastOccupiedTimestamp
	for _, potentialCop := range copResults {
		if potentialCop.LastOccupiedTimestamp <= minlastOccupiedTimestamp {
			chosenCop = *potentialCop
			minlastOccupiedTimestamp = potentialCop.LastOccupiedTimestamp
		}
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(chosenCop)
}

func (collections *Collections) freeCop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	now := time.Now()
	timestamp := now.Unix()
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	caseID := mux.Vars(r)["caseID"]
	filter := bson.D{primitive.E{Key: "case_id", Value: caseID}}
	var resultCase Case
	err := collections.CasesCollection.FindOne(context.TODO(), filter).Decode(&resultCase)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if (resultCase == Case{}) {
		w.WriteHeader(400)
		return
	}
	handlerCop := resultCase.HandledBy
	filter = bson.D{primitive.E{Key: "name", Value: handlerCop}}
	update := bson.M{"$set": bson.M{"status": "Unoccupied", "last_occupied_timestamp": timestamp}}
	updateResult, err := collections.CopsCollection.UpdateOne(context.TODO(), filter, update)
	if updateResult.ModifiedCount == 0 {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (collections *Collections) tagAsResolved(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	caseID := mux.Vars(r)["caseID"]
	filter := bson.M{
		"case_id": bson.M{
			"$eq": caseID, // check if name of cop matches
		},
	}
	currDateString := time.Now().Format("02-01-2006")
	update := bson.M{"$set": bson.M{"status": "Resolved", "date_recovered": currDateString}}
	updateResult, err := collections.CasesCollection.UpdateOne(context.TODO(), filter, update)
	if updateResult.ModifiedCount == 0 {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (collections *Collections) getUnassignedCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET,OPTIONS")
	if r.Method != http.MethodGet {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	//Array in which you can store the decoded documents
	var caseResults []*Case
	filter := bson.D{primitive.E{Key: "status", Value: "Unassigned"}}
	cur, err := collections.CasesCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {
		var potentialCase Case
		err := cur.Decode(&potentialCase)
		if err != nil {
			w.WriteHeader(500)
		}
		caseResults = append(caseResults, &potentialCase)
	}
	if err := cur.Err(); err != nil {
		w.WriteHeader(500)
	}
	cur.Close(context.TODO())
	if len(caseResults) == 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(Case{})
		return
	}
	var chosenUnassignedCase Case
	if err != nil {
		w.WriteHeader(500)
		return
	}

	//Assigning the earliest registered case
	oldestCaseTimestamp := caseResults[0].TimeStamp
	for _, oldestCase := range caseResults {
		if oldestCase.TimeStamp <= oldestCaseTimestamp {
			chosenUnassignedCase = *oldestCase
			oldestCaseTimestamp = oldestCase.TimeStamp
		}
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(chosenUnassignedCase)
}

func (collections *Collections) assignCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	now := time.Now()
	timestamp := now.Unix()
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	url := "http://localhost:8080/get_free_cop"
	var freeCop Cop
	res, err := http.Get(url)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if res.StatusCode == 400 {
		w.WriteHeader(400)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	json.Unmarshal(body, &freeCop)
	url = "http://localhost:8080/get_unassigned_case"
	var unassignedCase Case
	res, err = http.Get(url)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if res.StatusCode == 400 {
		w.WriteHeader(400)
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	json.Unmarshal(body, &unassignedCase)
	filter := bson.D{primitive.E{Key: "name", Value: freeCop.Name}}
	update := bson.M{"$set": bson.M{"last_occupied_timestamp": timestamp}}
	updateResult, err := collections.CopsCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	updateStatus := bson.M{"$set": bson.M{"status": "Occupied"}}
	updateResult, err = collections.CopsCollection.UpdateOne(context.TODO(), filter, updateStatus)
	if updateResult.ModifiedCount == 0 {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		return
	}
	filter = bson.D{primitive.E{Key: "case_id", Value: unassignedCase.CaseID}}
	updateCopDetails := bson.M{"$set": bson.M{"status": "Assigned to officer", "handled_by": freeCop.Name}}
	updateResult, err = collections.CasesCollection.UpdateOne(context.TODO(), filter, updateCopDetails)
	if updateResult.ModifiedCount == 0 {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (collections *Collections) resolveCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method != http.MethodPost {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	caseID := mux.Vars(r)["caseID"]
	url := "http://localhost:8080/free_cop/" + caseID
	client := &http.Client{}
	r, _ = http.NewRequest("POST", url, nil)
	res, err := client.Do(r)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if res.StatusCode == 400 {
		w.WriteHeader(400)
		return
	}
	url = "http://localhost:8080/tag_as_resolved/" + caseID
	client = &http.Client{}
	r, _ = http.NewRequest("POST", url, nil)
	res, err = client.Do(r)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if res.StatusCode == 400 {
		w.WriteHeader(400)
		return
	}
	url = "http://localhost:8080/assign_case"
	client = &http.Client{}
	r, _ = http.NewRequest("POST", url, nil)
	res, err = client.Do(r)
	w.WriteHeader(200)
}

func (collections *Collections) trackCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(405) // Return 405 Method Not Allowed.
		return
	}
	caseID := mux.Vars(r)["caseID"]
	filter := bson.D{primitive.E{Key: "case_id", Value: caseID}}
	var resultCase Case
	err := collections.CasesCollection.FindOne(context.TODO(), filter).Decode(&resultCase)
	if err != nil {
		var custMessage CustomMessage
		custMessage.CustomString = "Invalid case ID, please try again"
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(custMessage)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resultCase)

}

func main() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	database := client.Database("stolen_car_reports")
	listOfCopsCollection := database.Collection("list_of_cops")
	listOfCasesCollection := database.Collection("list_of_cases")
	router := mux.NewRouter().StrictSlash(true)
	collections := &Collections{CopsCollection: listOfCopsCollection, CasesCollection: listOfCasesCollection}
	router.HandleFunc("/report_stolen", collections.reportStolen).Methods("OPTIONS", "POST")
	router.HandleFunc("/add_cop", collections.addCop).Methods("OPTIONS", "POST")
	router.HandleFunc("/get_free_cop", collections.getFreeCop).Methods("OPTIONS", "GET")
	router.HandleFunc("/assign_case", collections.assignCase).Methods("OPTIONS", "POST")
	router.HandleFunc("/resolve_case/{caseID}", collections.resolveCase).Methods("OPTIONS", "POST")
	router.HandleFunc("/free_cop/{caseID}", collections.freeCop).Methods("OPTIONS", "POST")
	router.HandleFunc("/get_unassigned_case", collections.getUnassignedCase).Methods("OPTIONS", "GET")
	router.HandleFunc("/track_case/{caseID}", collections.trackCase).Methods("OPTIONS", "GET")
	router.HandleFunc("/tag_as_resolved/{caseID}", collections.tagAsResolved).Methods("OPTIONS", "POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
