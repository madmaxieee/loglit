package style

import (
	"strings"

	"github.com/madmaxieee/loglit/internal/utils"
	"github.com/pelletier/go-toml/v2"
)

type Highlight struct {
	Group     string
	Link      *string
	Fg        *string
	Bg        *string
	Italic    bool
	Bold      bool
	Underline bool
	ansi      *string
}

func (h *Highlight) UnmarshalText(text []byte) error {
	err := toml.Unmarshal(text, h)
	if err != nil {
		return err
	}
	if h.Link == nil {
		if h.Fg != nil {
			h.Fg = utils.Ptr(FgHex(*h.Fg))
		}
		if h.Bg != nil {
			h.Bg = utils.Ptr(BgHex(*h.Bg))
		}
	} else {
		h.Fg = nil
		h.Bg = nil
	}
	return nil
}

func (h Highlight) BuildAnsi() string {
	if h.ansi != nil {
		return *h.ansi
	}
	var b strings.Builder
	if h.Fg != nil {
		b.WriteString(*h.Fg)
	}
	if h.Bg != nil {
		b.WriteString(*h.Bg)
	}
	if h.Bold {
		b.WriteString(BoldAnsi)
	}
	if h.Italic {
		b.WriteString(ItalicAnsi)
	}
	if h.Underline {
		b.WriteString(UnderlineAnsi)
	}
	h.ansi = utils.Ptr(b.String())
	return *h.ansi
}

func (h Highlight) HasExtraStyle() bool {
	return h.Bold || h.Italic || h.Underline
}
