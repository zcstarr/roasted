package pkg

import (
	"log"
	"time"

	"github.com/zcstarr/roasted/sr700"
)

const (
	openRoastControlInterval = 250 * time.Millisecond
	roasterCommandDuration   = 10 * time.Second
)

// controlInterval is the OpenRoast PID loop period (tests may shorten this).
var controlInterval = openRoastControlInterval

// simpleStepSleep is the wall-clock sleep between simple recipe control ticks (tests may shorten).
var simpleStepSleep = time.Second

// Roaster drives an SR700 during recipe execution.
type Roaster interface {
	Connect() (*sr700.Program, error)
	Roast(fan sr700.Speed, heat sr700.Heat, duration time.Duration) (sr700.Temperature, error)
	Cool(fan sr700.Speed, duration time.Duration) (sr700.Temperature, error)
	Stop() (sr700.Temperature, error)
}

// RunSimpleRecipe executes a time-based simple recipe on roaster.
func RunSimpleRecipe(roaster Roaster, recipe *SimpleRecipe) error {
	for i, step := range recipe.Steps {
		secondsRemaining := step.Duration
		log.Printf("= Step %d - Heat: %v Fan: %v Cooling: %v Duration: %v", i, step.Heat, step.Fan, step.Cooling, step.Duration)

		for secondsRemaining > 0 {
			temp, err := runSimpleTick(roaster, step)
			logTemp(temp, secondsRemaining, err)

			sleep := min(secondsRemaining, 5)
			time.Sleep(simpleStepSleep * time.Duration(sleep))
			secondsRemaining -= sleep
		}
	}
	_, err := roaster.Stop()
	return err
}

func runSimpleTick(roaster Roaster, step SimpleRecipeStep) (sr700.Temperature, error) {
	if step.Cooling {
		return roaster.Cool(step.Fan, roasterCommandDuration)
	}
	return roaster.Roast(step.Fan, step.Heat, roasterCommandDuration)
}

// RunOpenRoastRecipe executes an OpenRoast recipe with PID thermostat control.
func RunOpenRoastRecipe(roaster Roaster, recipe *OpenRoastRecipe) error {
	if recipe.RoastName != "" {
		log.Printf("Recipe: %s", recipe.RoastName)
	}
	if recipe.Creator != "" {
		log.Printf("Creator: %s", recipe.Creator)
	}

	thermostat := NewThermostat()

	for i, step := range recipe.Steps {
		targetTemp := step.TargetTempF()
		fan := sr700.Speed(step.FanSpeed)
		sectionEnd := time.Now().Add(time.Duration(step.SectionTime) * time.Second)

		log.Printf("= Step %d - Target: %d F Fan: %v Cooling: %v SectionTime: %d s",
			i, targetTemp, fan, step.Cooling, step.SectionTime)

		thermostat.Reset()

		var (
			lastTemp      int
			lastTempValid bool
		)

		for time.Now().Before(sectionEnd) {
			secondsRemaining := int(time.Until(sectionEnd).Seconds()) + 1
			if secondsRemaining > step.SectionTime {
				secondsRemaining = step.SectionTime
			}

			var (
				temp sr700.Temperature
				err  error
			)
			if step.Cooling {
				temp, err = roaster.Cool(fan, roasterCommandDuration)
			} else {
				heatOn := true
				if lastTempValid {
					heatOn = thermostat.Tick(lastTemp, true, targetTemp)
				}
				heat := sr700.Cool
				if heatOn {
					heat = sr700.High
				}
				temp, err = roaster.Roast(fan, heat, roasterCommandDuration)
			}
			if temp.Valid() {
				lastTemp = tempF(temp)
				lastTempValid = true
			}
			logTemp(temp, secondsRemaining, err)

			sleep := controlInterval
			if remaining := time.Until(sectionEnd); remaining < sleep {
				sleep = remaining
			}
			if sleep > 0 {
				time.Sleep(sleep)
			}
		}
	}

	_, err := roaster.Stop()
	return err
}

func tempF(t sr700.Temperature) int {
	return int(t)
}

func logTemp(temp sr700.Temperature, secondsRemaining int, err error) {
	if err != nil {
		log.Println("error! ", err)
		return
	}
	if temp == sr700.TemperatureBelow150F {
		log.Printf("<- -- F / %d s", secondsRemaining)
		return
	}
	log.Printf("<- %v F / %d s", temp, secondsRemaining)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
