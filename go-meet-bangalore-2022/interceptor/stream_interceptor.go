package interceptor

import (
	"context"
	eventspb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &reqInterceptedServerStream{ServerStream: ss, ctx: ss.Context()})
	}
}

// reqInterceptedServerStream wraps around the embedded grpc.ServerStream, and intercepts req headers to set client
type reqInterceptedServerStream struct {
	grpc.ServerStream
	ctx          context.Context
	populatedCtx bool
}

func (ss *reqInterceptedServerStream) Context() context.Context {
	return ss.ctx
}

func (ss *reqInterceptedServerStream) RecvMsg(m interface{}) error {
	recvErr := ss.ServerStream.RecvMsg(m)
	if recvErr != nil {
		return recvErr
	}
	ss.populateAttributes(m)
	return nil
}

type RequestWithHeader interface {
	GetHeader() *eventspb.Header
}

const (
	field1Key = "field1"
	field2Key = "field2"
	field3Key = "field3"
	field4Key = "field4"
	field5Key = "field5"
	field6Key = "field6"
	field7Key = "field7"
	field8Key = "field8"
	field9Key = "field9"
)

func (ss *reqInterceptedServerStream) populateAttributes(m interface{}) {
	ctx := ss.Context()
	if reqWithHeader, ok := m.(RequestWithHeader); ok && reqWithHeader != nil {
		header := reqWithHeader.GetHeader()
		ctx = metadata.AppendToOutgoingContext(ctx, field1Key, header.GetField1())
		ctx = metadata.AppendToOutgoingContext(ctx, field2Key, header.GetField2())
		ctx = metadata.AppendToOutgoingContext(ctx, field3Key, header.GetField3())
		ctx = metadata.AppendToOutgoingContext(ctx, field4Key, header.GetField4())
		ctx = metadata.AppendToOutgoingContext(ctx, field5Key, header.GetField5())
		ctx = metadata.AppendToOutgoingContext(ctx, field6Key, header.GetField6())
		ctx = metadata.AppendToOutgoingContext(ctx, field7Key, header.GetField7())
		ctx = metadata.AppendToOutgoingContext(ctx, field8Key, header.GetField8())
		ctx = metadata.AppendToOutgoingContext(ctx, field9Key, header.GetField9())
		ss.ctx = ctx
		return
	}
}
