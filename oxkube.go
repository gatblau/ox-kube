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
	"github.com/sirupsen/logrus"
	"strings"
)

type OxKube struct {
	config *Config
	log    *logrus.Entry
}

func (k *OxKube) start() error {
	err := k.loadConfig()
	if err != nil {
		return err
	}
	switch k.config.Consumers.Consumer {
	case "webhook":
		wh := Webhook{
			log:    k.log,
			config: k.config.Consumers.Webhook,
		}
		wh.Start()
	case "broker":
		panic("Broker consumer is not implemented.")
	default:
		panic(fmt.Sprintf("Mode '%s' is not implemented.", k.config.Consumers.Consumer))
	}
	return nil
}

// load the configuration file
func (k *OxKube) loadConfig() error {
	// loads the configuration
	c, err := NewConfig()
	if err == nil {
		k.config = &c
	} else {
		return err
	}

	// adds the platform field to the logger
	k.log = logrus.WithFields(logrus.Fields{
		"Id": k.config.Id,
	})

	// try and parse the logging level in the configuration
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		// if the value was not recognised then return the error
		k.log.Errorf("Failed to recognise value LogLevel entry in the configuration: %s.", err)
		return err
	}
	// otherwise sets the logging level for the entire system
	logrus.SetLevel(level)
	k.log.Infof("%s has been set as the logger level.", strings.ToUpper(c.LogLevel))
	return nil
}
