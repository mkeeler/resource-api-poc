package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	resourcepb "github.com/mkeeler/resource-api-poc/proto/resource"
)

type resourceType struct {
	group   string
	version string
	rtype   string
}

type server struct {
	mu   sync.RWMutex
	data map[resourceType]map[string]*resourcepb.Resource
}

func NewServer() *server {
	return &server{
		data: make(map[resourceType]map[string]*resourcepb.Resource),
	}
}

func (s *server) Read(ctx context.Context, req *resourcepb.ReadRequest) (*resourcepb.ReadResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetId().GetType().Group, version: req.GetId().GetType().GetVersion(), rtype: req.GetId().GetType().GetType()}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &resourcepb.ReadResponse{Resource: s.data[t][req.GetId().GetName()]}, nil
}

func (s *server) Write(ctx context.Context, req *resourcepb.WriteRequest) (*resourcepb.WriteResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetResource().GetId().GetType().Group, version: req.GetResource().GetId().GetType().GetVersion(), rtype: req.GetResource().GetId().GetType().GetType()}
	s.mu.Lock()
	defer s.mu.Unlock()
	tstore, ok := s.data[t]
	if !ok {
		tstore = make(map[string]*resourcepb.Resource)
		s.data[t] = tstore
	}

	tstore[req.GetResource().GetId().GetName()] = req.GetResource()
	return &resourcepb.WriteResponse{Resource: req.Resource}, nil
}

func (s *server) List(ctx context.Context, req *resourcepb.ListRequest) (*resourcepb.ListResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetType().Group, version: req.GetType().GetVersion(), rtype: req.GetType().GetType()}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var resources []*resourcepb.Resource
	for _, r := range s.data[t] {
		resources = append(resources, r)
	}

	return &resourcepb.ListResponse{Resources: resources}, nil
}

func (s *server) Delete(ctx context.Context, req *resourcepb.DeleteRequest) (*resourcepb.DeleteResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetId().GetType().Group, version: req.GetId().GetType().GetVersion(), rtype: req.GetId().GetType().GetType()}
	s.mu.RLock()
	defer s.mu.RUnlock()
	tstore, ok := s.data[t]
	if !ok {
		return &resourcepb.DeleteResponse{}, nil
	}

	name := req.GetId().GetName()
	_, ok = tstore[name]
	if !ok {
		return &resourcepb.DeleteResponse{}, nil
	}

	delete(tstore, name)
	if len(tstore) == 0 {
		delete(s.data, t)
	}
	return &resourcepb.DeleteResponse{}, nil
}

func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	resourcepb.RegisterResourceServiceServer(s, NewServer())
	// Serve gRPC Server
	log.Println("Serving gRPC on 0.0.0.0:1234")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:1234",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = resourcepb.RegisterResourceServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":4321",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:4321")
	log.Fatalln(gwServer.ListenAndServe())
}
