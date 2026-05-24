package pkg

import "testing"

func TestPIDUpdateIncreasesWhenBelowTarget(t *testing.T) {
	pid := NewPID(defaultPIDKp, defaultPIDKi, defaultPIDKd, defaultHeaterSegments)
	out := pid.Update(200, 250)
	if out <= 0 {
		t.Fatalf("output = %v, want > 0 when below target", out)
	}
}

func TestPIDUpdateDecreasesWhenAboveTarget(t *testing.T) {
	pid := NewPID(defaultPIDKp, defaultPIDKi, defaultPIDKd, defaultHeaterSegments)
	pid.Update(300, 250)
	pid.Update(310, 250)
	out := pid.Update(320, 250)
	if out >= float64(defaultHeaterSegments) {
		t.Fatalf("output = %v, want reduced when above target", out)
	}
}

func TestHeatControllerBangBangDutyCycle(t *testing.T) {
	h := NewHeatController(8)
	h.SetHeatLevel(4)

	// First PWM cycle uses initial level 0; skip it to observe commanded level.
	for i := 0; i < 8; i++ {
		h.GenerateBangBangOutput()
	}

	var onCount int
	for i := 0; i < 8; i++ {
		if h.GenerateBangBangOutput() {
			onCount++
		}
	}
	if onCount != 4 {
		t.Fatalf("on pulses = %d, want 4 for level 4/8", onCount)
	}
}

func TestThermostatTickProducesOutput(t *testing.T) {
	ts := NewThermostat()
	var sawOn bool
	for i := 0; i < 32; i++ {
		if ts.Tick(200, true, 250) {
			sawOn = true
			break
		}
	}
	if !sawOn {
		t.Fatal("expected heat on when well below target")
	}
}

func TestThermostatResetClearsPID(t *testing.T) {
	ts := NewThermostat()
	ts.Tick(200, true, 250)
	ts.Reset()
	if ts.pid.Integrator != 0 {
		t.Fatalf("integrator = %v after reset", ts.pid.Integrator)
	}
}
