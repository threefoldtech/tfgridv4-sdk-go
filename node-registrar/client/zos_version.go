package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func (c RegistrarClient) GetZosVersion() (version ZosVersion, err error) {
	return c.getZosVersion()
}

func (c RegistrarClient) SetZosVersion(v string, safeToUpgrade bool) (err error) {
	return c.setZosVersion(v, safeToUpgrade)
}

func (c RegistrarClient) getZosVersion() (version ZosVersion, err error) {
	url, err := url.JoinPath(c.baseURL, "zos", "version")
	if err != nil {
		return version, errors.Wrap(err, "failed to construct registrar url")
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return version, err
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return version, errors.Wrapf(err, "failed to get zos version with status code %s", resp.Status)
	}

	defer resp.Body.Close()

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

func (c RegistrarClient) setZosVersion(v string, safeToUpgrade bool) (err error) {
	url, err := url.JoinPath(c.baseURL, "zos", "version")
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	version := ZosVersion{
		Version:       v,
		SafeToUpgrade: safeToUpgrade,
	}

	jsonData, err := json.Marshal(version)
	if err != nil {
		return errors.Wrap(err, "failed to marshal zos version")
	}

	encodedVersion := struct {
		Version string `json:"version"`
	}{
		Version: base64.StdEncoding.EncodeToString(jsonData),
	}

	jsonData, err = json.Marshal(encodedVersion)
	if err != nil {
		return errors.Wrap(err, "failed to marshal zos version in hex format")
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonData))
	if err != nil {
		return errors.Wrap(err, "failed to construct http request to the registrar")
	}

	req.Header.Set("X-Auth", c.signRequest())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request to get zos version from the registrar")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseResponseError(resp.Body)
	}

	return
}
