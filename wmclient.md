# wmclient
--
    import "scientiamobile/wmclient"


## Usage

#### func  GetAPIVersion

```go
func GetAPIVersion() string
```
GetAPIVersion returns the version number of WM Client API

#### type JSONDeviceData

```go
type JSONDeviceData struct {
	APIVersion   string            `json:"apiVersion"`
	Capabilities map[string]string `json:"capabilities"`
	Error        string            `json:"error, omitempty"`
	Mtime        int64             `json:"mtime"` 
	Ltime        string            `json:"ltime"` // time of last wurfl.xml file load
}
```

JSONDeviceData models a WURFL device data in JSON string only format

#### type JSONDeviceDataTyped

```go
type JSONDeviceDataTyped struct {
	APIVersion   string                 `json:"apiVersion"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Error        string                 `json:"error, omitempty"`
	Mtime        int64                  `json:"mtime"`
	Ltime        string                 `json:"ltime"`
}
```

JSONDeviceDataTyped models a WURFL device data in JSON typed format

#### type JSONDeviceOsVersions

```go
type JSONDeviceOsVersions struct {
	OsName    string `json:"device_os"`
	OsVersion string `json:"device_os_version"`
}
```

JSONDeviceOsVersions holds device os name and version

#### type JSONInfoData

```go
type JSONInfoData struct {
	WurflAPIVersion  string   `json:"wurfl_api_version"`
	WurflInfo        string   `json:"wurfl_info"`
	WmVersion        string   `json:"wm_version"`
	ImportantHeaders []string `json:"important_headers"`
	StaticCaps       []string `json:"static_caps"`
	VirtualCaps      []string `json:"virtual_caps"`
	Ltime            string   `json:"ltime"`
}
```

JSONInfoData - server and API informations

#### type JSONMakeModel

```go
type JSONMakeModel struct {
	BrandName     string `json:"brand_name"`
	ModelName     string `json:"model_name"`
	MarketingName string `json:"marketing_name,omitempty"`
}
```

JSONMakeModel models simple device "identity" data in JSON format

#### type JSONModelMktName

```go
type JSONModelMktName struct {
	ModelName     string `json:"model_name"`
	MarketingName string `json:"marketing_name,omitempty"`
}
```

JSONModelMktName holds model_name and marketing_name

#### type Request

```go
type Request struct {
	LookupHeaders  map[string]string `json:"lookup_headers"`
	RequestedCaps  []string          `json:"requested_caps"`
	RequestedVCaps []string          `json:"requested_vcaps, omitempty"`
	WurflID        string            `json:"wurfl_id, omitempty"`
	TacCode        string            `json:"tac_code, omitempty"`
}
```

Request - data object that is sent to the WM server in POST requests

#### type WmClient

```go
type WmClient struct {
	StaticCaps  []string
	VirtualCaps []string

	ImportantHeaders []string
}
```

WmClient holds http connection data to WM server and the list of static and
virtual capabilities it must return in response.

#### func  Create

```go
func Create(Scheme string, Host string, Port string, BaseURI string) (*WmClient, error)
```
Create : creates object, checks for server visibility

#### func (*WmClient) DestroyConnection

```go
func (c *WmClient) DestroyConnection()
```
DestroyConnection - Disposes resources used in connection to server and clears
cache and other shared data structures

#### func (*WmClient) GetActualCacheSizes

```go
func (c *WmClient) GetActualCacheSizes() (int, int)
```
GetActualCacheSizes return the values of cache size. The first value being the
device-id based cache, the second value being the size of the headers-based one

#### func (*WmClient) GetAllDeviceMakes

```go
func (c *WmClient) GetAllDeviceMakes() ([]string, error)
```
GetAllDeviceMakes returns a slice of all devices brand_name capabilities in WM
server

#### func (*WmClient) GetAllDevicesForMake

```go
func (c *WmClient) GetAllDevicesForMake(brandName string) ([]JSONModelMktName, error)
```
GetAllDevicesForMake returns a slice of an aggregate containing model_names and
marketing_names for the given brand_name

#### func (*WmClient) GetAllMakeModel

```go
func (c *WmClient) GetAllMakeModel() ([]JSONMakeModel, error)
```
GetAllMakeModel returns identity data for all devices in WM server Deprecated
since 1.2.0.0 in favour of GetAllDeviceMakes

#### func (*WmClient) GetAllOSes

```go
func (c *WmClient) GetAllOSes() ([]string, error)
```
GetAllOSes returns a slice of all devices device_os capabilities in WM server

#### func (*WmClient) GetAllVersionsForOS

```go
func (c *WmClient) GetAllVersionsForOS(osName string) ([]string, error)
```
GetAllVersionsForOS returns a slice of an aggregate containing device_os_version
for the given os_name

#### func (*WmClient) GetInfo

```go
func (c *WmClient) GetInfo() (*JSONInfoData, error)
```
GetInfo - Returns information about the running WM server and API

#### func (*WmClient) HasStaticCapability

```go
func (c *WmClient) HasStaticCapability(CapName string) bool
```
HasStaticCapability - returns true if the given CapName exist in this client'
static capability set, false otherwise

#### func (*WmClient) HasVirtualCapability

```go
func (c *WmClient) HasVirtualCapability(CapName string) bool
```
HasVirtualCapability - returns true if the given CapName exist in this client'
virtual capability set, false otherwise

#### func (*WmClient) LookupDeviceID

```go
func (c *WmClient) LookupDeviceID(deviceID string) (*JSONDeviceData, error)
```
LookupDeviceID - Searches WURFL device data using its wurfl_id value

#### func (*WmClient) LookupRequest

```go
func (c *WmClient) LookupRequest(request http.Request) (*JSONDeviceData, error)
```
LookupRequest - detects a device and returns its data in JSON format

#### func (*WmClient) LookupUserAgent

```go
func (c *WmClient) LookupUserAgent(userAgent string) (*JSONDeviceData, error)
```
LookupUserAgent - Searches WURFL device data using the given user-agent for
detection

#### func (*WmClient) SetCacheSize

```go
func (c *WmClient) SetCacheSize(uaMaxEntries int)
```
SetCacheSize : set UA cache size

#### func (*WmClient) SetHTTPTimeout

```go
func (c *WmClient) SetHTTPTimeout(connection int, transfer int)
```
SetHTTPTimeout sets the connection and transfer timeouts for this client in
seconds. This function should be called before performing any connection to WM
server

#### func (*WmClient) SetRequestedCapabilities

```go
func (c *WmClient) SetRequestedCapabilities(CapsList []string)
```
SetRequestedCapabilities - set the given capability names to the set they belong

#### func (*WmClient) SetRequestedStaticCapabilities

```go
func (c *WmClient) SetRequestedStaticCapabilities(CapsList []string)
```
SetRequestedStaticCapabilities - set list of standard static capabilities to
return

#### func (*WmClient) SetRequestedVirtualCapabilities

```go
func (c *WmClient) SetRequestedVirtualCapabilities(CapsList []string)
```
SetRequestedVirtualCapabilities - set list of virtual capabilities to return

#### func (*WmClient) SetupCache

```go
func (c *WmClient) SetupCache(deviceMaxEntries int, uaMaxEntries int)
```
SetupCache Deprecated: Use SetCacheSize()
