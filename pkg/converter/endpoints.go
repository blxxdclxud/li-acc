package converter

import (
	"fmt"
	"strings"
)

var ApiUrl = "https://api-server.compdf.com/server/v1"
var (
	EndpointToken        = "/oauth/token"
	EndpointUploadFile   = "/file/upload"
	EndpointConvert      = "/execute/start"
	EndpointGetConverted = "/file/fileInfo"
)

// Conversion defines converter's conversion kinds.
// e.g. from DOC to PDF, etc.
type Conversion struct {
	From string
	To   string
}

func (c Conversion) CreateTaskEndpoint() string {
	return fmt.Sprintf("/task/%s/%s", strings.ToLower(c.From), strings.ToLower(c.To))
}
