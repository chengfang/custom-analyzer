package analyzer

import (
	"fmt"
	"k8s.io/utils/strings/slices"
	"os"
	"context"
	"errors"
	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	updaterCommon "github.com/argoproj-labs/argocd-image-updater/pkg/common"
	"strings"
)

var ValidShortAnnKeys = []string{
	"allow-tags",
	"force-update",
	"git-branch",
	"git-repository",
	"helm.image-name",
	"helm.image-spec",
	"helm.image-tag",
	"ignore-tags",
	"image-list",
	"kustomize.image-name",
	"platforms",
	"pull-secret",
	"update-strategy",
	"write-back-method",
	"write-back-target",
}

// GetApplications gets all argocd applications that use argocd-image-updater.
// If an application contains the annotation argocd-image-updater.argoproj.io/image-list,
// it is considered using argocd-image-updater.
// If no argocd applications are found, a notFoundMessage is returned to indicate so.
// If there are argocd applications, but none of them are using argocd-image-updater, a
// notFoundMessage is returned to reflect that.
// An error is returned for any errors getting argocd applications from the cluster
func GetApplications (ctx context.Context) (apps []v1alpha1.Application, err error, notFoundMessage string)  {
	labelSelector := ""
	envAuthToken := os.Getenv("ARGOCD_TOKEN")
	if envAuthToken == "" {
		err = errors.New("ARGOCD_TOKEN environment variable not set")
		return nil, err, ""
	}

	envServer := os.Getenv("ARGOCD_SERVER")
	if envServer == "" {
		err = errors.New("ARGOCD_SERVER environment variable not set")
		return nil, err, ""
	}

	rOpts := &argocdclient.ClientOptions{
		ServerAddr:      envServer,
		PlainText:       true,
		Insecure:        true,
		AuthToken:		 envAuthToken,
	}

	client, err := argocdclient.NewClient(rOpts)
	if err != nil {
		return nil, err, ""
	}

	conn, appClient, err := client.NewApplicationClient()
	if err != nil {
		return nil, err, ""
	}
	defer conn.Close()

	applicationQuery := &application.ApplicationQuery{
		Name:            nil,
		Refresh:         nil,
		Projects:        nil,
		ResourceVersion: nil,
		Selector:        &labelSelector,
		Repo:            nil,
		AppNamespace:    nil,
		Project:         nil,
	}
	appList, err := appClient.List(context.TODO(), applicationQuery)
	if err != nil {
		return nil, err, ""
	}

	apps = appList.Items

	if len(apps) == 0 {
		return nil, nil, "No argocd applications are found"
	}
	return apps, nil, ""
}

func verifyApplications(apps []v1alpha1.Application) []*appResult {
	appResults := []*appResult{}
	for _, app := range apps {
		result := verifyApplication(&app)
		appResults = append(appResults, result)
	}
	return appResults
}

func verifyApplication(app *v1alpha1.Application) *appResult {
	annotations := app.GetAnnotations()
	ar := &appResult{
		namespace: app.Namespace,
		name:      app.Name,
	}

	if annotations == nil || len(annotations) == 0 {
		ar.message = "No annotations found"
		return ar
	}

	// check for any unrecognized annotation keys
	badAnnotations := []string{}
	for k, v := range annotations {
		prefix := updaterCommon.ImageUpdaterAnnotationPrefix + "/"
		if strings.HasPrefix(k, prefix) {
			trimPrefix := strings.TrimPrefix(k, prefix)
			before, after, found := strings.Cut(trimPrefix, ".")
			shortKey := before
			if found {
				shortKey = after
			}
			if slices.Contains(ValidShortAnnKeys, shortKey) {
				fmt.Printf("Verified annotation: %s: %s\n", k, v)
			} else {
				fmt.Printf("Unrecognized annotation: %s: %s\n", k, v)
				badAnnotations = append(badAnnotations, fmt.Sprintf("%s: %s", k, v))
			}
		} else {
			fmt.Printf("Ignoring annotation %s on application %s", k, app.Name)
		}
	}
	if len(badAnnotations) > 0 {
		ar.message = fmt.Sprintf("Unrecognized annotations: %s\n. Valid annotations are: %s",
			strings.Join(badAnnotations, ","), strings.Join(ValidShortAnnKeys, ","))
		ar.err = errors.New(ar.message)
		return ar
	}

	// check for image-list annotation
	annVal := annotations[updaterCommon.ImageUpdaterAnnotation]
	if annVal == "" {
		ar.ok = false
		suggestion := fmt.Sprintf("Add annotation %s to the application %s",
			updaterCommon.ImageUpdaterAnnotation, app.Name)
		ar.message = fmt.Sprintf("The required %s annotation not found.\nSuggestion:%s ",
			updaterCommon.ImageUpdaterAnnotation, suggestion)
		return ar
	}


	annVal = annotations[updaterCommon.ImageUpdaterAnnotation]

	return ar
}
