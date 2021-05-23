package main

import (
	//"fmt"
	"log"
	//"math/rand"
	"time"

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
	go api.Start()

	sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: ""}

	for {
		if sampleJob.RPCName == "test" {
			api.Jobs <- sampleJob
		}
		time.Sleep(time.Second * 2)
	}
}
