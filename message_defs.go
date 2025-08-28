package main

import (
	"time"
)

type AuctionObject struct {
	Id         string    `json:"id"`         // GUID for an auction
	ProductId  string    `json:"productid"`  // GUID for the product
	StartDate  time.Time `json:"startdate"`  // Start date of an auction
	EndDate    time.Time `json:"enddate"`    // End date of an auction
	WinningBid string    `json:"winningbid"` // Winning bid of an auction
}

type Response struct {
	GUID    string   `json:"guid"`
	Updated int      `json:"updated"`
	Errors  []string `json:"errors,omitempty"`
}

type BidObject struct {
	BidId        string    `json:"bidid"`        // GUID for the bid
	AuctionId    string    `json:"auctionid"`    // GUID for an auction
	Price        string    `json:"price"`        // Target price
	Quantity     int       `json:"quantity"`     // Quantity requested
	DeliveryDate time.Time `json:"deliverydate"` // Delivery date
	Onion        string    `json:"onion"`        // Onion address of the bid submitter
	ResponseDate time.Time `json:"responsedate"` // The time the bid was submitted
}

type BidResponse struct {
	AuctionID    string    `json:"auctionid"`    // GUID for the auction
	BidId        string    `json:"bidid"`        // GUID for the bid
	ResponseDate time.Time `json:"responsedate"` // end date
}
