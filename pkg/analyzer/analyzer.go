package analyzer

import (
	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	v1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"context"
	"fmt"
)

const ANALYZER_NAME = "image-updater-analyzer"

type Handler struct {
	rpc.CustomAnalyzerServiceServer
}
type Analyzer struct {
	Handler *Handler
}

func (a *Handler) Run(ctx context.Context, req *v1.RunRequest) (*v1.RunResponse, error) {
	fmt.Printf("Running %s\n", ANALYZER_NAME)

	// get all argocd applications that are using image-updater
	apps, err, message := GetApplications(ctx)
	if err != nil {
		return nil, err
	}
	if message != "" {
		response := &v1.RunResponse{
			Result: &v1.Result{
				Name:    ANALYZER_NAME,
				Details: message,
				Error: []*v1.ErrorDetail{
					&v1.ErrorDetail{
						Text: message,
					},
				},
			},
		}
		return response, nil
	}

	appResults := verifyApplications(apps)
	if len(appResults) == 0 {
		fmt.Println("Empty result from analyzing applications")
		response := &v1.RunResponse{
			Result: &v1.Result{
				Name:    ANALYZER_NAME,
				Details: "Empty result from analyzing applications",
			},
		}
		return response, nil
	}

	// create a consolidated RunResponse
	responseMessage := ""
	errorDetails := []*v1.ErrorDetail{}
	for _, appResult := range appResults {
		if appResult.err != nil {
			errorDetails = append(errorDetails, &v1.ErrorDetail{
				Text: appResult.err.Error(),
			})
		}
		responseMessage = responseMessage + appResult.toString() + "\n"
	}

	response := &v1.RunResponse{
		Result: &v1.Result{
			Name:    ANALYZER_NAME,
			Details: responseMessage,
			Error:   errorDetails,
		},
	}
	fmt.Printf("\n=========================\nConsolidated results: %s\n", responseMessage)

	return response, nil
}
