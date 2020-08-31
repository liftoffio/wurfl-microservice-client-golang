# Changelog

### 2.1.0
- Added `LookupHeaders(map[string]string` method to API

### 2.0.1
- Adapted unit tests to new server behavior of not returning error on empty user-agent or header. They are also adapted to run with different server configurations (ie AWS/Azure Professional edition, Docker server).

### 2.0.0
- Initial version: GetInfo, LookupUseragent, LookupRequest, LookupDeviceId and other functions to set cache and requested capabilitities
