package splitfile

import "fmt"

func FormatChunkFilename(mainFilename string, n int) string {
	return fmt.Sprintf("%s.split-%03d", mainFilename, n)
}

func FormatSplitfileFilename(mainFilename string) string {
	return fmt.Sprintf("%s.split", mainFilename)
}
