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
	grpcpipe "github.com/mkeeler/resource-api-poc/pkg/grpc-pipe"
	datapb "github.com/mkeeler/resource-api-poc/proto/data/v1alpha1"
	resourcepb "github.com/mkeeler/resource-api-poc/proto/resource/v1alpha1"
)

type resourceType struct {
	group   string
	version string
	kind    string
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

var dataType = resourceType{group: "g", version: "1", kind: "t"}

func (s *server) SetData(ctx context.Context, req *datapb.SetDataRequest) (*datapb.SetDataResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	tstore, ok := s.data[dataType]
	if !ok {
		tstore = make(map[string]*resourcepb.Resource)
		s.data[dataType] = tstore
	}

	r := &resourcepb.Resource{
		Id: &resourcepb.ResourceID{
			Type: &resourcepb.ResourceType{
				Group:   dataType.group,
				Version: dataType.version,
				Kind:    dataType.kind,
			},
			Name: req.GetName(),
		},
		Data: req.GetData().GetValue(),
	}

	tstore[req.GetName()] = r

	return &datapb.SetDataResponse{Resource: r}, nil
}

func (s *server) Read(ctx context.Context, req *resourcepb.ReadRequest) (*resourcepb.ReadResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetId().GetType().Group, version: req.GetId().GetType().GetVersion(), kind: req.GetId().GetType().GetKind()}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &resourcepb.ReadResponse{Resource: s.data[t][req.GetId().GetName()]}, nil
}

func (s *server) Write(ctx context.Context, req *resourcepb.WriteRequest) (*resourcepb.WriteResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	t := resourceType{group: req.GetResource().GetId().GetType().Group, version: req.GetResource().GetId().GetType().GetVersion(), kind: req.GetResource().GetId().GetType().GetKind()}
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

	t := resourceType{group: req.GetType().Group, version: req.GetType().GetVersion(), kind: req.GetType().GetKind()}
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

	t := resourceType{group: req.GetId().GetType().Group, version: req.GetId().GetType().GetVersion(), kind: req.GetId().GetType().GetKind()}
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

func (s *server) Watch(req *resourcepb.WatchRequest, srv resourcepb.ResourceService_WatchServer) error {
	if err := req.ValidateAll(); err != nil {
		return err
	}

	t := resourceType{group: req.GetId().GetType().Group, version: req.GetId().GetType().GetVersion(), kind: req.GetId().GetType().GetKind()}
	s.mu.RLock()
	defer s.mu.RUnlock()
	tstore, ok := s.data[t]
	if !ok {
		return nil
	}

	for _, rsc := range tstore {
		srv.Send(&resourcepb.WatchResponse{
			Operation: resourcepb.WatchResponse_OPERATION_UPSERT,
			Resource:  rsc,
		})
	}

	return nil
}

func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	appServer := NewServer()
	// Attach the Greeter service to the server
	resourcepb.RegisterResourceServiceServer(s, appServer)
	datapb.RegisterDataServiceServer(s, appServer)
	// Serve gRPC Server
	log.Println("Serving gRPC on 0.0.0.0:1234")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	pipeLn := grpcpipe.ListenPipe()
	go func() {
		log.Fatalln(s.Serve(pipeLn))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"pipe",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(pipeLn.DialContextWithoutNetwork),
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

	err = datapb.RegisterDataServiceHandler(context.Background(), gwmux, conn)
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
