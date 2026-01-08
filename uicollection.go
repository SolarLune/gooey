package gooey

import (
	"strconv"
)

// UICollection is a set of options for rendering UI elements together.
// While any element can be placed within a Collection, if you want to
// retrieve the state of the element or if the element needs to be
type UICollection struct {
	LayoutModifier ArrangeFunc
	Elements       []UIElement
}

// NewUICollection quickly creates a new UICollection out of a set of UIElements.
func NewUICollection(elements ...UIElement) UICollection {
	return UICollection{
		Elements: elements,
	}
}

func (c UICollection) WithLayoutModifier(modifier ArrangeFunc) UICollection {
	c.LayoutModifier = modifier
	return c
}

func (c UICollection) highlightable() bool {
	return false
}

func (c UICollection) ForAllElementsRecursive(forEach func(e UIElement) (UIElement, bool)) bool {
	for i, e := range c.Elements {
		switch element := e.(type) {
		case UICollection:
			return element.ForAllElementsRecursive(forEach)
		default:
			result, cancel := forEach(element)
			c.Elements[i] = result

			if !cancel {
				return false
			}
		}
	}
	return false
}

func (c UICollection) draw(dc DrawCall) {

	if c.LayoutModifier != nil {
		dc = c.LayoutModifier(dc)
	}

	for index, element := range c.Elements {
		dc.Instance.layout.add(dc.Instance.id+"__"+strconv.Itoa(index), element, dc)
		dc.Instance.layout.Advance(-1)
	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (c UICollection) AddTo(layout *Layout, id string) {
	layout.add(id, c, layout.newDefaultDrawcall())
}
