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
	"fmt"
        "math"
        
        "image"
	"os"
        _ "image/png"

	"github.com/cprevallet/baseballgui/trajectory"
        "github.com/faiface/pixel"
        "github.com/faiface/pixel/imdraw"
        "github.com/faiface/pixel/pixelgl"
        "golang.org/x/image/colornames"

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

func doCalc(initialAltitude float64, initialAngle float64, initialVelocity float64) (history []trajectory.TrajectoryPoint) {
	//fmt.Println("Baseball trajectory from Public Domain Aeronautical Software. Go version")
	var dt = 0.1          // time step
	var normalized = true //make initial altitude the reference point in results
	var k = 0
	// Instructions for use of this program: put your values for the three
	//  initial conditions on the next three lines. Recompile. Run bb.
	// Hints: Denver=1609  Mexico City=2420  La Paz=3650  Everest=8850
	/*
	   initialAltitude := 1609.0 // meters
	   initialAngle := 40.0      // degrees from horizontal
	   initialVelocity := 35.0   // m/s
	*/

	//fmt.Println("      t           x           y          vx           vy          ax         ay")
	history = trajectory.Trajectory(initialAltitude, initialVelocity,
		initialAngle, dt, normalized)
	for k = 0; k < len(history); k++ {
		fmt.Printf("%9.2f %11.2f %11.2f %11.2f %11.2f %11.2f %11.2f \n",
			history[k].Time,
			history[k].Position[0],
			history[k].Position[1],
			history[k].Velocity[0],
			history[k].Velocity[1],
			history[k].Acceleration[0],
			history[k].Acceleration[1])
	}
	fmt.Println("End of Baseball")
        return history
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
        inc := 0
	
        var Altitude float64 = 1609.0 //meters
        var Angle float64 = 40.0  // degrees from horizontal
        //var Velocity float64 = 35.0 // m/s
        // this isn't historically accurate, but useful to keep it within the screen resolution
        var Velocity float64 = 100.0 // m/s 
        
	trj := doCalc(Altitude, Angle, Velocity)
        var newtrj []trajectory.TrajectoryPoint

        pic, err := loadPicture("cannonball.png")
	if err != nil {
		panic(err)
        }
        cannonball := pixel.NewSprite(pic, pic.Bounds())
        pic2, err := loadPicture("cannon.png")
	if err != nil {
		panic(err)
        }
        cannon := pixel.NewSprite(pic2, pic2.Bounds())

        for !win.Closed() {
                win.Clear(colornames.Blue)
                imd.Clear()

                // Draw a cannonball sprite
        	mat := pixel.IM
        	mat = mat.Scaled(pixel.ZV, 0.1 )
        	mat = mat.Moved(pixel.V(trj[inc].Position[0], trj[inc].Position[1]))
                cannonball.Draw(win, mat)

                // Draw a cannon sprite
        	mat = pixel.IM
        	mat = mat.Scaled(pixel.ZV, 0.2 )
        	mat = mat.Rotated(pixel.ZV, (Angle-35.0)*math.Pi/180.0)
                cannon.Draw(win, mat)

                // Draw power graph
                launcherX := 40.0
                launcherY := 40.0
                power := Velocity / 5.0
                imd.Color = colornames.Red
                offset := 10.0
                imd.Push(pixel.V(launcherX + offset, launcherY + offset))
                imd.Push(pixel.V(launcherX + offset + power, launcherY + offset))
                imd.Line(1.0)
                // Draw the trajectory
                imd.Draw(win)
                win.Update()
                inc++
                if inc > len(trj)-1 {
                        inc = 0
                        if newtrj != nil {
                           trj = nil
                           trj = newtrj
                           newtrj = nil
                        }

                }
                // Accept keyboard input and update the trajectory.
                if win.Pressed(pixelgl.KeyRight) {
			Velocity += 1.0
	                newtrj = doCalc(Altitude, Angle, Velocity)
		}

                if win.Pressed(pixelgl.KeyLeft) {
			Velocity -= 1.0
	                newtrj = doCalc(Altitude, Angle, Velocity)
		}

                if win.Pressed(pixelgl.KeyUp) {
			Angle += 1.0
	                newtrj = doCalc(Altitude, Angle, Velocity)
		}

                if win.Pressed(pixelgl.KeyDown) {
			Angle -= 1.0
	                newtrj = doCalc(Altitude, Angle, Velocity)
		}

        }
}


func main() {
        pixelgl.Run(run)
}

