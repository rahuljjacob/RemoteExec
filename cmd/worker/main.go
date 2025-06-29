package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"remoteExec/internal/utils"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis container or service address
		// Password: "",         // Add if Redis requires a password
		DB: 0,
	})

	fmt.Println("Worker started. Waiting for jobs...")

	for {
		job, err := utils.PopJobFromQueue(rdb)
		if err != nil {
			log.Println("Error popping job:", err)
			time.Sleep(1 * time.Second) // Prevent tight loop on error
			continue
		}
		fmt.Println(job)
	}
}
