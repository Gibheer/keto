package client

import (
	"context"
	"fmt"
	"time"

	httpclient "github.com/ory/keto/internal/httpclient/client"
	"github.com/ory/keto/internal/httpclient/client/read"
)

type (
	clientSDK struct {
		ctx     context.Context
		conn    *httpclient.OryKeto
		timeout time.Duration
	}
)

func newSDKClient(ctx context.Context, remote string, timeout time.Duration) (client, error) {
	conn := httpclient.NewHTTPClientWithConfig(nil, &httpclient.TransportConfig{
		Host:    remote,
		Schemes: []string{"https", "http"},
	})
	return client(&clientSDK{conn: conn, ctx: ctx, timeout: timeout}), nil
}

func (c *clientSDK) Check(subject, relation, namespace, object string, maxDepth int32) (bool, error) {
	params := read.NewGetCheckParamsWithTimeout(c.timeout).
		WithNamespace(&namespace).
		WithObject(&object).
		WithRelation(&relation)
	params = params.WithSubjectID(&subject)

	resp, err := c.conn.Read.GetCheck(params)
	if err != nil {
		return false, fmt.Errorf("check request failed: %w", err)
	}
	return *resp.Payload.Allowed, nil
}

func (c *clientSDK) Close() {
	// noop, because SDK doesn't support close
}
