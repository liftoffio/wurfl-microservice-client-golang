/*
Copyright 2019 ScientiaMobile Inc. http://www.scientiamobile.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/wurfl/wurfl-microservice-client-golang/v2/scientiamobile/wmclient"
	"log"
	"net/http"
	"strings"
)

// This example assumes that you have a running Wurfl Microservice server
func main() {
	var err error

	// connect
	ClientConn, err := wmclient.Create("http", "localhost", "8080", "")
	if err != nil {
		log.Fatal("wmclient.Create returned :", err.Error())
	}

	CapsList := strings.Fields("is_tablet brand_name form_factor is_full_desktop")
	ClientConn.SetRequestedCapabilities(CapsList)

	// start server and process reqs

	fmt.Println("starting http server, port 9090 /detect")

	http.HandleFunc("/detect", func(w http.ResponseWriter, r *http.Request) {

		JsonDeviceData, callerr := ClientConn.LookupRequest(*r)
		if callerr != nil {
			log.Fatal("wmclient.LookupRequest returned :", callerr.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		jstr, _ := json.MarshalIndent(JsonDeviceData, "", "  ")
		w.Write(jstr)

	})

	log.Fatal(http.ListenAndServe(":9090", nil))

}
