package renderer

import (
	"reflect"
	"testing"
)

func TestStackOverlapRight(t *testing.T) {
	// Top: [10, 20]
	// Bottom: [15, 25]
	// Expected: [10, 20] (Top), [20, 25] (Bottom trimmed)

	top := MatchLayer{{Start: 10, End: 20, AnsiStart: "T", AnsiEnd: "t"}}
	bottom := MatchLayer{{Start: 15, End: 25, AnsiStart: "B", AnsiEnd: "b"}}

	result := Stack(top, bottom)

	expected := MatchLayer{
		{Start: 10, End: 20, AnsiStart: "T", AnsiEnd: "t"},
		{Start: 20, End: 25, AnsiStart: "B", AnsiEnd: "b"},
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d matches, got %d", len(expected), len(result))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Match %d mismatch:\nExpected: %+v\nGot:      %+v", i, expected[i], result[i])
		}
	}
}

func TestStackOverlapLeft(t *testing.T) {
	// Top: [20, 30]
	// Bottom: [15, 25]
	// Expected: [15, 20] (Bottom trimmed), [20, 30] (Top)

	top := MatchLayer{{Start: 20, End: 30, AnsiStart: "T", AnsiEnd: "t"}}
	bottom := MatchLayer{{Start: 15, End: 25, AnsiStart: "B", AnsiEnd: "b"}}

	result := Stack(top, bottom)

	expected := MatchLayer{
		{Start: 15, End: 20, AnsiStart: "B", AnsiEnd: "b"},
		{Start: 20, End: 30, AnsiStart: "T", AnsiEnd: "t"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Mismatch:\nExpected: %+v\nGot:      %+v", expected, result)
	}
}

func TestStackEnclosedTop(t *testing.T) {
	// Top: [20, 25]
	// Bottom: [15, 30]
	// Expected: [15, 20] (Bottom left), [20, 25] (Top), [25, 30] (Bottom right)

	top := MatchLayer{{Start: 20, End: 25, AnsiStart: "T", AnsiEnd: "t"}}
	bottom := MatchLayer{{Start: 15, End: 30, AnsiStart: "B", AnsiEnd: "b"}}

	result := Stack(top, bottom)

	expected := MatchLayer{
		{Start: 15, End: 20, AnsiStart: "B", AnsiEnd: "b"},
		{Start: 20, End: 25, AnsiStart: "T", AnsiEnd: "t"},
		{Start: 25, End: 30, AnsiStart: "B", AnsiEnd: "b"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Mismatch:\nExpected: %+v\nGot:      %+v", expected, result)
	}
}

func TestStackDisjoint(t *testing.T) {
	// Top: [10, 20]
	// Bottom: [30, 40]
	// Expected: [10, 20], [30, 40]

	top := MatchLayer{{Start: 10, End: 20, AnsiStart: "T", AnsiEnd: "t"}}
	bottom := MatchLayer{{Start: 30, End: 40, AnsiStart: "B", AnsiEnd: "b"}}

	result := Stack(top, bottom)

	expected := MatchLayer{
		{Start: 10, End: 20, AnsiStart: "T", AnsiEnd: "t"},
		{Start: 30, End: 40, AnsiStart: "B", AnsiEnd: "b"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Mismatch:\nExpected: %+v\nGot:      %+v", expected, result)
	}
}
