package utils

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"remoteExec/internal/models"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	ctx := context.Background()

	hashKey := "job:" + job.Id

	_, err := rdb.HSet(ctx, hashKey, map[string]any{
		"Language":       job.Language,
		"SourceCode":     job.SourceCode,
		"SubmissionTime": job.SubmissionTime,
		"Id":             job.Id,
		"Status":         "PENDING",
		"Output":         "",
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to store job metadata: %w", err)
	}

	_, err = rdb.RPush(ctx, "jobQueue", job.Id).Result()
	if err != nil {
		return fmt.Errorf("failed to queue job ID: %w", err)
	}

	err = rdb.Expire(ctx, hashKey, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

func PopJobFromRedisQueue(rdb *redis.Client, timeout time.Duration) (*models.RedisJob, error) {
	ctx := context.Background()

	result, err := rdb.BLPop(ctx, timeout, "jobQueue").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("error while popping job from queue: %w", err)
	}

	jobId := result[1]
	hashKey := "job:" + jobId

	jobData, err := rdb.HGetAll(ctx, hashKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error fetching job data: %w", err)
	}
	if len(jobData) == 0 {
		return nil, fmt.Errorf("job with ID %s not found in Redis", jobId)
	}

	job := &models.RedisJob{
		Id:             jobData["Id"],
		Language:       jobData["Language"],
		SourceCode:     jobData["SourceCode"],
		SubmissionTime: jobData["SubmissionTime"],
	}

	return job, nil
}

func UpdateHashValues(rdb *redis.Client, stdout string, stderr string, job *models.RedisJob) {

	hashKey := "job:" + job.Id
	if stderr != "" {
		rdb.HSet(ctx, hashKey, "Status", "FAILED", "Output", stderr)
	} else {
		rdb.HSet(ctx, hashKey, "Status", "SUCCESS", "Output", stdout)
	}
}

func PrintRedisJob(rdb *redis.Client, jobID string) {
	hashKey := "job:" + jobID

	jobData, err := rdb.HGetAll(ctx, hashKey).Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(jobData["Id"])
	fmt.Println(jobData["Language"])
	fmt.Println(jobData["SourceCode"])
	fmt.Println(jobData["SubmissionTime"])
	fmt.Println(jobData["Status"])
	fmt.Println(jobData["Output"])
}

func RunPythonJobContainer(
	job *models.RedisJob,
	imageName string,
	cli *client.Client,
) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		Cmd:          []string{"python3", "/main.py"},
		WorkingDir:   "/",
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		NetworkMode:    "none",
		ReadonlyRootfs: false,
		Privileged:     false,
		CapDrop:        []string{"ALL"},
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}
	defer cli.ContainerRemove(
		context.Background(),
		resp.ID,
		container.RemoveOptions{RemoveVolumes: true, Force: true},
	)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	code := []byte(job.SourceCode)

	tw.WriteHeader(&tar.Header{
		Name: "main.py",
		Mode: 0644,
		Size: int64(len(code)),
	})
	tw.Write(code)
	tw.Close()

	err = cli.CopyToContainer(ctx, resp.ID, "/", &buf, container.CopyToContainerOptions{})
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		panic(err)
	}
	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		panic(err)
	}
	defer out.Close()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, out)
	if err != nil {
		panic(err)
	}

	return stdout.String(), stderr.String(), nil
}
