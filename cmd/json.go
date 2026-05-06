package cmd

import (
	"encoding/json"
	"io"
	"time"
)

const machineSchemaVersion = 1

var nowUTC = func() time.Time {
	return time.Now().UTC()
}

type machineMeta struct {
	SchemaVersion int      `json:"schema_version"`
	GeneratedAt   string   `json:"generated_at"`
	ProjectRoot   string   `json:"project_root"`
	Warnings      []string `json:"warnings"`
}

func newMachineMeta(projectRoot string) machineMeta {
	return machineMeta{
		SchemaVersion: machineSchemaVersion,
		GeneratedAt:   nowUTC().Format(time.RFC3339),
		ProjectRoot:   projectRoot,
		Warnings:      []string{},
	}
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
