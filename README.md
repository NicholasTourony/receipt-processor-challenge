# Receipt Processor

## Overview
This is a Go-based web service that processes purchase receipts and calculates reward points based on various criteria.

## Features
- Submit a receipt for processing (`POST /receipts/process`)
- Retrieve points for a processed receipt (`GET /receipts/{id}/points`)

## Prerequisites
Ensure you have the following installed:
- [Go](https://go.dev/dl/)
- [Git](https://git-scm.com/downloads)

## Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/NicholasTourony/receipt-processor-challenge.git
   cd receipt-processor-challenge
   ```
2. Install dependencies:
    ```sh
   go mod tidy
   ```

## Running the Application
Start the server by running:
```sh
 go run server.go
```
The server will start and listen on **port 8080**.

## API Endpoints
### 1. Process a Receipt
**Endpoint:** `POST /receipts/process`

**Request Example:**
```sh
curl -X POST "http://localhost:8080/receipts/process" \
     -H "Content-Type: application/json" \
     -d '{
         "retailer": "Target",
         "purchaseDate": "2022-01-02",
         "purchaseTime": "13:13",
         "total": "1.25",
         "items": [
           { "shortDescription": "Pepsi - 12-oz", "price": "1.25" }
         ]
       }'
```

**Response Example:**
```json
{
  "id": "d23482df-5eb6-4add-93be-859da7c3e486"
}
```

### 2. Retrieve Points for a Receipt
**Endpoint:** `GET /receipts/{id}/points`

**Request Example:**
```sh
curl -X GET "http://localhost:8080/receipts/d23482df-5eb6-4add-93be-859da7c3e486/points"
```

**Response Example:**
```json
{
  "points": 31
}
```