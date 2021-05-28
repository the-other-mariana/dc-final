package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/the-other-mariana/dc-final/proto"
	"go.nanomsg.org/mangos"
	//"go.nanomsg.org/mangos/protocol/sub"
	"google.golang.org/grpc"
	//"dc-labs/mangos/protocol/respondent"
	"go.nanomsg.org/mangos/protocol/respondent"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"
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

	fmt.Printf("I will filter the following image: ")
	fmt.Printf(in.GetImg().Filepath + "\n")
	fmt.Printf("Using the following filter: ")
	fmt.Printf(in.GetImg().Filter + "\n")
	//filter := in.GetImg().Filter
	// download image from APIs endpoint

	//DownloadFile(in.GetImg().Filepath, in.Img.Index, in.Img.Workload, filter)

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

