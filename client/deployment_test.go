/*
 * Copyright (c) 2021 Angel Abad. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
