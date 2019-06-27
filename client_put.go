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
	_, success := c.putResource(c.getModel(), "data")
	return success
}

func (c *Client) putNamespace(event []byte) {
	// ensures the K8S cluster config item exists
	clusterKey, success := c.putResource(c.getClusterItem(event), "item")

	if !success {
		return
	}

	// gets the namespace item information
	item, err := c.getNamespaceItem(event)
	if err != nil {
		c.Log.Errorf("Failed to get Namespace information: %s", err)
		return
	}
	// push the item to the CMDB
	namespaceKey, success := c.putResource(item, "item")

	if !success {
		return
	}

	// push a link between items
	c.putResource(c.getLink(clusterKey, namespaceKey), "link")
}

func (c *Client) putPod(event []byte) {
	// gets the pod item information
	item, err := c.getPodItem(event)
	if err != nil {
		c.Log.Errorf("Failed to get POD information: %s.", err)
		return
	}
	// push the item to the CMDB
	podKey, success := c.putResource(item, "item")

	if !success {
		return
	}

	// ensure link between namespace and pod exist
	c.putResource(c.getLink(NS(event), podKey), "link")
}

func (c *Client) putService(event []byte) {
	// gets the service item information
	item, err := c.getServiceItem(event)
	if err != nil {
		c.Log.Errorf("Failed to get SERVICE information: %s.", err)
		return
	}
	// push the item to the CMDB
	c.putResource(item, "item")
}

func (c *Client) putResourceQuota(event []byte) {
}

func (c *Client) putPersistentVolume(event []byte) {
}

func (c *Client) putReplicationController(event []byte) {
}

func (c *Client) putIngress(event []byte) {
}
