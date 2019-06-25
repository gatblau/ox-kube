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
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Webhook struct {
	log    *logrus.Entry
	config WebhookConf
	ox     *Client
}

// launch a webhook on a TCP port listening for events
func (c *Webhook) Start(client *Client) {
	// set the ox client
	c.ox = client

	// creates an http server listening on the specified TCP port
	server := &http.Server{Addr: fmt.Sprintf(":%s", c.config.Port), Handler: nil}

	// registers web handlers
	c.log.Tracef("Registering handler for web root /")
	http.HandleFunc("/", c.rootHandler)

	c.log.Tracef("Registering handler for web path /%s.", c.config.Path)
	http.HandleFunc(fmt.Sprintf("/%s", c.config.Path), c.webhookHandler)

	// runs the server asynchronously
	go func() {
		c.log.Println(fmt.Sprintf("OxKube listening on :%s", c.config.Port))
		if err := server.ListenAndServe(); err != nil {
			c.log.Fatal(err)
		}
	}()

	// creates a channel to pass a SIGINT (ctrl+C) kernel signal with buffer capacity 1
	stop := make(chan os.Signal, 1)

	// sends any SIGINT signal to the stop channel
	signal.Notify(stop, os.Interrupt)

	// waits for the SIGINT signal to be raised (pkill -2)
	<-stop

	// gets a context with some delay to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// releases resources if main completes before the delay period elapses
	defer cancel()

	// on error shutdown
	if err := server.Shutdown(ctx); err != nil {
		c.log.Fatal(err)
		c.log.Println("Shutting down Webhook consumer.")
	}
}

func (c *Webhook) webhookHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case "GET":
		_, _ = io.WriteString(w, "OxKube webhook is ready.\n"+
			"Use an HTTP POST to send events.")
	case "POST":
		c.process(w, r)
	}
}

func (c *Webhook) rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case "GET":
		_, _ = io.WriteString(w, fmt.Sprintf("OxKube is ready.\n"+
			"POST events to webhook: /%s.", c.config.Path))
	case "POST":
	case "PUT":
	case "DELETE":
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("405 - Method Not Allowed"))
	}
}

func (c *Webhook) process(w http.ResponseWriter, r *http.Request) error {
	// get the request data
	event, err := c.getRequest(r)

	if err != nil {
		return err
	}

	result := gjson.GetBytes(event, "Change.kind")

	switch strings.ToLower(result.String()) {
	case "namespace":
		c.ox.putNamespace(event)
	case "pod":
		c.ox.putPod(event)
	case "service":
		c.ox.putService(event)
	case "resourcequota":
		c.ox.putResourceQuota(event)
	case "persistenvolume":
		c.ox.putPersistentVolume(event)
	case "ingress":
		c.ox.putIngress(event)
	case "replicationcontroller":
		c.ox.putReplicationController(event)
	}

	return nil
}

// unmarshal the http request into a json like structure
func (c *Webhook) getRequest(r *http.Request) ([]byte, error) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
