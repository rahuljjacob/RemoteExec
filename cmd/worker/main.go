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
	var curJob *models.RedisJob
	var stdout string
	var stderr string
	// var stderr string
	var err error

	timeout := 100 * time.Second

	curJob, err = utils.PopJobFromRedisQueue(rdb, timeout)

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	stdout, stderr, err = utils.RunPythonJobContainer(
		curJob,
		"python:3.11-alpine",
		apiClient,
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("==STDOUT==")
	fmt.Println(stdout)

	fmt.Println("==STDERR==")
	fmt.Println(stderr)

	utils.UpdateHashValues(rdb, stdout, stderr, curJob)

	// utils.PrintRedisJob(rdb, curJob.Id)
}
