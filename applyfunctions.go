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

	if !other.ToggledHighlightColor.IsZero() {
		s.ToggledHighlightColor = other.ToggledHighlightColor
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

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
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
func (s UIButtonGroup) Apply(other UIButtonGroup) UIButtonGroup {
	
	if other.MinimumToggled != 0 {
		s.MinimumToggled = other.MinimumToggled
	}

	if other.MaximumToggled != 0 {
		s.MaximumToggled = other.MaximumToggled
	}

	if other.DisallowUntoggling {
		s.DisallowUntoggling = other.DisallowUntoggling
	}

	if !other.BaseButton.IsZero() {
		s.BaseButton = other.BaseButton
	}

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
	}

	if other.Options != nil {
		s.Options = other.Options
	}

	if other.Pointer != 0 {
		s.Pointer = other.Pointer
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UICollection) Apply(other UICollection) UICollection {
	
	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
	}

	if other.Elements != nil {
		s.Elements = other.Elements
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UIColor) Apply(other UIColor) UIColor {
	
	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
	}

	if !other.FillColor.IsZero() {
		s.FillColor = other.FillColor
	}

	if other.OutlineThickness != 0 {
		s.OutlineThickness = other.OutlineThickness
	}

	if !other.OutlineColor.IsZero() {
		s.OutlineColor = other.OutlineColor
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UICustomDraw) Apply(other UICustomDraw) UICustomDraw {
	
	if other.DrawFunc != nil {
		s.DrawFunc = other.DrawFunc
	}

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
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

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
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

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
	}

	if other.Stretch != 0 {
		s.Stretch = other.Stretch
	}

	return s
}

// Apply copies the relevant non-zero elements from the other
// object into the calling object.
func (s UIImageLooping) Apply(other UIImageLooping) UIImageLooping {
	
	if other.Image != nil {
		s.Image = other.Image
	}

	if other.DrawOptions != nil {
		s.DrawOptions = other.DrawOptions
	}

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
	}

	if !other.Offset.IsZero() {
		s.Offset = other.Offset
	}

	if !other.Scale.IsZero() {
		s.Scale = other.Scale
	}

	if other.Rotation != 0 {
		s.Rotation = other.Rotation
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

	if other.NoWrap {
		s.NoWrap = other.NoWrap
	}

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
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

	if other.ArrangerModifier != nil {
		s.ArrangerModifier = other.ArrangerModifier
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

	if other.SliderHeadLerpPercentage != 0 {
		s.SliderHeadLerpPercentage = other.SliderHeadLerpPercentage
	}

	if other.StepSize != 0 {
		s.StepSize = other.StepSize
	}

	if other.Disabled {
		s.Disabled = other.Disabled
	}

	if other.Pointer != nil {
		s.Pointer = other.Pointer
	}

	return s
}

