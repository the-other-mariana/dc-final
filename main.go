package main

import (
	//"fmt"
	"log"
	//"math/rand"
	"time"
	"github.com/gin-gonic/gin"

	"github.com/the-other-mariana/dc-final/api"
	"github.com/the-other-mariana/dc-final/controller"
	"github.com/the-other-mariana/dc-final/scheduler"
)

func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")

	// Start Controller
	go controller.Start()

	// Start Scheduler
	
	go scheduler.Start(api.Jobs)
	// Send sample jobs
	//sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: "hello"}

	// API
	// Here's where your API setup will be
	router := gin.Default()
	
	router.GET("/login", api.Login)
	router.GET("/logout", api.Logout)
	router.GET("/status", api.Status)
	router.POST("/upload", api.Upload)

	router.GET("/workloads/test", api.Workloads)
	router.GET("/status/:worker", api.WorkerStatus)

	go router.Run(":8080")

	sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: ""}

	for {
		if sampleJob.RPCName == "test" {
			api.Jobs <- sampleJob
		}
		time.Sleep(time.Second * 2)
	}
}
