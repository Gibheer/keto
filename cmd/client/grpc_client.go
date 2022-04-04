package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rts "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

type (
	clientGRPC struct {
		ctx  context.Context
		conn *grpc.ClientConn
	}
)

func newGRPCClient(ctx context.Context, remote string, timeout time.Duration) (client, error) {
	creds := insecure.NewCredentials()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c, err := grpc.DialContext(
		ctx,
		remote,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithDisableHealthCheck())
	if err != nil {
		return nil, fmt.Errorf("could not open grpc connections: %w", err)
	}
	return &clientGRPC{conn: c, ctx: ctx}, nil
}

func (c *clientGRPC) Check(subject, relation, namespace, object string, maxDepth int32) (bool, error) {
	cl := rts.NewCheckServiceClient(c.conn)
	resp, err := cl.Check(c.ctx, &rts.CheckRequest{
		Subject:   rts.NewSubjectID(subject),
		Relation:  relation,
		Namespace: namespace,
		Object:    object,
		MaxDepth:  maxDepth,
	})
	if err != nil {
		return false, fmt.Errorf("check request failed: %w", err)
	}
	return resp.Allowed, nil
}

func (c *clientGRPC) Close() {
	c.conn.Close()
}
