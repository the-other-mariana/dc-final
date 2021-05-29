package scheduler

import (
	"context"
	"log"
	"time"
	"strconv"
	"path"
	"strings"
	"path/filepath"
	"fmt"

	"github.com/the-other-mariana/dc-final/controller"

	pb "github.com/the-other-mariana/dc-final/proto"
	"google.golang.org/grpc"
)

//const (
//	address     = "localhost:50051"
//	defaultName = "world"
//)

var jobsCount int

type Job struct {
	Address string
	RPCName string
	Info [4]string
}

func schedule(job Job, name string) {
	conn, err := grpc.Dial(job.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTaskClient(conn)
	controller.UpdateStatus(name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	if job.RPCName == "image" {
		wl := job.Info[2]
		pureName := filepath.Base(job.Info[0])[1:]
		fmt.Println(pureName)
		id := strings.Split(path.Base(pureName), "_")
		fmt.Println(id)
		id_int, _ := strconv.Atoi(id[0])
		img := pb.Image{
			Workload: wl, 
			Name: controller.Workloads[wl].Name,
			Index: int64(id_int), 
			Filepath: job.Info[0],
			Filter: job.Info[3],
		}
		r, err := c.FilterImage(ctx, &pb.ImgRequest{Name: "Image Filter", Img: &img })
		if err != nil {
			log.Fatalf("could not proccess image: %v", err)
		}
		log.Printf("Scheduler: RPC respose from %s : %s was filtered", job.Address, r.GetMessage())
	}
	controller.UpdateStatus(name)
	jobsCount++
}



func Start(jobs chan Job) error {
	jobsCount = 0
	for {
		job := <-jobs
		time.Sleep(time.Second * 5)
		minUsage := 999999
		minPort := 0
		worker := controller.Worker{}

		for _, w := range controller.Workers {
			if w.Usage < minUsage && w.Status == "free" {
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
