package plugin

import (
	"go.blockdaemon.com/bpm/sdk/pkg/node"
)

// DummyTester does nothing except panicking
//
// This Tester can be used if the plugin doesn't support testing
type DummyTester struct{}

func (t DummyTester) Test(currentNode node.Node) (bool, error) {
	panic("Not implemented")
}

func NewDummyTester() DummyTester {
	return DummyTester{}
}
