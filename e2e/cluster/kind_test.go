package cluster

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderManifests(t *testing.T) {
	// if os.Getenv("ACTIONS_RUNNER_DEBUG") == "true" {
	// 	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	// }
	tc := NewKind("foo", ".",
		Deploy(Deployment{ManifestFile: "pods.yml"}),
		Deploy(Deployment{Order: ExternalServices, ManifestFile: "sql"}),
		Override(Loki, Deployment{Order: ExternalServices, ManifestFile: "loki"}))

	// verify that deployments are overridden and/or inserted in proper order
	require.Equal(t, []Deployment{
		{Order: Preconditions, ManifestFile: path.Join(packageDir(), "base", "01-permissions.yml")},
		{Order: ExternalServices, ManifestFile: "sql"},
		{Order: ExternalServices, ManifestFile: "loki"},
		{Order: NetObservServices, ManifestFile: path.Join(packageDir(), "base", "03-flp.yml")},
		{Order: WithAgent, ManifestFile: path.Join(packageDir(), "base", "04-agent.yml")},
		{ManifestFile: "pods.yml"},
	}, tc.orderedManifests())
}
