package kube

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func NewCRIClient(addr string) (*CRIClient, error) {
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("unix", addr)
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	conn, err := grpc.DialContext(context.Background(), addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	runtimeServiceClient := runtimeapi.NewRuntimeServiceClient(conn)
	return &CRIClient{
		RuntimeServiceClient: runtimeServiceClient,
	}, nil
}

//func NewRemoteRuntimeService(endpoint string, connectionTimeout time.Duration, tp trace.TracerProvider) (internalapi.RuntimeService, error) {
//	klog.V(3).InfoS("Connecting to runtime service", "endpoint", endpoint)
//	addr, dialer, err := util.GetAddressAndDialer(endpoint)
//	if err != nil {
//		return nil, err
//	}
//	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
//	defer cancel()
//
//	var dialOpts []grpc.DialOption
//	dialOpts = append(dialOpts,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//		grpc.WithContextDialer(dialer),
//		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
//	if utilfeature.DefaultFeatureGate.Enabled(features.KubeletTracing) {
//		tracingOpts := []otelgrpc.Option{
//			otelgrpc.WithPropagators(tracing.Propagators()),
//			otelgrpc.WithTracerProvider(tp),
//		}
//		// Even if there is no TracerProvider, the otelgrpc still handles context propagation.
//		// See https://github.com/open-telemetry/opentelemetry-go/tree/main/example/passthrough
//		dialOpts = append(dialOpts,
//			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(tracingOpts...)),
//			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor(tracingOpts...)))
//	}
//
//	connParams := grpc.ConnectParams{
//		Backoff: backoff.DefaultConfig,
//	}
//	connParams.MinConnectTimeout = minConnectionTimeout
//	connParams.Backoff.BaseDelay = baseBackoffDelay
//	connParams.Backoff.MaxDelay = maxBackoffDelay
//	dialOpts = append(dialOpts,
//		grpc.WithConnectParams(connParams),
//	)
//
//	conn, err := grpc.DialContext(ctx, addr, dialOpts...)
//	if err != nil {
//		klog.ErrorS(err, "Connect remote runtime failed", "address", addr)
//		return nil, err
//	}
//
//	service := &CRIClient{
//		timeout:      connectionTimeout,
//		logReduction: logreduction.NewLogReduction(identicalErrorDelay),
//	}
//
//	if err := service.validateServiceConnection(ctx, conn, endpoint); err != nil {
//		return nil, fmt.Errorf("validate service connection: %w", err)
//	}
//
//	return service, nil
//}

type CRIClient struct {
	runtimeapi.RuntimeServiceClient
}
