package xrd

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
	"io"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

const autoGenHeader = "## WARNING: This file was autogenerated!\n" +
	"## Manual modifications will be overwritten\n" +
	"## Last Modification: %s.\n"

type OverrideField struct {
	Path     string `yaml:"path" json:"path"`
	Value    any    `yaml:"value,omitempty" json:"value,omitempty"`
	Override any    `yaml:"override,omitempty" json:"override,omitempty"`
	Ignore   bool   `yaml:"ignore" json:"ignore"`
}

type Generator struct {
	Group                string          `yaml:"group" json:"group"`
	Version              string          `yaml:"version" json:"version"`
	Kind                 string          `yaml:"kind" json:"kind"`
	Plural               string          `yaml:"plural" json:"plural"`
	ConnectionSecretKeys *[]string       `yaml:"connectionSecretKeys,omitempty" json:"connectionSecretKeys,omitempty"`
	Ignore               bool            `yaml:"ignore"`
	PatchExternalName    *bool           `yaml:"patchExternalName,omitempty" json:"patchExternalName,omitempty"`
	UIDFieldPath         *string         `yaml:"uidFieldPath,omitempty" json:"uidFieldPath,omitempty"`
	OverrideFields       []OverrideField `yaml:"overrideFields" json:"overrideFields"`
	ReadinessChecks      *bool           `yaml:"readinessChecks, omitempty" json:"readinessChecks,omitempty"`

	ManifestsPath string
	configPath    string
	tagType       string
	tagProperty   string
}
type jsonnetOutput map[string]any

func (g *Generator) Generate(crd []byte) error {
	vm := jsonnet.MakeVM()

	configJSON, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	vm.ExtVar("config", string(configJSON))
	vm.ExtVar("crd", string(crd))

	str, err := vm.EvaluateFile("./hack/xrd-gen.jsonnet")
	if err != nil {
		return fmt.Errorf("failed to evaluate jsonnet: %w", err)
	}
	out := make(jsonnetOutput)
	if err := json.Unmarshal([]byte(str), &out); err != nil {
		return fmt.Errorf("failed to unmarshal jsonnet: %w", err)
	}

	header := []byte(fmt.Sprintf(autoGenHeader, time.Now().Format(time.RFC3339)))
	var content []byte
	filename := searchMap(out, "name")
	if len(filename) > 0 {
		filename = fmt.Sprintf("%s/%s.yaml", g.ManifestsPath, filename)
	}

	for _, val := range out {
		body, err := yaml.Marshal(val)
		if err != nil {
			fmt.Printf("failed to marshal yaml: %v\n", err)
		}
		content = append(header, body...)
	}

	if err := os.WriteFile(filename, content, 0644); err != nil {
		fmt.Printf("failed to write file: %v\n", err)
	}
	return nil
}

func searchMap(m map[string]any, key string) string {
	for k, v := range m {
		if k == key {
			return v.(string)
		}
		if x, ok := v.(map[string]any); ok {
			return searchMap(x, key)
		}
	}
	return "default"
}

func OpenAndReadFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ReadToString(f)
}
func ReadToString(r io.Reader) (string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}