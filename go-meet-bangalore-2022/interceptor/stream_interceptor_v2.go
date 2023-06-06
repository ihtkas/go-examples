package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func StreamInterceptorV2() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &reqInterceptedServerStreamV2{ServerStream: ss, ctx: ss.Context()})
	}
}

// reqInterceptedServerStreamV2 wraps around the embedded grpc.ServerStream, and intercepts req headers to set client
type reqInterceptedServerStreamV2 struct {
	grpc.ServerStream
	ctx          context.Context
	populatedCtx bool
}

func (ss *reqInterceptedServerStreamV2) Context() context.Context {
	return ss.ctx
}

func (ss *reqInterceptedServerStreamV2) RecvMsg(m interface{}) error {
	recvErr := ss.ServerStream.RecvMsg(m)
	if recvErr != nil {
		return recvErr
	}
	if !ss.populatedCtx {
		ss.populateAttributes(m)
		ss.populatedCtx = true
	}
	return nil
}

func (ss *reqInterceptedServerStreamV2) populateAttributes(m interface{}) {
	ctx := ss.Context()
	if reqWithHeader, ok := m.(RequestWithHeader); ok && reqWithHeader != nil {
		header := reqWithHeader.GetHeader()
		ctx = metadata.AppendToOutgoingContext(ctx, field1Key, header.GetField1(), field2Key, header.GetField2(), field3Key,
			header.GetField3(), field4Key, header.GetField4(), field5Key, header.GetField5(), field6Key, header.GetField6(),
			field7Key, header.GetField7(), field8Key, header.GetField8(), field9Key, header.GetField9())
		ss.ctx = ctx
		return
	}
}
