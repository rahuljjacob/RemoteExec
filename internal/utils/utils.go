package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"remoteExec/internal/models"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CountLines(code string) int {
	if code == "" {
		return 0
	}

	lines := strings.Split(code, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func PackRedisJob(source *models.ExecuteRequest, target *models.RedisJob, job_id string) {
	target.SourceCode = source.SourceCode
	target.Language = source.Language
	target.SubmissionTime = time.Now().Format(time.UnixDate)
	target.Id = job_id
}

func PushJobToRedisQueue(rdb *redis.Client, job models.RedisJob) error {
	// Step 1: Store job data as hash
	hashKey := "job:" + job.Id

	_, err := rdb.HSet(ctx, hashKey, map[string]string{
		"Language":       job.Language,
		"SourceCode":     job.SourceCode,
		"SubmissionTime": job.SubmissionTime,
		"Id":             job.Id,
	}).Result()
	if err != nil {
		return err
	}

	// Step 2: Push job ID into queue (a Redis list)
	_, err = rdb.RPush(ctx, "jobQueue", job.Id).Result()
	return err
}

func PopJobFromQueue(rdb *redis.Client) (*models.RedisJob, error) {
	// Block until a job is available in the queue
	result, err := rdb.BLPop(ctx, 0*time.Second, "jobQueue").Result()
	if err != nil {
		return nil, err
	}
	jobID := result[1]

	// Fetch hash data using job ID
	hashKey := "job:" + jobID
	data, err := rdb.HGetAll(ctx, hashKey).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	rdb.Del(ctx, jobID)

	return &models.RedisJob{
		Language:       data["Language"],
		SourceCode:     data["SourceCode"],
		SubmissionTime: data["SubmissionTime"],
		Id:             data["Id"],
	}, nil
}
