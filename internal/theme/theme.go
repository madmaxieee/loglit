package theme

import (
	"fmt"

	"github.com/madmaxieee/loglit/internal/style"
	"github.com/madmaxieee/loglit/internal/utils"
)

type highlight = style.Highlight

type Theme struct {
	Name         string
	HighlightMap map[string]highlight
	linked       bool
}

func fg(raw string) *string {
	return utils.Ptr(style.FgHex(raw))
}

var DefaultTheme = Theme{
	Name:   "default",
	linked: false,
	HighlightMap: map[string]highlight{
		"Constant": {
			Group: "Constant",
			Fg:    fg("#FF966C"),
		},
		"Number": {
			Group: "Number",
			Link:  utils.Ptr("Constant"),
		},
		"Float": {
			Group: "Float",
			Link:  utils.Ptr("Number"),
		},
		"Special": {
			Group: "Special",
			Fg:    fg("#65BCFF"),
		},
		"Comment": {
			Group:  "Comment",
			Fg:     fg("#636DA6"),
			Italic: true,
		},
		"Boolean": {
			Group: "Boolean",
			Link:  utils.Ptr("Constant"),
		},
		"String": {
			Group: "String",
			Fg:    fg("#C3E88D"),
		},
		"Type": {
			Group: "Type",
			Fg:    fg("#65BCFF"),
		},
		"Operator": {
			Group: "Operator",
			Fg:    fg("#89DDFF"),
		},
		"Statement": {
			Group: "Statement",
			Fg:    fg("#C099FF"),
		},
		"Function": {
			Group: "Function",
			Fg:    fg("#82AAFF"),
		},
		"Underlined": {
			Group:     "Underlined",
			Underline: true,
		},
		"Label": {
			Group: "Label",
			Link:  utils.Ptr("Statement"),
		},
		"Structure": {
			Group: "Structure",
			Link:  utils.Ptr("Type"),
		},
		"ErrorMsg": {
			Group: "ErrorMsg",
			Fg:    fg("#C53B53"),
		},
		"WarningMsg": {
			Group: "WarningMsg",
			Fg:    fg("#FFC777"),
		},
		"Exception": {
			Group: "Exception",
			Link:  utils.Ptr("Statement"),
		},
		"Debug": {
			Group: "Debug",
			Fg:    fg("#FF966C"),
		},
		"LogGreen": {
			Group: "LogGreen",
			Fg:    fg("#C3E88D"),
		},
		"LogBlue": {
			Group: "LogBlue",
			Fg:    fg("#65BCFF"),
		},
	},
}

func GetDefaultTheme() Theme {
	return DefaultTheme
}

// TODO: handle cycle linking
func (t *Theme) Resolve(name string) error {
	if t.linked {
		return nil
	}

	hl, ok := t.HighlightMap[name]
	if !ok {
		return fmt.Errorf("highlight %q not found", name)
	}

	if hl.Link == nil {
		return nil
	}

	targetName := *hl.Link
	targetHl, ok := t.HighlightMap[targetName]
	if !ok {
		return fmt.Errorf("highlight link target %q not found", targetName)
	}

	if targetHl.Link != nil {
		if err := t.Resolve(targetName); err != nil {
			return err
		}
	}

	hl.Fg = targetHl.Fg
	hl.Bg = targetHl.Bg
	hl.Link = nil
	t.HighlightMap[name] = hl

	return nil
}

func (t *Theme) ResolveLinks() error {
	if t.linked {
		return nil
	}
	for name, hl := range t.HighlightMap {
		if hl.Link != nil {
			err := t.Resolve(name)
			if err != nil {
				return err
			}
		}
	}
	t.linked = true
	return nil
}

func (t *Theme) Insert(hl highlight) {
	t.HighlightMap[hl.Group] = hl
	if hl.Link != nil {
		t.linked = false
	}
}

func (t *Theme) GetHighlight(group string) (highlight, bool) {
	hl, ok := t.HighlightMap[group]
	return hl, ok
}
