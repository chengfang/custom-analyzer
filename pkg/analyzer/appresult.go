package analyzer

import (
	"fmt"
)

type appResult struct {
	namespace string
	name      string
	ok        bool
	message   string
	err       error
}

func (appResult *appResult) toString() string {
	if appResult.err != nil {
		return fmt.Sprintf("Encountered error while verifying %s/%s: %v", appResult.namespace, appResult.name, appResult.err)
	}
	return fmt.Sprintf("Verified application %s/%s: %s", appResult.namespace, appResult.name, appResult.message)
}