package v1beta2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/dsl"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netobserv/network-observability-operator/pkg/cluster"
)

var (
	log                    = logf.Log.WithName("flowcollector-resource")
	CurrentClusterInfo     *cluster.Info
	needPrivileged         = []AgentFeature{UDNMapping, NetworkEvents}
	neededOpenShiftVersion = map[AgentFeature]string{
		PacketDrop:    "4.14.0",
		UDNMapping:    "4.19.0",
		NetworkEvents: "4.19.0",
		EbpfManager:   "4.19.0",
	}
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateCreate(ctx context.Context, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate create", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, kerr.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.Validate(ctx, fc)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate update", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, kerr.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.Validate(ctx, fc)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	log.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *FlowCollector) Validate(ctx context.Context, fc *FlowCollector) (admission.Warnings, error) {
	var allW admission.Warnings
	var allE []error
	w, errs := r.validateAgent(ctx, &fc.Spec)
	allW, allE = collect(allW, allE, w, errs)
	w, errs = r.validateFLPConfig(ctx, &fc.Spec)
	allW, allE = collect(allW, allE, w, errs)
	w = r.warnLogLevels(&fc.Spec)
	allW, allE = collect(allW, allE, w, nil)
	return allW, errors.Join(allE...)
}

func collect(wPool admission.Warnings, errsPool []error, w admission.Warnings, errs []error) (admission.Warnings, []error) {
	if len(w) > 0 {
		wPool = append(wPool, w...)
	}
	if len(errs) > 0 {
		errsPool = append(errsPool, errs...)
	}
	return wPool, errsPool
}

func (r *FlowCollector) warnLogLevels(fc *FlowCollectorSpec) admission.Warnings {
	var w admission.Warnings
	if fc.Agent.EBPF.LogLevel == "debug" || fc.Agent.EBPF.LogLevel == "trace" {
		w = append(w, fmt.Sprintf("The log level for the eBPF agent is %s, which impacts performance and resource footprint.", fc.Agent.EBPF.LogLevel))
	}
	if fc.Processor.LogLevel == "debug" || fc.Processor.LogLevel == "trace" {
		w = append(w, fmt.Sprintf("The log level for the processor (flowlogs-pipeline) is %s, which impacts performance and resource footprint.", fc.Processor.LogLevel))
	}
	return w
}

// nolint:cyclop
func (r *FlowCollector) validateAgent(_ context.Context, fc *FlowCollectorSpec) (admission.Warnings, []error) {
	var warnings admission.Warnings
	for feat, minVersion := range neededOpenShiftVersion {
		if slices.Contains(fc.Agent.EBPF.Features, feat) {
			if CurrentClusterInfo != nil && CurrentClusterInfo.IsOpenShift() {
				// Make sure required version of ocp is installed
				ok, err := CurrentClusterInfo.OpenShiftVersionIsAtLeast(minVersion)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("Could not detect OpenShift cluster version: %s", err.Error()))
				} else if !ok {
					warnings = append(warnings, fmt.Sprintf("The %s feature requires OpenShift %s or above (version detected: %s)", feat, minVersion, CurrentClusterInfo.GetOpenShiftVersion()))
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("Unknown environment, cannot detect if the feature %s is supported", feat))
			}
		}
	}
	if !fc.Agent.EBPF.Privileged {
		for _, feat := range needPrivileged {
			if slices.Contains(fc.Agent.EBPF.Features, feat) {
				warnings = append(warnings, fmt.Sprintf("The %s feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged", feat))
			}
		}
	}

	if slices.Contains(fc.Agent.EBPF.Features, PacketDrop) &&
		!fc.Agent.EBPF.Privileged &&
		!slices.Contains(fc.Agent.EBPF.Features, EbpfManager) {
		warnings = append(warnings, "The PacketDrop feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged, or to use with eBPF Manager")
	}
	var errs []error
	if fc.Agent.EBPF.FlowFilter != nil && fc.Agent.EBPF.FlowFilter.Enable != nil && *fc.Agent.EBPF.FlowFilter.Enable {
		m := make(map[string]bool)
		for i := range fc.Agent.EBPF.FlowFilter.Rules {
			rule := fc.Agent.EBPF.FlowFilter.Rules[i]
			key := rule.CIDR + "-" + rule.PeerCIDR
			if found := m[key]; found {
				errs = append(errs, fmt.Errorf("flow filter rule CIDR and PeerCIDR %s already exists",
					key))
				break
			}
			m[key] = true
			errs = append(errs, validateFilter(&rule)...)
		}
		errs = append(errs, validateFilter(fc.Agent.EBPF.FlowFilter)...)
	}

	return warnings, errs
}

type filter interface {
	getCIDR() string
	getPorts() intstr.IntOrString
	getSrcPorts() intstr.IntOrString
	getDstPorts() intstr.IntOrString
}

func (f *EBPFFlowFilter) getCIDR() string {
	return f.CIDR
}

func (f *EBPFFlowFilter) getPorts() intstr.IntOrString {
	return f.Ports
}

func (f *EBPFFlowFilter) getSrcPorts() intstr.IntOrString {
	return f.SourcePorts
}

func (f *EBPFFlowFilter) getDstPorts() intstr.IntOrString {
	return f.DestPorts
}

func (f *EBPFFlowFilterRule) getCIDR() string {
	return f.CIDR
}

func (f *EBPFFlowFilterRule) getPorts() intstr.IntOrString {
	return f.Ports
}

func (f *EBPFFlowFilterRule) getSrcPorts() intstr.IntOrString {
	return f.SourcePorts
}

func (f *EBPFFlowFilterRule) getDstPorts() intstr.IntOrString {
	return f.DestPorts
}

func validateFilter[T filter](f T) []error {
	var errs []error

	cidr := f.getCIDR()
	if cidr != "" {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			errs = append(errs, err)
		}
	}
	ports := f.getPorts()
	hasPorts := ports.IntVal > 0 || ports.StrVal != ""
	if hasPorts {
		if err := validateFilterPortConfig(ports); err != nil {
			errs = append(errs, err)
		}
	}
	srcPorts := f.getSrcPorts()
	hasSrcPorts := srcPorts.IntVal > 0 || srcPorts.StrVal != ""
	if hasSrcPorts {
		if err := validateFilterPortConfig(srcPorts); err != nil {
			errs = append(errs, err)
		}
	}
	dstPorts := f.getDstPorts()
	hasDstPorts := dstPorts.IntVal > 0 || dstPorts.StrVal != ""
	if hasDstPorts {
		if err := validateFilterPortConfig(dstPorts); err != nil {
			errs = append(errs, err)
		}
	}
	if hasPorts && hasSrcPorts {
		errs = append(errs, errors.New("cannot configure agent filter with ports and sourcePorts, they are mutually exclusive"))
	}
	if hasPorts && hasDstPorts {
		errs = append(errs, errors.New("cannot configure agent filter with ports and destPorts, they are mutually exclusive"))
	}
	return errs
}

func validateFilterPortConfig(value intstr.IntOrString) error {
	if value.Type == intstr.Int {
		return nil
	}
	sVal := value.String()
	if strings.Contains(sVal, "-") {
		ps := strings.SplitN(sVal, "-", 2)
		if len(ps) != 2 {
			return fmt.Errorf("invalid ports range: expected two integers separated by '-' but found %s", sVal)
		}
		start, err := validatePortString(ps[0])
		if err != nil {
			return fmt.Errorf("start port in range: %w", err)
		}
		end, err := validatePortString(ps[1])
		if err != nil {
			return fmt.Errorf("end port in range: %w", err)
		}
		if start >= end {
			return fmt.Errorf("invalid port range: start is greater or equal to end")
		}
		return nil
	} else if strings.Contains(sVal, ",") {
		ps := strings.Split(sVal, ",")
		if len(ps) != 2 {
			return fmt.Errorf("invalid ports couple: expected two integers separated by ',' but found %s", sVal)
		}
		_, err := validatePortString(ps[0])
		if err != nil {
			return fmt.Errorf("first port: %w", err)
		}
		_, err = validatePortString(ps[1])
		if err != nil {
			return fmt.Errorf("second port: %w", err)
		}
		return nil
	}
	// Should be a single port then
	_, err := validatePortString(sVal)
	if err != nil {
		return err
	}
	return nil
}

func validatePortString(s string) (uint16, error) {
	p, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %w", err)
	}
	if p == 0 {
		return 0, fmt.Errorf("invalid port 0")
	}
	return uint16(p), nil
}

func (r *FlowCollector) validateFLPConfig(_ context.Context, fc *FlowCollectorSpec) (admission.Warnings, []error) {
	var errs []error
	var warnings admission.Warnings
	if fc.Processor.LogTypes != nil && *fc.Processor.LogTypes == LogTypeAll {
		warnings = append(warnings, "Enabling all log types (in spec.processor.logTypes) has a high impact on resources footprint")
	}
	if fc.Processor.LogTypes != nil && *fc.Processor.LogTypes != LogTypeFlows && fc.Loki.Enable != nil && !*fc.Loki.Enable {
		errs = append(errs, errors.New("enabling conversation tracking without Loki is not allowed, as it generates extra processing for no benefit"))
	}
	for i, filter := range fc.Processor.Filters {
		if _, err := dsl.Parse(filter.Query); err != nil {
			errs = append(errs, fmt.Errorf("cannot parse spec.processor.filters[%d].query: %w", i, err))
		}
	}
	return warnings, errs
}
