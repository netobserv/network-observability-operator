package status

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConditionsToString(conds []metav1.Condition) string {
	var strConds []string
	for _, cond := range conds {
		strConds = append(strConds, fmt.Sprintf("- %s: %s / %s / %s", cond.Type, string(cond.Status), cond.Reason, cond.Message))
	}
	return strings.Join(strConds, "\n")
}
