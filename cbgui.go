/*
 * Copyright (c) 2013-2014 Conformal Systems <info@conformal.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
//	"fmt"
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/cprevallet/baseballgui/trajectory"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/aquilax/go-perlin"
)

var speedFactor = 8.0   // Increasing this speeds up the trajectory display.
var xPctofScreen = 15.0 //what pct of the horizontal scale of the screen should the target use?
var yPctofScreen = 15.0 //what pct of the vertical scale of the screen should the target use?

// A Projectile is a sprite that has associated trajectory physics.
type Projectile struct {
	trj  trajectory.TrajectoryPoint // what's our position, velocity, and acceleration in time
	rect pixel.Rect                 // on screen position
	spr  *pixel.Sprite              // drawable frame of a picture
}

// A Target is a sprite that has associated perlin-based physics.
type Target struct {
	//Pos     pixel.Vec
	perlinx float64       // perlin horizontal coordinate index, left = -width/2, right = width/2
	rect    pixel.Rect    // on screen position
	spr     *pixel.Sprite // drawable frame of a picture
}

const (
	width  = 1024
	height = 768
	// must be around the 2/3 of the screen height
	verticalOffset = height * 2 / 3
	// Perlin noise provides variations in values between -1 and 1,
	// we multiply those so they're visible on screen
	scale            = height / 2 //adjust to taste
	waveLength       = width / 4  //adjust to taste
	alpha            = 2.
	beta             = 2.
	n                = 3
	maximumSeedValue = 300
        // inc = update perlin, controls speed of target
        inc = 6.0
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// FireProjectile provides starting values for a projectile with atmospheric drag.
func (p *Projectile) fireProjectile(
	initialAltitude float64, // meters
	initialAngle float64, // degrees from horizontal
	initialVelocity float64, // m/s
	pic pixel.Picture, // sprite image filename
) {

	//  Create the initial trajectory based on the angle of the object projecting it.
	position := [2]float64{0.0, initialAltitude}
	velocity := [2]float64{initialVelocity * math.Cos(initialAngle*math.Pi/180.0),
		initialVelocity * math.Sin(initialAngle*math.Pi/180.0)}
	acceleration := trajectory.Accel(0.0, position, velocity)
	p.trj = trajectory.TrajectoryPoint{Time: 0.0, Position: position,
		Velocity: velocity, Acceleration: acceleration}

	//  Create the drawable sprite
	p.spr = pixel.NewSprite(pic, pic.Bounds())
	p.rect = pixel.R(-1, -1, 1, 1)
        return
}

// UpdateTrajectory computes a trajectory, performing numerical solution of a set of
// ordinary differential equations with a fixed time step.
func (p *Projectile) updateTrajectory(dt float64) {
	// Update the trajectory.
	newTrajectory := trajectory.UpdateRK4(p.trj, dt)
	// What's the change in position?
	newVec := pixel.V(newTrajectory.Position[0]-p.trj.Position[0],
		newTrajectory.Position[1]-p.trj.Position[1])
	p.trj = newTrajectory
	// Update the rectangle so the sprite can redraw itself at
	// the new screen position.
	p.rect = p.rect.Moved(newVec)
}

// UpdateTarget moves the target to the new screen position.
func (t *Target) updateTarget(newPos pixel.Vec) {
	t.rect = t.rect.Moved(newPos.Sub(t.rect.Center()))
}

// Detect collision checks if a projectile has hit by rectangle positions.
func (t *Target) detectCollision(p []*Projectile) (hit bool) {
        hit = false
	//Intersect function requires normalized values
	tNormed := t.rect.Norm()
	//fmt.Println("...")
	for _, prj := range p {
		pNormed := prj.rect.Norm()
		if tNormed.Intersect(pNormed) != pixel.R(0, 0, 0, 0) {
			//fmt.Println("Hit!!!")
                        hit = true
		}
	}
        return hit
}

func run() {
	// One-time initialization section
	cfg := pixelgl.WindowConfig{
		Title:  "Cannonball trajectory visualization.",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(nil)

	var Altitude float64 = 0.0 //meters
	var Angle float64 = 90.0   // degrees from horizontal
	//var Velocity float64 = 35.0 // m/s
	// this isn't historically accurate, but useful to keep it within the screen resolution
	var Velocity float64 = 90.0 // m/s

	pic, err := loadPicture("cannonball.png")
	if err != nil {
		panic(err)
	}

	pic2, err := loadPicture("cannon.png")
	if err != nil {
		panic(err)
	}
	cannon := pixel.NewSprite(pic2, pic2.Bounds())

	pic3, err := loadPicture("haballoon.png")
	if err != nil {
		panic(err)
	}

	last := time.Now() //time of the start of the previous frame

	var Projs []*Projectile // a list of inflight projectiles
	var Targs []*Target     // a list of onscreen targets

	// Setup Perlin noise for the path of the balloon.
	var seed = rand.Int63n(maximumSeedValue)
	p := perlin.NewPerlin(alpha, beta, n, seed)

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		win.Clear(colornames.Blue)
		imd.Clear()

		// Update the projectile trajectory physics and draw the sprite.
		var keepProj []*Projectile
		for i, _ := range Projs {
			Projs[i].updateTrajectory(dt * speedFactor)
			Projs[i].spr.Draw(win, pixel.IM.
				Scaled(pixel.ZV, 0.1).
                                Moved(pixel.V(width/2,0)).
				Moved(Projs[i].rect.Center()),
			)
			if Projs[i].trj.Position[1] > 0.0 {
				keepProj = append(keepProj, Projs[i])
			}
		}
		// Remove elements that have left the screen.
		Projs = nil
		Projs = keepProj
		keepProj = nil

		// Draw a cannon sprite
		cannon.Draw(win, pixel.IM.
		        Scaled(pixel.ZV, 0.2).
        		Rotated(pixel.ZV, (Angle-35.0)*math.Pi/180.0).
                        Moved(pixel.V(width/2, 0)),
                        )

		// Create targets to shoot at.
		if Targs == nil {
			xPixels := xPctofScreen / 100.0 * pic3.Bounds().W()
			yPixels := yPctofScreen / 100.0 * pic3.Bounds().H()
			targ := &Target{
				spr:     pixel.NewSprite(pic3, pic3.Bounds()),
				rect:    pixel.R(-xPixels/2, -yPixels/2, xPixels/2, yPixels/2),
				perlinx: -width/2,
			}
			Targs = append(Targs, targ)
		}

		// Draw a target with motion using Perlin noise. Check for collisions.
		var keepTarg []*Target
		for _, targ := range Targs {
			targ.perlinx+= inc
			if targ.perlinx > width/2 {
				targ.perlinx = -width/2
			}
			position := pixel.V(targ.perlinx, p.Noise1D(targ.perlinx/waveLength)*scale+verticalOffset)
			targ.updateTarget(position)
			targ.spr.Draw(win, pixel.IM.
				ScaledXY(pixel.ZV, pixel.V(xPctofScreen/100.0, yPctofScreen/100.0)).
                                Moved(pixel.V(width/2,0)).
				Moved(targ.rect.Center()),
			)
			hit := targ.detectCollision(Projs)
                        if hit == false {keepTarg = append(keepTarg, targ)}
		}
		// Remove elements that have been hit.
		Targs = nil
		Targs = keepTarg
		keepTarg = nil

		// Draw power graph above the cannon.
		launcherX := 40.0
		launcherY := 40.0
		power := Velocity / 5.0
		imd.Color = colornames.Red
		offset := 10.0
		imd.Push(pixel.V(launcherX+offset, launcherY+offset))
		imd.Push(pixel.V(launcherX+offset+power, launcherY+offset))
		imd.Line(1.0)
		imd.Draw(win)

		// Display
		win.Update()

		// Accept keyboard input and calculate a new trajectory.
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			// Fire a new cannonball.
			cball := &Projectile{}
			cball.fireProjectile(
				Altitude,
				Angle,
				Velocity,
				pic)
			Projs = append(Projs, cball)
		}

		if win.Pressed(pixelgl.KeyRight) {
			Angle -= 1.0
		}

		if win.Pressed(pixelgl.KeyLeft) {
			Angle += 1.0
		}

		if win.Pressed(pixelgl.KeyUp) {
			Velocity += 1.0
		}

		if win.Pressed(pixelgl.KeyDown) {
			Velocity -= 1.0
		}

	}
}

func main() {
	pixelgl.Run(run)
}
