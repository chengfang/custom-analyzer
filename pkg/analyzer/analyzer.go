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

//func (a *Handler) RunError(context.Context, *v1.RunRequest) (*v1.RunResponse, error) {
//    response := &v1.RunResponse{
//        Result: &v1.Result{
//            Name:    "example",
//            Details: "example",
//            Error: []*v1.ErrorDetail{
//                &v1.ErrorDetail{
//                    Text: "This is an example error message!",
//                },
//            },
//        },
//    }
//
//    return response, nil
//}

func (a *Handler) Run(ctx context.Context, req *v1.RunRequest) (*v1.RunResponse, error) {
    println("Running analyzer")

    // get all argocd applications that are using image-updater
    apps, err, message := GetApplications(ctx)

    for _, app := range apps {
        // analyze each application and validate its annotations

    }
}

