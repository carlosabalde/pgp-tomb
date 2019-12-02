package parser

import (
	"fmt"
)

type Tree interface {
	Value() token
	Left() Tree
	Right() Tree
}

type tree struct {
	value token
	left  *tree
	right *tree
}

func (t *tree) Value() token {
	return t.value
}

func (t *tree) Left() Tree {
	return t.left
}

func (t *tree) Right() Tree {
	return t.right
}

func (t *tree) String() string {
	var left, right string
	if t.left != nil {
		left = t.left.String()
	}
	if t.right != nil {
		right = t.right.String()
	}
	return fmt.Sprintf("(%s, %s, %s)", left, t.value, right)
}

func newTree() *tree {
	return &tree{}
}
