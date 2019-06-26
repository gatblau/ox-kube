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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"strings"
)

const (
	Key         = "Object.metadata.name"
	Name        = "openshift.io/display-name"
	Description = "openshift.io/description"
	Created     = "Object.metadata.creationTimestamp"
	Requester   = "openshift.io/requester"
	MetaInfo    = "Object"
	Annotations = "Object.metadata.annotations"
	Cluster     = "Change.host"
)

type Client struct {
	Log    *logrus.Entry
	Token  string
	Config *Config
}

type MAP map[string]interface{}

// creates a new Onix REST web client
func NewClient(log *logrus.Entry, cfg *Config) (*Client, error) {
	client := new(Client)
	client.Log = log
	client.Config = cfg
	err := client.setAuthenticationToken()
	if err != nil {
		return client, err
	}
	return client, err
}

// Set up the authentication Token used by the client
func (c *Client) setAuthenticationToken() error {
	var err error = nil
	switch c.Config.Onix.AuthMode {
	case "basic":
		c.Log.Tracef("Setting basic authentication token.")
		c.Token = NewBasicToken(c.Config.Onix.Username, c.Config.Onix.Password)
	case "oidc":
		c.Log.Tracef("Requesting bearer authentication token.")
		c.Token, err = NewBearerToken(c.Config.Onix.TokeURI, c.Config.Onix.ClientId, c.Config.Onix.ClientSecret, c.Config.Onix.Username, c.Config.Onix.Password)
		if err != nil {
			c.Log.Errorf("Failed to authenticate with OpenId server.", err)
		} else {
			c.Log.Tracef("Bearer token acquired.")
		}
	case "none":
		c.Log.Tracef("No authentication is used to connect to the Onix CMDB.")
		c.Token = ""
	default:
		c.Log.Errorf("Cannot understand authentication mode selected: %s.", c.Config.Onix.AuthMode)
	}
	return err
}

// Make a generic HTTP request
func (c *Client) makeRequest(method string, resourceName string, key string, payload io.Reader) (*Result, error) {
	var (
		req *http.Request
		err error
	)

	// creates the request
	if len(key) > 0 {
		// with key
		req, err = http.NewRequest(method, fmt.Sprintf("%s/%s/%s", c.Config.Onix.URL, resourceName, key), payload)
	} else {
		// without key
		req, err = http.NewRequest(method, fmt.Sprintf("%s/%s", c.Config.Onix.URL, resourceName), payload)
	}
	// any errors are returned
	if err != nil {
		return &Result{Message: err.Error(), Error: true}, err
	}

	// requires a response in json format
	req.Header.Set("Content-Type", "application/json")

	// if an authentication Token has been specified then add it to the request header
	if c.Token != "" && len(c.Token) > 0 {
		req.Header.Set("Authorization", c.Token)
	}

	// submits the request
	response, err := http.DefaultClient.Do(req)

	// if the response contains an error then returns
	if err != nil {
		return &Result{Message: err.Error(), Error: true}, err
	}

	defer func() {
		if ferr := response.Body.Close(); ferr != nil {
			err = ferr
		}
	}()

	// decodes the response
	result := new(Result)
	err = json.NewDecoder(response.Body).Decode(result)

	// returns the result
	return result, err
}

// Make a GET HTTP request to the WAPI
func (c *Client) Get(resourceName string, key string) (interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", c.Config.Onix.URL, resourceName, key), nil)
	req.Header.Set("Content-Type", "application/json")
	// only add authorisation header if there is a token
	if len(c.Token) > 0 {
		req.Header.Set("Authorization", c.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer func() {
			if ferr := resp.Body.Close(); ferr != nil {
				err = ferr
			}
		}()
	}
	if err != nil {
		return nil, err
	}
	// if the response status is OK (200)
	if resp.StatusCode == 200 {
		switch {
		case resourceName == "item":
			result := new(Item)
			err = json.NewDecoder(resp.Body).Decode(result)
			return *result, err
		case resourceName == "itemtype":
			result := new(ItemType)
			err = json.NewDecoder(resp.Body).Decode(result)
			return *result, err
		case resourceName == "link":
			result := new(Link)
			err = json.NewDecoder(resp.Body).Decode(result)
			return *result, err
		case resourceName == "linktype":
			result := new(LinkType)
			err = json.NewDecoder(resp.Body).Decode(result)
			return *result, err
		case resourceName == "model":
			result := new(Model)
			err = json.NewDecoder(resp.Body).Decode(result)
			return *result, err
		}
		// if the response status is something other than not found
	} else if resp.StatusCode != 404 {
		// return an error with the status message
		return nil, errors.New(resp.Status)
	}
	// the model was not found
	return nil, nil
}

// Executes an HTTP PUT request to the Onix WAPI passing the following parameters:
// - payload: the payload object
// - resourceName: the WAPI resource name (e.g. item, itemtype, link, etc.)
func (c *Client) Put(payload Payload, resourceName string) (*Result, error) {
	// converts the passed-in payload to a bytes Reader
	bytes, err := payload.ToJSON()

	// any errors are returned immediately
	if err != nil {
		return nil, err
	}

	// make an http put request to the service
	return c.makeRequest(PUT, resourceName, payload.KeyValue(), bytes)
}

// Make a DELETE HTTP request to the WAPI
func (c *Client) Delete(payload Payload, resourceName string, result interface{}) (*Result, error) {
	// make an http put request to the service
	return c.makeRequest(DELETE, resourceName, payload.KeyValue(), nil)
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
}

func (c *Client) putService(event []byte) {
}

func (c *Client) putResourceQuota(event []byte) {
}

func (c *Client) putPersistentVolume(event []byte) {
}

func (c *Client) putReplicationController(event []byte) {
}

func (c *Client) putIngress(event []byte) {
}

// issues an http put request to the Onix CMDB passing the specified item
// returns the payload key and a success flag
func (c *Client) putResource(payload Payload, resourceName string) (string, bool) {
	result, err := c.Put(payload, resourceName)
	if err != nil {
		c.Log.Errorf("Failed to PUT %s: %s.", resourceName, err)
		return "", false
	}
	if result.Error {
		c.Log.Errorf("Failed to PUT %s: %s.", resourceName, result.Message)
		return "", false
	}
	if result.Changed {
		c.Log.Tracef("%s: %s successfully updated in Onix.", resourceName, payload.KeyValue())
		return payload.KeyValue(), true
	}
	c.Log.Tracef("%s: %s, Onix reports nothing to update.", resourceName, payload.KeyValue())
	return payload.KeyValue(), true
}

func (c *Client) getNamespaceItem(event []byte) (*Item, error) {
	cluster := gjson.GetBytes(event, Cluster)
	key := gjson.GetBytes(event, Key)
	annot := gjson.GetBytes(event, Annotations).Map()
	meta := gjson.GetBytes(event, MetaInfo)
	created := gjson.GetBytes(event, Created)
	item := &Item{
		Key:         fmt.Sprintf("k8s:cluster:%s:ns:%s", cluster.String(), key.String()),
		Name:        annot[Name].String(),
		Description: annot[Description].String(),
		Meta:        MAP{},
		Attribute:   MAP{},
		Type:        K8SNamespace,
	}
	item.Attribute["Requester"] = annot[Requester].String()
	item.Attribute["Created"] = created.String()

	err := json.Unmarshal([]byte(meta.String()), &item.Meta)
	if err != nil {
		c.Log.Errorf("Failed to unmarshal event metadata: %s.", err)
		return nil, err
	}
	return item, nil
}

func (c *Client) getClusterItem(event []byte) *Item {
	host := gjson.GetBytes(event, Cluster)
	return &Item{
		Key:         fmt.Sprintf("k8s:cluster:%s", host.String()),
		Name:        fmt.Sprintf("%s Container Platform", strings.ToUpper(host.String())),
		Description: "A Kubernetes Cluster instance.",
		Type:        K8SCluster,
	}
}

func (c *Client) getLink(startItem string, endItem string) Payload {
	return &Link{
		Key:          fmt.Sprintf("%s->%s", startItem, endItem),
		StartItemKey: startItem,
		EndItemKey:   endItem,
		Type:         K8SLink,
	}
}

func (c *Client) deleteNamespace(bytes []byte) {
	panic("deleteNamespace() not implemented")
}

func (c *Client) deletePod(bytes []byte) {
	panic("deletePod() not implemented")
}

func (c *Client) deleteService(bytes []byte) {
	panic("deleteService() not implemented")
}

func (c *Client) deleteResourceQuota(bytes []byte) {
	panic("deleteResourceQuota() not implemented")
}

func (c *Client) deletePersistentVolume(bytes []byte) {
	panic("deletePersistentVolume() not implemented")
}

func (c *Client) deleteIngress(bytes []byte) {
	panic("deleteIngress() not implemented")
}

func (c *Client) deleteReplicationController(bytes []byte) {
	panic("deleteReplicationController() not implemented")
}
