package pkg

// Thermostat replicates Openroast's PID + bang-bang heat modulation for the SR700.
// Defaults match freshroastsr700: Kp=0.06, Ki=0.0075, Kd=0.01, 8 segments.

const (
	defaultPIDKp          = 0.06
	defaultPIDKi          = 0.0075
	defaultPIDKd          = 0.01
	defaultHeaterSegments = 8
)

// PID is a discrete PID controller ported from freshroastsr700/pid.py.
type PID struct {
	Kp, Ki, Kd    float64
	Derivator     float64
	Integrator    float64
	OutputMax     float64
	OutputMin     float64
	integratorMax float64
	integratorMin float64
}

// NewPID returns a PID controller with Openroast default tuning.
func NewPID(kp, ki, kd float64, outputMax int) *PID {
	p := &PID{
		Kp:        kp,
		Ki:        ki,
		Kd:        kd,
		OutputMax: float64(outputMax),
		OutputMin: 0,
	}
	if ki > 0 {
		p.integratorMax = p.OutputMax / ki
		p.integratorMin = p.OutputMin / ki
	}
	return p
}

// Update calculates PID output for the given current and target temperatures (°F).
func (p *PID) Update(currentTemp, targetTemp int) float64 {
	err := float64(targetTemp - currentTemp)

	pValue := p.Kp * err
	dValue := p.Kd * (p.Derivator - float64(currentTemp))
	p.Derivator = float64(currentTemp)

	p.Integrator += err
	if p.Integrator > p.integratorMax {
		p.Integrator = p.integratorMax
	} else if p.Integrator < p.integratorMin {
		p.Integrator = p.integratorMin
	}
	iValue := p.Integrator * p.Ki

	output := pValue + iValue + dValue
	if output > p.OutputMax {
		output = p.OutputMax
	}
	if output < p.OutputMin {
		output = p.OutputMin
	}
	return output
}

// Reset clears integrator and derivative state (e.g. at section boundaries).
func (p *PID) Reset() {
	p.Integrator = 0
	p.Derivator = 0
}

// HeatController pulse-modulates bang-bang heat output across N segments.
type HeatController struct {
	numSegments  int
	outputArray  [][]bool
	heatLevel    int
	heatLevelNow int
	currentIndex int
}

// NewHeatController returns an 8-segment controller matching freshroastsr700.
func NewHeatController(segments int) *HeatController {
	h := &HeatController{
		numSegments: segments,
		outputArray: make([][]bool, segments+1),
	}
	for i := range h.outputArray {
		h.outputArray[i] = make([]bool, segments)
	}

	switch segments {
	case 4:
		h.outputArray[0] = []bool{false, false, false, false}
		h.outputArray[1] = []bool{true, false, false, false}
		h.outputArray[2] = []bool{true, false, true, false}
		h.outputArray[3] = []bool{true, true, true, false}
		h.outputArray[4] = []bool{true, true, true, true}
	case 8:
		h.outputArray[0] = []bool{false, false, false, false, false, false, false, false}
		h.outputArray[1] = []bool{true, false, false, false, false, false, false, false}
		h.outputArray[2] = []bool{true, false, false, false, true, false, false, false}
		h.outputArray[3] = []bool{true, false, false, true, false, false, true, false}
		h.outputArray[4] = []bool{true, false, true, false, true, false, true, false}
		h.outputArray[5] = []bool{true, true, false, true, true, false, true, false}
		h.outputArray[6] = []bool{true, true, true, false, true, true, true, false}
		h.outputArray[7] = []bool{true, true, true, true, true, true, true, false}
		h.outputArray[8] = []bool{true, true, true, true, true, true, true, true}
	default:
		for i := 0; i <= segments; i++ {
			for j := 0; j < segments; j++ {
				h.outputArray[i][j] = j < i
			}
		}
	}
	return h
}

// SetHeatLevel sets the desired output level (0..numSegments inclusive).
func (h *HeatController) SetHeatLevel(value float64) {
	level := int(value + 0.5)
	if level < 0 {
		level = 0
	} else if level > h.numSegments {
		level = h.numSegments
	}
	h.heatLevel = level
}

// AboutToRollover reports whether the next bang-bang tick picks up a new heat level.
func (h *HeatController) AboutToRollover() bool {
	return h.currentIndex >= h.numSegments
}

// GenerateBangBangOutput returns the on/off pulse for this control tick and advances.
func (h *HeatController) GenerateBangBangOutput() bool {
	if h.currentIndex >= h.numSegments {
		h.heatLevelNow = h.heatLevel
		h.currentIndex = 0
	}
	out := h.outputArray[h.heatLevelNow][h.currentIndex]
	h.currentIndex++
	return out
}

// Thermostat combines PID and bang-bang modulation for SR700 roast sections.
type Thermostat struct {
	pid    *PID
	heater *HeatController
}

// NewThermostat returns a thermostat with Openroast default tuning.
func NewThermostat() *Thermostat {
	return &Thermostat{
		pid:    NewPID(defaultPIDKp, defaultPIDKi, defaultPIDKd, defaultHeaterSegments),
		heater: NewHeatController(defaultHeaterSegments),
	}
}

// Reset clears PID state at section boundaries.
func (t *Thermostat) Reset() {
	t.pid.Reset()
	t.heater.SetHeatLevel(0)
}

// Tick runs one control-loop iteration and returns whether heat should be on.
func (t *Thermostat) Tick(currentTemp int, tempValid bool, targetTemp int) bool {
	if t.heater.AboutToRollover() && tempValid {
		output := t.pid.Update(currentTemp, targetTemp)
		t.heater.SetHeatLevel(output)
	}
	return t.heater.GenerateBangBangOutput()
}
