# Gooey

[![Go Reference](https://pkg.go.dev/badge/github.com/solarlune/gooey.svg)](https://pkg.go.dev/github.com/solarlune/gooey)

![buttonList](https://github.com/user-attachments/assets/88f3ea78-8dbb-437a-a522-47fcc6c8907f)

![buttonGrid](https://github.com/user-attachments/assets/e4381b7a-af91-42c0-b8f8-72346197d03c)

![buttonCircle](https://github.com/user-attachments/assets/6f3ad9d0-3ab6-4434-b134-14fe24227c1c)

**Gooey** is a pixel-focused immediate mode-ish GUI framework for game development, written in Go for Ebitengine.

## Why the name?

Next question.

## Why did you make Gooey?

There's a few UI solutions for Ebitengine out there, but I felt like they were a bit overly complicated or over-generic. I wanted something made for fixed resolution games that's easy-to-use and supports keyboards / gamepads well. This is what I came up with.

## How does it work?

Essentially, you:

1. Call `gooey.Begin()`. Pass an argument indicating settings, including what inputs to check for UI traversal.
1. Create `gooey.Layout`s. `Layouts` primarily contain UI elements in a set rectangle. You can give `Layout`s `Arranger`s to customize how and where UI elements draw to the screen.
1. Create `gooey` UI elements and add them to `Layout`s using unique IDs. The IDs are how the UI elements' internal state is stored.
1. Draw the GUI texture (`gooey.Texture()`) to the screen once finished.
1. Call `gooey.End()`.

That's it.

## Can you give an example?

Sure:

```go

package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/gooey"
)

type Game struct {}

func NewGame() *Game {

	// Initialize Gooey and create the GUI backbuffer image
	gooey.Init(640, 360)

	return &Game{}

}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {

	// Call `gooey.Begin()` at the beginning of the game frame.
	gooey.Begin(gooey.UpdateSettings{
		// Indicate what input to use for the "accept" / "OK" button.
		AcceptInput: ebiten.IsKeyPressed(ebiten.KeyX),
	})

	// Define a new Layout named "root" positioned at 0, 0 with a size of 640x360.
	area := gooey.NewLayout("root", 0, 0, 640, 360)

	// Create a new button with some sane default values,
	// and then set its graphics to a flat color.
	button := gooey.NewUIButton().WithGraphics(
		gooey.UIColor{Color: gooey.NewColor(1, 0, 0, 1)},
	)

	// UIButton.AddTo() returns a boolean indicating if the button is pressed.
	if button.AddTo(area, "testButton") {
		fmt.Println("The Test button was pressed.")
	}

	// You can also just create the button struct manually and fill its properties.

	// Draw the result.
	screen.Clear()
	screen.DrawImage(gooey.Texture(), nil)

	// At the end of each frame, finish using gooey.
	gooey.End()

	// And that's about it.

}

func (g *Game) Layout(w, h int) (int, int) { return 640, 360 }

func main() {
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}


```

You can also create UI element structs as bases and use their `With____` functions to make one-off adjustments to them. This creates a simple templating system to work with, like so:

```go

labelBase := gooey.UILabel{
	Alignment: gooey.AlignmentCenterLeft,
	LineSpacing: 1.25,
}.WithPadding(16)

// Make a copy of `labelBase` with its text set to "First", and then add it to a layout.
labelBase.WithText("First").AddTo(layout, "first label")

// Same thing with "Second" for the text.
labelBase.WithText("Second").AddTo(layout, "second label")

// Same thing with "Third" for the text.
labelBase.WithText("Third").AddTo(layout, "third label")

```

You can also use UI elements' `Apply()` functions to make many changes at once. Each non-zero property is copied from the passed UIElement object. This can be easier to read (though this creates an interim struct in the process).

```go

labelBase := gooey.UILabel{
	Alignment: gooey.AlignmentCenterLeft,
	LineSpacing: 1.25,
}.WithPadding(16)

overridden := labelBase.Apply(gooey.UILabel{
	Text: "Hello!\nHello!",
	LineSpacing: 1.9,
	PaddingTop: 8,
	PaddingBottom: 8,
})

// overridden:
//
// Text = "Hello\nHello!"
// LineSpacing = 1.9
// PaddingTop = 8
// PaddingBottom = 8
// PaddingLeft = 16
// PaddingRight = 16
// Alignment = gooey.AlignmentCenterLeft
```

See the `examples` folder for more examples of more complex concepts.

# What's in, what's not

Here's what's currently implemented and what has yet to be done:

- **UI Elements**
    - [x] Flat Color
    - [x] Button (clickable / pressable)
    - [x] Cyclical Button (selectable out of a set of options)
    - [x] Collection (draw multiple UI elements in a single space)
    - [x] Slider
    - [x] Image
    - [x] Text Label
        - [x] Typewriter effect
        - [ ] Editable labels
    - [x] Custom Draw Element
    - [x] Apply system to copy non-zero values to UI element structs
    - [ ] Radio buttons
        - [x] Button groups (similar to radio buttons, only a certain number can be toggled at a time)
    - [ ] Dropdown menu
- **Layout System**
    - [x] Layout modifier functions for overriding specific UI elements
    - [x] Layouts allow different methods of positioning and scaling UI elements
    - [x] Grid-based element arrangement system
    - [x] Custom element arrangement system
- **Highlighting system**
    - [x] Keyboard / gamepad / input-based highlighting
    - [x] Mouse input
    - [x] Switching between mouse and input-based highlighting (press an input to switch to input, click to switch to mouse)
    - [x] Hold input to repeat
    - [x] Custom highlighting system
        - [ ] More refinement here, maybe?
    - [x] Layout highlighting system to control which layouts can receive focus at any given time (e.g. you might have multiple layouts represent multiple menus that you walk through. Think of an RPG with an inventory. You might have a menu with different options like Items, Equipment, Key Items, etc. at the left, and then after making that selection, a larger list of items that you scroll through. You should have to select a menu option to view the items under that categorization, and so this would require different "levels" of highlighting.)
- **Scrolling system**
    - [x] Smooth linear automatic scrolling. When highlighting them, Gooey will scroll Layouts to them.
    - [ ] Fix scrolling to be more reliable / smoother
    - [ ] Scrollbars
    - [ ] Add custom scrolling
- **Debugging System**
    - [x] Debug display of Layouts
    - [ ] Debug display of elements drawn to Layouts
