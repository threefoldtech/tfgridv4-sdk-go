# Node Registrar client

To be able to use the node registrar you can use the following scripts.

## Account Management

- **Create a Seed:**

```go
// Generate new seed
seed := make([]byte, 32)
_, err := rand.Read(seed)
if err != nil {
    panic(err)
}

hexKey := hex.EncodeToString(seed)
fmt.Println("New Seed (Hex):", hexKey)

```

- **Parse The Seed**

```go
// Generate Key Pair
privateKey := ed25519.NewKeyFromSeed(seed)
publicKey := privateKey.Public().(ed25519.PublicKey)

fmt.Println("Private Key (Hex):", hex.EncodeToString(privateKey))
fmt.Println("Public Key (Hex):", hex.EncodeToString(publicKey))

```

- **Create Account**

```go
 url, err := url.JoinPath(registrarURL, "accounts")
 if err != nil {
  panic(err)
 }

 timestamp := time.Now().Unix()
 publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

 challenge := []byte(fmt.Sprintf("%d:%v", timestamp, publicKeyBase64))
 signature := ed25519.Sign(privateKey, challenge)

 data := map[string]any{
  "public_key":  publicKey,
  "signature":   signature,
  "timestamp":   timestamp,
  "rmb_enc_key": rmbEncKey,
  "relays":      relays,
 }

 var body bytes.Buffer
 err = json.NewEncoder(&body).Encode(data)
 if err != nil {
  panic(err)
 }

 resp, err := http.DefaultClient.Post(url, "application/json", &body)
 if err != nil {
  panic(err)
 }

 if resp.StatusCode != http.StatusCreated {
  panic(fmt.Errorf("account not created successfully"))
 }

 defer resp.Body.Close()

 var account map[string]any
 err = json.NewDecoder(resp.Body).Decode(&account)

 fmt.Println(account["twin_id"])
```

- **Get Account:**

```go
 url, err := url.JoinPath(registrarURL, "accounts")
 if err != nil {
  panic(err)
 }

 req, err := http.NewRequest("GET", url, nil)
 if err != nil {
  return
 }

 q := req.URL.Query()
 q.Add("twin_id", fmt.Sprint(twinID))
 req.URL.RawQuery = q.Encode()

 resp, err := http.DefaultClient.Do(req)
 if err != nil {
  panic(err)
 }
 defer resp.Body.Close()

 if resp.StatusCode != http.StatusNotFound {
  panic(fmt.Errorf("status code not ok"))
 }

 var account map[string]any
 err = json.NewDecoder(resp.Body).Decode(&account)
 fmt.Println(account)
```

## Farm Management

- **Create a Farm:**

```go
 url, err := url.JoinPath(registrarURL, "farms")
 if err != nil {
  panic(err)
 }

 data := map[string]any{
  "farm_name": farmName,
  "twin_id":   twinID,
  "dedicated": dedicated,
 }

 var body bytes.Buffer
 err = json.NewEncoder(&body).Encode(data)
 if err != nil {
  panic(err)
 }

 req, err := http.NewRequest("POST", url, &body)
 if err != nil {
  panic(err)
 }

 timestamp := time.Now().Unix()
 challenge := []byte(fmt.Sprintf("%d:%v", timestamp, twinID))
 signature := ed25519.Sign(privateKey, challenge)

 authHeader := fmt.Sprintf(
  "%s:%s",
  base64.StdEncoding.EncodeToString(challenge),
  base64.StdEncoding.EncodeToString(signature),
 )
 req.Header.Set("X-Auth", authHeader)
 req.Header.Set("Content-Type", "application/json")

 resp, err := http.DefaultClient.Do(req)
 if err != nil {
  panic(err)
 }

 defer resp.Body.Close()

 if resp.StatusCode != http.StatusCreated {
  panic(err)
 }

 result := struct {
  FarmID uint64 `json:"farm_id"`
 }{}

 if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
  panic(err)
 }

 fmt.Println(result.FarmID)
```

- **Get Farm:**

```go
    url, err := url.JoinPath(registrarURL, "farms", fmt.Sprint(farmID))
 if err != nil {
  panic(err)
 }
 resp, err := http.DefaultClient.Get(url)
 if err != nil {
  panic(err)
 }

 if resp.StatusCode != http.StatusOK {
  panic(err)
 }
 defer resp.Body.Close()

 var result map[string]any
 err = json.NewDecoder(resp.Body).Decode(&result)
 if err != nil {
  panic(err)
 }

 fmt.Println(result)
```

## Zos Version Management

- **Get Zos Version:**

```go
 url, err := url.JoinPath(registrarURL, "zos", "version")
 if err != nil {
  panic(err)
 }

 resp, err := http.DefaultClient.Get(url)
 if err != nil {
  panic(err)
 }

 if resp.StatusCode != http.StatusOK {
  panic(err)
 }

 defer resp.Body.Close()

 var versionString string
 err = json.NewDecoder(resp.Body).Decode(&versionString)
 if err != nil {
  panic(err)
 }

 versionBytes, err := base64.StdEncoding.DecodeString(versionString)
 if err != nil {
  panic(err)
 }

 correctedJSON := strings.ReplaceAll(string(versionBytes), "'", "\"")

 var version map[string]any
 err = json.NewDecoder(strings.NewReader(correctedJSON)).Decode(&version)
 if err != nil {
  panic(err)
 }

 fmt.Println(version)
```

- **Set Zos Version**
To set zos version you need to have the seed to the admin account

```go
 url, err := url.JoinPath(registrarURL, "zos", "version")
 if err != nil {
  panic(err)
 }

 v := "{'safe_to_upgrade': true, 'version':'v0.1.7'}"

 version := struct {
  Version string `json:"version"`
 }{
  Version: base64.StdEncoding.EncodeToString([]byte(v)),
 }

 body, err := json.Marshal(version)
 if err != nil {
  panic(err)
 }

 // Create auth headers
 timestamp := time.Now().Unix()
 challenge := []byte(fmt.Sprintf("%d:%v", timestamp, twinID))
 signature := ed25519.Sign(privateKey, challenge)

 req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
 if err != nil {
  panic(err)
 }

 // Set required headers
 authHeader := fmt.Sprintf(
  "%s:%s",
  base64.StdEncoding.EncodeToString(challenge),
  base64.StdEncoding.EncodeToString(signature),
 )
 req.Header.Set("X-Auth", authHeader)
 req.Header.Set("Content-Type", "application/json")

 resp, err := http.DefaultClient.Do(req)
 if err != nil {
  panic(err)
 }
 defer resp.Body.Close()

 if resp.StatusCode != http.StatusOK {
  panic(err)
 }

 fmt.Println("version updated successfully: ", string(body))
```

## Node Management

- **Get Node:**

```go
 url, err := url.JoinPath(registrarURL, "nodes", fmt.Sprint(nodeID))
 if err != nil {
  panic(err)
 }

 resp, err := http.DefaultClient.Get(url)
 if err != nil {
  panic(err)
 }

 if resp.StatusCode != http.StatusOK {
  panic(err)
 }
 defer resp.Body.Close()

 var result map[string]any
 err = json.NewDecoder(resp.Body).Decode(&result)
 if err != nil {
  panic(err)
 }

 fmt.Println(result)
```
