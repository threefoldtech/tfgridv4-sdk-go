package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// GetZosVersion gets zos version for specific network
func (c *RegistrarClient) GetZosVersion() (version ZosVersion, err error) {
	return c.getZosVersion()
}

// SetZosVersion sets zos version for specific network only valid for network admin
func (c *RegistrarClient) SetZosVersion(v string, safeToUpgrade bool) (err error) {
	return c.setZosVersion(v, safeToUpgrade)
}

func (c *RegistrarClient) getZosVersion() (version ZosVersion, err error) {
	url, err := url.JoinPath(c.baseURL, "zos", "version")
	if err != nil {
		return version, errors.Wrap(err, "failed to construct registrar url")
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return version, err
	}

	if resp == nil {
		return version, errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return version, errors.Wrapf(err, "failed to get zos version with status code %s", resp.Status)
	}

	var versionString string
	err = json.NewDecoder(resp.Body).Decode(&versionString)
	if err != nil {
		return version, err
	}

	versionBytes, err := base64.StdEncoding.DecodeString(versionString)
	if err != nil {
		return version, err
	}

	correctedJSON := strings.ReplaceAll(string(versionBytes), "'", "\"")

	err = json.NewDecoder(strings.NewReader(correctedJSON)).Decode(&version)
	if err != nil {
		return version, err
	}

	return
}

func (c *RegistrarClient) setZosVersion(v string, safeToUpgrade bool) (err error) {
	err = c.ensureTwinID()
	if err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "zos", "version")
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	version := ZosVersion{
		Version:       v,
		SafeToUpgrade: safeToUpgrade,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(version)
	if err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest("PUT", url, &body)
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
		return errors.Wrap(err, "failed to send request to get zos version from the registrar")
	}

	if resp == nil {
		return errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseResponseError(resp.Body)
	}

	return
}
