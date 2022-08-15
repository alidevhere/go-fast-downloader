package downloader

import (
	"fmt"
)

var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

func logErr(format string, a ...any) {

	s := fmt.Sprintf(format, a...)
	fmt.Println(string(Red), s, string(Reset))
}

func logInfo(format string, a ...any) {

	s := fmt.Sprintf(format, a...)
	fmt.Println(string(Blue), s, string(Reset))
}

func logSuccess(format string, a ...any) {

	s := fmt.Sprintf(format, a...)
	fmt.Println(string(Green), s, string(Reset))
}

func logWarn(format string, a ...any) {

	s := fmt.Sprintf(format, a...)
	fmt.Println(string(Yellow), s, string(Reset))
}
