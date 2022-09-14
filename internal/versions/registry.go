package versions

import (
	"fmt"
	"net/url"
	"sort"

	goversion "github.com/hashicorp/go-version"
	"github.com/ryan-jan/tfvc/internal/registry"
)

func Registry(client registry.Client, hostname, namespace, name, provider string) ([]*goversion.Version, error) {
	baseURL, err := client.DiscoverModules(hostname)
	if err != nil {
		return nil, fmt.Errorf("discover registry at %q: %w", hostname, err)
	}
	baseURLStruct, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse module registry url %q: %w", baseURL, err)
	}
	if baseURLStruct.Scheme == "" {
		baseURLStruct.Scheme = "https"
	}
	if baseURLStruct.Host == "" {
		baseURLStruct.Host = hostname
	}
	baseURL = baseURLStruct.String()
	versions, err := client.ListModuleVersions(baseURL, namespace, name, provider)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	out := make([]*goversion.Version, len(versions))
	for i, versionString := range versions {
		version, err := goversion.NewVersion(versionString)
		if err != nil {
			version = nil
		}
		out[i] = version
	}
	sort.Sort(goversion.Collection(out))
	return out, nil
}

func RegistryProvider(client registry.Client, namespace, name string) ([]*goversion.Version, error) {
	baseURL, err := client.DiscoverProviders(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("discover registry for providers: %s", err)
	}
	versions, err := client.ListProviderVersions(baseURL)
	if err != nil {
		return nil, fmt.Errorf("list provider versions: %w", err)
	}
	out := make([]*goversion.Version, len(versions))
	for i, versionString := range versions {
		version, err := goversion.NewVersion(versionString)
		if err != nil {
			version = nil
		}
		out[i] = version
	}
	sort.Sort(goversion.Collection(out))
	return out, nil
}
