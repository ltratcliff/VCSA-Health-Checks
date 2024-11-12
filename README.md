# MongoDB REST API with Go

This project demonstrates how to create a RESTful API in Go that connects to a MongoDB database, fetches data, sorts it, and returns it in JSON format.

## Project Structure
```aiignore
├── cmd
│ ├── api
│ │ └── main.go
│ ├── retrieve
│ │ └── main.go
│ └── scan
│     ├── inventory.yaml
│     └── main.go
├── go.mod
├── go.sum
└── vcsim
```

## Prerequisites

- Go 1.23.3 or later
- MongoDB instance running locally or accessible
- GoLand 2024.2.3 (IDE used for this project)

## Setup

1. **Clone the repository**:
    ```sh
    git clone https://github.com/yourusername/your-repo.git
    cd your-repo
    ```

2. **Install Go MongoDB driver**:
    ```sh
    go get go.mongodb.org/mongo-driver/mongo
    ```

3. **Replace placeholders**:
    - Update the MongoDB URI, database, and collection names in `retrieve/retrieve.go`.
    - Ensure the import path for the `retrieve` package in `main.go` matches your module path (`yourmodule/retrieve`).

4. **Initialize Go modules**:
    ```sh
    go mod tidy
    ```

## Running the Application

To run the application, use:

```sh
go run main.go
```

The server will start on port `8080`.

## Endpoints

### `/data`

Returns a list of documents from the `foo` collection in the specified MongoDB database, sorted by the `_id` field in ascending order and limited to 10 documents.

**Example Request**:

```http
GET /data HTTP/1.1
Host: localhost:8080
```

**Example Response**:

```json
[
    {
        "_id": "some_id",
        "field1": "value1",
        "field2": "value2"
    },
    ...
]
```

## Project Files

### `main.go`

This file sets up the HTTP server and defines the route that calls the `retrieve.GetData` function.

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"

    "yourmodule/retrieve" // Ensure this matches the actual module path
)

func handleGetData(w http.ResponseWriter, r *http.Request) {
    data, err := retrieve.GetData()
    if err != nil {
        http.Error(w, "Failed to retrieve data: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        http.Error(w, "Failed to write data: "+err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/data", handleGetData)

    log.Println("Server starting on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
```

### `retrieve/retrieve.go`

This file contains the logic to connect to MongoDB, fetch data, sort it, and return it.

```go
package retrieve

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
)

func GetData() ([]bson.M, error) {
    var results []bson.M

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
        return results, err
    }
    defer client.Disconnect(context.TODO())

    collection := client.Database("yourdatabase").Collection("foo")

    findOptions := options.Find()
    findOptions.SetSort(bson.D{{"_id", 1}}) // Sort by _id in ascending order
    findOptions.SetLimit(10)                // Limit the number of results to 10

    // Perform the query
    cursor, err := collection.Find(context.TODO(), bson.D{}, findOptions)
    if err != nil {
        return results, err
    }
    defer cursor.Close(context.TODO())

    // Iterate through the cursor and store the documents in results
    for cursor.Next(context.TODO()) {
        var result bson.M
        err := cursor.Decode(&result)
        if err != nil {
            return results, err
        }
        results = append(results, result)
    }

    if err := cursor.Err(); err != nil {
        return results, err
    }

    return results, nil
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
