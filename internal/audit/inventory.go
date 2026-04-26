package audit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	CategoryConfig   = "config"
	CategoryWorkflow = "workflow"
	CategoryRuntime  = "runtime"
	CategoryOther    = "other"
)

var (
	DefaultTargets = []string{".claude", ".codex", ".sango", ".hwt"}
	runtimeDirs    = []string{
		".sango/logs",
		".sango/pids",
		".sango/bare",
		".sango/work",
		".sango/template-cache",
		".sango/locks",
	}
)

type Options struct {
	Targets        []string
	IncludeRuntime bool
}

type Item struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Category string `json:"category"`
}

type Report struct {
	Root        string         `json:"root"`
	GeneratedAt time.Time      `json:"generated_at"`
	Targets     []string       `json:"targets"`
	Excluded    []string       `json:"excluded"`
	Items       []Item         `json:"items"`
	Summary     map[string]int `json:"summary"`
	Warnings    []string       `json:"warnings"`
}

func CollectInventory(root string, opts Options) (Report, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}

	targets := opts.Targets
	if len(targets) == 0 {
		targets = append([]string{}, DefaultTargets...)
	}

	report := Report{
		Root:        filepath.Clean(absRoot),
		GeneratedAt: time.Now().UTC(),
		Targets:     append([]string{}, targets...),
		Excluded:    []string{},
		Items:       []Item{},
		Summary: map[string]int{
			CategoryConfig:   0,
			CategoryWorkflow: 0,
			CategoryRuntime:  0,
			CategoryOther:    0,
		},
		Warnings: []string{},
	}
	if !opts.IncludeRuntime {
		report.Excluded = append(report.Excluded, runtimeDirs...)
	}

	for _, target := range targets {
		targetPath := filepath.Join(absRoot, target)
		stat, statErr := os.Stat(targetPath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				report.Warnings = append(report.Warnings, fmt.Sprintf("missing target: %s", target))
				continue
			}
			return report, statErr
		}

		if !stat.IsDir() {
			rel, relErr := filepath.Rel(absRoot, targetPath)
			if relErr != nil {
				return report, relErr
			}
			addItem(&report, filepath.ToSlash(rel), "file")
			continue
		}

		walkErr := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, relErr := filepath.Rel(absRoot, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)

			if !opts.IncludeRuntime && isRuntimePath(rel) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			kind := "file"
			if d.IsDir() {
				kind = "dir"
			}
			addItem(&report, rel, kind)
			return nil
		})
		if walkErr != nil {
			return report, walkErr
		}
	}

	sort.Slice(report.Items, func(i, j int) bool {
		return report.Items[i].Path < report.Items[j].Path
	})

	return report, nil
}

func RenderText(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "inventory root: %s\n", report.Root)
	fmt.Fprintf(&b, "targets: %s\n", strings.Join(report.Targets, ", "))
	if len(report.Excluded) > 0 {
		fmt.Fprintf(&b, "excluded: %s\n", strings.Join(report.Excluded, ", "))
	}
	fmt.Fprintf(&b, "items: %d\n", len(report.Items))
	fmt.Fprintf(&b, "- %s: %d\n", CategoryConfig, report.Summary[CategoryConfig])
	fmt.Fprintf(&b, "- %s: %d\n", CategoryWorkflow, report.Summary[CategoryWorkflow])
	fmt.Fprintf(&b, "- %s: %d\n", CategoryRuntime, report.Summary[CategoryRuntime])
	fmt.Fprintf(&b, "- %s: %d\n", CategoryOther, report.Summary[CategoryOther])
	for _, item := range report.Items {
		fmt.Fprintf(&b, "[%s] %s (%s)\n", item.Category, item.Path, item.Kind)
	}
	return strings.TrimSpace(b.String())
}

func addItem(report *Report, relPath, kind string) {
	category := classify(relPath)
	report.Items = append(report.Items, Item{
		Path:     relPath,
		Kind:     kind,
		Category: category,
	})
	report.Summary[category]++
}

func classify(relPath string) string {
	switch {
	case relPath == ".sango", strings.HasPrefix(relPath, ".sango/"):
		if isRuntimePath(relPath) {
			return CategoryRuntime
		}
		return CategoryConfig
	case relPath == ".codex", strings.HasPrefix(relPath, ".codex/"):
		return CategoryWorkflow
	case relPath == ".claude", strings.HasPrefix(relPath, ".claude/"):
		return CategoryWorkflow
	case relPath == ".hwt", strings.HasPrefix(relPath, ".hwt/"):
		return CategoryWorkflow
	default:
		return CategoryOther
	}
}

func isRuntimePath(relPath string) bool {
	clean := filepath.ToSlash(filepath.Clean(relPath))
	for _, dir := range runtimeDirs {
		if clean == dir || strings.HasPrefix(clean, dir+"/") {
			return true
		}
	}
	return false
}
