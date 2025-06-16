package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"

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
	style.TextColor = gooey.NewColor(0.1, 0.15, 0.2, 1)
	gooey.SetDefaultTextStyle(style)
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

	if inpututil.IsKeyJustPressed(ebiten.KeyR) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		gooey.Reset() // Reset resets the internal state of the areas and UI elements in use.
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.DebugDrawing = !g.DebugDrawing
	}

	// Update the highlight controls - this allows us to navigate the UI
	// using keyboard / gamepad presses.
	gooey.Update(gooey.HighlightControlSettings{
		RightInput: ebiten.IsKeyPressed(ebiten.KeyRight),
		LeftInput:  ebiten.IsKeyPressed(ebiten.KeyLeft),
		UpInput:    ebiten.IsKeyPressed(ebiten.KeyUp),
		DownInput:  ebiten.IsKeyPressed(ebiten.KeyDown),

		NextInput:   ebiten.IsKeyPressed(ebiten.KeyTab),
		PrevInput:   ebiten.IsKeyPressed(ebiten.KeyTab) && ebiten.IsKeyPressed(ebiten.KeyShift),
		AcceptInput: ebiten.IsKeyPressed(ebiten.KeyX),
		CancelInput: ebiten.IsKeyPressed(ebiten.KeyC),

		// If you can use the mouse or not
		// UseMouse: true,
	})

	g.Frame += 1.0 / 60.0

	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	gooey.Begin()
	defer gooey.End()

	// Draw a pattern behind the screen
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Scale(640/float64(g.BGPattern.Bounds().Dx()), 360/float64(g.BGPattern.Bounds().Dy()))
	opt.ColorScale.Scale(0.2, 0.2, 0.2, 1)
	screen.DrawImage(g.BGPattern, opt)

	// Perform an example
	g.ExampleSlider()

	screen.DrawImage(gooey.Texture(), nil)

	if g.DebugDrawing {
		gooey.DrawDebug(screen, false)
	}

}

func (g *Game) ExampleLabels() {

	a := gooey.NewArea("labels test", 0, 0, 256, 256)

	a.UITextbox("text", gooey.TextboxOptions{
		Background: gooey.NewBGWhite(),
		Text:       "This is just a simple test to show how labels work. They just draw over the previous element.",
	}.WithPadding(8))

	buttonStyle := gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		PaddingY:   16,
		PaddingX:   16,
	}

	labelStyle := gooey.LabelOptions{
		Background:       gooey.NewBGColor(gooey.NewColor(0.7, 0.8, 1, 1)),
		TextPaddingLeft:  8,
		TextPaddingRight: 8,
		// Color:            gooey.NewColor(0.7, 0.8, 1, 1),
		Anchor: gooey.AnchorBottomRight,
	}

	// We can change the offset to change where UI elements are created
	a.UIButton(0, buttonStyle.WithText("Button A"))
	a.UILabel(labelStyle.WithText("#1"))

	a.UIButton(1, buttonStyle.WithText("Button B"))
	a.UILabel(labelStyle.WithText("#2"))

	a.UIButton(2, buttonStyle.WithText("Button C"))
	a.UILabel(labelStyle.WithText("#3"))

	a.UIButton(3, buttonStyle.WithText("Button D"))
	a.UILabel(labelStyle.WithText("#4"))

}

// func (g *Game) ExampleCheckbox() {

// 	// Define an area for the GUI.
// 	a := gooey.NewArea("root", 0, 0, 128, 128)

// 	// Check to see if a button was pressed.
// 	if a.UICheckbox("checkbox", gooey.CheckboxOptions{
// 		Text:               "Test.",
// 		TextAnchorPosition: gooey.AnchorCenterLeft,
// 	}).Checked {
// 		fmt.Println("The Test checkbox is enabled.")
// 	}

// }

func (g *Game) ExampleSimple() {

	// Define an area for the GUI.
	a := gooey.NewArea("root", 0, 0, 640, 360)

	// Check to see if a button was pressed.
	if a.UIButton("testButton", gooey.ButtonOptions{Background: gooey.NewBGColor(gooey.NewColorFromColor(color.White)), Text: "Test."}) {
		fmt.Println("The Test button was pressed.")
	}
}

func (g *Game) ExampleSlider() {

	// Define an area for the GUI.
	a := gooey.NewArea("root", 32, 128, 640, 360)

	// Check to see if a button was pressed.

	opt := gooey.SliderOptions{
		Width:      512,
		Height:     64,
		Background: gooey.NewBGColor(gooey.NewColor(0.1, 0.2, 0.4, 1)),
		// LineImage:  gooey.NewBGImage(g.GUIImg.SubImage(image.Rect(0, 48, 24, 64)).(*ebiten.Image)),
		LineImage:  gooey.NewBGThreePatch(g.GUIImg.SubImage(image.Rect(0, 48, 48, 64)).(*ebiten.Image), true),
		HeadImage:  g.GUIImg.SubImage(image.Rect(48, 48, 64, 64)).(*ebiten.Image),
		HeadMargin: 16,
	}

	// a.Offset.Y += 96
	sliderValue := a.UISlider("slider", opt)

	a.WithAbsolutePositioning(func(areaRect *gooey.Rect) {
		state := gooey.State("slider").(*gooey.SliderState)
		// Reposition the area rectangle to place it where we want
		areaRect.X = state.HeadX - 8
		areaRect.Y = state.HeadY - 16
		a.UITextbox("slider head textbox", gooey.TextboxOptions{
			Text:      fmt.Sprintf("%0.2f", sliderValue),
			TextStyle: gooey.NewDefaultTextStyle().WithFGColor(gooey.NewColor(1, 1, 1, 1)),
		})
	})

	a.Offset = gooey.Position{}

}

func (g *Game) ExampleMovingButtons() {

	a := gooey.NewArea("root", 0, 0, 640, 360)

	for i := 0; i < 10; i++ {

		buttonName := fmt.Sprintf("Test button %d", i)

		// We can change the offset to change where UI elements are created
		a.WithOffset(
			float32(math.Sin((g.Frame+(float64(i)*0.1))*math.Pi))*100,
			0,
			func() {

				a.UIButton(buttonName, gooey.ButtonOptions{
					Background: gooey.NewBGWhite(),
					Text:       buttonName,
					PaddingY:   8,
					PaddingX:   8,
				})

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
	buttonOpt := gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		PaddingY:   8,
	}

	// a.HighlightControlBegin()

	if left.UIButton("A", buttonOpt.WithText("Button A")) {
		textStr = "You pressed the first button."
	}
	if left.UIButton("B", buttonOpt.WithText("Button B")) {
		textStr = "You pressed the second button, Button B - good work."
	}
	if left.UIButton("C", buttonOpt.WithText("Button C")) {
		textStr = "Wow, you pressed Button 3! Incredible!!!!"
	}

	// a.HighlightControlEnd()

	right.UITextbox("text", gooey.TextboxOptions{
		Background: gooey.NewBGWhite(),
		Text:       textStr,
		Height:     360 / 2,
	}.WithPadding(16))

}

func (g *Game) ExampleGrid() {

	a := gooey.NewArea("root", 0, 0, 300, 300)

	// a.LayoutGrid(3, 3, -1, 0)
	a.SetLayout(&gooey.LayoutGrid{
		DividerCount: 3,
	})

	for y := 0; y < 3; y++ {

		for x := 0; x < 3; x++ {

			a.UIButton(fmt.Sprintf("Button %d,%d", x, y), gooey.ButtonOptions{
				Background: gooey.NewBGWhite(),
				Text:       fmt.Sprintf("Bu %d,%d", x, y),
			})

		}

	}

	gooey.SetHighlightOptionsGrid(gooey.GridHighlightOptions{
		XCount:        3,
		YCount:        3,
		IDFunc:        func(x, y int) any { return fmt.Sprintf("Button %d,%d", x, y) },
		Bidirectional: true,
		Loop:          true,
	})

	// for y := 0; y < 3; y++ {

	// 	for x := 0; x < 3; x++ {

	// 		a.UIButton(fmt.Sprintf("Button %d,%d", x, y), gooey.ButtonOptions{
	// 			Background: gooey.NewBGWhite(),
	// 			Text:       fmt.Sprintf("Bu %d,%d", x, y),
	// 			// MinWidth:   96,
	// 		})

	// 		gooey.SetHighlightOptions(gooey.HighlightOptions{
	// 			Direction:     gooey.DirectionDown,
	// 			Bidirectional: true,
	// 			Loop:          true,
	// 			IDs: []any{
	// 				fmt.Sprintf("Button %d,0", x),
	// 				fmt.Sprintf("Button %d,1", x),
	// 				fmt.Sprintf("Button %d,2", x),
	// 			},
	// 		})

	// 	}

	// 	gooey.SetHighlightOptions(gooey.HighlightOptions{
	// 		Direction:     gooey.DirectionRight,
	// 		Bidirectional: true,
	// 		Loop:          true,
	// 		IDs: []any{
	// 			fmt.Sprintf("Button 0,%d", y),
	// 			fmt.Sprintf("Button 1,%d", y),
	// 			fmt.Sprintf("Button 2,%d", y),
	// 		},
	// 	})

	// }

}

var spin = false
var grow = false

func (g *Game) ExampleImage() {

	a := gooey.NewArea("example image", 0, 0, 200, 300)

	a.UITextbox("text", gooey.TextboxOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Here's an image test. The image fits in its space regardless of scaling and rotating.",
	}.WithPadding(8))

	opt := &ebiten.DrawImageOptions{}
	bounds := g.WalkingMonster.Bounds()

	opt.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)
	if grow {
		s := 1 + math.Sin(g.Frame*math.Pi)
		opt.GeoM.Scale(s, s)
	}
	if spin {
		opt.GeoM.Rotate(g.Frame * math.Pi * 5)
	}
	opt.GeoM.Translate(float64(bounds.Dx())/2, float64(bounds.Dy())/2)

	a.UIImage("img", gooey.ImageOptions{
		Image:            g.WalkingMonster,
		DrawImageOptions: opt,
	})

	if a.UIButton("spin", gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Toggle Spin",
		PaddingY:   16,
	}) {
		spin = !spin
	}

	if a.UIButton("scale", gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Toggle Grow",
		PaddingY:   16,
	}) {
		grow = !grow
	}

}

func (g *Game) ExampleSubAreas() {

	a := gooey.NewArea("root", 80, 0, 300, 100)

	a.UIButton("Area 1", gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Outside of area"},
	)
	a.UIButton("Area 2", gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Outside of area 2"},
	)

	// Create another area that exists "within" the owning area; the position is relative to the calling area
	sub := a.UIArea("subarea", 40, 0, 300-80, 100)
	for i := 0; i < 10; i++ {
		sub.UIButton(i, gooey.ButtonOptions{
			Background: gooey.NewBGWhite(),
			Text:       fmt.Sprintf("Test %d", i),
		})
	}
	a.UIButton("sub", gooey.ButtonOptions{
		Background: gooey.NewBGWhite(),
		Text:       "Sub Button",
	})
	sub.HandleScrolling()

	a.HandleScrolling()

}

func (g *Game) ExampleButtonList() {

	// TODO: Fix clicking also pressing the top button

	buttonStyle := gooey.ButtonOptions{
		Background: gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image)),
		Icon:       g.GUIImg.SubImage(image.Rect(32, 0, 48, 16)).(*ebiten.Image),
		IconAnchor: gooey.AnchorTextLeft,
		PaddingX:   32,
		// Width:      80, // If we don't set the width explicitly, the button will auto-expand for the text
	}

	// Vertical button list
	vertArea := gooey.NewArea("vertical list", 0, 32, 120, 300)
	vertArea.SetLayout(&gooey.LayoutGrid{DividerPadding: 4})

	verticals := []any{}
	for i := 0; i < 6; i++ {
		verticals = append(verticals, "v"+strconv.Itoa(i))
		vertArea.UIButton("v"+strconv.Itoa(i), buttonStyle.WithText(fmt.Sprintf("V Button %d", i)))
	}

	// Horizontal button list
	horizontalArea := gooey.NewArea("horizontal list", 32, 0, 600, 100)
	horizontalArea.SetLayout(&gooey.LayoutGrid{DividerPadding: 4, Direction: gooey.LayoutGridDirectionRight})

	horizontals := []any{}
	for i := 0; i < 4; i++ {
		horizontals = append(horizontals, "h"+strconv.Itoa(i))
		horizontalArea.UIButton("h"+strconv.Itoa(i), buttonStyle.WithText(fmt.Sprintf("Hori %d", i)))
	}

	// Set up the highlights

	// All vertical options lead into each other bidirectionally
	gooey.SetHighlightOptions(gooey.HighlightOptions{
		Direction:     gooey.DirectionDown,
		Bidirectional: true,
		Loop:          true,
		IDs:           verticals,
	})

	// Same for horizontal.
	gooey.SetHighlightOptions(gooey.HighlightOptions{
		Direction:     gooey.DirectionRight,
		Bidirectional: true,
		Loop:          true,
		IDs:           horizontals,
	})

	// OK, so now, all vertical options should allow you to press right and go to the first horizontal option.
	for _, v := range verticals {

		gooey.SetHighlightOptions(gooey.HighlightOptions{
			Direction:     gooey.DirectionRight,
			Bidirectional: false,
			Loop:          false,
			IDs:           []any{v, horizontals[0]},
		})

	}

	// And the same for the horizontal list.
	for _, h := range horizontals {

		gooey.SetHighlightOptions(gooey.HighlightOptions{
			Direction:     gooey.DirectionDown,
			Bidirectional: false,
			Loop:          false,
			IDs:           []any{h, verticals[0]},
		})

	}

}

// In this example, we create a custom UI element.
func (g *Game) ExampleCustomLayout() {

	area := gooey.NewArea("text", 0, 0, 192, 192)
	area.UITextbox("text", gooey.TextboxOptions{
		Background: gooey.NewBGWhite(),
		Text:       "This example just shows how a custom GUI element (in this case, a cross-shaped button element) could be made.",
	}.WithPadding(8))

	buttonPadArea := gooey.NewArea("button pad", 128, 128, 96, 96)

	// We call LayoutCustom, which allows us to force the size of UI elements from this point on - in this case, all buttons are 32x32.
	// This can be called multiple times freely.
	// buttonPadArea.LayoutCustom(32, 32)

	// Once we've done that, we can simply define the UI elements.

	// Create a draw image options object specifically for rotating the icon on the buttons.
	opt := &ebiten.DrawImageOptions{}

	bgi := gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image))
	buttonStyle := gooey.ButtonOptions{
		Background:      bgi,
		Icon:            g.GUIImg.SubImage(image.Rect(32, 16, 48, 32)).(*ebiten.Image),
		IconAnchor:      gooey.AnchorCenter,
		IconDrawOptions: opt,
	}

	moveSpd := float32(32)

	// Create a helper function to easily rotate the icon.
	rotateIcon := func(rotation float64) {

		// We set opt earlier as buttonStyle.IconDrawOptions, so setting its GeoM field here changes how the button icon renders afterwards.
		opt.GeoM.Translate(-8, -8)
		opt.GeoM.Rotate(rotation)
		opt.GeoM.Translate(8, 8)

	}

	// Up

	buttonPadArea.Offset = gooey.Position{X: 32, Y: 0}
	if buttonPadArea.UIButton("up", buttonStyle) {
		buttonPadArea.Rect.Y -= moveSpd
	}

	// Right

	buttonPadArea.Offset = gooey.Position{X: 64, Y: 32}

	// Rotate the icon by 90 degrees.
	rotateIcon(math.Pi / 2)

	if buttonPadArea.UIButton("right", buttonStyle) {
		buttonPadArea.Rect.X += moveSpd
	}

	// Down

	buttonPadArea.Offset = gooey.Position{X: 32, Y: 64}
	rotateIcon(math.Pi / 2)
	if buttonPadArea.UIButton("down", buttonStyle) {
		buttonPadArea.Rect.Y += moveSpd
	}

	// Left

	buttonPadArea.Offset = gooey.Position{X: 0, Y: 32}
	rotateIcon(math.Pi / 2)
	if buttonPadArea.UIButton("left", buttonStyle) {
		buttonPadArea.Rect.X -= moveSpd
	}

}

var textA = "Testing some long text out; it's pretty simple, and it should just work. You should be able to just type a lot and have it split... The core idea, anyway, is to make this as simple as possible. It simply draws a simple textbox in one location with one size and lets you do what you're going to do.\n\nIt handles new-line characters as well."
var textB = "You can dynamically change the text as necessary, of course, and it should also just work. This is just a test, of course, but it seems to be wroking fine - oops, I mean working fine. Thank you very much for your time."
var targetText *string = &textA

func (g *Game) ExampleButtonsAndTextbox() {

	// Buttons + textbox

	a := gooey.NewArea("buttons and textbox", 0, 0, 300, 300)

	a.UITextbox("text", gooey.TextboxOptions{
		Text:            *targetText,
		Background:      gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image)),
		TypewriterIndex: int(g.Frame * 60),
		TypewriterOn:    true,
		Height:          200, // Fix the height so changing the text doesn't change the height
	}.WithPadding(8))

	// Call HighlightControlBegin to set up switching between UI elements using input (see the Game.Update() function).
	// a.HighlightControlBegin()

	// Themed button graphical setup
	ninepatch := gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image))

	overlayPattern := gooey.NewBGImage(g.BGPattern)
	overlayPattern.DrawOptions.Address = ebiten.AddressRepeat
	overlayPattern.TileOverDst = true
	overlayPattern.DrawOptions.Blend = MultiplyBlendMode
	overlayPattern.ColorM.Scale(1, 1.5, 2, 1)
	overlayPattern.Background = ninepatch

	if a.UIButton("reset typewriter", gooey.ButtonOptions{
		Background: overlayPattern,
		Text:       "Reset Typewriter",
		PaddingY:   16,
	}) {
		g.Frame = 0
	}

	if a.UIButton("change text", gooey.ButtonOptions{
		Background: overlayPattern,
		Text:       "Change Text",
		PaddingY:   16,
	}) {
		g.Frame = 0
		if targetText == &textA {
			targetText = &textB
		} else {
			targetText = &textA
		}
	}

	if a.UIButton("move ui", gooey.ButtonOptions{
		Background: gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image)),
		Text:       "Move UI to Other Side",
		PaddingY:   16,
	}) {
		if a.Rect.X == 0 {
			a.Rect.X = 640 - a.Rect.W
		} else {
			a.Rect.X = 0
		}
	}

	gooey.SetHighlightOptions(gooey.HighlightOptions{
		Direction:     gooey.DirectionDown,
		Bidirectional: true,
		Loop:          true,
		IDs: []any{
			"reset typewriter",
			"change text",
			"move ui",
		},
		OnCancel: func() { gooey.SetHighlight("reset typewriter") },
	})

	gooey.SetDefaultHighlight("reset typewriter")

	// gooey.SetHighlightOptions("reset typewriter", "change text", gooey.DirectionDown, true)

	// a.HighlightControlEnd()

}

func (g *Game) ExampleEditableTextbox() {

	a := gooey.NewArea("editable textbox", 0, 0, 300, 300)

	t := gooey.TextboxOptions{
		Background: gooey.NewBGNinepatch(g.GUIImg.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image)),
		Text:       "Well, anyway, that's the idea. It should be rather simple. This\nIs\nA\nTest.",
		Editable:   true,
	}.WithPadding(8)

	if ebiten.IsKeyPressed(ebiten.KeyF5) {
		t.Text = "HUHUHHUHUHUHUH"
	}
	a.UITextbox("edit", t)

}

func (g *Game) Layout(w, h int) (int, int) {
	return 640, 360
}

func main() {

	g := NewGame()
	ebiten.RunGame(g)

}
