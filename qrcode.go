package main

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

func getQRString(toEncode string) string {

	qrc, err := qrcode.New(toEncode, qrcode.Low)

	if err != nil {
		logError(err.Error())
		return ""
	}

	var buf strings.Builder

	bm := qrc.Bitmap()

	for x := 4; x < len(bm)-4; x += 2 {
		for y := 4; y < len(bm[x])-4; y += 2 {
			buf.WriteString(getQRSquare(bm, x, y))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func getQRSquare(s [][]bool, x int, y int) string {
	a11 := s[y][x]
	a21 := s[y][x+1]
	a12 := s[y+1][x]
	a22 := s[y+1][x+1]
	if a11 {
		if a12 {
			if a21 {
				if a22 {
					return "██"
				}
				return "█▀"
			}
			if a22 {
				return "▀█"
			}
			return "▀▀"
		}
		if a21 {
			if a22 {
				return "█▄"
			}
			return "█ "
		}
		if a22 {
			return "▀▄"
		}
		return "▀ "
	}
	if a12 {
		if a21 {
			if a22 {
				return "▄█"
			}
			return "▄▀"
		}
		if a22 {
			return " █"
		}
		return " ▀"
	}
	if a21 {
		if a22 {
			return "▄▄"
		}
		return "▄ "
	}
	if a22 {
		return " ▄"
	}
	return "  "
}
