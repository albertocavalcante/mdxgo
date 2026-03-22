package syntax

// InsertBefore returns a new node with newChild inserted before the child at index i.
func InsertBefore(parent *GreenNode, i int, newChild GreenElement) *GreenNode {
	if i < 0 || i > len(parent.Children) {
		return nil
	}
	newChildren := make([]GreenElement, 0, len(parent.Children)+1)
	newChildren = append(newChildren, parent.Children[:i]...)
	newChildren = append(newChildren, newChild)
	newChildren = append(newChildren, parent.Children[i:]...)
	return NewGreenNode(parent.Kind, newChildren)
}

// InsertAfter returns a new node with newChild inserted after the child at index i.
func InsertAfter(parent *GreenNode, i int, newChild GreenElement) *GreenNode {
	if i < -1 || i >= len(parent.Children) {
		return nil
	}
	return InsertBefore(parent, i+1, newChild)
}

// RemoveChild returns a new node with the child at index i removed.
func RemoveChild(parent *GreenNode, i int) *GreenNode {
	if i < 0 || i >= len(parent.Children) {
		return nil
	}
	newChildren := make([]GreenElement, 0, len(parent.Children)-1)
	newChildren = append(newChildren, parent.Children[:i]...)
	newChildren = append(newChildren, parent.Children[i+1:]...)
	return NewGreenNode(parent.Kind, newChildren)
}

// TransformChildren returns a new node with each child transformed by fn.
// If fn returns the same element, it is shared (structural sharing).
// If fn returns nil, the child is removed.
func TransformChildren(parent *GreenNode, fn func(GreenElement) *GreenElement) *GreenNode {
	changed := false
	var newChildren []GreenElement

	for i, child := range parent.Children {
		result := fn(child)
		if result == nil {
			// Remove this child.
			if !changed {
				changed = true
				newChildren = make([]GreenElement, i, len(parent.Children))
				copy(newChildren, parent.Children[:i])
			}
			continue
		}
		if *result != child {
			if !changed {
				changed = true
				newChildren = make([]GreenElement, i, len(parent.Children))
				copy(newChildren, parent.Children[:i])
			}
		}
		if changed {
			newChildren = append(newChildren, *result)
		}
	}

	if !changed {
		return parent
	}
	return NewGreenNode(parent.Kind, newChildren)
}

// MapNodes recursively transforms all nodes in the tree using fn.
// fn receives each node and returns a replacement. If fn returns the
// same node, it is structurally shared.
func MapNodes(root *GreenNode, fn func(*GreenNode) *GreenNode) *GreenNode {
	changed := false
	newChildren := make([]GreenElement, len(root.Children))

	for i, child := range root.Children {
		if child.Node != nil {
			mapped := MapNodes(child.Node, fn)
			mapped = fn(mapped)
			if mapped != child.Node {
				newChildren[i] = NodeElement(mapped)
				changed = true
				continue
			}
		}
		newChildren[i] = child
	}

	result := root
	if changed {
		result = NewGreenNode(root.Kind, newChildren)
	}
	return fn(result)
}
