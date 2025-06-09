package analyzer

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	_ "github.com/argoproj-labs/argocd-image-updater"
)

// GetApplications gets all argocd applications that use argocd-image-updater.
// If an application contains the annotation argocd-image-updater.argoproj.io/image-list,
// it is considered using argocd-image-updater.
// If no argocd applications are found, a notFoundMessage is returned to indicate so.
// If there are argocd applications, but none of them are using argocd-image-updater, a
// notFoundMessage is returned to reflect that.
// An error is returned for any errors getting argocd applications from the cluster
func GetApplications (ctx context.Context) (apps []*v1alpha1.Application, err error, notFoundMessage string)  {

	return nil, nil, "No argocd applications are found"

}

func verifyApplication(application *v1alpha1.Application) (bool, string) {

}
