package incus

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/lxc/incus/v6/shared/api"
)

func (r *ProtocolIncus) GetInstanceDebugMemory(name string, filePath string, format string) error {
	path, v, err := r.instanceTypeToPath(api.InstanceTypeVM)
	if err != nil {
		return err
	}

	v.Set("path", filePath)
	v.Set("format", format)

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/1.0%s/%s/debug/memory?%s", r.httpBaseURL.String(), path, url.PathEscape(name), v.Encode())

	url, err = r.setQueryAttributes(url)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Send the request
	resp, err := r.DoHTTP(req)
	if err != nil {
		return err
	}

	// Check the return value for a cleaner error
	if resp.StatusCode != http.StatusOK {
		_, _, err := incusParseResponse(resp)
		if err != nil {
			return err
		}
	}

	return nil
}
