package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployments_Deduplicate(t *testing.T) {
	in := Deployments{
		Deployment{
			Namespace: "ns1",
			Name:      "deploy1",
		},
		Deployment{
			Namespace: "ns2",
			Name:      "deploy2",
		},
		Deployment{
			Namespace: "ns2",
			Name:      "deploy2",
		},
		Deployment{
			Namespace: "ns1",
			Name:      "deploy1",
		},
		Deployment{
			Namespace: "ns2",
			Name:      "deploy2",
		},
		Deployment{
			Namespace: "ns3",
			Name:      "deploy1",
		},
	}

	expected := Deployments{
		Deployment{
			Namespace: "ns1",
			Name:      "deploy1",
		},
		Deployment{
			Namespace: "ns2",
			Name:      "deploy2",
		},
		Deployment{
			Namespace: "ns3",
			Name:      "deploy1",
		},
	}

	in.deduplicate()
	assert.Equal(t, in, expected, "The structs slice is deduplicated fine.")
}
