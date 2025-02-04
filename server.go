package main

import (
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Receipt represents a purchase receipt containing transaction details.
// It includes information about the retailer, purchase date and time, items, and the total amount.
type Receipt struct {
	// Retailer is the name of the retailer or store the receipt is from.
	Retailer string `json:"retailer" binding:"required"`

	// PurchaseDate is the date of the purchase printed on the receipt (YYYY-MM-DD).
	PurchaseDate string `json:"purchaseDate" binding:"required"`

	// PurchaseTime is the time of the purchase printed on the receipt (24-hour HH:MM).
	PurchaseTime string `json:"purchaseTime" binding:"required"`

	// Items is a list of purchased items present on the receipt.
	Items []Item `json:"items" binding:"required,dive"`

	// Total is the total amount paid on the receipt (string).
	Total string `json:"total" binding:"required"`
}

// Item represents an individual product purchased on a receipt.
type Item struct {
	// ShortDescription is the short product description for the item.
	ShortDescription string `json:"shortDescription" binding:"required"`

	// Price is the total price paid for this item, represented as a string.
	Price string `json:"price" binding:"required"`
}

// ReceiptResponse represents the response returned when a receipt is successfully processed.
type ReceiptResponse struct {
	// ID is the unique identifier assigned to the receipt.
	ID string `json:"id"`
}

// PointsResponse represents the response returned when querying the points for a receipt.
type PointsResponse struct {
	// Points is the number of points awarded for the receipt.
	Points int `json:"points"`
}

// receiptStore stores receipts mapped by their unique ID.
var receiptStore = make(map[string]Receipt)

// storeLock to prevent concurrent access to receiptStore.
var storeLock sync.Mutex

// main initializes the Gin router, defines the API endpoints, and starts the server.
func main() {
	router := gin.Default()

	// Endpoint to submit a receipt for processing.
	router.POST("/receipts/process", processReceipt)

	// Endpoint to retrieve the points awarded for a receipt.
	router.GET("/receipts/:id/points", getPoints)

	// Start the HTTP server on port 8080.
	router.Run(":8080")
}

// processReceipt handles the submission of a receipt.
// It parses the JSON request body, generates a unique receipt ID,
// maps the ID to a receipt, and returns the ID.
func processReceipt(context *gin.Context) {
	var receipt Receipt

	// If the JSON is invalid, return a 400 Bad Request response.
	if err := context.ShouldBindJSON(&receipt); err != nil {
		log.Println("Failed to bind receipt JSON:", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid."})
		return
	}

	// Generate a unique identifier for the receipt.
	id := uuid.New().String()
	log.Println("Generated receipt ID:", id)

	// Lock receiptStore and map the ID to a receipt.
	storeLock.Lock()
	receiptStore[id] = receipt
	storeLock.Unlock()
	log.Println("Receipt stored successfully.")

	// Return the generated receipt ID in the response.
	context.JSON(http.StatusOK, ReceiptResponse{ID: id})
}

// getPoints handles the retrieval of points awarded for a given receipt.
// It retrieves the receipt using its unique ID, calculates the points, and returns the result.
func getPoints(context *gin.Context) {
	// Retrieve the receipt ID from the request URL parameters.
	id := context.Param("id")
	log.Println("Fetching points for receipt ID:", id)

	// Lock receiptStore and find the receipt by ID.
	storeLock.Lock()
	receipt, exists := receiptStore[id]
	storeLock.Unlock()

	// If the receipt does not exist, return a 404 Not Found response.
	if !exists {
		log.Println("No receipt found for ID:", id)
		context.JSON(http.StatusNotFound, gin.H{"error": "No receipt found for that ID"})
		return
	}

	// Calculate the points awarded for the receipt.
	points := calculatePoints(receipt)
	log.Println("Points calculated for receipt ID:", id, "Total Points:", points)

	// Return the calculated points in the response.
	context.JSON(http.StatusOK, PointsResponse{Points: points})
}

func calculatePoints(receipt Receipt) int {
	points := 0

	// One point for every alphanumeric character in the retailer name.

	// Compile regex for alphanumeric characters
	regex := regexp.MustCompile(`[a-zA-Z0-9]`)
	// Find all matches in the retailer name
	matches := regex.FindAllString(receipt.Retailer, -1)
	// Count the number of matches
	matchCount := len(matches)
	// Add to points
	points += matchCount
	log.Println("Added", matchCount, "points for retailer:", receipt.Retailer)

	// 50 points if the total is a round dollar amount with no cents.
	if strings.HasSuffix(receipt.Total, ".00") {
		points += 50
		log.Println("Added 50 points for round dollar amount:", receipt.Total)
	}

	// 25 points if the total is a multiple of 0.25.
	totalInCents, _ := strconv.Atoi(strings.ReplaceAll(receipt.Total, ".", ""))
	if totalInCents % 25 == 0 {
		points += 25
		log.Println("Added 25 points for total being multiple of 0.25:", receipt.Total)
	}

	// 5 points for every two items on the receipt.
	itemPoints := (len(receipt.Items) / 2) * 5
	points += itemPoints
	log.Println("Added", itemPoints, "points for item count.")

	// If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer.
	// The result is the number of points earned.
	for _, item := range receipt.Items {
		trimmedLen := len(strings.TrimSpace(item.ShortDescription))
		if trimmedLen%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			roundedPoints := int(math.Ceil(price * 0.2))
			points += roundedPoints
			log.Println("Added", roundedPoints, "points for item:", strings.TrimSpace(item.ShortDescription))
		}
	}

	// 6 points if the day in the purchase date is odd.
	dateParts := strings.Split(receipt.PurchaseDate, "-")
	if day, err := strconv.Atoi(dateParts[2]); err == nil && day%2 == 1 {
		points += 6
		log.Println("Added 6 points for odd purchase date:", receipt.PurchaseDate)
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	if purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime); err == nil {
		hour, min := purchaseTime.Hour(), purchaseTime.Minute()
		if (hour == 14 && min > 0) || (hour == 15) {
			points += 10
			log.Println("Added 10 points for purchase time:", receipt.PurchaseTime)
		}
	}
	
	log.Println("Final calculated points:", points)
	return points
}
