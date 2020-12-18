package main

import (
	"context"
	"fmt"
	"github.com/NYTimes/gziphandler"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-test/rpc"
)

const (
	defaultCorsHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web"
	flyHeaders         = "Fly-Client-IP, Fly-Forwarded-Port, Fly-Region, Via, X-Forwarded-For, X-Forwarded-Proto, X-Forwarded-SSL, X-Forwarded-Port"
)

type mainService struct {
}

func (m mainService) Clock(empty *rpc.Empty, server rpc.MainService_ClockServer) error {
	for {
		timeStr := time.Now().UTC().Format(time.RFC3339)
		err := server.Send(&rpc.Time{
			Timestamp: timeStr,
		})
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}

func (m mainService) Hello(ctx context.Context, empty *rpc.Empty) (*rpc.Greeting, error) {
	msg := "Hello World"
	return &rpc.Greeting{
		Message: msg,
	}, nil
}

// create a handler struct
type HttpHandler struct{} // implement `ServeHTTP` method on `HttpHandler` struct
func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { // create response binary data
	data := []byte("Hello World!") // slice of bytes    // write `data` to response
	_, _ = w.Write(data)
}

func main() {
	for _, pair := range os.Environ() {
		log.Print("====================")
		log.Print(pair)
		log.Print("====================")
	}
	testSecret := os.Getenv("TEST_SECRET")
	if testSecret == "" {
		log.Fatalf("unable to load TEST_SECRET: it's empty")
	}
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	service := mainService{}
	rpc.RegisterMainServiceService(server, &rpc.MainServiceService{Clock: service.Clock, Hello: service.Hello})
	reflection.Register(server)
	grpcWebServer := registerGrpcWebServer(server)
	handler := HttpHandler{}
	httpServer := createHttpHandler(log.New(os.Stdout, "grpcweb-test", 1), true, handler, grpcWebServer)

	log.Fatal(httpServer.Serve(listener))
	// log.Fatal(server.Serve(listener))
}

func createHttpHandler(logger *log.Logger, isGzipped bool, fileServer http.Handler, grpcWebServer *grpcweb.WrappedGrpcServer) *http.Server {
	return &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", fmt.Sprintf("%s,%s", defaultCorsHeaders, flyHeaders))
			logger.Printf("Serving Endpoint: %s", r.URL.Path)
			ct := r.Header.Get("content-type")
			if r.ProtoMajor == 2 && (strings.Contains(ct, "application/grpc") || strings.Contains(ct, "application/grpc-web")) {
				grpcWebServer.ServeHTTP(w, r)
			} else {
				if isGzipped {
					fileServer = gziphandler.GzipHandler(fileServer)
					fileServer.ServeHTTP(w, r)
				} else {
					fileServer.ServeHTTP(w, r)
				}
			}
		}), &http2.Server{}),
	}
}

func registerGrpcWebServer(srv *grpc.Server) *grpcweb.WrappedGrpcServer {
	return grpcweb.WrapServer(
		srv,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithAllowedRequestHeaders([]string{"Accept", "Cache-Control", "Keep-Alive", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-User-Agent", "X-Grpc-Web"}),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
		grpcweb.WithWebsockets(true),
	)
}
