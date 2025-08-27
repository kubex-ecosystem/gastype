// Package info manages modular control and configuration, with support for files separated by module.
package info

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Control represents the control configuration of a module.
type Control struct {
	Reference     Reference `json:"-"` // Usado internamente para nome do arquivo, nunca exportado
	SchemaVersion int       `json:"schema_version"`
	IPC           IPC       `json:"ipc"`
	Bitreg        Bitreg    `json:"bitreg"`
	KV            KV        `json:"kv"`
	Seq           int       `json:"seq"`
	EpochNS       int64     `json:"epoch_ns"`
}

func (c *Control) GetName() string    { return c.Reference.Name }
func (c *Control) GetVersion() string { return c.Reference.Version }

// LoadControlByModule loads the control from a specific module file.
func LoadControlByModule(module string) (*Control, error) {
	file := filepath.Join(".", fmt.Sprintf("control_%s.json", module))
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", file, err)
	}
	defer f.Close()
	var c Control
	dec := json.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("error decoding %s: %w", file, err)
	}
	c.Reference = Reference{Name: module}
	return &c, nil
}

// SaveControl saves the module control to a separate file.
func (c *Control) SaveControl(dir string) error {
	if c.Reference.Name == "" {
		return fmt.Errorf("Reference.Name cannot be empty to save the control")
	}
	file := filepath.Join(dir, fmt.Sprintf("control_%s.json", c.Reference.Name))
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", file, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	// Reference não é exportado
	return enc.Encode(c)
}
