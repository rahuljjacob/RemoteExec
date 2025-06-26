package models

type ExecuteRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
}

type RedisJob struct {
	Language       string
	SourceCode     string
	SubmissionTime string
	Id             string
}
