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
package wmclient

import (
	"bufio"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"flag"
	"github.com/stretchr/testify/require"
)

// NOTE: This tests assumes that you have a running Wurfl Microservice server. If not, most of the tests will fail

// If no value is provided to this flag, default ua.txt file is searched and, if not found, test is skipped.
// File must be a simple list of user-agents
var uafile = flag.String("f", "ua.txt", "name of ua real traffic ua file")

func TestCreateWithServerDown(t *testing.T) {
	_, err := Create("http", "localhost", "18080", "")
	require.NotNil(t, err)
}

func TestCreateWithEmptyServerValues(t *testing.T) {
	cl, err := Create("http", "", "", "")
	require.NotNil(t, err)
	require.Nil(t, cl)
	require.True(t, strings.Contains(err.Error(), "no Host"))
}

func TestCreateWithEmptySchemeValue(t *testing.T) {
	host, port := getHostPortFromEnv()
	cl, err := Create("", host, port, "")
	require.Nil(t, err)
	// client is created with default http scheme, thus no error
	require.NotNil(t, cl)
}

func TestCreateWithWrongSchemeValue(t *testing.T) {
	cl, err := Create("smtp", "localhost", "8080", "")
	require.NotNil(t, err)
	require.Nil(t, cl)
	require.True(t, strings.Contains(err.Error(), "unsupported protocol scheme"))
}

func TestCreateWithEmptyHost(t *testing.T) {
	_, port := getHostPortFromEnv()
	cl, err := Create("http", "", port, "")
	// This works because golang internal http client class assumes localhost when an empty host is provided
	require.NotNil(t, cl)
	require.Nil(t, err)
}

func TestCreateOk(t *testing.T) {
	client := createTestClient(t)
	require.True(t, len(client.ImportantHeaders) > 0)
	require.True(t, len(client.VirtualCaps) > 0)
	require.True(t, len(client.StaticCaps) > 0)
	client.DestroyConnection()
}

func TestCreateHttpsClientOk(t *testing.T) {
	t.SkipNow()
	client, err := Create("https", "wmserver-test", "8443", "")
	require.Nil(t, err)
	require.NotNil(t, client)
	require.True(t, len(client.ImportantHeaders) > 0)
	require.True(t, len(client.VirtualCaps) > 0)
	require.True(t, len(client.StaticCaps) > 0)
	client.DestroyConnection()
}

func TestHasStaticCapability(t *testing.T) {
	client := createTestClient(t)
	require.True(t, client.HasStaticCapability("brand_name"))
	require.True(t, client.HasStaticCapability("model_name"))
	require.True(t, client.HasStaticCapability("is_smarttv"))
	// this is a virtual capability, so it shouldn't be returned
	require.False(t, client.HasStaticCapability("is_app"))
	client.DestroyConnection()
}

func TestHasVirtualCapability(t *testing.T) {
	client := createTestClient(t)
	require.True(t, client.HasVirtualCapability("is_app"))
	require.True(t, client.HasVirtualCapability("is_smartphone"))
	require.True(t, client.HasVirtualCapability("form_factor"))
	require.True(t, client.HasVirtualCapability("is_app_webview"))
	// this is a static capability, so it shouldn't be returned
	require.False(t, client.HasVirtualCapability("brand_name"))
	require.False(t, client.HasVirtualCapability("is_wireless_device"))
}

func TestGetInfo(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.GetInfo()
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	require.NotEmpty(t, jsonData.WmVersion)
	require.True(t, len(jsonData.ImportantHeaders) > 0)
	require.True(t, len(jsonData.StaticCaps) > 0)
	require.True(t, len(jsonData.VirtualCaps) > 0)
	client.DestroyConnection()
}

func TestSingleLookupUserAgent(t *testing.T) {
	client := createTestCachedClient(t)
	internalTestLookupUserAgent(t, client)
	client.DestroyConnection()
}

func TestMultipleLookupUserAgent(t *testing.T) {
	client := createTestClient(t)
	client.SetCacheSize(1000)
	for i := 0; i < 50; i++ {
		internalTestLookupUserAgent(t, client)
	}
	client.DestroyConnection()

}

func TestSetRequestedCapabilities(t *testing.T) {
	client := createTestClient(t)
	// In both static and vcap lists we add 1 correct name, 1 non existent name and 1 name that belongs to a different set
	client.SetRequestedStaticCapabilities([]string{"brand_name", "invalid_name1", "is_ios"})
	client.SetRequestedVirtualCapabilities([]string{"is_ios", "invalid_name2", "brand_name"})

	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Version/10.0 Mobile/14D27 Safari/602.1"
	device, err := client.LookupUserAgent(context.Background(), ua)
	require.Nil(t, err)
	require.NotNil(t, device)
	// 1 cap, 1 vcap + wurfl_id
	require.Equal(t, 3, len(device.Capabilities))
	_, ok := device.Capabilities["invalid_name1"]
	require.False(t, ok) // this cap has been discarded because it does not exist
	client.SetRequestedStaticCapabilities(nil)
	device, _ = client.LookupUserAgent(context.Background(), ua)
	require.Equal(t, 2, len(device.Capabilities))
	client.SetRequestedVirtualCapabilities(nil)
	device, _ = client.LookupUserAgent(context.Background(), ua)
	// resetting all required caps arrays, ALL available caps are returned
	require.True(t, len(device.Capabilities) > 0, "len(device.Capabilities) > 0 failed")

	// Set and reset caps, All caps should be returned
	client.SetRequestedStaticCapabilities([]string{"brand_name", "invalid_name1", "is_ios"})
	client.SetRequestedVirtualCapabilities([]string{"is_ios", "invalid_name2", "brand_name"})
	client.SetRequestedCapabilities(nil)
	device, _ = client.LookupUserAgent(context.Background(), ua)

	require.True(t, len(device.Capabilities) > 0, "len(device.Capabilities) > 0 failed")

}

func TestResetCacheOnRequestedCapsChange(t *testing.T) {
	// Checks that cache is cleared whenever a reset of the requested static and/or virtual capabilities occur
	client := createTestClient(t)
	client.SetCacheSize(1000)
	reqCaps := []string{"brand_name", "is_wireless_device", "is_app"}
	client.SetRequestedCapabilities(reqCaps)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Version/10.0 Mobile/14D27 Safari/602.1"
	d, derr := client.LookupUserAgent(context.Background(), ua)
	require.NotNil(t, d)
	require.Nil(t, derr)
	dc, uac := client.GetActualCacheSizes()
	require.Equal(t, 0, dc)
	require.Equal(t, 1, uac)

	client.SetRequestedCapabilities(reqCaps)

	dc, uac = client.GetActualCacheSizes()
	require.Equal(t, 0, dc)
	require.Equal(t, 0, uac)

	d, _ = client.LookupUserAgent(context.Background(), ua)
	dc, uac = client.GetActualCacheSizes()
	require.Equal(t, 1, uac)
	reqCaps = []string{"brand_name", "is_wireless_device"}
	reqVCaps := []string{"is_app", "is_ios"}
	client.SetRequestedStaticCapabilities(reqCaps)
	client.SetRequestedVirtualCapabilities(reqVCaps)

	dc, uac = client.GetActualCacheSizes()
	require.Equal(t, 0, dc)
	require.Equal(t, 0, uac)

	client.DestroyConnection()

}

func internalTestLookupUserAgent(t *testing.T, client *WmClient) {

	ua := "Mozilla/5.0 (Linux; Android 7.0; SAMSUNG SM-G950F Build/NRD90M) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/5.2 Chrome/51.0.2704.106 Mobile Safari/537.36"
	jsonData, _ := client.LookupUserAgent(context.Background(), ua)
	require.NotNil(t, jsonData)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.True(t, len(did) > 0) // we just check that there are capabilities
	require.Equal(t, did["model_name"], "SM-G950F")
	require.Equal(t, "false", did["is_robot"])
	require.Equal(t, "false", did["is_full_desktop"])

}

func TestSingleLookupDeviceId(t *testing.T) {
	client := createTestCachedClient(t)
	internalTestLookupDeviceID(t, client)
	client.DestroyConnection()
}

func TestSingleLookupDeviceIdWithCacheExpiration(t *testing.T) {
	client := createTestCachedClient(t)

	d1, err := client.LookupDeviceID(context.Background(), "nokia_generic_series40")
	require.Nil(t, err)
	require.NotNil(t, d1)

	d2, err2 := client.LookupUserAgent(context.Background(), "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Version/10.0 Mobile/14D27 Safari/602.1")
	require.Nil(t, err2)
	require.NotNil(t, d2)

	dc, uac := client.GetActualCacheSizes()
	require.Equal(t, 1, dc)
	require.Equal(t, 1, uac)
	// In this first call, cache should not be cleaned, because ltime is the same as client's last load time
	client.clearCachesIfNeeded(d1.Ltime)
	require.Equal(t, 1, dc)
	require.Equal(t, 1, uac)

	// Now, date changes, so caches must be cleared
	client.clearCachesIfNeeded("2199-12-31")
	dc, uac = client.GetActualCacheSizes()
	require.Equal(t, 0, dc)
	require.Equal(t, 0, uac)

	// Load a device again
	d1, _ = client.LookupDeviceID(context.Background(), "nokia_generic_series40")
	d2, _ = client.LookupUserAgent(context.Background(), "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Version/10.0 Mobile/14D27 Safari/602.1")

	// caches are filled again
	dc, uac = client.GetActualCacheSizes()
	require.Equal(t, 1, dc)
	require.Equal(t, 1, uac)

	client.DestroyConnection()
}

func getHostPortFromEnv() (host string, port string) {
	var ok bool
	if host, ok = os.LookupEnv("TEST_WM_HOST"); !ok {
		host = "localhost"
	}
	if port, ok = os.LookupEnv("TEST_WM_PORT"); !ok {
		port = "8080"
	}
	return host, port
}

func createTestClient(t *testing.T) *WmClient {
	host, port := getHostPortFromEnv()
	client, cerr := Create("http", host, port, "")
	require.Nil(t, cerr)
	require.NotNil(t, client)
	return client
}

func createTestCachedClient(t *testing.T) *WmClient {
	client := createTestClient(t)
	client.SetCacheSize(1000)
	return client
}

func TestMultiThreadCachedClient(t *testing.T) {

	if _, err := os.Stat(*uafile); os.IsNotExist(err) {
		d, _ := os.Getwd()
		t.Skip(fmt.Sprintf("The specified UA file %s does not exist. Current directory is %s. Check if the 'f' flag is set correctly: %s", *uafile, d, err.Error()))
	}

	// This test can be very time consuming, to skip it do not provide the -f option value
	client, cerr := Create("http", "localhost", "8080", "")
	require.Nil(t, cerr)

	fmt.Printf("Ua file %s\n", *uafile)
	client.SetCacheSize(100000)

	done := make(chan bool)

	// generate parallel goroutines to run mtLookup
	for i := 0; i < 4; i++ {
		go mtLookup(t, client, done)
		fmt.Printf("goroutine %d started\n", i)
	}

	// wait for all go routines to finish
	for i := 0; i < 4; i++ {
		<-done
		fmt.Printf("goroutine %d terminated\n", i)
	}

	client.DestroyConnection()
}

func mtLookup(t *testing.T, client *WmClient, done chan bool) {
	var linecount int

	// read UA from file
	file, err := os.Open(*uafile)
	require.Nil(t, err)
	if err != nil {
		fmt.Printf("mtLookup open %s, error %s\n", *uafile, err)
	}

	scanner := bufio.NewScanner(file)

	linecount = 0

	for scanner.Scan() {
		ua := scanner.Text()
		if ua == "" {
			linecount++
			// skip empty lines
			continue
		}

		d, err := client.LookupUserAgent(context.Background(), ua)

		if err != nil {
			fmt.Printf("ClientConn.LookupUserAgent(%s) returned %s\n", ua, err.Error())
			break
		}

		d, err = client.LookupDeviceID(context.Background(), d.Capabilities["wurfl_id"])

		if err != nil {
			fmt.Printf("ClientConn.LookupDeviceID(%s) returned %s\n", ua, err.Error())
			break
		}

		// These calls with no assertions are just used to try to trigger race condition (if any)
		client.GetAllOSes()

		client.GetAllVersionsForOS("Android")

		client.GetInfo()

		client.GetAllDeviceMakes()

		client.GetAllDevicesForMake("Apple")

		client.GetActualCacheSizes()

		linecount++
	}
	file.Close()

	fmt.Printf("Lines read = %d\n", linecount)

	done <- true
}

func TestMultipleLookupDeviceId(t *testing.T) {

	client := createTestCachedClient(t)
	for i := 0; i < 50; i++ {
		internalTestLookupDeviceID(t, client)
	}
	client.DestroyConnection()
}

func internalTestLookupDeviceID(t *testing.T, client *WmClient) {

	jsonData, err := client.LookupDeviceID(context.Background(), "nokia_generic_series40")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.True(t, len(did) > 0)
	require.Equal(t, "true", did["is_mobile"])
	require.Equal(t, "Feature Phone", did["form_factor"])

}

func TestLookupDeviceIdWithSpecificCaps(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_smarttv"}
	reqvCaps := []string{"form_factor"}
	client.SetRequestedStaticCapabilities(reqCaps)
	client.SetRequestedVirtualCapabilities(reqvCaps)
	jsonData, err := client.LookupDeviceID(context.Background(), "generic_opera_mini_version1")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Opera", did["brand_name"])
	require.Equal(t, "false", did["is_smarttv"])
	require.Equal(t, 4, len(did))
	client.DestroyConnection()
}

func TestLookupDeviceIdWithSpecificCapsSingleMethods(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_smarttv", "is_smartphone", "form_factor"}
	client.SetRequestedCapabilities(reqCaps)
	jsonData, err := client.LookupDeviceID(context.Background(), "generic_opera_mini_version1")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Opera", did["brand_name"])
	require.Equal(t, "false", did["is_smarttv"])
	require.Equal(t, 5, len(did))
	client.DestroyConnection()
}

func TestLookupDeviceIdWithWrongSpecificCaps(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_smarttv", "nonexcap"}
	client.SetRequestedStaticCapabilities(reqCaps)
	jsonData, err := client.LookupDeviceID(context.Background(), "generic_opera_mini_version1")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Opera", did["brand_name"])
	require.Equal(t, "false", did["is_smarttv"])
	require.Equal(t, "", did["nonexcap"])
	require.Equal(t, 3, len(did)) // non existent cap now is discarded in SetRequiredStatic/VirtualCaps method
	client.DestroyConnection()
}

func TestLookupDeviceIdWithWrongId(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.LookupDeviceID(context.Background(), "nokia_generic_series40_wrong")
	require.NotNil(t, jsonData)
	require.NotNil(t, err)
	did := jsonData.Capabilities
	require.Nil(t, did)
	require.True(t, len(jsonData.APIVersion) > 0)
	require.True(t, len(jsonData.Error) == 0)
	require.NotNil(t, err)
	require.Equal(t, 0, len(did))
	require.True(t, jsonData.Mtime > 0)
	client.DestroyConnection()
}

func TestLookupDeviceIdWithEmptyId(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.LookupDeviceID(context.Background(), "")
	require.NotNil(t, jsonData)
	require.NotNil(t, err)
	did := jsonData.Capabilities
	require.Nil(t, did)
	require.True(t, len(jsonData.APIVersion) > 0)
	require.True(t, len(jsonData.Error) == 0)
	require.NotNil(t, err)
	require.Equal(t, 0, len(did))
	require.True(t, jsonData.Mtime > 0)
	client.DestroyConnection()

}

func TestLookupDeviceEmptyUseragent(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.LookupUserAgent(context.Background(), "")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.True(t, len(did) > 0)
	require.Equal(t, "generic", jsonData.Capabilities["wurfl_id"])
	require.True(t, len(jsonData.APIVersion) > 0)
	client.DestroyConnection()
}

func TestLookupDeviceuseragentWithSpecificCaps(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "marketing_name", "is_full_desktop", "model_name"}
	client.SetRequestedCapabilities(reqCaps)
	jsonData, err := client.LookupUserAgent(context.Background(), "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341")
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Nintendo", did["brand_name"])
	require.Equal(t, "Switch", did["model_name"])
	require.Equal(t, "false", did["is_full_desktop"])
	require.Equal(t, 5, len(did))
	client.DestroyConnection()
}

func TestLookupRequestWithSpecificCaps(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_full_desktop", "is_robot", "model_name"}
	client.SetRequestedCapabilities(reqCaps)

	url := "http://vimeo.com/api/v2/brad/info.json"
	request, err := http.NewRequest("GET", url, nil)
	if err == nil {
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Accept-Encoding", "gzip, deflate")
		request.Header.Add("User-Agent", "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341")
	}

	jsonData, err := client.LookupRequest(*request)

	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Nintendo", did["brand_name"])
	require.Equal(t, "Switch", did["model_name"])
	require.Equal(t, "false", did["is_robot"])
	require.Equal(t, "false", did["is_full_desktop"])
	require.Equal(t, 5, len(did))

	reqCaps = append(reqCaps, "is_smarttv")
	client.SetRequestedCapabilities(reqCaps)
	jsonData, err = client.LookupRequest(*request)

	require.Nil(t, err)
	did = jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, 6, len(did))

	client.DestroyConnection()

}

// WPC-170
func TestLookupHeaderMixedCase(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_wireless_device", "pointing_method", "model_name"}
	client.SetRequestedCapabilities(reqCaps)

	url := "http://mysite.com/api/v2/foo/info.json"
	request, err := http.NewRequest("GET", url, nil)
	if err == nil {
		request.Header.Add("Content-type", "application/json")
		request.Header.Add("X-uCBrowser-device-UA", "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B")
		request.Header.Add("user-agent", "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341")
	}

	require.Nil(t, err)

	jsonData, derr := client.LookupRequest(*request)
	require.NotNil(t, jsonData)
	require.Nil(t, derr)
}

func TestLookupRequestWithCache(t *testing.T) {
	client := createTestClient(t)
	client.SetCacheSize(100)

	url := "http://mysite.com/api/v2/foo/info.json"

	request, err := http.NewRequest("GET", url, nil)
	if err == nil {
		request.Header.Add("X-Requested-With", "json_client")
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Accept-Encoding", "gzip, deflate")
		request.Header.Add("X-UCBrowser-Device-UA", "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B")
		request.Header.Add("User-Agent", "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341")
	}

	var jsonData *JSONDeviceData
	var derr error

	for i := 0; i < 50; i++ {
		jsonData, derr = client.LookupRequest(*request)

		require.NotNil(t, jsonData)
		require.Nil(t, derr)
		did := jsonData.Capabilities
		require.NotNil(t, did)
		require.Equal(t, "Samsung", did["brand_name"])
		require.Equal(t, "GT-S5253", did["model_name"])
		require.Equal(t, "false", did["is_robot"])
		require.True(t, len(did) > 0)
		dCacheSize, uaCacheSize := client.GetActualCacheSizes()
		require.Equal(t, 0, dCacheSize)
		require.Equal(t, 1, uaCacheSize)

	}
	client.DestroyConnection()
}

func TestLookupHeadersOk(t *testing.T) {
	client := createTestClient(t)

	// Let's create test headers
	var headers = make(map[string]string, 4)
	headers["X-Requested-With"] = "json_client"
	headers["Content-Type"] = "application/json"
	headers["Accept-Encoding"] = "gzip, deflate"
	headers["X-UCBrowser-Device-UA"] = "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B"
	headers["User-Agent"] = "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341"

	var jsonData *JSONDeviceData
	var derr error
	jsonData, derr = client.LookupHeaders(context.Background(), headers)

	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Samsung", did["brand_name"])
	require.Equal(t, "GT-S5253", did["model_name"])
	require.Equal(t, "false", did["is_robot"])
	require.True(t, len(did) > 0)

	client.DestroyConnection()
}

func TestLookupHeadersWithMixedCase(t *testing.T) {
	client := createTestClient(t)

	// Let's create test headers
	var headers = make(map[string]string, 4)
	headers["X-RequesTed-With"] = "json_client"
	headers["Content-TYpe"] = "application/json"
	headers["Accept-EnCoding"] = "gzip, deflate"
	headers["X-UCBrowsEr-DeVice-UA"] = "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B"
	headers["UseR-AgEnt"] = "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341"

	var jsonData *JSONDeviceData
	var derr error
	jsonData, derr = client.LookupHeaders(context.Background(), headers)

	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Samsung", did["brand_name"])
	require.Equal(t, "GT-S5253", did["model_name"])
	require.Equal(t, "false", did["is_robot"])
	require.True(t, len(did) > 0)

	client.DestroyConnection()
}

func TestLookupHeadersWithMixedCase_CachedClient(t *testing.T) {
	client := createTestCachedClient(t)

	// Let's create test headers
	var headers = make(map[string]string, 4)
	headers["X-RequesTed-With"] = "json_client"
	headers["Content-TYpe"] = "application/json"
	headers["Accept-EnCoding"] = "gzip, deflate"
	headers["X-UCBrowsEr-DeVice-UA"] = "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B"
	headers["UseR-AgEnt"] = "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341"

	var jsonData *JSONDeviceData
	var derr error
	jsonData, derr = client.LookupHeaders(context.Background(), headers)

	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Samsung", did["brand_name"])
	require.Equal(t, "GT-S5253", did["model_name"])
	require.Equal(t, "false", did["is_robot"])
	require.True(t, len(did) > 0)
	_, c := client.GetActualCacheSizes()
	require.Equal(t, 1, c)

	// let's retry mixing header case in a different way (now we'll have a cache hit)
	headers = make(map[string]string, 4)
	headers["X-ReqUesTed-With"] = "json_client"
	headers["ConteNt-TYpe"] = "application/json"
	headers["AccepT-EnCoding"] = "gzip, deflate"
	headers["X-UCBrOwsEr-DeVice-UA"] = "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B"
	headers["UseR-AgeNT"] = "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341"

	jsonData, derr = client.LookupHeaders(context.Background(), headers)
	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	did = jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Samsung", did["brand_name"])
	// Same value as before, cache has been hit
	_, c = client.GetActualCacheSizes()
	require.Equal(t, 1, c)

	client.DestroyConnection()
}

func TestLookupHeadersWithNilOrEmptyMap(t *testing.T) {
	client := createTestClient(t)

	var jsonData *JSONDeviceData
	var derr error
	jsonData, derr = client.LookupHeaders(context.Background(), nil)

	// Passing a nil map should result in the creation of an empty request header map, thus in a "generic" device detection...
	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	require.Equal(t, "generic", jsonData.Capabilities["wurfl_id"])

	var headers = make(map[string]string, 0)
	jsonData, derr = client.LookupHeaders(context.Background(), headers)

	// ... the same result occurs if we pass an empty header map
	require.NotNil(t, jsonData)
	require.Nil(t, derr)
	require.Equal(t, "generic", jsonData.Capabilities["wurfl_id"])

	client.DestroyConnection()
}

func TestLookupHeadersWithSpecificCaps(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "marketing_name", "is_full_desktop", "model_name"}
	client.SetRequestedCapabilities(reqCaps)

	var headers = make(map[string]string, 4)
	headers["X-Requested-With"] = "json_client"
	headers["Content-Type"] = "application/json"
	headers["Accept-Encoding"] = "gzip, deflate"
	headers["X-UCBrowser-Device-UA"] = "Mozilla/5.0 (SAMSUNG; SAMSUNG-GT-S5253/S5253DDJI7; U; Bada/1.0; en-us) AppleWebKit/533.1 (KHTML, like Gecko) Dolfin/2.0 Mobile WQVGA SMM-MMS/1.2.0 OPN-B"
	headers["User-Agent"] = "Mozilla/5.0 (Nintendo Switch; WebApplet) AppleWebKit/601.6 (KHTML, like Gecko) NF/4.0.0.5.9 NintendoBrowser/5.1.0.13341"

	jsonData, err := client.LookupHeaders(context.Background(), headers)
	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.Equal(t, "Samsung", did["brand_name"])
	require.Equal(t, "GT-S5253", did["model_name"])
	require.Equal(t, "false", did["is_full_desktop"])
	require.Equal(t, 5, len(did))
	client.DestroyConnection()

}

func TestLookupRequestWithNoHeaders(t *testing.T) {
	client := createTestClient(t)
	reqCaps := []string{"brand_name", "is_wireless_device", "pointing_method", "model_name"}
	client.SetRequestedCapabilities(reqCaps)

	url := "http://vimeo.com/api/v2/brad/info.json"
	request, err := http.NewRequest("GET", url, nil)

	jsonData, err := client.LookupRequest(*request)

	require.NotNil(t, jsonData)
	require.Nil(t, err)
	did := jsonData.Capabilities
	require.NotNil(t, did)
	require.True(t, len(jsonData.Error) == 0)
	require.Equal(t, "generic", jsonData.Capabilities["wurfl_id"])
	client.DestroyConnection()
}

// Just to check that device error message has been created properly
func TestNewJSONDeviceDataWithError(t *testing.T) {
	jsonDataErr := JSONDeviceData{Error: "Error message", Mtime: time.Now().Unix(), APIVersion: "WURFL Microservice Client " + GetAPIVersion()}
	require.NotNil(t, jsonDataErr)
	require.Equal(t, "Error message", jsonDataErr.Error)
	// tests time has been set
	require.True(t, jsonDataErr.Mtime > 0)
	require.Equal(t, "WURFL Microservice Client "+GetAPIVersion(), jsonDataErr.APIVersion)

}

func TestDestroyConnection(t *testing.T) {
	client := createTestClient(t)
	res, err := client.GetInfo()
	require.NotNil(t, res)
	require.Nil(t, err)

	client.DestroyConnection()

	// This closure function tests that the client.GetInfo called after it panics
	defer func() {
		if r := recover(); r != nil {
			_, ok := r.(error)
			if !ok {
				require.Fail(t, "Test for DestroyConnection should have called panic()")
			}
		}

	}()
	client.GetInfo()
}

func TestGetAllDeviceMakes(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.GetInfo()
	if strings.Compare(jsonData.WmVersion, "1.2.0.0") == -1 {
		t.Skip("Endpoint available since 1.2.0.0")
	}
	mkMds, err := client.GetAllDeviceMakes()
	require.Nil(t, err)
	require.NotNil(t, mkMds)
	require.True(t, len(mkMds) > 2000)
	client.DestroyConnection()
}

func TestGetAllDevicesForMake(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.GetInfo()
	if strings.Compare(jsonData.WmVersion, "1.2.0.0") == -1 {
		t.Skip("Endpoint available since 1.2.0.0")
	}
	modelMktNames, err := client.GetAllDevicesForMake("Nokia")
	require.Nil(t, err)
	require.NotNil(t, modelMktNames)
	require.True(t, len(modelMktNames) > 700)
	require.NotNil(t, modelMktNames[0].ModelName)
	require.NotNil(t, modelMktNames[0].MarketingName)
	client.DestroyConnection()
}

func TestGetAllOses(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.GetInfo()
	if strings.Compare(jsonData.WmVersion, "1.2.0.0") == -1 {
		t.Skip("Endpoint available since 1.2.0.0")
	}
	deviceOses, err := client.GetAllOSes()
	require.Nil(t, err)
	require.NotNil(t, deviceOses)
	require.True(t, len(deviceOses) >= 30)
	client.DestroyConnection()
}

func TestGetAllVersionsForOS(t *testing.T) {
	client := createTestClient(t)
	jsonData, err := client.GetInfo()
	if strings.Compare(jsonData.WmVersion, "1.2.0.0") == -1 {
		t.Skip("Endpoint available since 1.2.0.0")
	}
	osVersions, err := client.GetAllVersionsForOS("Android")
	require.Nil(t, err)
	require.NotNil(t, osVersions)
	require.True(t, len(osVersions) > 30)
	for _, v := range osVersions {
		require.NotNil(t, v)
		// check that no empty version ended up in the output array
		require.True(t, v != "")
	}

	osVersions, err = client.GetAllVersionsForOS("iOS")
	require.Nil(t, err)
	require.NotNil(t, osVersions)
	require.True(t, len(osVersions) > 60)
	for _, v := range osVersions {
		require.NotNil(t, v)
		// check that no empty version ended up in the output array
		require.True(t, v != "")
	}

	client.DestroyConnection()
}

func TestLookupMatchingCacheWithAdditionalHeaders(t *testing.T) {
	client := createTestCachedClient(t)
	request, err := http.NewRequest("GET", "scientiamobile.com", nil)
	if err == nil {
		request.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 4.4.4; SmartTV Build/KTU84P), AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.132, _STB_C001_2017/0.9 (NETRANGEMMH, ExpressLuck, Wired)")
		request.Header.Add("x-requested-with", "org.mozilla.tv.firefox")
		d, err := client.LookupRequest(*request)
		require.Nil(t, err)
		require.NotNil(t, d)
		require.Equal(t, "generic_fire_os_ver4_0_tv", d.Capabilities["wurfl_id"])
	}

	// Create the same request, but without x-requested-with
	request, err = http.NewRequest("GET", "scientiamobile.com", nil)
	if err == nil {
		request.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 4.4.4; SmartTV Build/KTU84P), AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.132, _STB_C001_2017/0.9 (NETRANGEMMH, ExpressLuck, Wired)")
		d, err := client.LookupRequest(*request)
		require.Nil(t, err)
		require.NotNil(t, d)
		require.Equal(t, "generic_smarttv_browser", d.Capabilities["wurfl_id"])
	}

	// And again with the old one
	request, err = http.NewRequest("GET", "scientiamobile.com", nil)
	if err == nil {
		request.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 4.4.4; SmartTV Build/KTU84P), AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.132, _STB_C001_2017/0.9 (NETRANGEMMH, ExpressLuck, Wired)")
		request.Header.Add("x-requested-with", "org.mozilla.tv.firefox")
		d, err := client.LookupRequest(*request)
		require.Nil(t, err)
		require.NotNil(t, d)
		require.Equal(t, "generic_fire_os_ver4_0_tv", d.Capabilities["wurfl_id"])
	}
}

func TestMd5KeyCreation(t *testing.T) {
	m := make(map[string]string)
	m["User-Agent"] = "Mozilla/5.0 (Linux; Android 4.4.4; SmartTV Build/KTU84P), AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.132, _STB_C001_2017/0.9"
	m["x-requested-with"] = "org.mozilla.firefox.tv"
	client := createTestCachedClient(t)
	md5k := client.getUserAgentCacheKey(m)
	require.NotNil(t, md5k)
	require.Equal(t, 32, len(md5k))
}

func TestCacheUsage(t *testing.T) {
	client := createTestClient(t)
	ua := "Mozilla/5.0 (Linux; Android 4.4.4; SmartTV Build/KTU84P), AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.132, _STB_C001_2017/0.9 (NETRANGEMMH, ExpressLuck, Wired)"

	// perform 10K detections without cache
	lookupCount := 10000
	start := time.Now().UnixNano()
	for i := 0; i < lookupCount; i++ {
		client.LookupUserAgent(context.Background(), ua)
	}
	totalDetectionTime := time.Now().UnixNano() - start
	avgDetectionTime := float64(totalDetectionTime) / float64(lookupCount)

	// Add cache
	client.SetCacheSize(10000)
	// Fill cache
	for i := 0; i < lookupCount; i++ {
		client.LookupUserAgent(context.Background(), ua)
	}

	// Measure cache
	start = time.Now().UnixNano()
	for i := 0; i < lookupCount; i++ {
		client.LookupUserAgent(context.Background(), ua)
	}
	totalCacheTime := time.Now().UnixNano() - start
	avgCacheTime := float64(totalCacheTime) / float64(lookupCount)
	// cache time must be at least an order of magnitude faster than non cached detection
	assert.True(t, avgDetectionTime > avgCacheTime*10)

}
