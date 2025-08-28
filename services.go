package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var auctionCollection string
var bidCollection string

// connectMongo opens a MongoDB connection.
func connectMongo(ctx context.Context) *mongo.Client {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatalln("MONGO_URI is empty")
	}

	auctionCollection = os.Getenv("AUCTION_COLLECTION_NAME")
	if auctionCollection == "" {
		log.Fatalln("auctionCollectionName is empty")
	}

	bidCollection = os.Getenv("BID_COLLECTION_NAME")
	if bidCollection == "" {
		log.Fatalln("bidCollectionName is empty")
	}

	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo.Connect: %v", err)
	}

	if err := cl.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo.Ping: %v", err)
	}
	return cl
}

func CreateAuction(w http.ResponseWriter, r *http.Request) {
	var auctionObj AuctionObject
	if err := json.NewDecoder(r.Body).Decode(&auctionObj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Set the context and timeout
	ctx2, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add auction object to auctionCollection
	_, err := client.Database(databaseName).
		Collection(auctionCollection).
		InsertOne(ctx2, auctionObj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction %s was added to auctionTable\n", auctionObj.Id)
	jsonResponse(w, auctionObj.Id)
}

func GetAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	// Set the context and timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"id": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", auctionId, err)
		}
		return
	}

	// Send JSON response
	jsonResponse2(w, http.StatusOK, auctionObj)
}

func GetAuctionList(w http.ResponseWriter, r *http.Request) {
	// Set the context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//
	cursor, err := client.Database(databaseName).
		Collection(auctionCollection).
		Find(ctx, bson.M{})
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []AuctionObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d auctions", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}

func StartAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	// Set the context and timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"id": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", auctionId, err)
		}
		return
	}

	// Update startDate to reflect that the auction has started
	auctionObj.StartDate = time.Now()

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"id": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s was started", auctionId)
	w.WriteHeader(http.StatusOK)
}

func StopAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	// Set the context and timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"id": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for auctionId: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for auctionId %s: %v", auctionId, err)
		}
		return
	}

	// Update endDate to reflect that the auction has stopped
	auctionObj.EndDate = time.Now()

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"id": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s was stopped", auctionId)
	w.WriteHeader(http.StatusOK)
}

func GetAuctionWinner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	// Set the context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"auctionid": auctionId}).
		Decode(&auctionObj)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}

	// Send JSON response
	jsonResponse2(w, http.StatusOK, auctionObj.WinningBid)
}

func SetAuctionWinner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	bidId := strings.TrimSpace(vars["bidId"])
	if bidId == "" {
		http.Error(w, "missing bidId", http.StatusBadRequest)
		log.Printf("Missing bidId in request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"id": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for auctionId: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for auctionId %s: %v", auctionId, err)
		}
		return
	}

	// Update WinningBid to reflect that the auction has stopped
	auctionObj.WinningBid = bidId

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"auctionid": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s WinningBid was updated", auctionId)
	w.WriteHeader(http.StatusOK)
}

func AddBid(w http.ResponseWriter, r *http.Request) {
	var bidObj BidObject
	if err := json.NewDecoder(r.Body).Decode(&bidObj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add bid object to bidCollection
	_, err := client.Database(databaseName).
		Collection(bidCollection).
		InsertOne(ctx, bidObj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Bid %s was added to bid collection\n", bidObj.BidId)

	jsonResponse(w, bidObj.BidId)
}

func GetBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidId := strings.TrimSpace(vars["bidId"])
	if bidId == "" {
		http.Error(w, "missing bidId", http.StatusBadRequest)
		log.Printf("Missing bidId in request")
		return
	}

	// Set the context and timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var bidObj BidObject
	err := client.Database(databaseName).
		Collection(bidCollection).
		FindOne(ctx, bson.M{"bidid": bidId}).
		Decode(&bidObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", bidId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", bidId, err)
		}
		return
	}

	// Send JSON response
	jsonResponse2(w, http.StatusOK, bidObj)
}

func GetBidList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	filter := bson.M{"auctionid": auctionId}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cursor, err := client.Database(databaseName).
		Collection(bidCollection).
		Find(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []BidObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d bids", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}
