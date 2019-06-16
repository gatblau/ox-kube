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

import "fmt"

/*
	oxkube is a simple http server process which exposes a webhook
*/
func main() {
	// determines the operation mode i.e. webhook or msg broker
	mode := getMode()

	switch mode {
	case "webhook":
		runWebhook()
	case "broker":
		panic("Mode 'broker' is not implemented.")
	default:
		panic(fmt.Sprintf("Mode '%s' is not implemented.", getMode()))
	}
}
