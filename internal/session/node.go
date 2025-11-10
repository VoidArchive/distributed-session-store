package session

import (
	"context"
	"sync"
	"time"

	pb "github.com/voidarchive/distributed-session-store/proto"
)

type Node struct {
	pb.UnimplementedSessionServiceServer
	store     map[string]*Session
	mu        sync.RWMutex
	isLeader  bool
	followers []string // follower addresses for replication
}

func NewLeaderNode() *Node {
	return &Node{
		store:    make(map[string]*Session),
		isLeader: true,
	}
}

func NewFollowerNode() *Node {
	return &Node{
		store:    make(map[string]*Session),
		isLeader: false,
	}
}

func (n *Node) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if !n.isLeader {
		return &pb.SetResponse{Success: false}, nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	expiresAt := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)

	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	n.store[req.SessionId] = &Session{
		ID:        req.SessionId,
		Data:      data,
		ExpiresAt: expiresAt,
	}

	// TODO: Async replication to followers

	return &pb.SetResponse{Success: true}, nil
}

func (n *Node) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	sess, exists := n.store[req.SessionId]
	if !exists || time.Now().After(sess.ExpiresAt) {
		return &pb.GetResponse{Found: false}, nil
	}

	data := make(map[string]string)
	for k, v := range sess.Data {
		data[k] = v.(string)
	}

	return &pb.GetResponse{
		SessionId: sess.ID,
		Data:      data,
		ExpiresAt: sess.ExpiresAt.Unix(),
		Found:     true,
	}, nil
}

func (n *Node) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if !n.isLeader {
		return &pb.DeleteResponse{Success: false}, nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.store, req.SessionId)

	// TODO: Async replication to followers

	return &pb.DeleteResponse{Success: true}, nil
}
