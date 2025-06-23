package main

import (
    "errors"
    "fmt"
    "net"
    "net/http"

    rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
    "github.com/chengfang/custom-analyzer/pkg/analyzer"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

func main() {
    var err error
    address := fmt.Sprintf(":%s", "8085")
    fmt.Printf("starting %s at %s\n", analyzer.ANALYZER_NAME, address)

    lis, err := net.Listen("tcp", address)
    if err != nil {
        panic(err)
    }
    grpcServer := grpc.NewServer()
    reflection.Register(grpcServer)

    aa := analyzer.Analyzer{}
    rpc.RegisterCustomAnalyzerServiceServer(grpcServer, aa.Handler)
    if err := grpcServer.Serve(
        lis,
    ); err != nil && !errors.Is(err, http.ErrServerClosed) {
        return
    }
}

