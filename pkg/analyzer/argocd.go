package analyzer

import (
	"fmt"
	"k8s.io/utils/strings/slices"
	//"os"
	"context"
	"errors"
	updaterCommon "github.com/argoproj-labs/argocd-image-updater/pkg/common"
	updaterKube "github.com/argoproj-labs/argocd-image-updater/pkg/kube"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const KUBE_ANNOTATION_PREFIX = "kubectl.kubernetes.io/"

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
	kubeConfigPath := ""
	labelSelector := ""
	argocdNamespace := "argocd"
	appListOpts := v1.ListOptions{LabelSelector: labelSelector}

	fmt.Printf("Creating Kubernetes client with kubeConfigPath: %s and argocdNamespace: %s\n", kubeConfigPath, argocdNamespace)
	kubernetesClient, err := updaterKube.NewKubernetesClient(ctx, kubeConfigPath, argocdNamespace)
	if err != nil {
		return nil, err, notFoundMessage
	}

	fmt.Printf("Getting application list\n")
	appList, err := kubernetesClient.ApplicationsClientset.ArgoprojV1alpha1().Applications(argocdNamespace).List(context.TODO(), appListOpts)
	if err != nil {
		return nil, err, notFoundMessage
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
		appResults = verifyApplication(&app, appResults)
	}
	return appResults
}

func verifyApplication(app *v1alpha1.Application, appResults []*appResult) []*appResult {
	fmt.Printf("\nVerifying application %s\n-------------------------\n", app.Name)
	annotations := app.GetAnnotations()

	if annotations == nil || len(annotations) == 0 {
		ar := &appResult{
			namespace: app.Namespace,
			name:      app.Name,
			message:   "No annotations found",
			ok:        false,
		}
		appResults = append(appResults, ar)
		return appResults
	}

	// check for any unrecognized annotation keys
	badAnnotations := []string{}
	for k, v := range annotations {
		if strings.HasPrefix(k, KUBE_ANNOTATION_PREFIX) {
			// ignore any system level annotation
			continue
		}
		prefix := updaterCommon.ImageUpdaterAnnotationPrefix + "/"
		if strings.HasPrefix(k, prefix) {
			trimPrefix := strings.TrimPrefix(k, prefix)
			before, after, found := strings.Cut(trimPrefix, ".")
			shortKey := before
			if found {
				shortKey = after
			}
			if slices.Contains(ValidShortAnnKeys, shortKey) {
				fmt.Printf("\u2713 Verified annotation: %s: %s\n", k, v)
			} else {
				fmt.Printf("\u2717 Unrecognized annotation: %s: %s\n", k, v)
				badAnnotations = append(badAnnotations, fmt.Sprintf("%s: %s", k, v))
			}
		} else {
			fmt.Printf("Ignoring annotation %s on application %s", k, app.Name)
		}
	}
	if len(badAnnotations) > 0 {
		ar := &appResult{
			namespace: app.Namespace,
			name:      app.Name,
			ok:        false,
		}
		ar.message = fmt.Sprintf("Unrecognized annotations: %s in application %s/%s\n. Valid annotations are: %s",
			strings.Join(badAnnotations, ", "), app.Namespace, app.Name, strings.Join(ValidShortAnnKeys, ", "))
		ar.err = errors.New(ar.message)
		appResults = append(appResults, ar)
	}

	// check for image-list annotation
	annVal := annotations[updaterCommon.ImageUpdaterAnnotation]
	if annVal == "" {
		ar := &appResult{
			namespace: app.Namespace,
			name:      app.Name,
			ok:        false,
		}
		msg := fmt.Sprintf("The required annotation not found.\nSuggestion:\n Add annotation %s to the application %s/%s",
			updaterCommon.ImageUpdaterAnnotation, app.Namespace, app.Name)
		fmt.Printf("\u2717 %s\n",msg)
		ar.message = msg
		ar.err = errors.New(msg)
		appResults = append(appResults, ar)
	}
	return appResults
}
