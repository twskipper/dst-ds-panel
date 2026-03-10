package dst

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dst-ds-panel/internal/model"
)

func WriteModOverrides(shardDir string, mods []model.Mod) error {
	var entries []string
	for _, mod := range mods {
		opts := serializeConfigOptions(mod.Config)
		entries = append(entries, fmt.Sprintf(
			`  ["workshop-%s"] = { enabled = %t, configuration_options = { %s } }`,
			mod.WorkshopID, mod.Enabled, opts,
		))
	}

	content := "return {\n" + strings.Join(entries, ",\n") + "\n}\n"
	return os.WriteFile(filepath.Join(shardDir, "modoverrides.lua"), []byte(content), 0644)
}

func serializeConfigOptions(config map[string]any) string {
	if len(config) == 0 {
		return ""
	}
	var parts []string
	for k, v := range config {
		parts = append(parts, fmt.Sprintf(`["%s"] = %s`, k, luaValue(v)))
	}
	return strings.Join(parts, ", ")
}

func luaValue(v any) string {
	switch val := v.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	case string:
		return fmt.Sprintf("%q", val)
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", v))
	}
}

func WriteModsSetup(clusterDir string, mods []model.Mod) error {
	var lines []string
	for _, mod := range mods {
		if mod.Enabled {
			lines = append(lines, fmt.Sprintf(`ServerModSetup("%s")`, mod.WorkshopID))
		}
	}
	return os.WriteFile(filepath.Join(clusterDir, "mods_setup.lua"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// Match workshop mod entries - handles both formats:
//   ["workshop-123"] = { enabled = true, ... }
//   ["workshop-123"] = { configuration_options={...}, enabled=true }
var workshopIDRegex = regexp.MustCompile(`\["workshop-(\d+)"\]`)

// Find enabled=true or enabled=false within a mod block
var enabledRegex = regexp.MustCompile(`enabled\s*=\s*(true|false)`)

func ReadModOverrides(shardDir string) ([]model.Mod, error) {
	data, err := os.ReadFile(filepath.Join(shardDir, "modoverrides.lua"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	content := string(data)
	var mods []model.Mod

	// Find all workshop mod IDs and their positions
	idMatches := workshopIDRegex.FindAllStringSubmatchIndex(content, -1)
	for i, idMatch := range idMatches {
		workshopID := content[idMatch[2]:idMatch[3]]

		// Determine the block for this mod entry (from this match to the next one, or end)
		blockStart := idMatch[1]
		blockEnd := len(content)
		if i+1 < len(idMatches) {
			blockEnd = idMatches[i+1][0]
		}
		block := content[blockStart:blockEnd]

		// Find enabled status in this block
		enabled := true // default to enabled if not found
		if em := enabledRegex.FindStringSubmatch(block); em != nil {
			enabled = em[1] == "true"
		}

		mods = append(mods, model.Mod{
			WorkshopID: workshopID,
			Name:       "Workshop-" + workshopID,
			Enabled:    enabled,
			Config:     map[string]any{},
		})
	}

	return mods, nil
}
