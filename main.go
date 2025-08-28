package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	client       *mongo.Client
	databaseName = "digitalthread"
)

func init() {
	_ = godotenv.Load()
}

// jsonResponse sets header and writes JSON.
func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonResponse2(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func main() {
	ctx := context.Background()
	client = connectMongo(ctx)
	defer client.Disconnect(ctx)

	r := mux.NewRouter()
	// Auction functions
	r.HandleFunc("/createAuction", CreateAuction).Methods("POST")
	r.HandleFunc("/getAuction/{auctionId}", GetAuction).Methods("GET")
	r.HandleFunc("/getAuctionList", GetAuctionList).Methods("GET")

	r.HandleFunc("/startAuction/{auctionId}", StartAuction).Methods("PUT")
	r.HandleFunc("/stopAuction/{auctionId}", StopAuction).Methods("PUT")
	r.HandleFunc("/getAuctionWinner/{auctionId}", GetAuctionWinner).Methods("GET")

	// Bid functions
	r.HandleFunc("/addBid", AddBid).Methods("POST")
	r.HandleFunc("/getBid/{bidId}", GetBid).Methods("GET")
	r.HandleFunc("/getBidList/{auctionId}", GetBidList).Methods("GET")

	port := os.Getenv("AUCTION_PORT")
	if port == "" {
		log.Println("Error: AUCTION_PORT is empty")
		return
	}

	log.Printf("Bid Manager is listening on :%s", port) // 8080
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
