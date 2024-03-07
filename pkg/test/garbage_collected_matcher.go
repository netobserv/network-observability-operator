package test

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/onsi/gomega/format"
	gomegatypes "github.com/onsi/gomega/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GarbageCollectedMatcher struct {
	ref client.Object
}

func (m *GarbageCollectedMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return false, fmt.Errorf("expected a client.Object assignable resource, got nil")
	}
	actualObj, ok := actual.(client.Object)
	if !ok {
		return false, fmt.Errorf("expected a client.Object assignable resource, got %v", reflect.TypeOf(actual))
	}
	owners := actualObj.GetOwnerReferences()
	if len(owners) == 0 {
		return false, fmt.Errorf("expected some owner references, got none")
	}
	owner := owners[0]
	typez := reflect.TypeOf(m.ref).String()
	versionType := strings.Split(typez, ".")
	if owner.Kind != versionType[1] {
		return false, fmt.Errorf("wrong Kind, expected %s, got %s", versionType[1], owner.Kind)
	}
	// For some reason APIVersion can be found in ObjectMeta.ManagedFields[0] and not in TypeMeta
	var expectedAPIVersion string
	fields := m.ref.GetManagedFields()
	if len(fields) > 0 {
		expectedAPIVersion = fields[0].APIVersion
	}
	if owner.APIVersion != expectedAPIVersion {
		return false, fmt.Errorf("wrong APIVersion, expected %s, got %s", expectedAPIVersion, owner.APIVersion)
	}
	if owner.Name != m.ref.GetName() {
		return false, fmt.Errorf("wrong Name, expected %s, got %s", m.ref.GetName(), owner.Name)
	}
	if string(owner.UID) != string(m.ref.GetUID()) {
		return false, fmt.Errorf("wrong UID, expected %v, got %v", m.ref.GetUID(), owner.UID)
	}
	return true, nil
}

func (m *GarbageCollectedMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have as owner reference", m.ref)
}

func (m *GarbageCollectedMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have as owner reference", m.ref)
}

func BeGarbageCollectedBy(ref client.Object) gomegatypes.GomegaMatcher {
	return &GarbageCollectedMatcher{
		ref: ref,
	}
}
