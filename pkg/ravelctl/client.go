package ravelctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alexisbouchez/ravel/api"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Namespaces

func (c *Client) ListNamespaces() ([]api.Namespace, error) {
	var result []api.Namespace
	err := c.do("GET", "/namespaces", nil, &result)
	return result, err
}

func (c *Client) CreateNamespace(name string) (*api.Namespace, error) {
	var result api.Namespace
	err := c.do("POST", "/namespaces", map[string]string{"name": name}, &result)
	return &result, err
}

func (c *Client) DeleteNamespace(name string) error {
	return c.do("DELETE", "/namespaces/"+name, nil, nil)
}

// Fleets

func (c *Client) ListFleets(namespace string) ([]api.Fleet, error) {
	var result []api.Fleet
	err := c.do("GET", "/fleets?namespace="+url.QueryEscape(namespace), nil, &result)
	return result, err
}

func (c *Client) CreateFleet(namespace, name string) (*api.Fleet, error) {
	var result api.Fleet
	err := c.do("POST", "/fleets?namespace="+url.QueryEscape(namespace), map[string]string{"name": name}, &result)
	return &result, err
}

func (c *Client) DeleteFleet(namespace, name string) error {
	return c.do("DELETE", "/fleets/"+name+"?namespace="+url.QueryEscape(namespace), nil, nil)
}

// Machines

func (c *Client) ListMachines(namespace, fleet string) ([]api.Machine, error) {
	var result []api.Machine
	err := c.do("GET", fmt.Sprintf("/fleets/%s/machines?namespace=%s", fleet, url.QueryEscape(namespace)), nil, &result)
	return result, err
}

func (c *Client) GetMachine(namespace, fleet, id string) (*api.Machine, error) {
	var result api.Machine
	err := c.do("GET", fmt.Sprintf("/fleets/%s/machines/%s?namespace=%s", fleet, id, url.QueryEscape(namespace)), nil, &result)
	return &result, err
}

func (c *Client) CreateMachine(namespace, fleet string, req *api.CreateMachinePayload) (*api.Machine, error) {
	var result api.Machine
	err := c.do("POST", fmt.Sprintf("/fleets/%s/machines?namespace=%s", fleet, url.QueryEscape(namespace)), req, &result)
	return &result, err
}

func (c *Client) StartMachine(namespace, fleet, id string) error {
	return c.do("POST", fmt.Sprintf("/fleets/%s/machines/%s/start?namespace=%s", fleet, id, url.QueryEscape(namespace)), nil, nil)
}

func (c *Client) StopMachine(namespace, fleet, id string) error {
	return c.do("POST", fmt.Sprintf("/fleets/%s/machines/%s/stop?namespace=%s", fleet, id, url.QueryEscape(namespace)), nil, nil)
}

func (c *Client) DeleteMachine(namespace, fleet, id string, force bool) error {
	path := fmt.Sprintf("/fleets/%s/machines/%s?namespace=%s", fleet, id, url.QueryEscape(namespace))
	if force {
		path += "&force=true"
	}
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) GetMachineLogs(namespace, fleet, id string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/fleets/%s/machines/%s/logs?namespace=%s", c.baseURL, fleet, id, url.QueryEscape(namespace)), nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Gateways

func (c *Client) ListGateways(namespace, fleet string) ([]api.Gateway, error) {
	var result []api.Gateway
	err := c.do("GET", fmt.Sprintf("/fleets/%s/gateways?namespace=%s", fleet, url.QueryEscape(namespace)), nil, &result)
	return result, err
}

func (c *Client) CreateGateway(namespace, fleet, name string, targetPort int) (*api.Gateway, error) {
	var result api.Gateway
	err := c.do("POST", fmt.Sprintf("/fleets/%s/gateways?namespace=%s", fleet, url.QueryEscape(namespace)),
		map[string]interface{}{"name": name, "target_port": targetPort}, &result)
	return &result, err
}

func (c *Client) DeleteGateway(namespace, fleet, name string) error {
	return c.do("DELETE", fmt.Sprintf("/fleets/%s/gateways/%s?namespace=%s", fleet, name, url.QueryEscape(namespace)), nil, nil)
}

// Nodes

func (c *Client) ListNodes() ([]api.Node, error) {
	var result []api.Node
	err := c.do("GET", "/nodes", nil, &result)
	return result, err
}
