package main

import (
    "context"
    "fmt"
    "github.com/go-redis/redis/v8"
    "runtime"
    "sync"
)

var ctx = context.Background()
var workListKey = "work-item-list"

func connectToDB() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr: "redis:6379",
        Password: "",
        DB: 0,
    })

    connectionAttempts := 0
    for {
        pong, err := rdb.Ping(ctx).Result()
        if err == nil {
            fmt.Printf("Pong received: %s\n", pong)
            break
        } else {
            connectionAttempts++
            fmt.Printf("Connection attempt %d to redis failed, retrying...\n", connectionAttempts)
        }
    }

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

    // Report number entries we actually made
    var length int64 = 0
    length, err = rdb.LLen(ctx, workListKey).Result()
    if err != nil {
        fmt.Println("Could not read length of work item list.")
    } else {
        fmt.Printf("Created %d work items out of %d.\n", length, numItems)
    }
}

// Each worker works until the list of work to do is empty.
// Redis fulfils all interprocess communication needs.
func worker(ctx context.Context, workerId int) {
    // Each worker can have its own database connection
    rdb := connectToDB()
    isAlive := true

    for isAlive {
        result, err := rdb.LPop(ctx, workListKey).Result()
        fmt.Println("result", result, "error", err)
        if err != nil {
            if err.Error() == "redis: nil" {
                fmt.Printf("Worker %d couldn't find any more work.\n", workerId)
                break
            } else {
                fmt.Printf("Worker %d couldn't get an item from the work list.\n", workerId)
                break
            }
        }
    }
}

func spawnWorkers(ctx context.Context) {
    numWorkers := runtime.GOMAXPROCS(0) - 1
    var wg sync.WaitGroup

    fmt.Printf("Spawning %d workers.\n", numWorkers)
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            worker(ctx, id)
        }(i)
    }
    wg.Wait()
}

func main() {
    rdb := connectToDB()
    createWork(rdb, ctx, 100)
    spawnWorkers(ctx)
}

