// Package trajectory calculates the trajectory of a baseball with air drag.
// Trajectory is the only publicly visible function guaranteeing an explicit interface.
package trajectory

import (
	//"fmt"
	"math"
)

type TrajectoryPoint struct {
	Time         float64
	Position     [2]float64
	Velocity     [2]float64
	Acceleration [2]float64
}

// Package conversion factor constants:
var ft2meters = 0.3048  // convert feet to meters
var g = 9.8066          // acceleration of gravity, m/s**2
var lbs2kg = 0.45359237 // weight in pounds to mass in kg.
var rhozero = 1.2250    // density of air at sealevel, kg/cu.m

// Baseball constants - by the rules, the circumference of a baseball must
// be no less than 9 inches and no more than 9.25 inches. The weight of a
// baseball must be no less than 5.00 ounces and no more than 5.25 ounces.
// I assume my ideal ball lies exactly in the middle.
// In SI units:
/*
var diam = (9.125 / (12 * math.Pi)) * ft2meters // diameter of a baseball (m)
var mass = (5.125 / 16) * lbs2kg                // mass of a baseball (kg)
var sref = 0.25 * math.Pi * diam * diam         // frontal area (sq.m)
*/
// British cannonballs in SI units.
// https://hypertextbook.com/facts/2009/JenniferChung.shtml
var diam =  4.95 / 12 * ft2meters               // diameter of a cannonball (m)
var mass = 5.4                                  // mass of a cannonball (kg)
var sref = 0.25 * math.Pi * diam * diam         // frontal area (sq.m)

// accel computes the acceleration (vector) for a spherical projectile
// moving through a viscous medium. Assume Mach number is small enough
// that wave drag may be neglected. Ignore added mass term.
// NOTE - position has units of meters, but first argument to simpleAtmosphere
//   is in kilometers. Be sure to remember to multiply by 0.001
func accel(time float64, position [2]float64, velocity [2]float64) (acceleration [2]float64) {
	vertical := [2]float64{0.0, 1.0}
	var drag, unitVelocity [2]float64
	vsq := 0.0
	for _, v := range velocity {
		vsq += math.Pow(v, 2.0)
	}
	vmag := math.Sqrt(vsq)
	for i, v := range velocity {
		unitVelocity[i] = v / vmag
	}
	// unitVelocity=velocity/vmag
	sigma, _, theta := simpleAtmosphere(0.001 * position[1])
	// first arg is altitude in kilometers
	density := sigma * rhozero
        // q represents the dynamic pressure
	q := 0.5 * density * vsq
	reynolds := density * vmag * diam / viscosity(theta)
	cd := cdSphere(reynolds)
	dragMagnitude := cd * q * sref
	for i, _ := range acceleration {
		drag[i] = -dragMagnitude * unitVelocity[i]
		acceleration[i] = drag[i]/mass - g*vertical[i]
	}
	return acceleration
}

// cdSphere computes the drag coefficient of a sphere as a function of
//  Reynolds number. Assumes Mach number is small.
//  Taken from Chow, "Computational Aerodynamics"
//  r = Reynolds number
//  d = drag coefficient based on cross-section area
func cdSphere(r float64) (d float64) {
	if r <= 0 {
		d = 0
	} else if r <= 1.0 {
		d = 24.0 / r
	} else if r <= 400.0 {
		d = 24 * math.Pow(r, -0.646)
	} else if r <= 3E5 {
		d = 0.5
	} else if r <= 2E6 {
		d = 3.66E-4 * math.Pow(r, 0.4275)
	} else {
		d = 0.18
	}
	return d
}

// correctFinalPosition insures final point of a trajectory so that its y-coordinate
// is exactly equal to initial altitude. Assume that the altitudes of a1
// and a2 straddle initialAltitude.
func correctFinalPosition(initialAltitude float64, a1 TrajectoryPoint, a2 TrajectoryPoint)(corrected TrajectoryPoint) {
	fraction := (initialAltitude - a1.Position[1]) / (a2.Position[1] - a1.Position[1])
	corrected.Time = a1.Time + fraction*(a2.Time-a1.Time)
	for i := 0; i < 2; i++ {
		corrected.Position[i] = a1.Position[i] + fraction*(a2.Position[i]-a1.Position[i])
		corrected.Velocity[i] = a1.Velocity[i] + fraction*(a2.Velocity[i]-a1.Velocity[i])
		corrected.Acceleration[i] = a1.Acceleration[i] + 
                  fraction*(a2.Acceleration[i]-a1.Acceleration[i])
	}
	return
}

// viscosity computes air viscosity using Sutherland's formula.
// Returns viscosity in kg/(meter-sec)
// theta = temperature/sea-level temperature
func viscosity(theta float64) (visc float64) {
	var betavisc = 1.458E-6 // viscosity term, N sec/(sq.m sqrt(deg K)
	var suth = 110.4        // Sutherland's constant, deg K
	var tzero = 288.15      // temperature at sealevel, deg K
        temp := tzero * theta   // temp = temperature in deg Kelvin
	visc = betavisc * math.Sqrt(temp*temp*temp) / (temp + suth)
	return visc
}

// simpleAtmosphere computes the characteristics of the lower atmosphere.
// NOTES-Correct to 20 km. Only approximate above there
//   alt    = geometric altitude, km.
//   sigma  = density/sea-level standard density
//   delta  = pressure/sea-level standard pressure
//   theta  = temperature/sea-level standard temperature
func simpleAtmosphere(alt float64) (sigma float64, delta float64, theta float64) {
	rearth := 6369.0                   // radius of the Earth (km)
	gmr := 34.163195                   // gas constant
	h := alt * rearth / (alt + rearth) // convert geometric to geopotential altitude
	if h < 11.0 {
		theta = 1.0 + (-6.5/288.15)*h // Troposphere
		delta = math.Pow(theta, gmr/6.5)
	} else {
		theta = 216.65 / 288.15 // Stratosphere
		delta = 0.2233611 * math.Exp(-gmr*(h-11.0)/216.65)
	}
	sigma = delta / theta
	return
}

//  baseballKutta advances one time-like step in a trajectory. This is a system
//  of four first order ordinary differential equations. Use fourth-order
//  Runge-Kutta equation to advance one time step.
//  p1 = current position
//  h  = step to be taken in time
//  p2 = next position
//  Ref:
//  https://en.wikipedia.org/wiki/Runge%E2%80%93Kutta_methods#The_Runge%E2%80%93Kutta_method
func baseballKutta(p1 TrajectoryPoint, h float64) (p2 TrajectoryPoint) {
	var dx1, dx2, dx3, dx4 [2]float64
	var dv1, dv2, dv3, dv4 [2]float64

	//start of interval
	t := p1.Time
	x := p1.Position
	v := p1.Velocity
	a := accel(t, x, v)
	for i := 0; i < 2; i++ {
		dx1[i] = h * v[i]
		dv1[i] = h * a[i]
	}

	//midpoint of interval w/ slopes dx1/dv1
	var x2, v2 [2]float64
	for i := 0; i < 2; i++ {
		x2[i] = x[i] + dx1[i]/2.0
		v2[i] = v[i] + dv1[i]/2.0
	}
	a = accel(t+h/2.0, x2, v2)
	for i := 0; i < 2; i++ {
		dx2[i] = h * (v[i] + dv1[i]/2.0)
		dv2[i] = h * a[i]
	}

	//midpoint of interval w/ slopes dx2/dv2
	var x3, v3 [2]float64
	for i := 0; i < 2; i++ {
		x3[i] = x[i] + dx2[i]/2.0
		v3[i] = v[i] + dv2[i]/2.0
	}
	a = accel(t+h/2.0, x3, v3)
	for i := 0; i < 2; i++ {
		dx3[i] = h * (v[i] + dv2[i]/2.0)
		dv3[i] = h * a[i]
	}

	// endpoint of inteval
	var x4, v4 [2]float64
	for i := 0; i < 2; i++ {
		x4[i] = x[i] + dx3[i]
		v4[i] = v[i] + dv3[i]
	}
	a = accel(t+h, x4, v4)
	for i := 0; i < 2; i++ {
		dx4[i] = h * (v[i] + dv3[i])
		dv4[i] = h * a[i]
	}

	p2.Time = t + h
	for i := 0; i < 2; i++ {
		p2.Position[i] = p1.Position[i] + (dx1[i]+dx2[i]+dx2[i]+dx3[i]+dx3[i]+dx4[i])/6.0
		p2.Velocity[i] = p1.Velocity[i] + (dv1[i]+dv2[i]+dv2[i]+dv3[i]+dv3[i]+dv4[i])/6.0
	}
	p2.Acceleration = accel(p2.Time, p2.Position, p2.Velocity)
	return
}

// Trajectory computes a trajectory, performing numerical solution of a set of
// ordinary differential equations with a fixed time step. Halt the
// calculation when the altitude is less than the initial altitude and
// correct the final point to have the same altitude as the initial altitude.
// initialAltitude = meters
// initialTheta    = degrees from horizontal
// initialVelocity = m/s
// normalized      = make the output positions relative ref: initalAltitude = 0
func Trajectory(initialAltitude float64, initialVelocity float64, initialTheta float64, dt float64, normalized bool) (history []TrajectoryPoint) {
        // Initialize vectors.
	t := 0.0
	position := [2]float64{0.0, initialAltitude}
	velocity := [2]float64{initialVelocity * math.Cos(initialTheta*math.Pi/180.0),
		initialVelocity * math.Sin(initialTheta*math.Pi/180.0)}
	acceleration := accel(t, position, velocity)
	initialTrajectory := TrajectoryPoint{Time: t, Position: position,
            Velocity: velocity, Acceleration: acceleration}
	history = append(history, initialTrajectory)
        // Perform the Runge-Kutta.
	k := 0
	cond := true
	for ok := true; ok; ok = cond {
		newTrajectory := baseballKutta(history[k], dt)
		k++
		history = append(history, newTrajectory)
		cond = (newTrajectory.Position[1] > initialAltitude)
	}
        // Interpolate the last element at the initial altitude.
	corrected := correctFinalPosition(initialAltitude,
                           history[len(history)-2], history[len(history)-1])
	//Replace the last element.
	history = history[:len(history)-1]
	history = append(history, corrected)
	if normalized {
		for i := len(history) - 1; i > -1; i-- {
			history[i].Position[1] = history[i].Position[1] - history[0].Position[1]
		}
	}
	return history
}
