package server

import "grammar/token"

func PosToPosition(offset token.Pos, srcRunes []rune) Position {
	line, col := token.FindLineAndCol(int(offset), srcRunes)
	return Position{Line: line - 1, Character: col - 1}
}

func PositionToPos(pos Position, srcRunes []rune) token.Pos {
	if pos.Line < 0 {
		return 0
	}
	lineStart := 0
	for i := 0; i < pos.Line; i++ {
		found := false
		for j := lineStart; j < len(srcRunes); j++ {
			if srcRunes[j] == '\n' {
				lineStart = j + 1
				found = true
				break
			}
		}
		if !found {
			return token.Pos(len(srcRunes))
		}
	}
	if lineStart+pos.Character > len(srcRunes) {
		return token.Pos(len(srcRunes))
	}
	return token.Pos(lineStart + pos.Character)
}

