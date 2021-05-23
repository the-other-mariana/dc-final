package scheduler

import (
	"context"
	"log"
	"time"
	"strconv"

	"github.com/the-other-mariana/dc-final/controller"

	pb "github.com/the-other-mariana/dc-final/proto"
	"google.golang.org/grpc"
)

//const (
//	address     = "localhost:50051"
//	defaultName = "world"
//)

type Job struct {
	Address string
	RPCName string
}

func schedule(job Job, name string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(job.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	controller.UpdateStatus(name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: job.RPCName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, r.GetMessage())
	controller.UpdateStatus(name)
	jobsCount++
}

var jobsCount int

func Start(jobs chan Job) error {
	jobsCount = 0
	for {
		job := <-jobs
		time.Sleep(time.Second * 5)
		minUsage := 99999
		minPort := 0
		worker := controller.Worker{}

		for _, w := range controller.Workers {
			if w.Usage < minUsage {
				minPort = w.Port
				minUsage = w.Usage
				worker = w
			}
		}
		controller.UpdateUsage(worker.Name)
		controller.Register(worker.Name, jobsCount)
		if minPort == 0 {
			return nil
		}

		job.Address = "localhost:" + strconv.Itoa(minPort)
		schedule(job, worker.Name)
	}
	return nil
}
