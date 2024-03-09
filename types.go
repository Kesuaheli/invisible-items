package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RawVersionManifest struct {
	Latest   map[string]string `json:"latest"`
	Versions []VersionManifest `json:"versions"`
}

type DataDownload struct {
	SHA1 string `json:"sha1"`
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type VersionManifest struct {
	ID   string    `json:"id"`
	Type string    `json:"type"`
	Time time.Time `json:"releaseTime"`
	DataDownload
}

type Version struct {
	AssetIndex DataDownload `json:"assetIndex"`
	Downloads  JARDownload  `json:"downloads"`
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Time       time.Time    `json:"releaseTime"`
}

type JARDownload struct {
	Client         DataDownload `json:"client"`
	ClientMappings DataDownload `json:"client_mappings"`
	Server         DataDownload `json:"server"`
	ServerMappings DataDownload `json:"server_mappings"`
}

type AssetIndex struct {
	Objects map[string]Asset `json:"objects"`
}

type Asset struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

type Client struct {
	rawData []byte
	files   []*zip.File
}

type ClientVersions struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	WorldVersion    int    `json:"world_version"`
	Series          string `json:"series"`
	ProtocolVersion int    `json:"protocol_version"`
	PackVersion     struct {
		Resource int `json:"resource"`
		Data     int `json:"data"`
	} `json:"pack_version"`
	BuildTime     time.Time `json:"build_time"`
	JavaComponent string    `json:"java_component"`
	JavaVersion   int       `json:"java_version"`
	Stable        bool      `json:"stable"`
	UseEditor     bool      `json:"use_editor"`
}

type PackMCMeta struct {
	PackMCMetaPack `json:"pack"`
}

type PackMCMetaPack struct {
	PackFormat  int    `json:"pack_format"`
	Description string `json:"description,omitempty"`
}

type Model struct {
	Parent    string          `json:"parent,omitempty"`
	Display   interface{}     `json:"display,omitempty"`
	Textures  interface{}     `json:"textures,omitempty"`
	GUILight  interface{}     `json:"gui_light,omitempty"`
	Elements  []interface{}   `json:"elements,omitempty"`
	Overrides []ModelOverride `json:"overrides,omitempty"`
}

type ModelOverride struct {
	Predicate ModelOverridePredicate `json:"predicate"`
	Model     string                 `json:"model"`
}

type ModelOverridePredicate struct {
	CustomModelData uint `json:"custom_model_data"`
}

func GetVersion(url, id string) (Version, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	if resp.StatusCode != 200 {
		return Version{}, fmt.Errorf("get version: expected status 200 but got %d", resp.StatusCode)
	}

	manifestData, err := io.ReadAll(resp.Body)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	var manifest RawVersionManifest
	err = json.Unmarshal(manifestData, &manifest)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	if id == "release" || id == "snapshot" {
		id = manifest.Latest[id]
	}

	var vManifest VersionManifest
	var found bool
	for _, ver := range manifest.Versions {
		if ver.ID == id {
			vManifest = ver
			found = true
			break
		}
	}
	if !found {
		return Version{}, fmt.Errorf("get version: could not find specified version \"%s\"", id)
	}

	req, err = http.NewRequest(http.MethodGet, vManifest.URL, nil)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	if resp.StatusCode != 200 {
		return Version{}, fmt.Errorf("get version: expected status 200 but got %d", resp.StatusCode)
	}

	versionData, err := io.ReadAll(resp.Body)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	hasher := sha1.New()
	hasher.Write(versionData)
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != vManifest.SHA1 {
		return Version{}, fmt.Errorf("get version: SHA1 has validation failure: expexted \"%s\" but got \"%s\"", vManifest.SHA1, hash)
	}

	var version Version
	err = json.Unmarshal(versionData, &version)
	if err != nil {
		return Version{}, fmt.Errorf("get version: %v", err)
	}

	return version, nil
}

func (v Version) GetAssets() (map[string]Asset, error) {
	req, err := http.NewRequest(http.MethodGet, v.AssetIndex.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get assets: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get assets: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get assets: expected status 200 but got %d", resp.StatusCode)
	}

	assetData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get assets: %v", err)
	}

	hasher := sha1.New()
	hasher.Write(assetData)
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != v.AssetIndex.SHA1 {
		return nil, fmt.Errorf("get assets: SHA1 has validation failure: expexted \"%s\" but got \"%s\"", v.AssetIndex.SHA1, hash)
	}

	var assetIndex AssetIndex
	err = json.Unmarshal(assetData, &assetIndex)
	if err != nil {
		return nil, fmt.Errorf("get assets: %v", err)
	}

	return assetIndex.Objects, nil
}

func (v Version) GetClient() (*Client, error) {
	req, err := http.NewRequest(http.MethodGet, v.Downloads.Client.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get client: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get client: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get client: expected status 200 but got %d", resp.StatusCode)
	}

	clientData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get client: %v", err)
	}

	hasher := sha1.New()
	hasher.Write(clientData)
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != v.Downloads.Client.SHA1 {
		return nil, fmt.Errorf("get client: SHA1 has validation failure: expexted \"%s\" but got \"%s\"", v.Downloads.Client.SHA1, hash)
	}
	if len(clientData) != v.Downloads.Client.Size {
		return nil, fmt.Errorf("get client: SHA1 matches but client has %db of data but expected %db", len(clientData), v.Downloads.Client.Size)
	}
	return &Client{rawData: clientData}, nil
}

func (c Client) Versions() *ClientVersions {
	if c.files == nil {
		zipReader, _ := zip.NewReader(bytes.NewReader(c.rawData), int64(len(c.rawData)))
		c.files = zipReader.File
	}

	for _, f := range c.files {
		if f.Name == "version.json" {
			fileReader, _ := f.Open()
			data, _ := io.ReadAll(fileReader)
			clientVersions := &ClientVersions{}
			json.Unmarshal(data, clientVersions)
			return clientVersions
		}
	}
	return nil
}

func (c Client) GetFiles(pathPrefix string) map[string][]byte {
	if c.files == nil {
		zipReader, _ := zip.NewReader(bytes.NewReader(c.rawData), int64(len(c.rawData)))
		c.files = zipReader.File
	}

	files := make(map[string][]byte)
	for _, f := range c.files {
		if name, ok := strings.CutPrefix(f.Name, pathPrefix); ok {
			//fmt.Printf("File: \"%s\" (%db)\n", name, f.UncompressedSize64)
			fileReader, _ := f.Open()
			buf, _ := io.ReadAll(fileReader)
			files[name] = buf
		}
	}
	return files
}

func (c Client) GetItemModels() map[string]*Model {
	iteModels := make(map[string]*Model)
	for name, file := range c.GetFiles("assets/minecraft/models/item/") {
		if strings.HasPrefix(name, "template_") ||
			name == "air" {
			continue
		}
		itemModel := &Model{}
		json.Unmarshal(file, itemModel)
		iteModels[name] = itemModel
	}
	return iteModels
}
