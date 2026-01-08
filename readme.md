# Gooey

Gooey is a pixel-focused immediate-ish GUI for games, written in Go for Ebitengine.

## Why the name?

Next question.

## Why did you make Gooey?

There's a few UI solutions for Ebitengine out there, but I felt like they were a bit overly complicated or over-generic. I wanted something made for fixed resolution games that's easy-to-use and supports keyboards / gamepads well. This is what I came up with.

## How does it work?

Essentially, you:

1. Call `gooey.Begin()`. Pass an argument indicating settings, including what inputs to check for UI traversal.
1. Create a `gooey.Layout`. `Layouts` control where and how UI elements are rendered in a set rectangle. You can have many `Layouts` and can use a `Layouter` to customize how those elements draw to the screen.
1. Create `gooey` UI elements and add them to the `Layout` using unique IDs. The IDs are how the UI elements' internal state is stored.
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
		AcceptInput: ebiten.IsKeyPressed(ebiten.KeyX), // Indicate what input to use for the "accept" / "OK" button.
	})

	// At the end of each frame, finish using `gooey`.
	defer gooey.End()

	// Define an area named "root" with its X, Y, W, and H.
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

	// Draw the result.
	screen.Clear()
	screen.DrawImage(gooey.Texture(), nil)

	// And that's about it.

}

func (g *Game) Layout(w, h int) (int, int) { return 640, 360 }

func main() {
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}


```

You can also reuse UI element structs and `With____` functions to make templates to work with, like so:

```go

labelBase := gooey.UILabel{
	Anchor: gooey.AnchorCenterLeft,
	LineSpacing: 1.25,
}.WithPadding(16)

///

// Make a copy of `labelBase` with its text set to "First", and then add it to a layout.
labelBase.WithText("First").AddTo(layout, "first label")

labelBase.WithText("Second").AddTo(layout, "second label")

labelBase.WithText("Third").AddTo(layout, "third label")

```

See the `examples` folder for examples of more complex concepts.

# What's in, what's not

Here's what's currently implemented and what has yet to be done:

-   **UI Elements**
    -   [x] Flat Color
    -   [x] Button (clickable / pressable)
    -   [x] Cyclical Button (selectable out of a set of options)
    -   [x] Collection (draw multiple UI elements in a single space)
    -   [x] Slider
    -   [x] Image
    -   [x] Text Label
        -   [x] Typewriter effect
        -   [ ] Editable labels
    -   [x] Custom Draw Element
    -   [ ] Radio buttons
    -   [ ] Dropdown menu
-   **Layout System**
    -   [x] Default full layout system
    -   [x] Layout modifier functions for UI elements
    -   [x] Grid-based layout system
    -   [x] Custom layout system
-   **Highlighting system**
    -   [x] Keyboard / gamepad / input-based highlighting
    -   [x] Mouse input
    -   [x] Switching between mouse and input-based highlighting (press an input to switch to input, click to switch to mouse)
    -   [x] Hold input to repeat
    -   [x] Custom highlighting system
    -   [ ] A more refined custom highlighting system
    -   [ ] Layout highlighting system to control which layouts can receive focus at any given time (e.g. you might have multiple layouts to represent multiple menus that you walk through; think of an RPG with an inventory. You might have a menu with different options Items, Equipment, Key Items, etc. at the left, then a larger list of items; you should have to select a menu option to view the items under that categorization.)
-   **Scrolling system**
    -   [x] Smooth linear automatic scrolling. When highlighting them, Gooey will scroll Layouts to them.
    -   [ ] Fix scrolling to be more reliable / smoother
    -   [ ] Scrollbars
    -   [ ] Add custom scrolling
