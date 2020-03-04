package driver

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	"github.com/graphql-editor/stucco/pkg/types"
)

// WorkerClient creates new protobuf for communication with workers
type WorkerClient interface {
	New(url string) driver.Driver
}

// ProtobufClient is a worker client using protobuf protocol
type ProtobufClient struct {
	*http.Client
}

func (p ProtobufClient) httpClient() *http.Client {
	if p.Client == nil {
		return http.DefaultClient
	}
	return p.Client
}

// New returns new driver using protobuf protocol
func (p ProtobufClient) New(u string) driver.Driver {
	return &protohttp.Client{
		Client: p.httpClient(),
		URL:    u,
	}
}

// Driver implements stucco driver interface calling protobuf workers over http
// with configurable workers base url
type Driver struct {
	WorkerClient
}

func normalizeFuncName(fn string) string {
	fn = strings.ReplaceAll(fn, ".", "_")
	fn = strings.ReplaceAll(fn, "/", "_")
	fn = strings.ToUpper(fn)
	return fn
}

func (d *Driver) newClient(url string) driver.Driver {
	workerClient := d.WorkerClient
	if workerClient == nil {
		workerClient = &ProtobufClient{}
	}
	return workerClient.New(url)
}

func (d *Driver) baseURL(f types.Function) (us string, err error) {
	baseURL := os.Getenv("STUCCO_WORKER_BASE_URL")
	if funcURL := os.Getenv("STUCCO_" + normalizeFuncName(f.Name) + "_URL"); funcURL != "" {
		baseURL = funcURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	u.Path = path.Join(u.Path, f.Name)
	us = u.String()
	return
}

func (d *Driver) SetSecrets(in driver.SetSecretsInput) driver.SetSecretsOutput {
	// noop, secrets must be sed during deployment
	return driver.SetSecretsOutput{}
}

func (d *Driver) functionClient(f types.Function) (client driver.Driver, derr *driver.Error) {
	url, err := d.baseURL(f)
	if err != nil {
		derr = &driver.Error{
			Message: err.Error(),
		}
		return
	}
	client = d.newClient(url)
	return
}

func (d *Driver) FieldResolve(in driver.FieldResolveInput) driver.FieldResolveOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.FieldResolveOutput{
			Error: err,
		}
	}
	return client.FieldResolve(in)
}

func (d *Driver) InterfaceResolveType(in driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.InterfaceResolveTypeOutput{
			Error: err,
		}
	}
	return client.InterfaceResolveType(in)
}

func (d *Driver) ScalarParse(in driver.ScalarParseInput) driver.ScalarParseOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarParseOutput{
			Error: err,
		}
	}
	return client.ScalarParse(in)
}
func (d *Driver) ScalarSerialize(in driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarSerializeOutput{
			Error: err,
		}
	}
	return client.ScalarSerialize(in)
}
func (d *Driver) UnionResolveType(in driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.UnionResolveTypeOutput{
			Error: err,
		}
	}
	return client.UnionResolveType(in)
}
func (d *Driver) Stream(in driver.StreamInput) driver.StreamOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.StreamOutput{
			Error: err,
		}
	}
	return client.Stream(in)
}
