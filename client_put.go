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
func (c *Client) modelExists() (bool, error) {
	model, err := c.getResource("model", K8SModel, nil)
	if err != nil {
		return false, err
	}
	return model != nil, nil
}

func (c *Client) putModel() bool {
	_, result, _ := c.putResource(c.getModel(), "data")
	return result.Error
}

func (c *Client) putNamespace(event []byte) (*Result, error) {
	// ensures the K8S cluster config item exists
	clusterKey, result, err := c.putResource(c.getClusterItem(event), "item")

	if result.Error {
		return result, err
	}

	// gets the namespace item information
	item, err := item(event, K8SNamespace, "ns")
	if err != nil {
		c.Log.Errorf("Failed to get Namespace information: %s", err)
		return result, err
	}
	// push the item to the CMDB
	namespaceKey, result, err := c.putResource(item, "item")

	if result.Error {
		return result, err
	}

	// push a link between items
	_, result, err = c.putResource(c.getLink(clusterKey, namespaceKey), "link")
	return result, err
}

func (c *Client) putPod(event []byte) (*Result, error) {
	// gets the pod item information
	pod, err := item(event, K8SPod, "pod")
	if err != nil {
		c.Log.Errorf("Failed to get POD information: %s.", err)
		return nil, err
	}
	// push the item to the CMDB
	podKey, result, err := c.putResource(pod, "item")

	if check(result, err) {
		return result, err
	}

	// ensure link between namespace and pod exist
	_, result, err = c.putResource(c.getLink(NS(event), podKey), "link")

	// link the pod with services
	_, _ = c.linkPodToK8SObject(K8SService, pod)

	// link the pod with replication controllers
	_, _ = c.linkPodToK8SObject(K8SReplicationController, pod)

	return result, err
}

func (c *Client) putService(event []byte) (*Result, error) {
	// gets the service item information
	item, err := item(event, K8SService, "svc")
	if err != nil {
		c.Log.Errorf("Failed to get SERVICE information: %s.", err)
		return nil, err
	}
	// push the item to the CMDB
	_, result, err := c.putResource(item, "item")

	// check if there are pods that should be linked to this service
	_, _ = c.linkK8SObjectToPods(item)

	return result, err
}

func (c *Client) putReplicationController(event []byte) (*Result, error) {
	// gets the service item information
	item, err := item(event, K8SReplicationController, "rc")
	if err != nil {
		c.Log.Errorf("Failed to get REPLICATION CONTROLLER information: %s.", err)
		return nil, err
	}
	// push the item to the CMDB
	_, result, err := c.putResource(item, "item")

	// check if there are pods that should be linked to this replication controller
	_, _ = c.linkK8SObjectToPods(item)

	return result, err
}

func (c *Client) putPersistentVolume(event []byte) (*Result, error) {
	// gets the persistent volume item information
	item, err := item(event, K8SPersistentVolume, "pv")
	if err != nil {
		c.Log.Errorf("Failed to get PERSISTENT VOLUME information: %s.", err)
		return nil, err
	}
	// push the volume to the CMDB
	_, result, err := c.putResource(item, "item")

	return result, err
}

func (c *Client) putResourceQuota(event []byte) {
}

func (c *Client) putIngress(event []byte) {
}

// link the passed in pod with any K8S objects in the namespace
// by matching the objects selectors with the pod labels
func (c *Client) linkPodToK8SObject(itemType string, pod *Item) (*Result, error) {
	// now link the pod with any matching services
	filters := map[string]string{
		"type":  itemType,
		"attrs": fmt.Sprintf("namespace,%s", pod.Attribute["namespace"].(string)),
	}
	// query services in the namespace first: /item?type=K8SService&attrs=namespace,value
	objects, err := c.getResource("item", "", filters)

	if err != nil {
		return nil, err
	}
	// unwraps the response into a list of service items
	k8sObjs := objects.(*ResultList).Values
	for _, k8sObj := range k8sObjs {
		// for each k8s object check if the selectors match the pod labels
		if selector, ok := k8sObj.Meta["selector"]; ok {
			selectors := selector.(map[string]interface{})
			for selectorKey, selectorValue := range selectors {
				// if the pod label matches the service descriptor
				if pod.Attribute[selectorKey] == selectorValue {
					// link the k8s object with the pod
					_, result, err := c.putResource(c.getLink(pod.Key, k8sObj.Key), "link")
					if err != nil || result.Error {
						return result, err
					}
				}
			}
		}
	}
	return &Result{}, nil
}

// link the passed-in K8S object with any existing pods in the namespace
// by matching the pods labels with the object selectors
func (c *Client) linkK8SObjectToPods(k8sObj *Item) (*Result, error) {
	// now link the pod with any matching services
	filters := map[string]string{
		"type":  K8SPod,
		"attrs": fmt.Sprintf("namespace,%s", k8sObj.Attribute["namespace"].(string)),
	}
	// get pods in the namespace first
	podsObj, err := c.getResource("item", "", filters)

	if err != nil {
		return nil, err
	}
	// unwraps the response into a list of pod items
	pods := podsObj.(*ResultList).Values
	for _, pod := range pods {
		if selector, ok := k8sObj.Meta["selector"]; ok {
			selectors := selector.(map[string]interface{})
			for selectorKey, selectorValue := range selectors {
				// if the pod label matches the object service descriptor
				if pod.Attribute[selectorKey] == selectorValue {
					// link the object with the pod
					_, result, err := c.putResource(c.getLink(pod.Key, k8sObj.Key), "link")
					if err != nil || result.Error {
						return result, err
					}
				}
			}
		}
	}
	return &Result{}, nil
}

// link the passed-in pod to any persistent volume via pod's PVCs
func (c *Client) linkPodToPvs(pod *Item) (*Result, error) {
	if volume, ok := pod.Meta["volumes"]; ok {
		volumes := volume.([]map[string]interface{})
		for _, volumeMap := range volumes {
			for k, v := range volumeMap {
				if k == "persistentVolumeClaim" {
					pvc := v.(map[string]interface{})
					claim := pvc["claimName"]
					c.Log.Tracef("Found PVC %s", claim)
				}
			}
		}
	}
	return &Result{}, nil
}
