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

func (self *tree) Value() token {
	return self.value
}

func (self *tree) Left() Tree {
	return self.left
}

func (self *tree) Right() Tree {
	return self.right
}

func (self *tree) String() string {
	var left, right string
	if self.left != nil {
		left = self.left.String()
	}
	if self.right != nil {
		right = self.right.String()
	}
	return fmt.Sprintf("(%s, %s, %s)", left, self.value, right)
}

func newTree() *tree {
	return &tree{}
}
