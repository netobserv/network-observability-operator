package cluster

import (
	"context"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	apix "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestIsOpenShiftVersionLessThan(t *testing.T) {
	info := Info{openShiftVersion: semver.New("4.14.9"), ready: true}
	b, _, err := info.IsOpenShiftVersionLessThan("4.15.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.15.0")
	b, _, err = info.IsOpenShiftVersionLessThan("4.15.0")
	assert.NoError(t, err)
	assert.False(t, b)
}

func TestIsOpenShiftVersionAtLeast(t *testing.T) {
	info := Info{openShiftVersion: semver.New("4.14.9"), ready: true}
	b, _, err := info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.14.0")
	b, _, err = info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.12.0")
	b, _, err = info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.False(t, b)
}

func TestGetCRDProperty(t *testing.T) {
	crd := apix.CustomResourceDefinition{
		Spec: apix.CustomResourceDefinitionSpec{
			Versions: []apix.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha1",
					Schema: &apix.CustomResourceValidation{
						OpenAPIV3Schema: &apix.JSONSchemaProps{
							Properties: map[string]apix.JSONSchemaProps{
								"spec": {},
							},
						},
					},
				},
				{
					Name: "v1",
					Schema: &apix.CustomResourceValidation{
						OpenAPIV3Schema: &apix.JSONSchemaProps{
							Properties: map[string]apix.JSONSchemaProps{
								"spec": {
									Properties: map[string]apix.JSONSchemaProps{
										"foo": {
											Properties: map[string]apix.JSONSchemaProps{
												"bar": {},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.True(t, hasCRDProperty(context.Background(), &crd, "", "spec.foo.bar"))
	assert.False(t, hasCRDProperty(context.Background(), &crd, "v1alpha1", "spec.foo.bar"))
	assert.True(t, hasCRDProperty(context.Background(), &crd, "v1", "spec.foo.bar"))
	assert.False(t, hasCRDProperty(context.Background(), &crd, "", "spec.foo.bar.baz"))
	assert.False(t, hasCRDProperty(context.Background(), &crd, "v2", "spec"))
}
