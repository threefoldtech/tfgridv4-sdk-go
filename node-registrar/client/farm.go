package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var ErrorFarmNotFround = fmt.Errorf("failed to get requested farm from node regiatrar")

func (c RegistrarClient) CreateFarm(farmName string, twinID uint64, dedicated bool) (farmID uint64, err error) {
	return c.createFarm(farmName, twinID, dedicated)
}

func (c RegistrarClient) UpdateFarm(farmID uint64, opts ...UpdateFarmOpts) (err error) {
	return c.updateFarm(farmID, opts)
}

func (c RegistrarClient) GetFarm(id uint64) (farm Farm, err error) {
	return c.getFarm(id)
}

func (c RegistrarClient) ListFarms(opts ...ListFarmOpts) (farms []Farm, err error) {
	return c.listFarms(opts...)
}

type farmCfg struct {
	farmName  string
	farmID    uint64
	twinID    uint64
	dedicated bool
	page      uint32
	size      uint32
}

type (
	ListFarmOpts   func(*farmCfg)
	UpdateFarmOpts func(*farmCfg)
)

func ListFarmWithName(name string) ListFarmOpts {
	return func(n *farmCfg) {
		n.farmName = name
	}
}

func ListFarmWithFarmID(id uint64) ListFarmOpts {
	return func(n *farmCfg) {
		n.farmID = id
	}
}

func ListFarmWithTwinID(id uint64) ListFarmOpts {
	return func(n *farmCfg) {
		n.twinID = id
	}
}

func ListFarmWithDedicated() ListFarmOpts {
	return func(n *farmCfg) {
		n.dedicated = true
	}
}

func ListFarmWithPage(page uint32) ListFarmOpts {
	return func(n *farmCfg) {
		n.page = page
	}
}

func ListFarmWithSize(size uint32) ListFarmOpts {
	return func(n *farmCfg) {
		n.size = size
	}
}

func UpdateFarmWithName(name string) UpdateFarmOpts {
	return func(n *farmCfg) {
		n.farmName = name
	}
}

func UpdateFarmWithDedicated() UpdateFarmOpts {
	return func(n *farmCfg) {
		n.dedicated = true
	}
}

func (c RegistrarClient) createFarm(farmName string, twinID uint64, dedicated bool) (farmID uint64, err error) {
	err = c.ensureTwinID()
	if err != nil {
		return farmID, errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "farms")
	if err != nil {
		return farmID, errors.Wrap(err, "failed to construct registrar url")
	}

	data := Farm{
		FarmName:  farmName,
		TwinID:    twinID,
		Dedicated: dedicated,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return farmID, errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return farmID, errors.Wrap(err, "failed to construct http request to the registrar")
	}

	req.Header.Set("X-Auth", c.signRequest(time.Now().Unix()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return farmID, errors.Wrap(err, "failed to send request to create farm")
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

func (c RegistrarClient) updateFarm(farmID uint64, opts []UpdateFarmOpts) (err error) {
	err = c.ensureTwinID()
	if err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(farmID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	var body bytes.Buffer
	data := parseUpdateFarmOpts(opts)

	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request to the registrar")
	}

	req.Header.Set("X-Auth", c.signRequest(time.Now().Unix()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request to update farm")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return fmt.Errorf("failed to create farm with status code %s", resp.Status)
	}

	return
}

func (c RegistrarClient) getFarm(id uint64) (farm Farm, err error) {
	url, err := url.JoinPath(c.baseURL, "farms", fmt.Sprint(id))
	if err != nil {
		return farm, errors.Wrap(err, "failed to construct registrar url")
	}
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return farm, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return farm, ErrorFarmNotFround
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return farm, errors.Wrapf(err, "failed to get farm with status code %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&farm)
	if err != nil {
		return farm, err
	}

	return
}

func (c RegistrarClient) listFarms(opts ...ListFarmOpts) (farms []Farm, err error) {
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

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return farms, errors.Wrapf(err, "failed to get list farms with status code %s", resp.Status)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&farms)
	if err != nil {
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
	cfg := farmCfg{
		farmName:  "",
		dedicated: false,
	}

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

	return data
}
