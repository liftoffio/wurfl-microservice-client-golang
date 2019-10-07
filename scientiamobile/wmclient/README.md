# wmclient
--
    import "github.com/wurfl/wurfl-microservice-client-golang/v2/scientiamobile/wmclient"


## Usage

```go
const UserAgent = "User-Agent"
```
UserAgent is the User-Agent header name

#### type JSONDeviceData

```go
type JSONDeviceData struct {
	APIVersion   string            `json:"apiVersion"`
	Capabilities map[string]string `json:"capabilities"`
	Error        string            `json:"error, omitempty"`
	Mtime        int64             `json:"mtime"`
}
```

JSONDeviceData models a WURFL device data in JSON format

#### func  NewJSONDeviceData

```go
func NewJSONDeviceData() JSONDeviceData
```
NewJSONDeviceData - factory method

#### func  NewJSONDeviceDataWithError

```go
func NewJSONDeviceDataWithError(errorMsg string) JSONDeviceData
```
NewJSONDeviceDataWithError - factory method: creates device data object that
contains only an error message

#### type JSONInfoData

```go
type JSONInfoData struct {
	WurflAPIVersion string `json:"wurfl_api_version"`
	WurflInfo       string `json:"wurfl_info"`
	WmVersion      string `json:"wm_version"`
}
```

JSONInfoData - server and API informations

#### type Request

```go
type Request struct {
	LookupHeaders map[string]string `json:"lookup_headers"`
	RequestedCaps []string          `json:"requested_caps"`
	WurflID       string            `json:"wurfl_id, omitempty"`
	TacCode       string            `json:"tac_code, omitempty"`
}
```

Request - data object that is sent to the WM server in POST requests

#### func  NewRequest

```go
func NewRequest() Request
```
NewRequest - creates a new empty Request struct

#### type WmClient

```go
type WmClient struct {
}
```

WmClient holds http connection data to server and the list capability it
must return in response

#### func  Create

```go
func Create(Scheme string, Host string, Port string, BaseURI string)) (*WmClient, error)
```
Create : creates object, checks for server visibility

#### func (*WmClient) DestroyConnection

```go
func (c *WmClient) DestroyConnection()
```
DestroyConnection - Disposes resources used in connection to server, if any

#### func (*WmClient) GetInfo

```go
func (c *WmClient) GetInfo() (*JSONInfoData, error)
```
GetInfo - Returns information about the running WM server and API

#### func (*WmClient) LookupDeviceID

```go
func (c *WmClient) LookupDeviceID(deviceID string) (*JSONDeviceData, error)
```
LookupDeviceID - Searches WURFL device data using its wurfl_id value

#### func (*WmClient) LookupRequest

```go
func (c *WmClient) LookupRequest(request Request) (*JSONDeviceData, error)
```
LookupRequest - detects a device and returns its data in JSON format

#### func (*WmClient) LookupUserAgent

```go
func (c *WmClient) LookupUserAgent(userAgent string) (*JSONDeviceData, error)
```
LookupUserAgent - Searches WURFL device data using the given user-agent for
detection

#### func (*WmClient) SetRequestedCapabilities

```go
func (c *WmClient) SetRequestedCapabilities(CapsList []string)
```
SetRequestedCapabilities - set list of caps to return, both caps and vcap are
the same
