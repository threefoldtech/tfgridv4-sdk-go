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

// UpdateFarm updates an existing farm's configuration
func (c *RegistrarClient) UpdateFarm(farmID uint64, update FarmUpdate) (err error) {
	return c.updateFarm(farmID, update)
}

// GetFarm get a farm using its farmID
func (c *RegistrarClient) GetFarm(farmID uint64) (farm Farm, err error) {
	return c.getFarm(farmID)
}

// ListFarms gets a list of farms using filter options
func (c *RegistrarClient) ListFarms(filter FarmFilter) (farms []Farm, err error) {
	return c.listFarms(filter)
}

// ApproveNodes approves multiple nodes for a specific farm
func (c *RegistrarClient) ApproveNodes(farmID uint64, nodeIDs []uint64) error {
	return c.approveNodes(farmID, nodeIDs)
}

func (c *RegistrarClient) approveNodes(farmID uint64, nodeIDs []uint64) error {
	if err := c.ensureTwinID(); err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(farmID), "approve")
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	data := struct {
		NodeIDs []uint64 `json:"node_ids"`
	}{
		NodeIDs: nodeIDs,
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(data); err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("POST", url, &body)
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
		return errors.Wrap(err, "failed to send request to approve nodes")
	}

	if resp == nil {
		return errors.New("failed to approve nodes, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return errors.Wrapf(err, "failed to approve nodes with status code %s", resp.Status)
	}

	return nil
}

// FarmUpdate represents the data needed to update an existing farm
type FarmUpdate struct {
	FarmName       *string
	StellarAddress *string
	Dedicated      *bool
}

// FarmFilter represents filtering options for listing farms
type FarmFilter struct {
	FarmID    *uint64
	FarmName  *string
	TwinID    *uint64
	Dedicated *bool
	Page      *uint32
	Size      *uint32
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

func (c *RegistrarClient) updateFarm(farmID uint64, update FarmUpdate) (err error) {
	if err = c.ensureTwinID(); err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(farmID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	var body bytes.Buffer
	data := parseUpdateFarmOpts(update)

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
		return errors.Wrapf(err, "failed to update farm with status code %s", resp.Status)
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

func (c *RegistrarClient) listFarms(filter FarmFilter) (farms []Farm, err error) {
	url, err := url.JoinPath(c.baseURL, "farms")
	if err != nil {
		return farms, errors.Wrap(err, "failed to construct registrar url")
	}

	data := parseListFarmOpts(filter)

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

func parseListFarmOpts(filter FarmFilter) map[string]any {
	data := map[string]any{}

	if filter.FarmName != nil && *filter.FarmName != "" {
		data["farm_name"] = *filter.FarmName
	}

	if filter.FarmID != nil {
		data["farm_id"] = *filter.FarmID
	}

	if filter.TwinID != nil {
		data["twin_id"] = *filter.TwinID
	}

	if filter.Dedicated != nil {
		data["dedicated"] = *filter.Dedicated
	}

	page := uint32(1)
	if filter.Page != nil {
		page = *filter.Page
	}
	data["page"] = page

	size := uint32(50)
	if filter.Size != nil {
		size = *filter.Size
	}
	data["size"] = size

	return data
}

func parseUpdateFarmOpts(update FarmUpdate) map[string]any {
	data := map[string]any{}

	if update.FarmName != nil {
		data["farm_name"] = *update.FarmName
	}

	if update.StellarAddress != nil {
		data["stellar_address"] = *update.StellarAddress
	}

	if update.Dedicated != nil {
		data["dedicated"] = *update.Dedicated
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
