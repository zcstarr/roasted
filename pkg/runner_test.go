package pkg

import (
	"testing"
	"time"

	"github.com/zcstarr/roasted/sr700"
)

type roastCall struct {
	fan     sr700.Speed
	heat    sr700.Heat
	cooling bool
}

type fakeRoaster struct {
	calls    []roastCall
	temp     sr700.Temperature
	tempStep int
}

func (f *fakeRoaster) Connect() (*sr700.Program, error) {
	return &sr700.Program{}, nil
}

func (f *fakeRoaster) Roast(fan sr700.Speed, heat sr700.Heat, _ time.Duration) (sr700.Temperature, error) {
	f.calls = append(f.calls, roastCall{fan: fan, heat: heat})
	if heat == sr700.High {
		f.temp += sr700.Temperature(f.tempStep)
	}
	return f.temp, nil
}

func (f *fakeRoaster) Cool(fan sr700.Speed, _ time.Duration) (sr700.Temperature, error) {
	f.calls = append(f.calls, roastCall{fan: fan, cooling: true})
	return f.temp, nil
}

func (f *fakeRoaster) Stop() (sr700.Temperature, error) {
	return f.temp, nil
}

func TestRunSimpleRecipeExecutesSteps(t *testing.T) {
	simpleStepSleep = time.Millisecond
	t.Cleanup(func() { simpleStepSleep = time.Second })

	r := &SimpleRecipe{
		Steps: []SimpleRecipeStep{
			{Heat: sr700.High, Fan: 5, Duration: 1},
			{Fan: 9, Duration: 1, Cooling: true},
		},
	}
	roaster := &fakeRoaster{temp: 200, tempStep: 5}
	if err := RunSimpleRecipe(roaster, r); err != nil {
		t.Fatal(err)
	}
	if len(roaster.calls) < 2 {
		t.Fatalf("calls = %d, want at least 2", len(roaster.calls))
	}
	if roaster.calls[0].heat != sr700.High {
		t.Fatalf("first heat = %v", roaster.calls[0].heat)
	}
	if !roaster.calls[len(roaster.calls)-1].cooling {
		t.Fatal("expected final cooling call")
	}
}

func TestRunOpenRoastRecipeSectionsAndCooling(t *testing.T) {
	controlInterval = time.Millisecond
	t.Cleanup(func() { controlInterval = openRoastControlInterval })

	r := &OpenRoastRecipe{
		RoastName: "test",
		Steps: []OpenRoastRecipeStep{
			{TargetTemp: 250, FanSpeed: 7, SectionTime: 1},
			{FanSpeed: 9, SectionTime: 1, Cooling: true},
		},
	}
	roaster := &fakeRoaster{temp: 200, tempStep: 10}
	if err := RunOpenRoastRecipe(roaster, r); err != nil {
		t.Fatal(err)
	}

	var roastSteps, coolSteps int
	var sawHighHeat bool
	for _, c := range roaster.calls {
		if c.cooling {
			coolSteps++
		} else {
			roastSteps++
			if c.heat == sr700.High {
				sawHighHeat = true
			}
		}
	}
	if roastSteps == 0 {
		t.Fatal("expected roast calls in first section")
	}
	if coolSteps == 0 {
		t.Fatal("expected cool calls in cooling section")
	}
	if !sawHighHeat {
		t.Fatal("expected PID to turn heat on when below target")
	}
}

func TestRunOpenRoastRecipeUsesDefaultTargetTemp(t *testing.T) {
	controlInterval = time.Millisecond
	t.Cleanup(func() { controlInterval = openRoastControlInterval })

	r := &OpenRoastRecipe{
		Steps: []OpenRoastRecipeStep{
			{FanSpeed: 5, SectionTime: 1},
		},
	}
	roaster := &fakeRoaster{temp: 140, tempStep: 5}
	if err := RunOpenRoastRecipe(roaster, r); err != nil {
		t.Fatal(err)
	}
	if len(roaster.calls) == 0 {
		t.Fatal("expected roaster calls")
	}
}
