package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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

	err = json.Unmarshal(bodyBytes, &version)
	if err != nil {
		// try decoding base64 version
		var versionString string
		err = json.Unmarshal(bodyBytes, &versionString)
		if err != nil {
			return version, err
		}

		decodedVersion, err := base64.StdEncoding.DecodeString(versionString)
		if err != nil {
			return version, err
		}

		correctedJSON := strings.ReplaceAll(string(decodedVersion), "'", "\"")

		err = json.Unmarshal([]byte(correctedJSON), &version)

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

	sendRequest := func(body bytes.Buffer) (resp *http.Response, err error) {
		req, err := http.NewRequest("PUT", url, &body)
		if err != nil {
			return resp, errors.Wrap(err, "failed to construct http request to the registrar")
		}

		authHeader, err := c.signRequest(time.Now().Unix())
		if err != nil {
			return resp, errors.Wrap(err, "failed to sign request")
		}
		req.Header.Set("X-Auth", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return resp, errors.Wrap(err, "failed to send request to get zos version from the registrar")
		}
		return resp, nil
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

	resp, err := sendRequest(body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		// fallback to old encoded format
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

		resp, err = sendRequest(*bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
	}

	if resp.StatusCode != http.StatusOK {
		return parseResponseError(resp.Body)
	}

	return
}
