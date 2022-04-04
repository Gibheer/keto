package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ory/keto/internal/expand"
	httpclient "github.com/ory/keto/internal/httpclient/client"
	"github.com/ory/keto/internal/httpclient/client/read"
	"github.com/ory/keto/internal/httpclient/models"
	"github.com/ory/keto/internal/relationtuple"
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

func (c *clientSDK) Expand(relation, namespace, object string, maxDepth int32) (*expand.Tree, error) {
	params := &read.GetExpandParams{
		Relation:  relation,
		Namespace: namespace,
		Object:    object,
		Context:   c.ctx,
	}
	resp, err := c.conn.Read.GetExpand(params)
	if err != nil {
		return nil, fmt.Errorf("expand request failed: %s", err)
	}

	return sdkConvertExpandTree(resp.Payload), nil
}

func sdkConvertExpandTree(mt *models.ExpandTree) *expand.Tree {
	et := &expand.Tree{
		Type: expand.NodeType(*mt.Type),
	}
	if mt.SubjectSet != nil {
		et.Subject = &relationtuple.SubjectSet{
			Namespace: *mt.SubjectSet.Namespace,
			Object:    *mt.SubjectSet.Object,
			Relation:  *mt.SubjectSet.Relation,
		}
	} else {
		et.Subject = &relationtuple.SubjectID{ID: mt.SubjectID}
	}

	if et.Type != expand.Leaf && len(mt.Children) != 0 {
		et.Children = make([]*expand.Tree, len(mt.Children))
		for i, c := range mt.Children {
			et.Children[i] = sdkConvertExpandTree(c)
		}
	}
	return et
}

func (c *clientSDK) Close() {
	// noop, because SDK doesn't support close
}
