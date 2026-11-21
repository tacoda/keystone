package main

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestLadderSteps(t *testing.T) {
	up, ok := ladderStep("corpus", +1)
	if !ok || up != primitive.KindReference {
		t.Errorf("corpus+1 = %s,%v", up, ok)
	}
	up, ok = ladderStep("reference", +1)
	if !ok || up != primitive.KindRule {
		t.Errorf("reference+1 = %s,%v", up, ok)
	}
	if _, ok := ladderStep("rule", +1); ok {
		t.Error("rule is the top of the ladder")
	}
	down, ok := ladderStep("rule", -1)
	if !ok || down != primitive.KindReference {
		t.Errorf("rule-1 = %s,%v", down, ok)
	}
	if _, ok := ladderStep("corpus", -1); ok {
		t.Error("corpus is the bottom of the ladder")
	}
	if _, ok := ladderStep("hook", +1); ok {
		t.Error("lateral kinds are not ladder kinds")
	}
}
