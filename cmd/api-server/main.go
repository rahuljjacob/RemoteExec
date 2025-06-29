package main

import (
	"context"
	"fmt"

	"remoteExec/internal/models"
	"remoteExec/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func main() {
	fmt.Println("Starting Router")
	r := setupRouter()
	r.Run(":8080")
}

func setupRouter() *gin.Engine {

	router := gin.Default()

	router.POST("/execute", executeHandler)
	router.GET("/status/:job_id", statusHandler)

	return router
}

func executeHandler(c *gin.Context) {

	var req models.ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if utils.CountLines(req.SourceCode) > 75 || len(req.SourceCode) > 10000 {
		c.JSON(400, gin.H{"error": "Source Code Exceeds Limit"})
		return
	}

	//temp language check??
	if req.Language != "python" {
		c.JSON(400, gin.H{"error": "Only Supports Python Execution"})
		return

	}

	job_id := uuid.New().String()

	c.JSON(200, gin.H{
		"message":     "Code submission received (not yet implemented)",
		"language":    req.Language,
		"source_code": req.SourceCode,
		"id":          job_id,
	})

	var queueElement models.RedisJob

	utils.PackRedisJob(&req, &queueElement, job_id)

	err := utils.PushJobToRedisQueue(rdb, queueElement)

	if err != nil {
		panic(err)
	}
}

func statusHandler(c *gin.Context) {
	jobID := c.Param("job_id")

	c.JSON(200, gin.H{"job_id": jobID, "status": "Status check not yet implemented"})
}
