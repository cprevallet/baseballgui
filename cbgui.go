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
        "time"
	"os"

	"github.com/cprevallet/baseballgui/trajectory"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

        "github.com/aquilax/go-perlin"
)

var speedFactor = 8.0 // Increasing this speeds up the trajectory display.

// A Projectile is a sprite that has associated trajectory physics.
type Projectile struct {
        trj     trajectory.TrajectoryPoint // what's our position, velocity, and acceleration in time
        rect    pixel.Rect                  // on screen position
        spr     *pixel.Sprite               // drawable frame of a Picture
}
type Target struct {
        Pos     pixel.Vec
        Spr     *pixel.Sprite               // drawable frame of a Picture
        Mat     pixel.Matrix                // linear transformation for movement, rotation, etc.
}

const (
	width  = 1024
	height = 768
	// must be around the 2/3 of the screen height
	verticalOffset = height * 2 / 3
	// Perlin noise provides variations in values between -1 and 1,
	// we multiply those so they're visible on screen
	scale            = height/2  //adjust to taste
	waveLength       = width/4   //adjust to taste
	alpha            = 2.
	beta             = 2.
	n                = 3
	maximumSeedValue = 300
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

// FireProjectile provides starting values for a projectile.
func (p *Projectile) fireProjectile(
                initialAltitude float64, // meters 
                initialAngle float64,    // degrees from horizontal
                initialVelocity float64, // m/s
                pic pixel.Picture,      // sprite image filename
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
        return 
}

// UpdateProjectile computes a trajectory, performing numerical solution of a set of
// ordinary differential equations with a fixed time step.
func (p *Projectile) updateProjectile(dt float64) {
        // Update the matrix to move the sprite on the screen.
        newTrajectory := trajectory.UpdateRK4(p.trj, dt)
        // What's the change?
        newVec := pixel.V(newTrajectory.Position[0] - p.trj.Position[0],
                newTrajectory.Position[1] - p.trj.Position[1])
        // Update the moved matrix.
        p.trj = newTrajectory
        p.rect = p.rect.Moved(newVec)
}


// InitTarget provides starting values for a target.
func initTarget(
                pic pixel.Picture,      // sprite image filename
                newVec pixel.Vec,       // coordinate to move to
        ) (target Target) {

        //  Create the drawable sprite 
        sprite := pixel.NewSprite(pic, pic.Bounds())

        // Start with the identity matrix and scale based on the picture.
        mat := pixel.IM
        mat = mat.Scaled(pixel.ZV, 0.2)
        mat = mat.Moved(newVec)
        target = Target{newVec, sprite, mat}
        return target
}

// UpdateTarget moves the target around in a random walk. 
func updateTarget(tar *Target, newPos pixel.Vec) {
        // What's the change?
        // Update the moved matrix.
        newVec := pixel.V(newPos.X-tar.Pos.X, newPos.Y-tar.Pos.Y)
        tar.Pos.X = newPos.X
        tar.Pos.Y = newPos.Y
        //fmt.Println(tar.Pos)
        tar.Mat = tar.Mat.Moved(newVec)
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
	// inc := 0

	var Altitude float64 = 0.0 //meters
	var Angle float64 = 40.0      // degrees from horizontal
	//var Velocity float64 = 35.0 // m/s
	// this isn't historically accurate, but useful to keep it within the screen resolution
	var Velocity float64 = 100.0 // m/s

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
        targ := initTarget(pic3, win.Bounds().Center())

        last := time.Now() //time of the start of the previous frame

        var inFlight []*Projectile

        // Setup Perlin noise for the path of the balloon.
        var seed = rand.Int63n(maximumSeedValue)
        p := perlin.NewPerlin(alpha, beta, n, seed)
        var xpos float64 = 0.


	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		win.Clear(colornames.Blue)
		imd.Clear()

                // Update the projectile trajectories and draw the sprite.
                var keepProj []*Projectile
                for i, _ := range inFlight {
                    inFlight[i].updateProjectile(dt*speedFactor)
                    inFlight[i].spr.Draw(win, pixel.IM.
                                    Scaled(pixel.ZV, 0.1).
                                    Moved(inFlight[i].rect.Center()),
                    )
                    if inFlight[i].trj.Position[1] > 0.0 { keepProj = append(keepProj, inFlight[i]) }
                }
                // Remove elements that have left the screen.
                inFlight = nil
                inFlight = keepProj
                keepProj = nil


		// Draw a cannon sprite
                mat := pixel.IM
		mat = mat.Scaled(pixel.ZV, 0.2)
		mat = mat.Rotated(pixel.ZV, (Angle-35.0)*math.Pi/180.0)
		cannon.Draw(win, mat)

                // Update and draw a target
                xpos++
                if xpos > width {xpos = 0 }
                position := pixel.V(xpos, p.Noise1D(xpos/waveLength)*scale + verticalOffset)
                updateTarget(&targ, position)
                targ.Spr.Draw(win, targ.Mat)

		// Draw power graph
		launcherX := 40.0
		launcherY := 40.0
		power := Velocity / 5.0
		imd.Color = colornames.Red
		offset := 10.0
		imd.Push(pixel.V(launcherX+offset, launcherY+offset))
		imd.Push(pixel.V(launcherX+offset+power, launcherY+offset))
		imd.Line(1.0)
		// Draw the trajectory
		imd.Draw(win)
		win.Update()

		// Accept keyboard input and calculate a new trajectory.
                if win.JustPressed(pixelgl.MouseButtonLeft) {
                    // Initialize our cannonball.
                    cball := &Projectile{}
                    cball.fireProjectile(
                            Altitude,
                            Angle,
                            Velocity,
                            pic)
                    inFlight = append(inFlight,cball)
		}

		if win.Pressed(pixelgl.KeyRight) {
			Velocity += 1.0
		}

		if win.Pressed(pixelgl.KeyLeft) {
			Velocity -= 1.0
		}

		if win.Pressed(pixelgl.KeyUp) {
			Angle += 1.0
		}

		if win.Pressed(pixelgl.KeyDown) {
			Angle -= 1.0
		}

	}
}


func main() {
	pixelgl.Run(run)
}
