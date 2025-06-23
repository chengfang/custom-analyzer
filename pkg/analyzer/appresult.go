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
		return fmt.Sprintf("encountered error while verifying %s/%s: %v", appResult.namespace, appResult.name, appResult.err)
	}
	if !appResult.ok {
		return fmt.Sprintf("problems found in application %s/%s: %s", appResult.namespace, appResult.name, appResult.message)
	}
	if appResult.ok {
		return fmt.Sprintf("application %s/%s properly configured", appResult.namespace, appResult.name)
	}
	return fmt.Sprintf("verified application %s/%s: %s", appResult.namespace, appResult.name, appResult.message)
}