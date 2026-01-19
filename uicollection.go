package gooey

import (
	"strconv"
)

// UICollection is a set of options for rendering UI elements together.
// While any element can be placed within a Collection, if you want to
// retrieve the state of the element or if the element needs to be
type UICollection struct {
	ArrangerModifier ArrangeFunc // A customizeable modifier that alters the location where the UI element is going to render.
	Elements         []UIElement
}

// NewUICollection quickly creates a new UICollection out of a set of UIElements.
func NewUICollection(elements ...UIElement) UICollection {
	return UICollection{
		Elements: elements,
	}
}

func (c *UICollection) Add(elements ...UIElement) {
	c.Elements = append(c.Elements, elements...)
}

func (c UICollection) WithArrangerModifier(modifier ArrangeFunc) UICollection {
	c.ArrangerModifier = modifier
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

func (c UICollection) draw(dc *DrawCall) {

	if c.ArrangerModifier != nil {
		c.ArrangerModifier(dc)
	}

	for index, element := range c.Elements {
		dc.Instance.layout.add(dc.Instance.id+"__"+strconv.Itoa(index), element, dc.Clone())
		dc.Instance.layout.Advance(-1)
	}

}

// AddTo adds the UI element to the given Layout.
// The id string should be unique and is used to identify and keep track of its location and internal state, if it saves any such state.
func (c UICollection) AddTo(layout *Layout, id string) {
	layout.add(id, c, layout.newDefaultDrawcall())
}
