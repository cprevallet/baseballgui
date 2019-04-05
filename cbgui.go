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
        "time"
	"os"
        "log"

        "runtime/pprof"
        "flag"

	"github.com/cprevallet/baseballgui/trajectory"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var timefactor = 10.0 //projectile update time is a multiple of the frame rate, depends on cpu speed

type Projectile struct {
        Trj      trajectory.TrajectoryPoint // what's our position, velocity, and acceleration in time
        Spr     *pixel.Sprite               // drawable frame of a Picture
}

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

// InitProjectile provides starting values for a projectile.
func initProjectile(
                initialAltitude float64, // meters 
                initialAngle float64,    // degrees from horizontal
                initialVelocity float64, // m/s
                pic pixel.Picture,      // sprite image filename
        ) (projectile Projectile) {

        //  Create the initial trajectory based on the angle of the object projecting it.
        position := [2]float64{0.0, initialAltitude}
        velocity := [2]float64{initialVelocity * math.Cos(initialAngle*math.Pi/180.0),
                initialVelocity * math.Sin(initialAngle*math.Pi/180.0)}
        acceleration := trajectory.Accel(0.0, position, velocity)
        trj := trajectory.TrajectoryPoint{Time: 0.0, Position: position,
            Velocity: velocity, Acceleration: acceleration}

        //  Create the drawable sprite 
        sprite := pixel.NewSprite(pic, pic.Bounds())
        projectile = Projectile{trj,sprite}
        return projectile
}

// UpdateProjectile computes a trajectory, performing numerical solution of a set of
// ordinary differential equations with a fixed time step.
func updateProjectile(prj *Projectile, dt float64) {
        newTrajectory := trajectory.UpdateRK4(prj.Trj, dt)
        prj.Trj = newTrajectory
}

func run() {
	// One-time initialization section
	cfg := pixelgl.WindowConfig{
		Title:  "Cannonball trajectory visualization.",
		Bounds: pixel.R(0, 0, 1024, 768),
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
	var Velocity float64 = 55.0 // m/s

	// var trj []trajectory.TrajectoryPoint
	// var newtrj []trajectory.TrajectoryPoint

	pic, err := loadPicture("cannonball.png")
	if err != nil {
		panic(err)
	}
	// cannonball := pixel.NewSprite(pic, pic.Bounds())

        // Initialize our cannonball.
        cball := initProjectile(
                Altitude,
                Angle,
                Velocity,
                pic)

	pic2, err := loadPicture("cannon.png")
	if err != nil {
		panic(err)
	}
	cannon := pixel.NewSprite(pic2, pic2.Bounds())

        last := time.Now() //time of the start of the previous frame

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		win.Clear(colornames.Blue)
		imd.Clear()

		// Draw a cannonball sprite
		mat := pixel.IM
                /*
		if (trj != nil) && (inc < len(trj)) {
			cannonball.Draw(win, mat)
		}
                */
                updateProjectile(&cball, dt*timefactor)
                mat = mat.Scaled(pixel.ZV, 0.1)
                //fmt.Println(cball.Trj.Position)
                mat = mat.Moved(pixel.V(cball.Trj.Position[0], cball.Trj.Position[1]))
                cball.Spr.Draw(win, mat)

		// Draw a cannon sprite
		mat = pixel.IM
		mat = mat.Scaled(pixel.ZV, 0.2)
		mat = mat.Rotated(pixel.ZV, (Angle-35.0)*math.Pi/180.0)
		cannon.Draw(win, mat)

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
		// Update the current position of the cannonball.
                /*
		if (trj != nil) && (inc < len(trj)) {
			inc++
		}
		if (trj != nil) && (inc == len(trj)) {
			trj = nil
			newtrj = nil
			inc = 0
		}
		if (newtrj != nil) && (trj == nil) {
			trj = newtrj
		}
                */

		// Accept keyboard input and calculate a new trajectory.

		if win.Pressed(pixelgl.KeySpace) {
                        /*
			newtrj = doCalc(Altitude, Angle, Velocity)
                        */

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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
	pixelgl.Run(run)
}
