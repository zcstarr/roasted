// Package pkg contains recipe types and loaders for the SR700 roaster.
//
//go:generate go run ../tools/schemagen -src . -out ../schemas
package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zcstarr/roasted/sr700"
)

// Format identifies which recipe JSON schema a file uses.
type Format int

const (
	FormatSimple Format = iota
	FormatOpenRoast
)

func (f Format) String() string {
	switch f {
	case FormatOpenRoast:
		return "openroast"
	default:
		return "simple"
	}
}

// Recipe is a roast program loaded from JSON.
type Recipe interface {
	Format() Format
}

// OpenRoastRecipeStep is one step of an OpenRoast-format recipe.
type OpenRoastRecipeStep struct {
	// TargetTemp is the target bean temperature in degrees Fahrenheit.
	TargetTemp int `json:"targetTemp,omitempty" jsonschema:"minimum=150,maximum=550"`
	// FanSpeed selects the fan power (1 = lowest, 9 = highest).
	FanSpeed int `json:"fanSpeed" jsonschema:"minimum=1,maximum=9,enum=1,enum=2,enum=3,enum=4,enum=5,enum=6,enum=7,enum=8,enum=9"`
	// SectionTime is the duration of this step in seconds.
	SectionTime int `json:"sectionTime" jsonschema:"minimum=1"`
	// Cooling runs the roaster in cooling mode (fan only, no heat) for this step.
	Cooling bool `json:"cooling,omitempty"`
}

// OpenRoastBeanSource describes where a bean was purchased.
type OpenRoastBeanSource struct {
	Reseller string `json:"reseller,omitempty"`
	Link     string `json:"link,omitempty"`
}

// OpenRoastBean is the bean metadata block of an OpenRoast recipe.
type OpenRoastBean struct {
	Region  string              `json:"region,omitempty"`
	Source  OpenRoastBeanSource `json:"source,omitempty"`
	Country string              `json:"country,omitempty"`
}

// OpenRoastDescription is the roast description block of an OpenRoast recipe.
type OpenRoastDescription struct {
	// RoastType is a free-form label such as "Light", "Medium", or "Full City".
	RoastType string `json:"roastType,omitempty"`
	// Description is free-form notes about the roast.
	Description string `json:"description,omitempty"`
}

// OpenRoastRecipe is the recipe format used by the Openroast project.
type OpenRoastRecipe struct {
	// Creator is the author of the recipe.
	Creator string `json:"creator,omitempty"`
	// RoastName is a human-friendly name for the roast.
	RoastName string `json:"roastName,omitempty"`
	// Steps is the ordered list of roast steps to execute.
	Steps []OpenRoastRecipeStep `json:"steps" jsonschema:"minItems=1"`
	// Bean is metadata about the coffee bean being roasted.
	Bean OpenRoastBean `json:"bean,omitempty"`
	// TotalTime is the total roast duration in seconds (informational; the sum of step times is what is executed).
	TotalTime        int                  `json:"totalTime,omitempty" jsonschema:"minimum=0"`
	RoastDescription OpenRoastDescription `json:"roastDescription,omitempty"`
}

// SimpleRecipeStep is one segment of a simple roast program.
//
// Heat is ignored when Cooling is true; the roaster runs fan-only in cooling
// mode regardless of the heat setting.
type SimpleRecipeStep struct {
	// Heat selects the heating element power.
	// 0 = Cool, 1 = Low (~390F), 2 = Medium (~455F), 3 = High (~490F).
	Heat sr700.Heat `json:"heat" jsonschema:"minimum=0,maximum=3,enum=0,enum=1,enum=2,enum=3"`
	// Fan selects the fan power (1 = lowest, 9 = highest).
	Fan sr700.Speed `json:"fan" jsonschema:"minimum=1,maximum=9,enum=1,enum=2,enum=3,enum=4,enum=5,enum=6,enum=7,enum=8,enum=9"`
	// Duration is the length of this step in seconds.
	Duration int `json:"duration" jsonschema:"minimum=1"`
	// Cooling runs the roaster in cooling mode (fan only, no heat) for this step.
	Cooling bool `json:"cooling,omitempty"`
}

// SimpleRecipe is the recipe format consumed by the roasted CLI.
type SimpleRecipe struct {
	// Steps is the ordered list of roast steps to execute. Must contain at least one step.
	Steps []SimpleRecipeStep `json:"steps" jsonschema:"minItems=1"`
}

// Format reports the simple recipe schema.
func (SimpleRecipe) Format() Format { return FormatSimple }

// Format reports the OpenRoast recipe schema.
func (OpenRoastRecipe) Format() Format { return FormatOpenRoast }

// TargetTempF returns the PID setpoint for a step, defaulting to 150°F when omitted.
func (s OpenRoastRecipeStep) TargetTempF() int {
	if s.TargetTemp == 0 {
		return 150
	}
	return s.TargetTemp
}

// LoadSimpleRecipe reads and parses a SimpleRecipe JSON file from path.
func LoadSimpleRecipe(path string) (*SimpleRecipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var recipe SimpleRecipe
	if err := json.Unmarshal(data, &recipe); err != nil {
		return nil, err
	}
	return &recipe, nil
}

// LoadOpenRoastRecipe reads and parses an OpenRoastRecipe JSON file from path.
func LoadOpenRoastRecipe(path string) (*OpenRoastRecipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var recipe OpenRoastRecipe
	if err := json.Unmarshal(data, &recipe); err != nil {
		return nil, err
	}
	return &recipe, nil
}

// LoadRecipe reads a recipe file and auto-detects simple vs OpenRoast format.
func LoadRecipe(path string) (Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	format, err := detectFormat(path, data)
	if err != nil {
		return nil, err
	}
	switch format {
	case FormatOpenRoast:
		var recipe OpenRoastRecipe
		if err := json.Unmarshal(data, &recipe); err != nil {
			return nil, err
		}
		if len(recipe.Steps) == 0 {
			return nil, fmt.Errorf("recipe has no steps")
		}
		return &recipe, nil
	default:
		var recipe SimpleRecipe
		if err := json.Unmarshal(data, &recipe); err != nil {
			return nil, err
		}
		if len(recipe.Steps) == 0 {
			return nil, fmt.Errorf("recipe has no steps")
		}
		return &recipe, nil
	}
}

func detectFormat(path string, data []byte) (Format, error) {
	if strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".openroast.json") {
		return FormatOpenRoast, nil
	}

	var peek struct {
		Steps []json.RawMessage `json:"steps"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return 0, fmt.Errorf("parse recipe: %w", err)
	}
	if len(peek.Steps) == 0 {
		return 0, fmt.Errorf("recipe has no steps")
	}

	var firstStep struct {
		FanSpeed json.RawMessage `json:"fanSpeed"`
		Heat     json.RawMessage `json:"heat"`
	}
	if err := json.Unmarshal(peek.Steps[0], &firstStep); err != nil {
		return 0, fmt.Errorf("parse first step: %w", err)
	}
	if len(firstStep.FanSpeed) > 0 {
		return FormatOpenRoast, nil
	}
	if len(firstStep.Heat) > 0 {
		return FormatSimple, nil
	}
	return FormatSimple, nil
}
