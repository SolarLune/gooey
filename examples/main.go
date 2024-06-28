package main

import (
	"embed"
	"fmt"
	"image"
	"math"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/solarlune/gooey"
)

//go:embed *.png *.ttf
var assets embed.FS

var MultiplyBlendMode = ebiten.Blend{
	BlendFactorSourceRGB:        ebiten.BlendFactorDestinationColor,
	BlendFactorSourceAlpha:      ebiten.BlendFactorDestinationColor,
	BlendFactorDestinationRGB:   ebiten.BlendFactorZero,
	BlendFactorDestinationAlpha: ebiten.BlendFactorZero,
	BlendOperationRGB:           ebiten.BlendOperationAdd,
	BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

type Game struct {
	Frame float64
	Font  text.Face

	GUIImg         *ebiten.Image
	BGPattern      *ebiten.Image
	WalkingMonster *ebiten.Image

	DebugDrawing bool
}

func NewGame() *Game {
	ebiten.SetWindowTitle("Gooey Example")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	g := &Game{}

	g.GUIImg = g.LoadImage("gui.png")
	g.BGPattern = g.LoadImage("rocktexture.png")
	g.WalkingMonster = g.LoadImage("walkingMonster.png")

	fontFile, err := assets.Open("Greenscr.ttf")
	if err != nil {
		panic(err)
	}
	face, err := text.NewGoTextFaceSource(fontFile)
	if err != nil {
		panic(err)
	}

	g.Font = &text.GoTextFace{
		Source: face,
		Size:   12,
	}

	gooey.Init(640, 360) // Create backbuffer
	style := gooey.NewDefaultTextStyle()
	style.Font = g.Font
	style.OutlineColor = gooey.NewColor(0, 0, 0, 1)
	style.FGColor = gooey.NewColor(0.1, 0.15, 0.2, 1)
	gooey.SetTextStyle(style)
	// gooey.SetFont(g.Font) // Set the font used for text

	return g
}

func (g *Game) LoadImage(path string) *ebiten.Image {
	image, _, err := ebitenutil.NewImageFromFileSystem(assets, path)
	if err != nil {
		panic(err)
	}
	return image
}

func (g *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		gooey.Reset() //
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.DebugDrawing = !g.DebugDrawing
	}

	// Update the highlight controls - this allows us to navigate the UI
	// using keyboard presses.
	gooey.HighlightControlUpdate(gooey.HighlightControlSettings{
		RightInput:  inpututil.IsKeyJustPressed(ebiten.KeyRight),
		LeftInput:   inpututil.IsKeyJustPressed(ebiten.KeyLeft),
		UpInput:     inpututil.IsKeyJustPressed(ebiten.KeyUp),
		DownInput:   inpututil.IsKeyJustPressed(ebiten.KeyDown),
		AcceptInput: inpututil.IsKeyJustPressed(ebiten.KeyX),
	})

	g.Frame += 1.0 / 60.0

	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	gooey.Clear()

	// Draw a pattern behind the screen
	opt := &ebiten.DrawImageOptions{}
	opt.ColorScale.Scale(0.2, 0.2, 0.2, 1)
	screen.DrawImage(g.BGPattern, opt)
	opt.GeoM.Translate(512, 0)
	screen.DrawImage(g.BGPattern, opt)

	// Perform an example
	g.ExampleButtonsAndTextbox()

	screen.DrawImage(gooey.Texture(), nil)

	if g.DebugDrawing {
		gooey.DrawDebug(screen)
	}

}

func (g *Game) ExampleSimple() {

	// Define an area for the GUI.
	a := gooey.NewArea("root", 0, 0, 640, 360)

	// Check to see if a button was pressed.
	if a.UIButton("testButton", gooey.ButtonOptions{Label: "Test."}) {
		fmt.Println("The Test button was pressed.")
	}
}

func (g *Game) ExampleMovingButtons() {

	a := gooey.NewArea("root", 0, 0, 640, 360)

	for i := 0; i < 10; i++ {
		// We can change the offset to change where UI elements are created
		a.Offset.X = float32(math.Sin((g.Frame+(float64(i)*0.1))*math.Pi)) * 100
		buttonName := fmt.Sprintf("Test button %d", i)
		a.UIButton(buttonName, gooey.ButtonOptions{
			Label:    buttonName,
			PaddingY: 8,
			PaddingX: 8,
		})
	}

}

var textStr = "This is a text box on the right-hand side, with a list of buttons on the left."

func (g *Game) ExampleSplit() {

	a := gooey.NewArea("root", 0, 0, 640, 360)

	// Split is just an easy way to split an area into two; you could just manually create two Areas.
	left, right := a.Split(0.4, true)

	// left.ExpandMode = gooey.ExpandFill

	// You can easily create a ButtonOptions struct to share properties and then
	// use modifier functions to set individual properties
	buttonOpt := gooey.ButtonOptions{PaddingY: 8}

	if left.UIButton("A", buttonOpt.WithLabel("Button A")) {
		textStr = "You pressed the first button."
	}
	if left.UIButton("B", buttonOpt.WithLabel("Button B")) {
		textStr = "You pressed the second button, Button B - good work."
	}
	if left.UIButton("C", buttonOpt.WithLabel("Button C")) {
		textStr = "Wow, you pressed Button 3! Incredible!!!!"
	}

	right.UITextbox(gooey.TextboxOptions{
		Text:    textStr,
		Padding: 16,
		Height:  360 / 2,
	})

}

func (g *Game) ExampleGrid() {

	a := gooey.NewArea("root", 0, 0, 300, 100)

	grid := a.UIGrid("grid", 3, 3, 0, 0, 300, 300)

	for i, c := range grid.Cells {
		c.LayoutFill() // Set the layout to fill up the cell
		c.UIButton(fmt.Sprintf("Button %d", i), gooey.ButtonOptions{
			Label: fmt.Sprintf("Button %d", i),
		})
	}

}

var spin = false

func (g *Game) ExampleImage() {

	a := gooey.NewArea("an area", 0, 0, 200, 250)

	a.UITextbox(gooey.TextboxOptions{
		Padding: 8,
		Text:    "Here's an image test. You can change its size and the 'space' will react accordingly, unless you use an Area.Layout* function to fix the size.",
	})

	opt := &ebiten.DrawImageOptions{}
	bounds := g.WalkingMonster.Bounds()

	// Move it to the center, scale, rotate, and then move it back
	opt.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)
	opt.GeoM.Scale(2, 2)
	if spin {
		opt.GeoM.Rotate(g.Frame * math.Pi * 5)
	}
	opt.GeoM.Translate(float64(bounds.Dx())/2, float64(bounds.Dy())/2)

	// opt.GeoM.Translate(math.Sin(g.Frame*math.Pi)*32, math.Cos(g.Frame*math.Pi)*32)

	a.UIImage(gooey.ImageOptions{
		Image:            g.WalkingMonster,
		DrawImageOptions: opt,
	})

	if a.UIButton("spin", gooey.ButtonOptions{
		Label:    "Toggle Spin",
		PaddingY: 16,
	}) {
		spin = !spin
	}

}

func (g *Game) ExampleSubAreas() {

	a := gooey.NewArea("root", 0, 0, 300, 100)

	a.UIButton("Area 1", gooey.ButtonOptions{Label: "Outside of area"})
	a.UIButton("Area 2", gooey.ButtonOptions{Label: "Outside of area 2"})

	// Create another area that exists "within" the owning area
	sub := a.UIArea("subarea", 40, 0, 100, 100)
	for i := 0; i < 10; i++ {
		sub.UIButton(i, gooey.ButtonOptions{
			Label: fmt.Sprintf("Test %d", i),
		})
	}
	a.UIButton("sub", gooey.ButtonOptions{
		Label: "Sub Button",
	})
	sub.HandleScrolling()

	a.HandleScrolling()

}

func (g *Game) ExampleButtonList() {

	// Right-flowing button list

	a := gooey.NewArea("root", 0, 0, 640, 360)

	a.LayoutRow(6, 4)

	for i := 0; i < 8; i++ {
		a.UIButton(i, gooey.ButtonOptions{
			Icon:       g.GUIImg.SubImage(image.Rect(32, 0, 48, 16)).(*ebiten.Image),
			IconAnchor: gooey.AnchorTextRight,
			Ninepatch:  g.GUIImg.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image),
			Label:      fmt.Sprintf("Test %d", i),
			PaddingX:   64,
		})
	}

	a.HandleScrolling()

}

// In this example, we create a custom UI element.
func (g *Game) ExampleCustomLayout() {

	area := gooey.NewArea("text", 0, 0, 192, 192)

	area.UITextbox(gooey.TextboxOptions{
		Padding: 8,
		Text:    "This example just shows how a custom GUI element (in this case, a cross-shaped button element) could be made.",
	})

	buttonPad := gooey.NewArea("button pad", 128, 128, 96, 96)

	// Reset offset because we'll be manipulating it manually to do an unusual UI layout
	buttonPad.Offset.X = 0
	buttonPad.Offset.Y = 0

	// We call LayoutCustom, which allows us to force the size of UI elements from this point on - in this case, all buttons are 32x32.
	// This can be called multiple times freely.
	buttonPad.LayoutCustom(32, 32)

	// Once we've done that, we can simply define the UI elements.

	// Up Button

	// We set the Offset to control where the UI elements are created relative to the area's origin.
	buttonPad.Offset.X = 32
	buttonPad.Offset.Y = 0

	opt := &ebiten.DrawImageOptions{}

	buttonStyle := gooey.ButtonOptions{
		Icon:            g.GUIImg.SubImage(image.Rect(32, 16, 48, 32)).(*ebiten.Image),
		IconAnchor:      gooey.AnchorCenter,
		Ninepatch:       g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image),
		IconDrawOptions: opt,
	}

	moveSpd := float32(32)

	// Up

	if buttonPad.UIButton("up", buttonStyle) {
		buttonPad.Rect.Y -= moveSpd
	}

	// Right

	// Rotate the icon first
	opt.GeoM.Translate(-8, -8)
	opt.GeoM.Rotate(math.Pi / 2)
	opt.GeoM.Translate(8, 8)

	// Set the offset
	buttonPad.Offset.X = 64
	buttonPad.Offset.Y = 32

	// Render the button and check
	if buttonPad.UIButton("right", buttonStyle) {
		buttonPad.Rect.X += moveSpd
	}

	// Down

	buttonPad.Offset.X = 32
	buttonPad.Offset.Y = 64

	opt.GeoM.Translate(-8, -8)
	opt.GeoM.Rotate(math.Pi / 2)
	opt.GeoM.Translate(8, 8)

	if buttonPad.UIButton("down", buttonStyle) {
		buttonPad.Rect.Y += moveSpd
	}

	// Left

	buttonPad.Offset.X = 0
	buttonPad.Offset.Y = 32

	opt.GeoM.Translate(-8, -8)
	opt.GeoM.Rotate(math.Pi / 2)
	opt.GeoM.Translate(8, 8)

	if buttonPad.UIButton("left", buttonStyle) {
		buttonPad.Rect.X -= moveSpd
	}

}

func (g *Game) ExampleButtonsAndTextbox() {

	// Buttons + textbox

	a := gooey.NewArea("buttons and textbox", 0, 0, 300, 300)

	a.UITextbox(gooey.TextboxOptions{
		Text: "Testing some long text out; it's pretty simple, and it should just work. You should be able to just type a lot and have it split... The core idea, anyway, is to make this as simple as possible. It simply draws a simple textbox in one location with one size and lets you do what you're going to do.\n\nIt handles new-line characters as well.",
		// Height:          256,
		Ninepatch:       g.GUIImg.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image),
		Padding:         8,
		TypewriterIndex: int(g.Frame * 60),
		TypewriterOn:    true,
	})

	// Themed button
	opt := &ebiten.DrawImageOptions{}

	// Multiply blend mode
	opt.Blend = MultiplyBlendMode
	opt.ColorScale.Scale(1, 1.5, 2, 1)

	// Call HighlightControlBegin to set up switching between UI elements using input (see the Game.Update() function).
	a.HighlightControlBegin()

	if a.UIButton("reset typewriter", gooey.ButtonOptions{
		Label:                "Reset Typewriter",
		PaddingY:             16,
		Ninepatch:            g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image),
		BGPatternDrawOptions: opt,
		BGPattern:            g.BGPattern,
	}) {
		g.Frame = 0
	}

	if a.UIButton("move ui", gooey.ButtonOptions{
		Label:     "Move UI to Other Side",
		PaddingY:  16,
		Ninepatch: g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image),
	}) {
		if a.Rect.X == 0 {
			a.Rect.X = 640 - a.Rect.W
		} else {
			a.Rect.X = 0
		}
	}

	a.HandleScrolling()

	a.HighlightControlEnd()

}

func (g *Game) Layout(w, h int) (int, int) {
	return 640, 360
}

func main() {

	g := NewGame()
	ebiten.RunGame(g)

}
