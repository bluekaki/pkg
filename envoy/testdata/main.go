package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"net"
	"net/http"

	"github.com/bluekaki/pkg/envoy/controlplane/cluster"
	"github.com/bluekaki/pkg/envoy/controlplane/listener"
	log "github.com/bluekaki/pkg/envoy/controlplane/logger"
	"github.com/bluekaki/pkg/envoy/controlplane/router"
	"github.com/bluekaki/pkg/envoy/controlplane/secret"
	"github.com/bluekaki/pkg/shutdown"
	vvs "github.com/bluekaki/pkg/vv/builder/server"
	"github.com/bluekaki/pkg/vv/proposal"
	"github.com/bluekaki/pkg/zaplog"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

//go:embed crypto/cert.pem
var certPEM string

//go:embed crypto/privkey.pem
var privkeyPEM string

//go:embed httpfilter.wasi
var httpfilterWASI []byte

func main() {
	logger, err := zaplog.NewJSONLogger(zaplog.WithInfoLevel())
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	go httpServer()

	ctx, cancel := context.WithCancel(context.Background())

	cache := cache.NewSnapshotCache(true, new(cache.IDHash), log.New(logger))
	if err := cache.SetSnapshot(ctx, "edge1", newSnapshot(logger)); err != nil {
		logger.Fatal("set snapshot cache err", zap.Error(err))
	}

	ads := server.NewServer(ctx, cache, &test.Callbacks{})

	hcheck := health.NewServer()
	hcheck.SetServingStatus("grpc.health.v1.Health", healthpb.HealthCheckResponse_SERVING)

	register := func(server *grpc.Server) {
		healthpb.RegisterHealthServer(server, hcheck)

		discoverygrpc.RegisterAggregatedDiscoveryServiceServer(server, ads)
		endpointservice.RegisterEndpointDiscoveryServiceServer(server, ads)
		clusterservice.RegisterClusterDiscoveryServiceServer(server, ads)
		routeservice.RegisterRouteDiscoveryServiceServer(server, ads)
		listenerservice.RegisterListenerDiscoveryServiceServer(server, ads)
		secretservice.RegisterSecretDiscoveryServiceServer(server, ads)
		runtimeservice.RegisterRuntimeDiscoveryServiceServer(server, ads)
	}

	notify := func(msg *proposal.AlertMessage) {
		logger.Error("got err", zap.Any("msg", json.RawMessage(msg.Marshal())))
	}

	srv := vvs.New(logger, notify, register,
		vvs.WithDisableMessageValitator(),

		vvs.WithIgnoreFileDescriptor(healthpb.File_grpc_health_v1_health_proto),

		vvs.WithIgnoreFileDescriptor(discoverygrpc.File_envoy_service_discovery_v3_ads_proto),
		vvs.WithIgnoreFileDescriptor(discoverygrpc.File_envoy_service_discovery_v3_discovery_proto),
		vvs.WithIgnoreFileDescriptor(endpointservice.File_envoy_service_endpoint_v3_eds_proto),
		vvs.WithIgnoreFileDescriptor(endpointservice.File_envoy_service_endpoint_v3_leds_proto),
		vvs.WithIgnoreFileDescriptor(clusterservice.File_envoy_service_cluster_v3_cds_proto),
		vvs.WithIgnoreFileDescriptor(routeservice.File_envoy_service_route_v3_rds_proto),
		vvs.WithIgnoreFileDescriptor(routeservice.File_envoy_service_route_v3_srds_proto),
		vvs.WithIgnoreFileDescriptor(listenerservice.File_envoy_service_listener_v3_lds_proto),
		vvs.WithIgnoreFileDescriptor(secretservice.File_envoy_service_secret_v3_sds_proto),
		vvs.WithIgnoreFileDescriptor(runtimeservice.File_envoy_service_runtime_v3_rtds_proto),
	)

	listener, err := net.Listen("tcp", "0.0.0.0:18000")
	if err != nil {
		logger.Fatal("new grpc listener err", zap.Error(err))
	}

	go func() {
		logger.Info("server trying to listen on " + listener.Addr().String())
		if err := srv.Serve(listener); err != nil {
			logger.Fatal("start server err", zap.Error(err))
		}
	}()

	shutdown.NewHook().Close(func() {
		logger.Info("shutdown being")

		cancel()
		srv.GracefulStop()
		hcheck.Shutdown()

		logger.Info("shutdown success")
	})
}

func newSnapshot(logger *zap.Logger) cache.Snapshot {
	const version = "15"

	snapshot, err := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ListenerType: {
				listener.NewHTTP_GRPC("edge_ingress_443", 443,
					listener.WithTLS("*.minami.cc", secret.NewTlsCertificate(certPEM, privkeyPEM, ""), false),
					listener.WithMaxConnections(100),
					listener.WithHttpWasmFilter("http://local.minami.cc:64320/httpfilter.wasi", "9063153f86266b9e61205ea5813bb97f4a83061940b5bce24ec5fb80f9003a26"),
					listener.WithVia("envoy@"+version),
				),
			},

			resource.RouteType: {
				router.New(
					router.NewHTTPRoute("rt", "/", "rt_cluster", router.WithAuthority("rt.minami.cc")),
				),
			},

			resource.ClusterType: {
				// --------------------- required  -------------------------
				// the wasm_cluster must be decalred for wasm filter
				cluster.New("wasm_cluster", []*cluster.Target{
					{Host: "local.minami.cc", Port: 64320},
				}, cluster.WithTCPHealthCheck()),

				// declared by http wasm filter for do log
				cluster.New("self_http_cluster", []*cluster.Target{
					{Host: "local.minami.cc", Port: 64320},
				}, cluster.WithTCPHealthCheck()),

				// --------------------- manually  -------------------------
				cluster.New("rt_cluster", []*cluster.Target{
					{Host: "www.rt.com", Port: 443, TLS: true},
				}, cluster.WithTCPHealthCheck()),
			},
		})

	if err != nil {
		logger.Fatal("create snapshot err", zap.Error(err))
	}

	if err = snapshot.Consistent(); err != nil {
		logger.Fatal("snapshot not consistent", zap.Error(err))
	}

	return snapshot
}

func httpServer() {
	// for health check
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {})

	// for http wasm filter log
	http.HandleFunc("/log", func(w http.ResponseWriter, req *http.Request) {})

	http.HandleFunc("/httpfilter.wasi", func(w http.ResponseWriter, req *http.Request) {
		w.Write(httpfilterWASI)
	})

	if err := http.ListenAndServe(":64320", nil); err != nil {
		panic(err)
	}
}
