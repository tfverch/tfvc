package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Client struct {
	HTTP *http.Client
}

var errNoModuleRegistryHost = errors.New("no module registry host specified")

// DiscoverModules obtains the module index base URL for the given hostname.
// ref.: https://www.terraform.io/docs/registry/api.html#service-discovery
func (c *Client) DiscoverModules(hostname string) (string, error) {
	var response struct {
		ModulesV1 *string `json:"modules.v1"`
	}
	resp, err := c.HTTP.Get(fmt.Sprintf("https://%s/.well-known/terraform.json", hostname))
	if err != nil {
		return "", fmt.Errorf("discover registry: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode registry response: %w", err)
	}
	if response.ModulesV1 == nil {
		return "", fmt.Errorf("%w at %q", errNoModuleRegistryHost, hostname)
	}
	return *response.ModulesV1, nil
}

// ListModuleVersions lists the available module versions for the a specific module.
// ref.: https://www.terraform.io/docs/registry/api.html#list-available-versions-for-a-specific-module
func (c *Client) ListModuleVersions(baseURL, namespace, name, provider string) ([]string, error) {
	url := fmt.Sprintf("%s%s/%s/%s/versions", baseURL, namespace, name, provider)
	var response struct {
		Modules []struct {
			Versions []struct {
				Version string `json:"version"`
			} `json:"versions"`
		} `json:"modules"`
	}
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %w", url, err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode registry response: %w", err)
	}
	var versions []string
	for _, m := range response.Modules {
		for _, v := range m.Versions {
			versions = append(versions, v.Version)
		}
	}
	return versions, nil
}

func (c *Client) DiscoverProviders(namespace, name string) (string, error) {
	url := fmt.Sprintf("https://registry.terraform.io/v2/providers?filter[namespace]=%s&filter[name]=%s", namespace, name)
	var response struct {
		Data []struct {
			Attributes struct {
				FullName string `json:"full-name"`
			} `json:"attributes"`
			Id string `json:"id"`
		} `json:"data"`
	}
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET %q: %w", url, err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode provider registry response: %s", err)
	}
	baseUrl := fmt.Sprintf("https://registry.terraform.io/v2/providers/%s/provider-versions", response.Data[0].Id)
	return baseUrl, nil
}

func (c *Client) ListProviderVersions(url string) ([]string, error) {
	var response struct {
		Data []struct {
			Attributes struct {
				Version string `json:"version"`
			} `json:"attributes"`
			Id string `json:"id"`
		} `json:"data"`
	}
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %w", url, err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode provider registry response: %w", err)
	}
	var versions []string
	for _, d := range response.Data {
		versions = append(versions, d.Attributes.Version)
	}
	return versions, nil
}
