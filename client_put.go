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
	model, err := c.getResource("model", K8SModel)
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
	item, err := item(event, K8SPod, "pod")
	if err != nil {
		c.Log.Errorf("Failed to get POD information: %s.", err)
		return nil, err
	}
	// push the item to the CMDB
	podKey, result, err := c.putResource(item, "item")

	if result.Error {
		return result, err
	}

	// ensure link between namespace and pod exist
	_, result, err = c.putResource(c.getLink(NS(event), podKey), "link")

	// now link the pod with any matching services
	// 1. query services in the namespace first: /item?type=K8SService&attrs=namespace,value
	// 2. for each service query the selector: meta.selector => key, value
	// 3. find pods with such selectors: item?type=K8SPod&attrs=selector_key,selector_value|namespace,value
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
	return result, err
}

func (c *Client) putResourceQuota(event []byte) {
}

func (c *Client) putPersistentVolume(event []byte) {
}

func (c *Client) putReplicationController(event []byte) {
}

func (c *Client) putIngress(event []byte) {
}
