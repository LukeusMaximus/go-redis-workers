package main

import (
    "context"
    "fmt"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var workListKey = "work-item-list"

func connectToDB() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })
    return rdb
}

// Create a number of work items in the DB for workers to do
func createWork(rdb *redis.Client, ctx context.Context, numItems int) {
    // Clear any existing work item list
    err := rdb.Del(ctx, workListKey).Err()
    if err != nil {
        fmt.Println("Failed to delete existing work item list")
    }

    for i := 0; i < numItems; i++ {
        // Add a number of work items to the database
        // Each work item is a map describing the work to be done
        workItemName := fmt.Sprintf("work-item-%d", i)
        err := rdb.HSet(ctx, workItemName, map[string]interface{}{
            "function": "sleep",
            "duration": 10,
        }).Err()
        if err != nil {
            fmt.Printf("Could not add %s to DB.\n", workItemName)
            continue
        }

        // Add the new work item to the list of work items so it can be consumed later
        err = rdb.RPush(ctx, workListKey, workItemName).Err()
        if err != nil {
            fmt.Printf("Could not add %s to work list.\n", workItemName)
            continue
        }
    }

    var length int64 = 0
    length, err = rdb.LLen(ctx, workListKey).Result()
    if err != nil {
        fmt.Println("Could not read length of work item list.")
    } else {
        fmt.Printf("Created %d work items out of %d.\n", length, numItems)
    }
}

func worker() {

}

func spawnWorkers() {
    for 
}

func main() {
    rdb := connectToDB()
    createWork(rdb, ctx, 100)
}

