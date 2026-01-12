package gooey
// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UIButton) Apply(other UIButton) UIButton {
	
	if !other.BaseColor.IsZero() {
		s.BaseColor = other.BaseColor
	}

	if !other.HighlightColor.IsZero() {
		s.HighlightColor = other.HighlightColor
	}

	if !other.PressedColor.IsZero() {
		s.PressedColor = other.PressedColor
	}

	if !other.DisabledColor.IsZero() {
		s.DisabledColor = other.DisabledColor
	}

	if other.Toggleable {
		s.Toggleable = other.Toggleable
	}

	if other.Disabled {
		s.Disabled = other.Disabled
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if other.Graphics != nil {
		s.Graphics = other.Graphics
	}

	if other.Pointer != nil {
		s.Pointer = other.Pointer
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UICollection) Apply(other UICollection) UICollection {
	
	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if other.Elements != nil {
		s.Elements = other.Elements
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UIColor) Apply(other UIColor) UIColor {
	
	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if !other.Color.IsZero() {
		s.Color = other.Color
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UICustomDraw) Apply(other UICustomDraw) UICustomDraw {
	
	if other.DrawFunc != nil {
		s.DrawFunc = other.DrawFunc
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if other.Highlightable {
		s.Highlightable = other.Highlightable
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UICycleButton) Apply(other UICycleButton) UICycleButton {
	
	if other.Options != nil {
		s.Options = other.Options
	}

	if !other.BaseColor.IsZero() {
		s.BaseColor = other.BaseColor
	}

	if !other.HighlightColor.IsZero() {
		s.HighlightColor = other.HighlightColor
	}

	if !other.DisabledColor.IsZero() {
		s.DisabledColor = other.DisabledColor
	}

	if !other.GraphicsButtonBaseColor.IsZero() {
		s.GraphicsButtonBaseColor = other.GraphicsButtonBaseColor
	}

	if !other.GraphicsButtonHighlightColor.IsZero() {
		s.GraphicsButtonHighlightColor = other.GraphicsButtonHighlightColor
	}

	if !other.GraphicsButtonPressedColor.IsZero() {
		s.GraphicsButtonPressedColor = other.GraphicsButtonPressedColor
	}

	if !other.GraphicsButtonDisabledColor.IsZero() {
		s.GraphicsButtonDisabledColor = other.GraphicsButtonDisabledColor
	}

	if other.ClickZoneSize != 0 {
		s.ClickZoneSize = other.ClickZoneSize
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if other.Disabled {
		s.Disabled = other.Disabled
	}

	if other.Vertical {
		s.Vertical = other.Vertical
	}

	if other.GraphicsBody != nil {
		s.GraphicsBody = other.GraphicsBody
	}

	if other.GraphicsButtonPrevious != nil {
		s.GraphicsButtonPrevious = other.GraphicsButtonPrevious
	}

	if other.GraphicsButtonNext != nil {
		s.GraphicsButtonNext = other.GraphicsButtonNext
	}

	if other.Pointer != nil {
		s.Pointer = other.Pointer
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UIImage) Apply(other UIImage) UIImage {
	
	if other.Image != nil {
		s.Image = other.Image
	}

	if other.DrawOptions != nil {
		s.DrawOptions = other.DrawOptions
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if other.Stretch != 0 {
		s.Stretch = other.Stretch
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UILabel) Apply(other UILabel) UILabel {
	
	if other.Text != "" {
		s.Text = other.Text
	}

	if other.Alignment != 0 {
		s.Alignment = other.Alignment
	}

	if other.DrawOptions != nil {
		s.DrawOptions = other.DrawOptions
	}

	if other.TypewriterIndex != 0 {
		s.TypewriterIndex = other.TypewriterIndex
	}

	if other.LineSpacing != 0 {
		s.LineSpacing = other.LineSpacing
	}

	if other.MaxCharCount != 0 {
		s.MaxCharCount = other.MaxCharCount
	}

	if other.PaddingTop != 0 {
		s.PaddingTop = other.PaddingTop
	}

	if other.PaddingLeft != 0 {
		s.PaddingLeft = other.PaddingLeft
	}

	if other.PaddingRight != 0 {
		s.PaddingRight = other.PaddingRight
	}

	if other.PaddingBottom != 0 {
		s.PaddingBottom = other.PaddingBottom
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if !other.OverrideTextStyle.IsZero() {
		s.OverrideTextStyle = other.OverrideTextStyle
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UISlider) Apply(other UISlider) UISlider {
	
	if other.Background != nil {
		s.Background = other.Background
	}

	if other.SliderGraphics != nil {
		s.SliderGraphics = other.SliderGraphics
	}

	if other.LayoutModifier != nil {
		s.LayoutModifier = other.LayoutModifier
	}

	if !other.BaseColor.IsZero() {
		s.BaseColor = other.BaseColor
	}

	if !other.HighlightColor.IsZero() {
		s.HighlightColor = other.HighlightColor
	}

	if !other.DisabledColor.IsZero() {
		s.DisabledColor = other.DisabledColor
	}

	if other.ClickPadding != 0 {
		s.ClickPadding = other.ClickPadding
	}

	if other.SliderHeadLerpPercentage != 0 {
		s.SliderHeadLerpPercentage = other.SliderHeadLerpPercentage
	}

	if other.StepSize != 0 {
		s.StepSize = other.StepSize
	}

	if other.Disabled {
		s.Disabled = other.Disabled
	}

	return s
}

