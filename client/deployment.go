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

// Deployment stores deployment namespace and name
type Deployment struct {
	Namespace string
	Name      string
}

// Deployments is an slice of deployments
type Deployments []Deployment

func (d *Deployments) deduplicate() {
	keys := make(map[Deployment]bool)
	list := Deployments{}
	for _, entry := range *d {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	*d = list
}
