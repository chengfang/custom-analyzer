package analyzer

import (
    "context"
    "fmt"

    rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
    v1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
    "github.com/ricochet2200/go-disk-usage/du"
)

type Handler struct {
    rpc.CustomAnalyzerServiceServer
}
type Analyzer struct {
    Handler *Handler
}

func (a *Handler) RunError(context.Context, *v1.RunRequest) (*v1.RunResponse, error) {
    response := &v1.RunResponse{
        Result: &v1.Result{
            Name:    "example",
            Details: "example",
            Error: []*v1.ErrorDetail{
                &v1.ErrorDetail{
                    Text: "This is an example error message!",
                },
            },
        },
    }

    return response, nil
}

func (a *Handler) Run(context.Context, *v1.RunRequest) (*v1.RunResponse, error) {
    println("Running analyzer")
    usage := du.NewDiskUsage("/")
    diskUsage := int((usage.Size() - usage.Free()) * 100 / usage.Size())
    return &v1.RunResponse{
        Result: &v1.Result{
            Name:    "diskuse",
            Details: fmt.Sprintf("Disk usage is %d", diskUsage),
            Error: []*v1.ErrorDetail{
                {
                    Text: fmt.Sprintf("Disk usage is %d", diskUsage),
                },
            },
        },
    }, nil
}

