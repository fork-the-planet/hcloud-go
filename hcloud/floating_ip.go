package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// FloatingIP represents a Floating IP in the Hetzner Cloud.
type FloatingIP struct {
	ID           int64
	Description  string
	Created      time.Time
	IP           net.IP
	Network      *net.IPNet
	Type         FloatingIPType
	Server       *Server
	DNSPtr       map[string]string
	HomeLocation *Location
	Blocked      bool
	Protection   FloatingIPProtection
	Labels       map[string]string
	Name         string
}

// DNSPtrForIP returns the reverse DNS pointer of the IP address.
// Deprecated: Use GetDNSPtrForIP instead.
func (f *FloatingIP) DNSPtrForIP(ip net.IP) string {
	return f.DNSPtr[ip.String()]
}

// FloatingIPProtection represents the protection level of a Floating IP.
type FloatingIPProtection struct {
	Delete bool
}

// FloatingIPType represents the type of Floating IP.
type FloatingIPType string

// Floating IP types.
const (
	FloatingIPTypeIPv4 FloatingIPType = "ipv4"
	FloatingIPTypeIPv6 FloatingIPType = "ipv6"
)

// changeDNSPtr changes or resets the reverse DNS pointer for an IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (f *FloatingIP) changeDNSPtr(ctx context.Context, client *Client, ip net.IP, ptr *string) (*Action, *Response, error) {
	reqBody := schema.FloatingIPActionChangeDNSPtrRequest{
		IP:     ip.String(),
		DNSPtr: ptr,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/floating_ips/%d/actions/change_dns_ptr", f.ID)
	req, err := client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.FloatingIPActionChangeDNSPtrResponse{}
	resp, err := client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// GetDNSPtrForIP searches for the dns assigned to the given IP address.
// It returns an error if there is no dns set for the given IP address.
func (f *FloatingIP) GetDNSPtrForIP(ip net.IP) (string, error) {
	dns, ok := f.DNSPtr[ip.String()]
	if !ok {
		return "", DNSNotFoundError{ip}
	}

	return dns, nil
}

// FloatingIPClient is a client for the Floating IP API.
type FloatingIPClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a Floating IP by its ID. If the Floating IP does not exist,
// nil is returned.
func (c *FloatingIPClient) GetByID(ctx context.Context, id int64) (*FloatingIP, *Response, error) {
	reqPath := fmt.Sprintf("/floating_ips/%d", id)

	respBody, resp, err := getRequest[schema.FloatingIPGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return FloatingIPFromSchema(respBody.FloatingIP), resp, nil
}

// GetByName retrieves a Floating IP by its name. If the Floating IP does not exist, nil is returned.
func (c *FloatingIPClient) GetByName(ctx context.Context, name string) (*FloatingIP, *Response, error) {
	return firstByName(name, func() ([]*FloatingIP, *Response, error) {
		return c.List(ctx, FloatingIPListOpts{Name: name})
	})
}

// Get retrieves a Floating IP by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Floating IP by its name. If the Floating IP does not exist, nil is returned.
func (c *FloatingIPClient) Get(ctx context.Context, idOrName string) (*FloatingIP, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// FloatingIPListOpts specifies options for listing Floating IPs.
type FloatingIPListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l FloatingIPListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Floating IPs for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *FloatingIPClient) List(ctx context.Context, opts FloatingIPListOpts) ([]*FloatingIP, *Response, error) {
	reqPath := fmt.Sprintf("/floating_ips?%s", opts.values().Encode())

	respBody, resp, err := getRequest[schema.FloatingIPListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.FloatingIPs, FloatingIPFromSchema), resp, nil
}

// All returns all Floating IPs.
func (c *FloatingIPClient) All(ctx context.Context) ([]*FloatingIP, error) {
	return c.AllWithOpts(ctx, FloatingIPListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Floating IPs for the given options.
func (c *FloatingIPClient) AllWithOpts(ctx context.Context, opts FloatingIPListOpts) ([]*FloatingIP, error) {
	return iterPages(func(page int) ([]*FloatingIP, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// FloatingIPCreateOpts specifies options for creating a Floating IP.
type FloatingIPCreateOpts struct {
	Type         FloatingIPType
	HomeLocation *Location
	Server       *Server
	Description  *string
	Name         *string
	Labels       map[string]string
}

// Validate checks if options are valid.
func (o FloatingIPCreateOpts) Validate() error {
	switch o.Type {
	case FloatingIPTypeIPv4, FloatingIPTypeIPv6:
		break
	default:
		return errors.New("missing or invalid type")
	}
	if o.HomeLocation == nil && o.Server == nil {
		return errors.New("one of home location or server is required")
	}
	return nil
}

// FloatingIPCreateResult is the result of creating a Floating IP.
type FloatingIPCreateResult struct {
	FloatingIP *FloatingIP
	Action     *Action
}

// Create creates a Floating IP.
func (c *FloatingIPClient) Create(ctx context.Context, opts FloatingIPCreateOpts) (FloatingIPCreateResult, *Response, error) {
	if err := opts.Validate(); err != nil {
		return FloatingIPCreateResult{}, nil, err
	}

	reqBody := schema.FloatingIPCreateRequest{
		Type:        string(opts.Type),
		Description: opts.Description,
		Name:        opts.Name,
	}
	if opts.HomeLocation != nil {
		reqBody.HomeLocation = Ptr(opts.HomeLocation.Name)
	}
	if opts.Server != nil {
		reqBody.Server = Ptr(opts.Server.ID)
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return FloatingIPCreateResult{}, nil, err
	}

	req, err := c.client.NewRequest(ctx, "POST", "/floating_ips", bytes.NewReader(reqBodyData))
	if err != nil {
		return FloatingIPCreateResult{}, nil, err
	}

	var respBody schema.FloatingIPCreateResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return FloatingIPCreateResult{}, resp, err
	}
	var action *Action
	if respBody.Action != nil {
		action = ActionFromSchema(*respBody.Action)
	}
	return FloatingIPCreateResult{
		FloatingIP: FloatingIPFromSchema(respBody.FloatingIP),
		Action:     action,
	}, resp, nil
}

// Delete deletes a Floating IP.
func (c *FloatingIPClient) Delete(ctx context.Context, floatingIP *FloatingIP) (*Response, error) {
	reqPath := fmt.Sprintf("/floating_ips/%d", floatingIP.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// FloatingIPUpdateOpts specifies options for updating a Floating IP.
type FloatingIPUpdateOpts struct {
	Description string
	Labels      map[string]string
	Name        string
}

// Update updates a Floating IP.
func (c *FloatingIPClient) Update(ctx context.Context, floatingIP *FloatingIP, opts FloatingIPUpdateOpts) (*FloatingIP, *Response, error) {
	reqBody := schema.FloatingIPUpdateRequest{
		Description: opts.Description,
		Name:        opts.Name,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	reqPath := fmt.Sprintf("/floating_ips/%d", floatingIP.ID)

	respBody, resp, err := putRequest[schema.FloatingIPUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return FloatingIPFromSchema(respBody.FloatingIP), resp, nil
}

// Assign assigns a Floating IP to a server.
func (c *FloatingIPClient) Assign(ctx context.Context, floatingIP *FloatingIP, server *Server) (*Action, *Response, error) {
	reqBody := schema.FloatingIPActionAssignRequest{
		Server: server.ID,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/floating_ips/%d/actions/assign", floatingIP.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.FloatingIPActionAssignResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Unassign unassigns a Floating IP from the currently assigned server.
func (c *FloatingIPClient) Unassign(ctx context.Context, floatingIP *FloatingIP) (*Action, *Response, error) {
	var reqBody schema.FloatingIPActionUnassignRequest
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/floating_ips/%d/actions/unassign", floatingIP.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.FloatingIPActionUnassignResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeDNSPtr changes or resets the reverse DNS pointer for a Floating IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (c *FloatingIPClient) ChangeDNSPtr(ctx context.Context, floatingIP *FloatingIP, ip string, ptr *string) (*Action, *Response, error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return nil, nil, InvalidIPError{ip}
	}
	return floatingIP.changeDNSPtr(ctx, c.client, net.ParseIP(ip), ptr)
}

// FloatingIPChangeProtectionOpts specifies options for changing the resource protection level of a Floating IP.
type FloatingIPChangeProtectionOpts struct {
	Delete *bool
}

// ChangeProtection changes the resource protection level of a Floating IP.
func (c *FloatingIPClient) ChangeProtection(ctx context.Context, floatingIP *FloatingIP, opts FloatingIPChangeProtectionOpts) (*Action, *Response, error) {
	reqBody := schema.FloatingIPActionChangeProtectionRequest{
		Delete: opts.Delete,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/floating_ips/%d/actions/change_protection", floatingIP.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.FloatingIPActionChangeProtectionResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}
