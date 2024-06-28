# Gooey

Gooey is a pixel-focused immediate-mode GUI for games, written in Go for Ebitengine.

## Why the name

Next question.

## How does it work?

Pretty easily:

```go

type Game struct {}

func (g *Game) Init() {
    gooey.Init(640, 360)  // Create GUI backbuffer
}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {

    gooey.Clear() // Clear the GUI backbuffer.

	a := gooey.NewArea("root", 0, 0, 640, 360) // Define an area, with a given ID and X, Y, W, and H.

	// Draw and check to see if a button was pressed.
	if a.UIButton("testButton", gooey.ButtonOptions{Label: "Test."}) { 
		fmt.Println("The Test button was pressed.")
	}

	// Draw the result.
	screen.DrawImage(gooey.Texture(), nil)

}

func (g *Game) Layout(w, h int) (int, int) { return 640, 360 }

func main() {
	if err := n.RunGame(&Game{}); err != nil { panic(err) }
}

```