package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	//"strconv"

	pb "github.com/the-other-mariana/dc-final/proto"
	"go.nanomsg.org/mangos"
	//"go.nanomsg.org/mangos/protocol/sub"
	"google.golang.org/grpc"
	//"dc-labs/mangos/protocol/respondent"
	"go.nanomsg.org/mangos/protocol/respondent"
	"github.com/the-other-mariana/dc-final/controller"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
    "github.com/anthonynsimon/bild/imgio"
)

var (
	defaultRPCPort = 50051
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedTaskServer
}

var (
	controllerAddress = ""
	workerName        = ""
	tags              = ""
	status            = ""
	workDone          = 0
	usage             = 0
	port              = 0
	jobsDone          = 0
)

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("RPC: Received: %v", in.GetName())
	if in.GetName() == "test" {
		workDone += 1
		log.Printf("RPC [Worker] %+v: testing...", workerName)
		usage += 1
		status = "Running"
		usage -= 1
		return &pb.HelloReply{Message: "Hello, " + workerName + " in test"}, nil
	} else {
		workDone += 1
		log.Printf("[Worker] %+v: calling", workerName)
		usage += 1
		status = "Running"
		return &pb.HelloReply{Message: "Hello " + workerName}, nil
	}	
}


func (s *server) FilterImage(ctx context.Context, in *pb.ImgRequest) (*pb.ImgReply, error) {

	msg := fmt.Sprintf("I will filter the following image: %v with filter: %v \n", in.GetImg().Filepath, in.GetImg().Filter)
	fmt.Printf(msg)

	//DownloadFile(in.GetImg().Filepath, in.Img.Index, in.Img.Workload, filter)
	
	

	if in.GetImg().Filter == "grayscale" {

		newFilename := fmt.Sprintf("%v_%v", in.Img.Index, in.Img.Name) + ".png"
		resultsFolder := "./public/results/" + in.Img.Name + "/"
		newResultsPath := path.Join(resultsFolder, newFilename)

		img, err := imgio.Open(in.GetImg().Filepath)

		if err != nil {
			return &pb.ImgReply{Message: "Bild lib could not open image " + workerName}, nil 
		}

		filtered := effect.Grayscale(img)
		if err := imgio.Save(newResultsPath, filtered, imgio.PNGEncoder()); err != nil {
			fmt.Println(err)
			return &pb.ImgReply{Message: "Grayscale error " + workerName}, nil
		}

	} else if in.GetImg().Filter == "blur" {
		newFilename := fmt.Sprintf("%v_%v", in.Img.Index, in.Img.Name) + ".png"
		resultsFolder := "./public/results/" + in.Img.Name + "/"
		newResultsPath := path.Join(resultsFolder, newFilename)

		img, err := imgio.Open(in.GetImg().Filepath)

		if err != nil {
			return &pb.ImgReply{Message: "Bild lib could not open image " + workerName}, nil 
		}

		filtered := blur.Gaussian(img, 3.0)

		if err := imgio.Save(newResultsPath, filtered, imgio.PNGEncoder()); err != nil {
			fmt.Println(err)
			return &pb.ImgReply{Message: "Blur error " + workerName}, nil
		}

		updatedWL := controller.Workload{}
		prev := controller.Workloads[in.Img.Workload]
		updatedWL = controller.Workload{
			Id: prev.Id,
			Filter: prev.Filter,
			Name: prev.Name,
			Status: "completed",
			Jobs: prev.Jobs + 1,
			Imgs: prev.Imgs,
		}

		updatedWL.Imgs = append(prev.Imgs, newFilename)
		controller.Workloads[in.Img.Workload] = updatedWL

	} else {
        return &pb.ImgReply{Message: "Required filter not supported by " + workerName}, nil
	}
	
	return &pb.ImgReply{Message: "The image was proccesed by " + workerName}, nil

}

// ./worker --controller <host>:<port> --worker-name <node_name> --tags <tag1>,<tag2>...
func init() {
	flag.StringVar(&controllerAddress, "controller", "tcp://localhost:40899", "Controller address")
	flag.StringVar(&workerName, "worker-name", "hard-worker", "Worker Name")
	flag.StringVar(&tags, "tags", "gpu,superCPU,largeMemory", "Comma-separated worker tags")
}

// joinCluster is meant to join the controller message-passing server
func joinCluster() {
	var sock mangos.Socket
	var err error
	var msg []byte

	if sock, err = respondent.NewSocket(); err != nil {
		die("can't get new sub socket: %s", err.Error())
	}

	log.Printf("Connecting to controller on: %s", controllerAddress)
	if err = sock.Dial(controllerAddress); err != nil {
		die("can't dial on sub socket: %s", err.Error())
	}
	for {
		if msg, err = sock.Recv(); err != nil {
			die("Cannot recv: %s", err.Error())
		}
		info := fmt.Sprintf("%v %v %v %v %v %v", workerName, status, usage, tags, defaultRPCPort, jobsDone)
		if err = sock.Send([]byte(info)); err != nil {
			die("Cannot send: %s", err.Error())
		}
		log.Printf("Message-Passing: Worker(%s): Received %s\n", workerName, string(msg))
	}
}

func getAvailablePort() int {
	port := defaultRPCPort
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			port = port + 1
			continue
		}
		ln.Close()
		break
	}
	return port
}

func main() {
	flag.Parse()

	// Subscribe to Controller
	go joinCluster()

	// Setup Worker RPC Server
	rpcPort := getAvailablePort()
	defaultRPCPort = rpcPort
	log.Printf("Starting RPC Service on localhost:%v", rpcPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTaskServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

