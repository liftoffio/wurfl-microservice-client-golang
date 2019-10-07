# WM Golang Client API

Golang implementation of the WM Client api

Example api use looking up a single UserAgent :

```
package main

import (
	"fmt"
	"log"
	"scientiamobile/wmclient"
	"strings"
)

func main() {
	var err error

	// First we need to create a WM client instance, to connect to our WM server API at the specified host and port.
	ClientConn, err := wmclient.Create("localhost", "8080")
	if err != nil {
		// problems such as network errors  or internal server problems
		log.Fatal("wmclient.Create returned :", err.Error())
	}

	// set the capabilities we want to receive from WM server
	CapsList := strings.Fields("is_mobile is_tablet is_smartphone model_name brand_name")
	ClientConn.SetRequestedCapabilities(CapsList)

	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3"

	// Perform a device detection calling WM server API
	JsonDeviceData, callerr := ClientConn.LookupUserAgent(ua)
	if callerr != nil {
		log.Fatal("wmclient.Create returned %s\n", err.Error())
	}

	// Let's get the device capabilities and print some of them
	if JsonDeviceData.Capabilities["is_smartphone"] == "true" {
		fmt.Println("This is a is_smartphone")
	}

	fmt.Println(JsonDeviceData)
	fmt.Println(JsonDeviceData.Capabilities["wurfl_id"])
}
```

Another example of API use inside a golang http server using LookupRequest() :

```
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"scientiamobile/wmclient"
	"strings"
)

func main() {
	var err error

	// First we need to create a WM client instance, to connect to our WM server API at the specified host and port.
	ClientConn, err := wmclient.Create("localhost", "8080")
	if err != nil {
		// problems such as network errors  or internal server problems
		log.Fatal("wmclient.Create returned :", err.Error())
	}

	CapsList := strings.Fields("is_tablet brand_name form_factor is_full_desktop")
	ClientConn.SetRequestedCapabilities(CapsList)

	// start server and process reqs

	fmt.Println("starting http server, port 9090 /detect")

	http.HandleFunc("/detect", func(w http.ResponseWriter, r *http.Request) {

		// Perform a device detection calling WM server API
		JsonDeviceData, callerr := ClientConn.LookupRequest(*r)
		if callerr != nil {
			log.Fatal("wmclient.LookupRequest returned :", callerr.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		
		if JsonDeviceData.Capabilities["is_smartphone"] == "true" {
			fmt.Println("This is a is_smartphone")
		}

		
		// return json data
		jstr, _ := json.MarshalIndent(JsonDeviceData, "", "  ")
		w.Write(jstr)

	})

	log.Fatal(http.ListenAndServe(":9090", nil))

}
```


# wmclient APIs

See [wmclient.md](wmclient.md)
