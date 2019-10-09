# WM Golang Client API

Golang implementation of the WM Client api

Example api usage :

```go
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	 "github.com/wurfl/wurfl-microservice-client-golang/v2/scientiamobile/wmclient"
)

// Implements sort.Interface for []wmclient.JSONModelMktName
type ByModelName []wmclient.JSONModelMktName

func (c ByModelName) Len() int {
	return len(c)
}

func (c ByModelName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ByModelName) Less(i, j int) bool {
	return c[i].ModelName < c[j].ModelName
}

func main() {
	var err error

	// First we need to create a WM client instance, to connect to our WM server API at the specified host and port.
	ClientConn, err := wmclient.Create("http", "localhost", "80", "")
	if err != nil {
		// problems such as network errors  or internal server problems
		log.Fatal("wmclient.Create returned :", err.Error())
	}
	// By setting the cache size we are also activating the caching option in WM client. In order to not use cache, you just to need to omit setCacheSize call
	ClientConn.SetCacheSize(100000)

	// We ask Wm server API for some Wm server info such as server API version and info about WURFL API and file used by WM server.
	info, ierr := ClientConn.GetInfo()
	if ierr != nil {
		fmt.Println("Error getting server info: " + ierr.Error())
	} else {
		fmt.Println("WURFL Microservice information:")
		fmt.Println("Server version: " + info.WmVersion)
		fmt.Println("WURFL API version: " + info.WurflAPIVersion)
		fmt.Printf("WURFL file info: %s \n", info.WurflInfo)
	}

	// set the capabilities we want to receive from WM server
	// Static capabilities
	sCapsList := strings.Fields("model_name brand_name")
	// Virtual capabilities
	vCapsList := strings.Fields("is_smartphone form_factor")
	ClientConn.SetRequestedStaticCapabilities(sCapsList)
	ClientConn.SetRequestedVirtualCapabilities(vCapsList)

	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3"

	// Perform a device detection calling WM server API
	JSONDeviceData, callerr := ClientConn.LookupUserAgent(ua)

	if callerr != nil {
		// Applicative error, ie: invalid input provided
		log.Fatalf("Error getting device data %s\n", callerr.Error())
	}

	// Let's get the device capabilities and print some of them
	fmt.Printf("WURFL device id : %s\n", JSONDeviceData.Capabilities["wurfl_id"])

	// print brand & model (static capabilities)
	fmt.Printf("This device is a : %s %s\n", JSONDeviceData.Capabilities["brand_name"], JSONDeviceData.Capabilities["model_name"])

	// check if device is a smartphone (a virtual capability)
	if JSONDeviceData.Capabilities["is_smartphone"] == "true" {
		fmt.Println("This is a smartphone")
	}
	fmt.Printf("This device form_factor is %s\n", JSONDeviceData.Capabilities["form_factor"])

	// Get all the device manufacturers, and print the first twenty
	deviceMakes, err := ClientConn.GetAllDeviceMakes()
	if err != nil {
		log.Fatalf("Error getting device data %s\n", err.Error())
	}

	var limit = 20
	fmt.Printf("Print the first %d Brand of %d\n", limit, len(deviceMakes))

	// Sort the device manufacturer names
	sort.Strings(deviceMakes)

	for _, deviceMake := range deviceMakes[0:limit] {
		fmt.Printf(" - %s\n", deviceMake)
	}

	// Now call the WM server to get all device model and marketing names produced by Apple
	fmt.Println("Print all Model for the Apple Brand")
	modelMktNames, err := ClientConn.GetAllDevicesForMake("Apple")

	if err != nil {
		log.Fatalf("Error getting device data %s\n", err.Error())
	}

	// Sort modelMktNames objects by their model name
	sort.Sort(ByModelName(modelMktNames))

	for _, modelMktName := range modelMktNames {
		fmt.Printf(" - %s %s\n", modelMktName.ModelName, modelMktName.MarketingName)
	}

	// Now call the WM server to get all operative system names
	fmt.Println("Print the list of OSes")
	oses, err := ClientConn.GetAllOSes()

	if err != nil {
		log.Fatalf("Error getting device data %s\n", err.Error())
	}

	// Sort and print all OS names
	sort.Strings(oses)

	for _, os := range oses {
		fmt.Printf(" - %s\n", os)
	}

	// Let's call the WM server to get all version of the Android OS
	fmt.Println("Print all versions for the Android OS")
	versions, err := ClientConn.GetAllVersionsForOS("Android")

	if err != nil {
		log.Fatalf("Error getting device data %s\n", err.Error())
	}

	// Sort all Android version numbers and print them.
	sort.Strings(versions)

	for _, version := range versions {
		fmt.Printf(" - %s\n", version)
	}
}
```


# wmclient APIs

See [wmclient.md](wmclient.md)
