package peer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/vedhavyas/go-subkey/v2"
)

var (
	errNoCache = fmt.Errorf("not cached")
)

// TwinDB is used to get Twin instances
type TwinDB interface {
	Get(id uint32) (Twin, error)
	GetByPk(pk []byte) (uint32, error)
}

// Twin is used to store a twin id and its public key
type Twin struct {
	ID        uint32
	PublicKey []byte
	Relay     *string // TODO: multiple relays (slice?)
	E2EKey    []byte
	Timestamp uint64
}

type RegistrarTwin struct {
	TwinID    uint64   `json:"twin_id"`
	Relays    []string `json:"relays"`
	RMBEncKey string   `json:"rmb_enc_key"`
	PublicKey string   `json:"public_key"`
}

type twinDB struct {
	httpClient   *http.Client
	registrarUrl string
}

// NewTwinDB creates a new twinDBImpl instance, with a non expiring cache.
func NewTwinDB(registrarUrl string) TwinDB {
	return &twinDB{
		httpClient:   &http.Client{},
		registrarUrl: registrarUrl,
	}
}

type updateTwin struct {
	Relays    []string `json:"relays"`
	RMBEncKey string   `json:"rmb_enc_key"`
}

// GetTwin gets Twin from cache if present. if not, gets it from substrate client and caches it.
func (t *twinDB) Get(id uint32) (Twin, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v1/accounts?twin_id=%v", t.registrarUrl, id),
		nil,
	)
	if err != nil {
		return Twin{}, errors.Wrap(err, "could not create new request")
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return Twin{}, errors.Wrapf(err, "could not get twin with id %d", id)
	}
	defer resp.Body.Close()

	var registrarTwin RegistrarTwin
	if err = json.NewDecoder(resp.Body).Decode(&registrarTwin); err != nil {
		return Twin{}, err
	}

	var relay *string

	if len(registrarTwin.Relays) > 0 {
		relay = &registrarTwin.Relays[0] // TODO: will relays be a slice???
	}

	pk, err := base64.StdEncoding.DecodeString(registrarTwin.PublicKey)
	if err != nil {
		return Twin{}, err
	}

	e2ePK, err := base64.StdEncoding.DecodeString(registrarTwin.RMBEncKey)
	if err != nil {
		return Twin{}, err
	}

	twin := Twin{
		ID:        id,
		PublicKey: pk,
		Relay:     relay,
		E2EKey:    e2ePK,
	}

	return twin, nil
}

func (t *twinDB) GetByPk(pk []byte) (uint32, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v1/accounts", t.registrarUrl),
		nil,
	)
	if err != nil {
		return 0, errors.Wrap(err, "could not create new request")
	}

	q := req.URL.Query()
	q.Add("public_key", base64.StdEncoding.EncodeToString(pk))
	req.URL.RawQuery = q.Encode()

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "could not get twin")
	}
	defer resp.Body.Close()

	var registrarTwin RegistrarTwin
	if err = json.NewDecoder(resp.Body).Decode(&registrarTwin); err != nil {
		return 0, err
	}

	return uint32(registrarTwin.TwinID), nil
}

func UpdateTwin(twinID uint32, registrarUrl string, kp subkey.KeyPair, rmbEncKey []byte, relays []string) error {
	client := &http.Client{}

	timestamp := time.Now().Unix()
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, twinID))
	signature, err := kp.Sign(challenge)
	if err != nil {
		return err
	}

	updates := updateTwin{
		Relays:    relays,
		RMBEncKey: base64.StdEncoding.EncodeToString(rmbEncKey),
	}

	jsonData, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf("%s/v1/accounts/%v", registrarUrl, twinID),
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return errors.Wrap(err, "could not create new request")
	}

	authHeader := fmt.Sprintf(
		"%s:%s",
		base64.StdEncoding.EncodeToString(challenge),
		base64.StdEncoding.EncodeToString(signature),
	)

	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp == nil {
		return errors.New("failed to update twin, no response received")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with %d", resp.StatusCode)
	}

	return nil
}

// if ttl == 0, then the data will stay forever
type inMemoryCache struct {
	cache map[uint32]Twin
	inner TwinDB
	m     sync.RWMutex
	ttl   uint64
}

func newInMemoryCache(inner TwinDB, ttl uint64) TwinDB {
	return &inMemoryCache{
		cache: make(map[uint32]Twin),
		inner: inner,
		ttl:   ttl,
	}
}

func (twin *Twin) isExpired(ttl uint64) bool {
	age := uint64(time.Now().Unix()) - twin.Timestamp
	if ttl != 0 && age > ttl {
		log.Trace().Uint64("age", age).Msg("twin cache hit but expired")
		return true
	}
	return false
}

func (m *inMemoryCache) Get(id uint32) (twin Twin, err error) {
	m.m.RLock()
	twin, ok := m.cache[id]
	m.m.RUnlock()
	if ok && !twin.isExpired(m.ttl) {
		return twin, nil
	}
	twin, err = m.inner.Get(id)
	if err != nil {
		return Twin{}, errors.Wrapf(err, "could not get twin with id %d", id)
	}
	twin.Timestamp = uint64(time.Now().Unix())
	m.m.Lock()
	m.cache[id] = twin
	m.m.Unlock()

	return twin, nil
}

func (m *inMemoryCache) GetByPk(pk []byte) (uint32, error) {
	return m.inner.GetByPk(pk)
}

type tmpCache struct {
	base  string
	ttl   uint64
	inner TwinDB
}

func newTmpCache(ttl uint64, inner TwinDB) (TwinDB, error) {
	path := filepath.Join(os.TempDir(), "rmb-cache")
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	return &tmpCache{
		base:  path,
		ttl:   ttl,
		inner: inner,
	}, nil
}

func (r *tmpCache) get(path string) (twin Twin, err error) {
	data, err := os.ReadFile(path)

	if os.IsNotExist(err) {
		return twin, errNoCache
	} else if err != nil {
		return twin, err
	}

	err = json.Unmarshal(data, &twin)
	if err != nil {
		// we return an errNoCache so we don't
		// crash on file corruption
		return twin, errNoCache
	}
	if twin.isExpired(r.ttl) {
		return twin, errNoCache
	}

	log.Trace().Msg("twin cache hit")
	return twin, nil
}

func (r *tmpCache) set(path string, twin Twin) error {
	data, err := json.Marshal(twin)

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (r *tmpCache) Get(id uint32) (twin Twin, err error) {
	path := filepath.Join(r.base, fmt.Sprint(id))

	twin, err = r.get(path)
	if err == errNoCache {
		twin, err = r.inner.Get(id)
		if err != nil {
			return twin, err
		}
		// set cache
		twin.Timestamp = uint64(time.Now().Unix())
		if err := r.set(path, twin); err != nil {
			log.Error().Err(err).Msg("failed to warm up cache")
		}
		return twin, nil
	} else if err != nil {
		return twin, err
	}

	return twin, nil
}

func (r *tmpCache) GetByPk(pk []byte) (uint32, error) {
	return r.inner.GetByPk(pk)
}
