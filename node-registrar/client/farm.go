package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

var ErrorFarmNotFound = fmt.Errorf("failed to get requested farm from node registrar")

// CreateFarm create new farm on the registrar with uniqe name.
func (c *RegistrarClient) CreateFarm(farmName, stellarAddr string, dedicated bool) (farmID uint64, err error) {
	return c.createFarm(farmName, stellarAddr, dedicated)
}

// UpdateFarm update farm configuration (farmName, stellarAddress, dedicated).
func (c *RegistrarClient) UpdateFarm(farmID uint64, opts ...UpdateFarmOpts) (err error) {
	return c.updateFarm(farmID, opts)
}

// GetFarm get a farm using its farmID
func (c *RegistrarClient) GetFarm(farmID uint64) (farm Farm, err error) {
	return c.getFarm(farmID)
}

// ListFarms get a list of farm using ListFarmOpts
func (c *RegistrarClient) ListFarms(opts ...ListFarmOpts) (farms []Farm, err error) {
	return c.listFarms(opts...)
}

type farmCfg struct {
	farmName       string
	farmID         uint64
	twinID         uint64
	dedicated      bool
	stellarAddress string
	page           uint32
	size           uint32
}

type (
	ListFarmOpts   func(*farmCfg)
	UpdateFarmOpts func(*farmCfg)
)

// ListFarmWithName lists farms with farm name
func ListFarmWithName(name string) ListFarmOpts {
	return func(n *farmCfg) {
		n.farmName = name
	}
}

// ListFarmWithFarmID lists farms with farmID
func ListFarmWithFarmID(id uint64) ListFarmOpts {
	return func(n *farmCfg) {
		n.farmID = id
	}
}

// ListFarmWithTwinID lists farms with twinID
func ListFarmWithTwinID(id uint64) ListFarmOpts {
	return func(n *farmCfg) {
		n.twinID = id
	}
}

// ListFarmWithDedicated lists dedicated farms
func ListFarmWithDedicated() ListFarmOpts {
	return func(n *farmCfg) {
		n.dedicated = true
	}
}

// ListFarmWithPage lists farms in a certain page
func ListFarmWithPage(page uint32) ListFarmOpts {
	return func(n *farmCfg) {
		n.page = page
	}
}

// ListFarmWithPage lists size number of farms
func ListFarmWithSize(size uint32) ListFarmOpts {
	return func(n *farmCfg) {
		n.size = size
	}
}

// UpdateFarmWithName update farm name
func UpdateFarmWithName(name string) UpdateFarmOpts {
	return func(n *farmCfg) {
		n.farmName = name
	}
}

// UpdateFarmWithName set farm status to dedicated
func UpdateFarmWithDedicated() UpdateFarmOpts {
	return func(n *farmCfg) {
		n.dedicated = true
	}
}

// UpdateFarmWithName set farm status to dedicated
func UpdateFarmWithStellarAddress(address string) UpdateFarmOpts {
	return func(n *farmCfg) {
		n.stellarAddress = address
	}
}

func (c *RegistrarClient) createFarm(farmName, stellarAddr string, dedicated bool) (farmID uint64, err error) {
	if err := c.ensureTwinID(); err != nil {
		return farmID, errors.Wrap(err, "failed to ensure twin id")
	}

	if err = validateStellarAddress(stellarAddr); err != nil {
		return
	}

	url, err := url.JoinPath(c.baseURL, "farms")
	if err != nil {
		return farmID, errors.Wrap(err, "failed to construct registrar url")
	}

	data := Farm{
		FarmName:       farmName,
		TwinID:         c.twinID,
		Dedicated:      dedicated,
		StellarAddress: stellarAddr,
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(data); err != nil {
		return farmID, errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return farmID, errors.Wrap(err, "failed to construct http request to the registrar")
	}

	authHeader, err := c.signRequest(time.Now().Unix())
	if err != nil {
		return farmID, errors.Wrap(err, "failed to sign request")
	}
	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return farmID, errors.Wrap(err, "failed to send request to create farm")
	}

	if resp == nil {
		return farmID, errors.New("failed to create farm, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		err = parseResponseError(resp.Body)
		return farmID, fmt.Errorf("failed to create farm with status code %s", resp.Status)
	}

	result := struct {
		FarmID uint64 `json:"farm_id"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return farmID, errors.Wrap(err, "failed to decode response body")
	}

	return result.FarmID, nil
}

func (c *RegistrarClient) updateFarm(farmID uint64, opts []UpdateFarmOpts) (err error) {
	if err = c.ensureTwinID(); err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(farmID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	var body bytes.Buffer
	data := parseUpdateFarmOpts(opts)

	if stellarAddr, ok := data["stellar_address"]; ok {
		if err = validateStellarAddress(stellarAddr.(string)); err != nil {
			return
		}
	}

	if err = json.NewEncoder(&body).Encode(data); err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request to the registrar")
	}

	authHeader, err := c.signRequest(time.Now().Unix())
	if err != nil {
		return errors.Wrap(err, "failed to sign request")
	}
	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request to update farm")
	}

	if resp == nil {
		return errors.New("failed to update farm, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return errors.Wrapf(err, "failed to create farm with status code %s", resp.Status)
	}

	return
}

func (c *RegistrarClient) getFarm(id uint64) (farm Farm, err error) {
	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(id))
	if err != nil {
		return farm, errors.Wrap(err, "failed to construct registrar url")
	}
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return farm, err
	}

	if resp == nil {
		return farm, errors.New("failed to get farm, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return farm, ErrorFarmNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return farm, errors.Wrapf(err, "failed to get farm with status code %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&farm); err != nil {
		return farm, err
	}

	return
}

func (c *RegistrarClient) listFarms(opts ...ListFarmOpts) (farms []Farm, err error) {
	url, err := url.JoinPath(c.baseURL, "farms")
	if err != nil {
		return farms, errors.Wrap(err, "failed to construct registrar url")
	}

	data := parseListFarmOpts(opts)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return farms, errors.Wrap(err, "failed to construct http request to the registrar")
	}

	q := req.URL.Query()

	for key, val := range data {
		q.Add(key, fmt.Sprint(val))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return farms, errors.Wrap(err, "failed to send request to list farm")
	}

	if resp == nil {
		return farms, errors.New("failed to list farms, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return farms, errors.Wrapf(err, "failed to get list farms with status code %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&farms); err != nil {
		return farms, errors.Wrap(err, "failed to decode response body")
	}

	return
}

func parseListFarmOpts(opts []ListFarmOpts) map[string]any {
	cfg := farmCfg{
		farmName:  "",
		farmID:    0,
		twinID:    0,
		dedicated: false,
		page:      1,
		size:      50,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	data := map[string]any{}

	if len(cfg.farmName) != 0 {
		data["farm_name"] = cfg.farmName
	}

	if cfg.farmID != 0 {
		data["farm_id"] = cfg.farmID
	}

	if cfg.twinID != 0 {
		data["twin_id"] = cfg.twinID
	}

	if cfg.dedicated {
		data["dedicated"] = true
	}

	data["page"] = cfg.page
	data["size"] = cfg.size

	return data
}

func parseUpdateFarmOpts(opts []UpdateFarmOpts) map[string]any {
	cfg := farmCfg{}

	for _, opt := range opts {
		opt(&cfg)
	}

	data := map[string]any{}

	if len(cfg.farmName) != 0 {
		data["farm_name"] = cfg.farmName
	}

	if cfg.dedicated {
		data["dedicated"] = true
	}

	if len(cfg.stellarAddress) != 0 {
		data["stellar_address"] = cfg.stellarAddress
	}

	return data
}

// validateStellarAddress ensures that the address is valid stellar address
func validateStellarAddress(stellarAddr string) error {
	stellarAddr = strings.TrimSpace(stellarAddr)
	if len(stellarAddr) != 56 {
		return fmt.Errorf("invalid stellar address %s, address length should be 56 characters", stellarAddr)
	}
	if stellarAddr[0] != 'G' {
		return fmt.Errorf("invalid stellar address %s, address should should start with 'G'", stellarAddr)
	}

	if strings.Compare(stellarAddr, strings.ToUpper(stellarAddr)) != 0 {
		return fmt.Errorf("invalid stellar address %s, address should be all uppercase", stellarAddr)
	}

	// check if not alphanumeric
	for _, c := range stellarAddr {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			return fmt.Errorf("invalid stellar address %s, address character should be alphanumeric only", stellarAddr)
		}
	}
	return nil
}
