// Package pkg contains recipe types and loaders for the SR700 roaster.
//
//go:generate go run ../tools/schemagen -src . -out ../schemas
package pkg

import (
	"encoding/json"
	"os"

	"github.com/zcstarr/roasted/sr700"
)

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

// LoadSimpleRecipe reads and parses a SimpleRecipe JSON file from path.
func LoadSimpleRecipe(path string) (*SimpleRecipe, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var recipe SimpleRecipe
	err = json.NewDecoder(f).Decode(&recipe)
	return &recipe, err
}
