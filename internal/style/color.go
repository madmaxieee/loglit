package style

import "fmt"

const ESCAPE = "\033["

const ResetAllAnsi = ESCAPE + "0m"

const ResetFgAnsi = ESCAPE + "39m"

const ResetBgAnsi = ESCAPE + "49m"

func Fg(r int, g int, b int) string {
	return fmt.Sprintf("%s38;2;%d;%d;%dm", ESCAPE, r, g, b)
}

func FgHex(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return Fg(r, g, b)
}

func Bg(r int, g int, b int) string {
	return fmt.Sprintf("%s48;2;%d;%d;%dm", ESCAPE, r, g, b)
}

func BgHex(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return Bg(r, g, b)
}
