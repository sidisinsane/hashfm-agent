package model_test

import (
	"testing"

	"github.com/sidisinsane/hashfm-agent/internal/model"
)

func TestBlock_IsMulti_WithMulti(t *testing.T) {
	block := model.Block{
		Multi: []model.Command{
			{Description: "First", Usage: "foo.sh first", Exits: map[string]string{"0": "success"}},
			{Description: "Second", Usage: "foo.sh second", Exits: map[string]string{"0": "success"}},
		},
	}
	if !block.IsMulti() {
		t.Error("expected IsMulti() = true for block with Multi set")
	}
}

func TestBlock_IsMulti_WithSingle(t *testing.T) {
	block := model.Block{
		Single: &model.Command{
			Description: "Does a thing",
			Usage:       "foo.sh <arg>",
			Exits:       map[string]string{"0": "success"},
		},
	}
	if block.IsMulti() {
		t.Error("expected IsMulti() = false for block with Single set")
	}
}

// IsMulti checks for a non-nil Multi slice. An empty non-nil slice is
// considered multi — the validator in pipeline rejects it before it can
// reach a generator, so this state should never occur in practice.
func TestBlock_IsMulti_EmptyNonNilSlice(t *testing.T) {
	block := model.Block{
		Multi: []model.Command{},
	}
	if !block.IsMulti() {
		t.Error("expected IsMulti() = true for non-nil empty Multi slice")
	}
}