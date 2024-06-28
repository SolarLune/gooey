# Gooey

Gooey is a pixel-focused immediate-mode GUI for games, written in Go for Ebitengine.

## Why the name

Next question

## How does it work?

Pretty easily:

```go

type Game struct {
    DebugDrawing bool
}

func (g *Game) Init() {

    gooey.Init(640, 360)  // Create GUI backbuffer
    gooey.SetFont(g.Font) // Set the font used for text rendering

}

func (g *Game) Draw(screen *ebiten.Image) {

    gooey.Clear()

	g.ExampleSplit()

	screen.DrawImage(gooey.Texture(), nil)

	if g.DebugDrawing {
		gooey.DrawDebug(screen)
	}

}

```