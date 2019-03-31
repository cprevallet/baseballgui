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
	"github.com/cprevallet/baseballgui/trajectory"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"strconv"

        "github.com/faiface/pixel"
        "github.com/faiface/pixel/imdraw"
        "github.com/faiface/pixel/pixelgl"
        "golang.org/x/image/colornames"

)

var history []trajectory.TrajectoryPoint

func main() {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.Add(windowWidget())
	win.ShowAll()

	gtk.Main()
}

func windowWidget() *gtk.Widget {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	inputs := []*gtk.Entry{}
	//labels := []*gtk.Entry{}

	for i := 0; i < 3; i++ {
		entry, err := gtk.EntryNew()
		if err != nil {
			log.Fatal("Unable to create entry:", err)
		}
		inputs = append(inputs, entry)
		grid.Add(inputs[i])
		inputs[i].SetHExpand(true)
	}

	calcbtn, err := gtk.ButtonNewWithLabel("Calculate")
	grid.Add(calcbtn)

	calcbtn.Connect("clicked", func() {
		// pull the arguments out of the widgets as float64
		var args []float64
		for i := 0; i < 3; i++ {
			v, _ := inputs[i].GetText()
			if a, err := strconv.ParseFloat(v, 64); err == nil {
				args = append(args, a)
			}
		}
		if len(args) == 3 {
			doCalc(args[0], args[1], args[2])
		}
	})

	return &grid.Container.Widget
}

func doCalc(initialAltitude float64, initialAngle float64, initialVelocity float64) {
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
        pixelgl.Run(run)
}

func run() {
        cfg := pixelgl.WindowConfig{
                Title:  "Baseball trajectory visualization.",
                Bounds: pixel.R(0, 0, 1024, 768),
                VSync:  true,
        }
        win, err := pixelgl.NewWindow(cfg)
        if err != nil {
                panic(err)
        }

        imd := imdraw.New(nil)
        inc := 0

        for !win.Closed() {
                if inc > len(history)-1 {
                        inc = 0
                }
                win.Clear(colornames.Blue)
                imd.Clear()
                imd.Color = colornames.Limegreen
                imd.Push(pixel.V(history[inc].Position[0], history[inc].Position[1]))
                imd.Circle(5, 0)
                imd.Draw(win)
                win.Update()
                inc++
        }
}

