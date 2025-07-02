package main

import (
	"context"
	"fmt"
	"time"

	// "remoteExec/internal/models"
	"remoteExec/internal/models"
	"remoteExec/internal/utils"

	// "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func main() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	timeout := 100 * time.Second

	for {
		var curJob *models.RedisJob
		var stdout string
		var stderr string
		var err error

		curJob, err = utils.PopJobFromRedisQueue(rdb, timeout)
		if err != nil {
			fmt.Println("Error popping job:", err)
			continue
		}
		if curJob == nil {
			continue
		}

		stdout, stderr, err = utils.RunPythonJobContainer(
			curJob,
			"python:3.11-alpine",
			apiClient,
		)

		if err != nil {
			fmt.Println("Error running container:", err)
			continue
		}

		utils.UpdateHashValues(rdb, stdout, stderr, curJob)

		// utils.PrintRedisJob(rdb, curJob.Id)
	}
}
