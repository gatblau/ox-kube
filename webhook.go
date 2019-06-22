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
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Webhook struct {
	log    *logrus.Entry
	config WebhookConf
}

// launch a webhook on a TCP port listening for events
func (c *Webhook) Start() {
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
		io.WriteString(w, "OxKube webhook is ready.\n"+
			"Use an HTTP POST to send events.")
	case "POST":

	}
}

func (c *Webhook) rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	switch r.Method {
	case "GET":
		io.WriteString(w, fmt.Sprintf("OxKube is ready.\n"+
			"POST events to webhook: /%s.", c.config.Path))
	case "POST":
	case "PUT":
	case "DELETE":
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
	}
}
