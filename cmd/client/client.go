package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	//"github.com/ory/keto/internal/expand"
	//"github.com/ory/keto/internal/relationtuple"
	//"github.com/ory/keto/internal/x"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	clientTypeGRPC = "grpc"
	clientTypeREST = "rest"

	FlagReadRemote    = "read-remote"
	FlagWriteRemote   = "write-remote"
	FlagClientType    = "client-type"
	FlagClientTimeout = "client-timeout"

	EnvClientType    = "KETO_CLIENT_TYPE"
	EnvReadRemote    = "KETO_READ_REMOTE"
	EnvWriteRemote   = "KETO_WRITE_REMOTE"
	EnvClientTimeout = "KETO_CLIENT_TIMEOUT"

	ModeReadOnly  readWriteFlag = 0
	ModeReadWrite readWriteFlag = 1
)

type (
	client interface {
		// run check if the combination of subject, relation, namespace and object is allowed
		Check(subject, relation, namespace, object string, maxDepth int32) (bool, error)
		Close()
	}

	readWriteFlag uint
)

// RegisterRemoteURLFlags adds flags for connection handling and the protocol.
func RegisterRemoteURLFlags(flags *pflag.FlagSet) {
	flags.Int(FlagClientTimeout, 30, "client timeout in seconds")
	flags.String(FlagClientType, "grpc", fmt.Sprintf("Choose the client library to use: %s, %s, %s", clientTypeGRPC, clientTypeREST))
	flags.String(FlagReadRemote, "127.0.0.1:4466", "Remote address of the read API endpoint.")
	flags.String(FlagWriteRemote, "127.0.0.1:4467", "Remote address of the write API endpoint.")
}

// FromCmd extracts all flags from cobra.cmd to create the new client instance.
func FromCmd(cmd *cobra.Command, readWrite readWriteFlag, ctx context.Context) (client, error) {
	clientType, err := cmd.Flags().GetString(FlagClientType)
	if err != nil {
		return nil, fmt.Errorf("could not resolve client-type flag: %w", err)
	}
	if env, isSet := os.LookupEnv(EnvClientType); isSet {
		clientType = env
	}

	timeout, err := cmd.Flags().GetInt(FlagClientType)
	if err != nil {
		return nil, fmt.Errorf("could not resolve client-timeout flag: %w", err)
	}
	if env, isSet := os.LookupEnv(EnvClientType); isSet {
		timeout, err = strconv.Atoi(env)
		if err != nil {
			return nil, fmt.Errorf("timeout set in environment is not a valid integer")
		}
	}

	var connString string
	switch readWrite {
	case ModeReadOnly:
		connString, err = cmd.Flags().GetString(FlagReadRemote)
		if err != nil {
			return nil, fmt.Errorf("could not resolve read-remote flag: %w", err)
		}
		if remote, isSet := os.LookupEnv(EnvReadRemote); isSet {
			connString = remote
		}
		if connString == "" {
			return nil, fmt.Errorf("read address is set to empty")
		}
	case ModeReadWrite:
		connString, err = cmd.Flags().GetString(FlagWriteRemote)
		if err != nil {
			return nil, fmt.Errorf("could not resolve write-remote flag: %w", err)
		}
		if remote, isSet := os.LookupEnv(EnvWriteRemote); isSet {
			connString = remote
		}
		if connString == "" {
			return nil, fmt.Errorf("write address is set to empty")
		}
	default:
		return nil, fmt.Errorf("unknown readWriteFlag value")
	}
	if host, port, err := net.SplitHostPort(connString); err != nil {
		return nil, fmt.Errorf("remote address must consist of <host>:<port> or [<host>]:<port> in case of IPv6: %w", err)
	} else if host == "" {
		return nil, fmt.Errorf("remote address contains no host")
	} else if port == "" {
		return nil, fmt.Errorf("remote address contains no port")
	}

	return New(ctx, clientType, connString, time.Duration(timeout)*time.Second)
}

// New creates a new client based on the set flags.
func New(ctx context.Context, clientType, connStr string, timeout time.Duration) (client, error) {
	var newClient func(ctx context.Context, remote string, timeout time.Duration) (client, error)
	switch clientType {
	case clientTypeGRPC:
		newClient = newGRPCClient
	case clientTypeREST:
		newClient = newSDKClient
	default:
		return nil, fmt.Errorf("unknown client type '%s'", clientType)
	}

	c, err := newClient(ctx, connStr, timeout)
	if err != nil {
		return nil, fmt.Errorf("could not open client: %w", err)
	}

	return c, nil
}
