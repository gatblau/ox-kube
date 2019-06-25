/*
   Onix Kube - Copyright (c) 2019 by www.gatblau.org

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software distributed under
   the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
   either express or implied.
   See the License for the specific language governing permissions and limitations under the License.

   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/
package main

import (
	"fmt"
)

const (
	K8SModel                 = "KUBE"
	K8SCluster               = "K8SCluster"
	K8SNamespace             = "K8SNamespace"
	K8SResourceQuota         = "K8SResourceQuota"
	K8SPod                   = "K8SPod"
	K8SService               = "K8SService"
	K8SIngress               = "K8SIngress"
	K8SReplicationController = "K8SReplicationController"
	K8SPersistentVolume      = "K8SPersistentVolume"
	K8SLink                  = "K8SLink"
)

// checks the kube model is defined in Onix
func ModelExists(ox *Client) (bool, error) {
	model, err := ox.Get("model", "kube")
	if err != nil {
		return false, err
	}
	return model != nil, nil
}

func CreateModel(c *Client) (*Result, error) {
	// defines the kube meta model
	modelData := &Data{
		Models: []Model{
			Model{
				Key:         K8SModel,
				Name:        "Kubernetes Resource Model",
				Description: "Defines the item and link types that describe Kubernetes resources in a given Namespace.",
			},
		},
		ItemTypes: []ItemType{
			ItemType{
				Key:         K8SCluster,
				Name:        "Kuebernetes Cluster",
				Description: "An open-source system for automating deployment, scaling, and management of containerized applications.",
				Model:       K8SModel,
			},
			ItemType{
				Key:         K8SNamespace,
				Name:        "Namespace",
				Description: "A way to divide cluster resources between multiple users or teams providing virtual areas to deploy project resources.",
				Model:       K8SModel,
			},
			ItemType{
				Key:         K8SResourceQuota,
				Name:        "Resource Quota",
				Description: "A set of constraints that limit aggregate resource consumption per namespace.",
				Model:       K8SModel,
			},
			ItemType{
				Key:         K8SPod,
				Name:        "Pod",
				Description: "Encapsulates an application’s container (or, in some cases, multiple containers), storage resources, a unique network IP, and options that govern how the container(s) should run.",
				Model:       K8SModel,
			},
			ItemType{
				Key:         K8SService,
				Name:        "Service",
				Description: "Exposes an application running on a set of Pods as a network service.",
				Model:       K8SModel,
			},
			ItemType{
				Key:  K8SIngress,
				Name: "Ingress (Route)",
				Description: "Exposes HTTP and HTTPS routes from outside the cluster to services within the cluster.\n" +
					"Traffic routing is controlled by rules defined on the Ingress resource.",
				Model: K8SModel,
			},
			ItemType{
				Key:         K8SReplicationController,
				Name:        "Replication Controller",
				Description: "Ensures that a specified number of pod replicas are running at any one time.",
				Model:       K8SModel,
			},
			ItemType{
				Key:         K8SPersistentVolume,
				Name:        "Persistent Volume",
				Description: "A piece of storage in the cluster against which, claims can be made by pods.",
				Model:       K8SModel,
			},
		},
		LinkTypes: []LinkType{
			LinkType{
				Key:         K8SLink,
				Name:        "Kubernetes Resource Link Type",
				Description: "Links Kubernetes resources.",
				Model:       K8SModel,
			},
		},
		LinkRules: []LinkRule{
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SCluster, K8SNamespace),
				Name:             "K8S Cluster to Namespace Rule",
				Description:      "A cluster contains one or more namespaces.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SCluster,
				EndItemTypeKey:   K8SNamespace,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SNamespace, K8SResourceQuota),
				Name:             "K8S Namespace to Resource Quota Rule",
				Description:      "A namespace has a resource quota.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SNamespace,
				EndItemTypeKey:   K8SResourceQuota,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SNamespace, K8SPod),
				Name:             "K8S Namespace to Pod Rule",
				Description:      "A namespace contains one or more pods.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SNamespace,
				EndItemTypeKey:   K8SPod,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SPod, K8SPersistentVolume),
				Name:             "K8S Pod to Persistent Volume Rule",
				Description:      "A pod uses one or more persistent volumes.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SPod,
				EndItemTypeKey:   K8SPersistentVolume,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SPod, K8SReplicationController),
				Name:             "K8S Pod to Replication Controller Rule",
				Description:      "A pod is controlled by a replication controller.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SPod,
				EndItemTypeKey:   K8SReplicationController,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SPod, K8SService),
				Name:             "K8S Pod to Service Rule",
				Description:      "A pod is accessed by a service.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SPod,
				EndItemTypeKey:   K8SService,
			},
			LinkRule{
				Key:              fmt.Sprintf("%s->%s", K8SService, K8SIngress),
				Name:             "K8S Service to Ingress Rule",
				Description:      "A service is published via an Ingress route.",
				LinkTypeKey:      K8SLink,
				StartItemTypeKey: K8SService,
				EndItemTypeKey:   K8SIngress,
			},
		},
	}
	// create the model
	result, err := c.Put(modelData, "data")
	return result, err
}
