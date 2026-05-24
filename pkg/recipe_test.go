package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOpenRoastRecipe(t *testing.T) {
	path := filepath.Join("..", "testdata", "diedrich-style.openroast.json")
	recipe, err := LoadOpenRoastRecipe(path)
	if err != nil {
		t.Fatal(err)
	}
	if recipe.RoastName != "Diedrich Style (truncated)" {
		t.Fatalf("roastName = %q", recipe.RoastName)
	}
	if len(recipe.Steps) != 4 {
		t.Fatalf("steps = %d, want 4", len(recipe.Steps))
	}
	if recipe.Steps[0].FanSpeed != 9 || recipe.Steps[0].TargetTemp != 250 {
		t.Fatalf("first step = %+v", recipe.Steps[0])
	}
	if !recipe.Steps[3].Cooling {
		t.Fatal("expected cooling on last step")
	}
}

func TestLoadRecipeAutoDetect(t *testing.T) {
	openPath := filepath.Join("..", "testdata", "diedrich-style.openroast.json")
	r, err := LoadRecipe(openPath)
	if err != nil {
		t.Fatal(err)
	}
	if r.Format() != FormatOpenRoast {
		t.Fatalf("format = %v, want openroast", r.Format())
	}

	simplePath := filepath.Join("..", "sr700-fan-test.recipe.json")
	r, err = LoadRecipe(simplePath)
	if err != nil {
		t.Fatal(err)
	}
	if r.Format() != FormatSimple {
		t.Fatalf("format = %v, want simple", r.Format())
	}
}

func TestDetectFormatByExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.openroast.json")
	if err := os.WriteFile(path, []byte(`{"steps":[{"fanSpeed":5,"sectionTime":10}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	format, err := detectFormat(path, []byte(`{"steps":[{"fanSpeed":5,"sectionTime":10}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if format != FormatOpenRoast {
		t.Fatalf("format = %v", format)
	}
}

func TestOpenRoastStepTargetTempF(t *testing.T) {
	if got := (OpenRoastRecipeStep{}).TargetTempF(); got != 150 {
		t.Fatalf("default target = %d, want 150", got)
	}
	if got := (OpenRoastRecipeStep{TargetTemp: 300}).TargetTempF(); got != 300 {
		t.Fatalf("target = %d, want 300", got)
	}
}

func TestLoadRecipeEmptySteps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.recipe.json")
	if err := os.WriteFile(path, []byte(`{"steps":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadRecipe(path)
	if err == nil {
		t.Fatal("expected error for empty steps")
	}
}
