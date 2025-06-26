package utils

import (
	"strings"
	"time"

	"remoteExec/internal/models"
)

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
