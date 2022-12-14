// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: data/v1alpha1/data.proto

package datav1alpha1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DataServiceClient is the client API for DataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataServiceClient interface {
	SetData(ctx context.Context, in *SetDataRequest, opts ...grpc.CallOption) (*SetDataResponse, error)
}

type dataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataServiceClient(cc grpc.ClientConnInterface) DataServiceClient {
	return &dataServiceClient{cc}
}

func (c *dataServiceClient) SetData(ctx context.Context, in *SetDataRequest, opts ...grpc.CallOption) (*SetDataResponse, error) {
	out := new(SetDataResponse)
	err := c.cc.Invoke(ctx, "/data.v1alpha1.DataService/SetData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataServiceServer is the server API for DataService service.
// All implementations should embed UnimplementedDataServiceServer
// for forward compatibility
type DataServiceServer interface {
	SetData(context.Context, *SetDataRequest) (*SetDataResponse, error)
}

// UnimplementedDataServiceServer should be embedded to have forward compatible implementations.
type UnimplementedDataServiceServer struct {
}

func (UnimplementedDataServiceServer) SetData(context.Context, *SetDataRequest) (*SetDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetData not implemented")
}

// UnsafeDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataServiceServer will
// result in compilation errors.
type UnsafeDataServiceServer interface {
	mustEmbedUnimplementedDataServiceServer()
}

func RegisterDataServiceServer(s grpc.ServiceRegistrar, srv DataServiceServer) {
	s.RegisterService(&DataService_ServiceDesc, srv)
}

func _DataService_SetData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataServiceServer).SetData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/data.v1alpha1.DataService/SetData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataServiceServer).SetData(ctx, req.(*SetDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataService_ServiceDesc is the grpc.ServiceDesc for DataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "data.v1alpha1.DataService",
	HandlerType: (*DataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetData",
			Handler:    _DataService_SetData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "data/v1alpha1/data.proto",
}
