package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/voidarchive/distributed-session-store/internal/session"
	pb "github.com/voidarchive/distributed-session-store/proto"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 8080, "node port")
	isLeader := flag.Bool("leader", false, "is this node the leader")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	var node *session.Node
	if *isLeader {
		node = session.NewLeaderNode()
	} else {
		node = session.NewFollowerNode()
	}

	pb.RegisterSessionServiceServer(grpcServer, node)

	log.Printf("Node starting on port %d (leader: %v)", *port, *isLeader)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
