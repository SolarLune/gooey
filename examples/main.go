package main

import (
	"embed"
	"fmt"
	"image"
	"math"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/gooey"
	"golang.org/x/image/font/basicfont"
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
	ExampleIndex int
	Examples     []func(screen *ebiten.Image)
	Frame        float64
	Font         text.Face

	GUIImg         *ebiten.Image
	RockTexture    *ebiten.Image
	WalkingMonster *ebiten.Image

	DebugDrawing bool
}

func NewGame() *Game {
	ebiten.SetWindowTitle("Gooey Example")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	g := &Game{}

	g.GUIImg = g.LoadImage("gui.png")
	g.RockTexture = g.LoadImage("rocktexture.png")
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
	style := gooey.NewTextStyle()
	style.Font = g.Font
	style.OutlineColor = gooey.NewColor(0, 0, 0, 1)
	style.ShadowLength = 4
	style.TextColor = gooey.NewColor(1, 1, 1, 1)
	// style.TextColor = gooey.NewColor(0.1, 0.15, 0.2, 1)
	style.OutlineThickness = 1
	gooey.SetDefaultTextStyle(style)
	// gooey.SetFont(g.Font) // Set the font used for text

	g.Examples = []func(screen *ebiten.Image){
		g.ExampleColor,
		g.ExampleButton,
		g.ExampleButtonList,
		g.ExampleScrollableButtonGrid,
		g.ExampleCustomLayout,
		g.ExampleSliders,
		g.ExampleCustomDraw,
		g.ExampleLayoutMap,
		g.ExampleHighlightToggle,
	}

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

	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.DebugDrawing = !g.DebugDrawing
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.ExampleIndex++
	} else if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.ExampleIndex--
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		StartProfiling()
	}

	if g.ExampleIndex < 0 {
		g.ExampleIndex += len(g.Examples)
	} else if g.ExampleIndex >= len(g.Examples) {
		g.ExampleIndex = 0
	}

	g.Frame += 1.0 / 60.0

	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	// To use gooey, just call gooey.Begin() / gooey.End() and issue UIElement calls in-between.
	// That is done in the example functions.
	gooey.Begin(
		gooey.UpdateSettings{
			RightInput: ebiten.IsKeyPressed(ebiten.KeyRight),
			LeftInput:  ebiten.IsKeyPressed(ebiten.KeyLeft),
			UpInput:    ebiten.IsKeyPressed(ebiten.KeyUp),
			DownInput:  ebiten.IsKeyPressed(ebiten.KeyDown),

			NextInput:   ebiten.IsKeyPressed(ebiten.KeyTab),
			PrevInput:   ebiten.IsKeyPressed(ebiten.KeyTab) && ebiten.IsKeyPressed(ebiten.KeyShift),
			AcceptInput: ebiten.IsKeyPressed(ebiten.KeyX),
			CancelInput: ebiten.IsKeyPressed(ebiten.KeyC),

			UseMouse:       true,
			LeftMouseClick: ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		},
	)

	defer gooey.End()

	// Update the highlight controls - this allows us to navigate the UI
	// using keyboard / gamepad presses.

	// Draw a pattern behind the screen
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Scale(640/float64(g.RockTexture.Bounds().Dx()), 360/float64(g.RockTexture.Bounds().Dy()))
	opt.ColorScale.Scale(0.2, 0.2, 0.2, 1)
	screen.DrawImage(g.RockTexture, opt)

	// Here is where the gooey UI elements are created and logic is done.
	g.Examples[g.ExampleIndex](screen)

	screen.DrawImage(gooey.Texture(), nil)

	if g.DebugDrawing {
		gooey.DrawDebug(screen, true)
	}

	g.drawtext(screen, 0, 0, "Q / E : Go to different examples\nF1 : Toggle debug view\nArrow keys + X : Make selections\nTab / Shift+Tab : Next / Prev.")

}

func (g *Game) ExampleColor(screen *ebiten.Image) {

	// Define a layout for the GUI.
	l := gooey.NewLayout("Simple Color", 0, 0, 500, 200)

	// Center it.
	l.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	// Create a Color UI element and add it to the layout.
	flat := gooey.UIColor{FillColor: gooey.NewColor(1, 0, 0, 1)}
	flat.AddTo(l, "flat color")

	g.drawtext(gooey.Texture(), 250, 0,
		`Flat Color: This is a simple example showing how a 
	flat color can be displayed in a Layout area.`)

}

func (g *Game) ExampleButton(screen *ebiten.Image) {

	// Define a layout for the GUI.
	layout := gooey.NewLayout("Example Simple Button", 0, 0, 500, 200)

	layout.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	buttonOpt := gooey.NewUIButton().WithGraphics(

		gooey.NewUICollection(

			gooey.UIImage{
				Image:   gooey.SubImage(g.GUIImg, 0, 24, 24, 24),
				Stretch: gooey.StretchModeNinepatch,
			},

			gooey.UILabel{
				Text:      "Hello, there! Welcome to Gooey - it should be an easy solution for games made in Ebitengine.\n\n... Hopefully.",
				Alignment: gooey.AlignmentCenterCenter,
			},
		),
	)

	if buttonOpt.AddTo(layout, "button") {
		fmt.Println("Hello!")
	}

	g.drawtext(gooey.Texture(), 250, 0,
		`Simple Button: This example simply displays 
		a button with a ninepatch background 
		and some static text.`)

}

func (g *Game) ExampleButtonList(screen *ebiten.Image) {

	layout := gooey.NewLayout("Example Button List", 0, 0, 500, 200)

	layout.SetArranger(gooey.ArrangerGrid{
		ElementSize:    gooey.Vector2{500, 24},
		ElementPadding: gooey.Vector2{8, 8},
	})

	layout.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	for i := 0; i < 4; i++ {

		iconOpt := &ebiten.DrawImageOptions{}
		iconOpt.GeoM.Scale(0.6, 0.6)
		iconOpt.GeoM.Translate(-52, 0)

		button := gooey.NewUIButton().WithGraphics(

			gooey.NewUICollection(

				// Outer frame
				gooey.UIImage{
					Image:   gooey.SubImage(g.GUIImg, 0, 24, 24, 24),
					Stretch: gooey.StretchModeNinepatch,
				},

				// Label
				gooey.UILabel{
					Alignment: gooey.AlignmentCenterCenter,
					Text:      "Button #" + strconv.Itoa(i),
				},

				// Icon
				gooey.UIImage{
					Image:       gooey.SubImage(g.GUIImg, 32, 16, 16, 16),
					DrawOptions: iconOpt,
				},
			),
		).AddTo(layout, "button_"+strconv.Itoa(i))

		if button {
			fmt.Println("You pressed button #", i, "!")
		}

	}

	opt := &ebiten.DrawImageOptions{}
	opt.ColorScale.Scale(0, 0.5, 1, 1)

	cyclebuttonIcon := gooey.SubImage(g.GUIImg, 48, 16, 16, 16)

	cycle := gooey.UICycleButton{
		BaseColor:      gooey.NewColor(0.6, 0.6, 0.6, 1),
		HighlightColor: gooey.NewColor(1, 1, 1, 1),
		DisabledColor:  gooey.NewColor(0.2, 0.2, 0.2, 1),

		GraphicsButtonBaseColor:      gooey.NewColor(0.6, 0.6, 0.6, 1),
		GraphicsButtonHighlightColor: gooey.NewColor(1, 1, 1, 1),
		GraphicsButtonPressedColor:   gooey.NewColor(0.2, 0.2, 0.2, 1),
		GraphicsButtonDisabledColor:  gooey.NewColor(0.2, 0.2, 0.2, 1),

		Options: []string{"Option #1", "Option #2", "Option #3"},

		GraphicsBody: gooey.NewUICollection(
			gooey.UIImage{
				DrawOptions: opt,
				Image:       gooey.SubImage(g.GUIImg, 0, 0, 24, 24),
				Stretch:     gooey.StretchModeNinepatch,
			},
			gooey.UILabel{
				Alignment: gooey.AlignmentCenterCenter,
			},
		),

		GraphicsButtonPrevious: gooey.UIImage{
			Image: cyclebuttonIcon,
		}.WithTransform(gooey.Vector2{}, gooey.Vector2{-1, 1}, 0),

		GraphicsButtonNext: gooey.UIImage{
			Image: cyclebuttonIcon,
		}.WithTransform(gooey.Vector2{}, gooey.Vector2{1, 1}, 0),
	}

	first := cycle.AddTo(layout, "cycle button")

	cycle.WithDisabled(first == 0).AddTo(layout, "sub-cycle button")

	g.drawtext(gooey.Texture(), 250, 0,
		`Button List: Layouts allow you to easily
	change the positioning and size of UI elements,
	which will stretch to fit. In this example, a grid
	layout function is used to create a vertical list
	of buttons.`)

}

func (g *Game) ExampleScrollableButtonGrid(screen *ebiten.Image) {

	// Layouts will be scrollable automatically if they extend horizontally to the right
	// or vertically to the bottom beyond the Layout's Rect.
	// To turn off this scrolling behavior, set Layout.AutoScrollSpeed to 0.

	layout := gooey.NewLayout("Example Scrollable Button List", 0, 0, 500, 200)

	layout.SetArranger(gooey.ArrangerGrid{
		ElementSize:    gooey.Vector2{64, 64},
		ElementPadding: gooey.Vector2{4, 4},
		ElementCount:   6,
		// NoCenterElements:   true,
	})

	layout.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	for i := 0; i < 60; i++ {

		button := gooey.NewUIButton().WithGraphics(

			gooey.NewUICollection(

				// Outer frame
				gooey.UIImage{
					Image:   gooey.SubImage(g.GUIImg, 0, 24, 24, 24),
					Stretch: gooey.StretchModeNinepatch,
				},

				// Label
				gooey.UILabel{
					Alignment: gooey.AlignmentCenterCenter,
					Text:      "#" + strconv.Itoa(i),
				},
			),
		).AddTo(layout, "button_"+strconv.Itoa(i))

		if button {
			fmt.Println("You pressed button #", i, "!")
		}

	}

	opt := &ebiten.DrawImageOptions{}
	opt.ColorScale.Scale(0, 0.5, 1, 1)

	g.drawtext(gooey.Texture(), 250, 0,
		`Scrollable Button List: Layouts scroll
	to focus on highlighted UI elements automatically
	if the layouts extend horizontally to the right
	or vertically from the bottom.
	`)

}

func (g *Game) ExampleCustomLayout(screen *ebiten.Image) {

	// Define a layout for the GUI.
	layout := gooey.NewLayout("Example Custom Layout", 0, 0, 500, 200)
	layout.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	totalButtons := 8
	circleSize := 60.0

	layout.SetCustomArranger(func(drawCall *gooey.DrawCall) {

		buttonSize := float32(32)

		buttonRect := gooey.Rect{
			X: drawCall.Rect.Center().X - float32(math.Cos(float64(drawCall.ElementIndex)/float64(totalButtons)*math.Pi*2)*circleSize),
			Y: drawCall.Rect.Center().Y - float32(math.Sin(float64(drawCall.ElementIndex)/float64(totalButtons)*math.Pi*2)*circleSize),
			W: buttonSize,
			H: buttonSize,
		}

		// fmt.Println(buttonRect)
		if drawCall.IsHighlighted() {
			buttonRect.W *= 2
			buttonRect.X -= buttonRect.W / 4
			buttonRect.Y -= 8
		}

		buttonRect = drawCall.Instance.PrevRect().Lerp(buttonRect, 0.1)

		drawCall.Rect = buttonRect

	})

	highlightingOrder := []string{}

	for i := 0; i < totalButtons; i++ {

		id := "button_" + strconv.Itoa(i)

		button := gooey.NewUIButton().WithGraphics(

			gooey.NewUICollection(

				gooey.UIImage{
					Image:   gooey.SubImage(g.GUIImg, 0, 24, 24, 24),
					Stretch: gooey.StretchModeNinepatch,
				},

				gooey.UIImage{
					Image: g.WalkingMonster,
				},
			),
		).AddTo(layout, id)

		highlightingOrder = append(highlightingOrder, id)

		if button {
			fmt.Println("You pressed button #", i, "!")
		}

	}

	layout.CustomHighlightingOrder = highlightingOrder

	g.drawtext(gooey.Texture(), 250, 0,
		`Custom Layout: Layouts can also be modified
		with custom layout functions to change where elements
		are positioned programmatically.
		In this example, the buttons are given an icon 
		and arrayed around a circle.`)

}

func (g *Game) ExampleSliders(screen *ebiten.Image) {

	// Define a layout for the GUI.
	layout := gooey.NewLayout("Example Sliders", 0, 0, 500, 200)
	layout.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	layout.SetCustomArranger(func(drawCall *gooey.DrawCall) {
		sliderHeight := float32(32)
		sliderRect := gooey.Rect{drawCall.Rect.X, drawCall.Rect.Y + (float32(drawCall.ElementIndex) * sliderHeight), drawCall.Rect.W, sliderHeight}
		drawCall.Rect = sliderRect
	})

	for i := 0; i < 6; i++ {

		sliderObj := &ebiten.DrawImageOptions{}
		sliderObj.GeoM.Scale(0.5, 0.5)

		labelOpt := gooey.UILabel{
			Alignment: gooey.AlignmentCenterCenter,
			Text:      "Huh?",
			// Move the label a bit lower; this could also be done with the
			// DrawOptions property.
			ArrangerModifier: func(drawCall *gooey.DrawCall) {
				drawCall.Rect = drawCall.Rect.Move(0, 16)
			},
		}

		if element := layout.UIElement("slider" + strconv.Itoa(i)); element != nil {
			labelOpt.Text = strconv.FormatFloat(float64(element.State().(*gooey.SliderState).Percentage), 'f', 1, 32)
		}

		gooey.UISlider{

			SliderHeadLerpPercentage: 0.2,

			Background: gooey.UIImage{
				Image:   g.GUIImg.SubImage(image.Rect(0, 48, 48, 64)).(*ebiten.Image),
				Stretch: gooey.StretchModeThreepatch,
			},

			// You could just use an image, but I want an image and a label, so we'll go with a collection
			// to draw them together.
			// NewUICollection() quickly creates a UI collection from the supplied UI elements.
			SliderGraphics: gooey.NewUICollection(
				gooey.UIImage{
					Image:       g.GUIImg.SubImage(image.Rect(48, 48, 64, 64)).(*ebiten.Image),
					DrawOptions: sliderObj,
				},

				labelOpt,
			),
		}.AddTo(layout, "slider"+strconv.Itoa(i))

	}

	g.drawtext(gooey.Texture(), 250, 0,
		`Sliders: Sliders are vertical or horizontal
		elements that are draggable across a certain range.`)

}

func (g *Game) ExampleCustomDraw(screen *ebiten.Image) {

	// Define a layout for the GUI.
	// layout := gooey.NewLayout("Example Custom Draw", 0, 0, 500, 200)
	// layout.AlignToScreenbuffer(gooey.AnchorCenter, 0)

	baseRect := gooey.Rect{0, 0, 500, 200}.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	leftRect, rightRect := baseRect.Split(0.5, true)

	leftLayout := gooey.NewLayoutFromRect("CD-L", leftRect)
	rightLayout := gooey.NewLayoutFromRect("CD-R", rightRect)

	rightLayout.Rect.Y = 100 + float32(math.Sin(g.Frame*math.Pi*1.276)*50)

	colorShift := 0.0

	circle := gooey.UICustomDraw{
		DrawFunc: func(screen *ebiten.Image, dc *gooey.DrawCall) {
			center := dc.Rect.Center()
			radius := 16 + (float32(math.Sin(g.Frame*math.Pi)) * 8)
			color := gooey.NewColorFromHSV((g.Frame+colorShift)/10, 1, 1).ToNRGBA64()
			vector.FillCircle(screen, center.X, center.Y, radius, color, true)
			vector.StrokeCircle(screen, center.X, center.Y, radius+8, 4, color, true)
		},
	}

	circle.AddTo(leftLayout, "custom draw-l")

	colorShift += 2
	circle.AddTo(rightLayout, "custom draw-r")

	g.drawtext(gooey.Texture(), 250, 0,
		`Custom Draw: In this example, a customized UI
		element is drawn using the UICustomDraw object.
		Each custom draw object can be independent, 
		inherit relevant properties from previous UI
		elements in a stack (e.g. UICollections),
		and be altered by external variables.`)

}

func (g *Game) ExampleLayoutMap(screen *ebiten.Image) {

	base := gooey.Rect{X: 0, Y: 0, W: 500, H: 200}.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	layouts := gooey.NewLayoutsFromStrings("Layout", base,
		"aaaabbcccc",
		"aaaabbcccc",
		"aaaa gddd ",
		"eee ffdddh",
	)

	for _, l := range layouts {

		if gooey.NewUIButton().WithGraphics(
			gooey.NewUICollection(

				// Outer frame
				gooey.UIImage{
					Image:   g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image),
					Stretch: gooey.StretchModeNinepatch,
				},

				// Label
				gooey.UILabel{
					Alignment: gooey.AlignmentCenterCenter,
					Text:      l.ID,
				},
			),
		).AddTo(l, "button_"+l.ID) {
			fmt.Println("You pressed " + l.ID + "!")
		}

	}

	g.drawtext(gooey.Texture(), 250, 0,
		`Layout Map: In this example, we're creating
		several layouts from a starting rectangle and 
		a set of strings indicating position and proportions.
		This can be useful for creating an overall interface.
		`)

}

var highlightToggleSelection = [3]int{-1, -1, -1}

func (g *Game) ExampleHighlightToggle(screen *ebiten.Image) {

	base := gooey.Rect{X: 0, Y: 0, W: 500, H: 200}.AlignToScreenbuffer(gooey.AlignmentCenterCenter, 0)

	// Create the layouts.
	layouts := gooey.NewLayoutsFromStrings("highlight toggle", base,
		"aaa   cccc",
		"aaa   cccc",
		"aaa   cccc",
		"bbb   cccc",
		"bbb   cccc",
	)

	// Create the button style.
	buttonStyle := gooey.NewUIButton().WithGraphics(
		gooey.NewUICollection(

			// Outer frame
			gooey.UIImage{
				Image:   g.GUIImg.SubImage(image.Rect(0, 24, 24, 48)).(*ebiten.Image),
				Stretch: gooey.StretchModeNinepatch,
			},

			// Label
			gooey.UILabel{
				Alignment: gooey.AlignmentCenterCenter,
			},
		),
	)

	// Options A

	optionsA := layouts['a']

	gooey.UIColor{
		OutlineColor:     gooey.NewColor(1, 1, 1, 1),
		OutlineThickness: 2,
	}.AddTo(optionsA, "bg")

	// You can also use the gooey.ContainerSize constant for ArrangerGrid's ElementSize to easily do fractional scaling.
	grid := gooey.ArrangerGrid{ElementSize: gooey.Vector2{X: 0, Y: 32}}

	optionsA.SetArranger(grid.WithOuterPadding(8).WithElementSize(0, gooey.ContainerSize/4))

	optionsB := layouts['b'].SetArranger(grid)

	optionsC := layouts['c'].SetArranger(grid.WithElementSize(0, 32))

	// A button group is a set of toggleable buttons that only has one available at a time
	buttonGroup := gooey.UIButtonGroup{
		BaseButton: buttonStyle,
		// MinimumToggled: 1,
		MaximumToggled:     1,
		DisallowUntoggling: true,
	}

	// We create a page to indicate highlighting flow - A > B > C
	page := gooey.NewPage(optionsA, optionsB, optionsC)

	optionsAResult := buttonGroup.WithOptions("First", "Second", "Third", "Fourth").AddTo(optionsA, "menu a")
	if optionsAResult.SelectionMade() {
		highlightToggleSelection[0] = optionsAResult.FirstSelected()
		page.Advance(1)
	}

	// Options B

	optionsBResult := buttonGroup.WithOptions("Option A", "Option B").AddTo(optionsB, "menuoptions-b-choice a")
	if optionsBResult.SelectionMade() {
		highlightToggleSelection[1] = optionsBResult.FirstSelected()
		page.Advance(1)
	}

	// Options C

	for i := 0; i < 40; i++ {
		n := strconv.Itoa(i)
		if buttonStyle.WithText("Item #"+n).AddTo(optionsC, "menuoptions-c-item "+n) {
			highlightToggleSelection[2] = i
			fmt.Println("You pressed:", highlightToggleSelection[0], highlightToggleSelection[1], highlightToggleSelection[2])
		}
	}

	// Syntactic sugar for just checking the input you're already passing earlier.
	if gooey.InputPressedCancel() {
		page.Advance(-1)
		// if !optionsB.HighlightingLocked {
		// 	optionsBResult.SetAllSelected(false)
		// } else if !optionsA.HighlightingLocked {
		// 	optionsAResult.SetAllSelected(false)
		// }
	}

	g.drawtext(gooey.Texture(), 250, 0,
		`Highlighting Pages: This example shows
how highlighting pages work. Highlighting advances through menus / layouts
in sequence, before ending at the last one.

Pressing cancel (C) goes back one menu.`)

}

func (g *Game) Layout(w, h int) (int, int) {
	return 640, 360
}

var defaultFont = text.NewGoXFace(basicfont.Face7x13)

func (g *Game) drawtext(screen *ebiten.Image, x, y float32, txt string) {

	txt = strings.ReplaceAll(txt, "\t", "")

	opt := &text.DrawOptions{}
	for index, str := range strings.Split(txt, "\n") {
		opt.GeoM.Translate(float64(x)+4, float64(y)+4)
		opt.GeoM.Translate(0, float64(index*14))
		opt.ColorScale.Scale(0, 0, 0, 1)
		text.Draw(screen, str, defaultFont, opt)
		opt.GeoM.Translate(-1, -1)
		opt.ColorScale.Reset()
		text.Draw(screen, str, defaultFont, opt)
		opt.GeoM.Reset()
	}
}

func main() {

	g := NewGame()
	ebiten.RunGame(g)

}

func StartProfiling() {
	outFile, err := os.Create("./cpu.out")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Beginning CPU profiling...")
	pprof.StartCPUProfile(outFile)
	go func() {
		time.Sleep(5 * time.Second)
		pprof.StopCPUProfile()
		fmt.Println("CPU profiling finished.")
	}()
}
