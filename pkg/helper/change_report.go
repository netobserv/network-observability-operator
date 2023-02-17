package helper

import (
	"context"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ChangeReport is a logging utility for troubleshooting operator issues
type ChangeReport struct {
	title  string
	report []string
}

func NewChangeReport(title string) ChangeReport {
	return ChangeReport{title: title}
}

func (r *ChangeReport) Add(msg string) {
	r.report = append(r.report, msg)
}

func (r *ChangeReport) Check(msg string, change bool) bool {
	if change {
		r.report = append(r.report, msg)
	}
	return change
}

func (r *ChangeReport) LogIfNeeded(ctx context.Context) {
	if len(r.report) > 0 {
		log.FromContext(ctx).Info(r.String())
	}
}

func (r *ChangeReport) String() string {
	sb := strings.Builder{}
	sb.WriteString(r.title)
	sb.WriteRune(':')
	if len(r.report) > 0 {
		for _, str := range r.report {
			sb.WriteRune('<')
			sb.WriteString(str)
		}
	} else {
		sb.WriteString("no change")
	}
	return sb.String()
}
