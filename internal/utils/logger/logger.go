package logger

import (
	"fmt"

	"github.com/fatih/color"
)

func Info(format string, args ...any) {
	color.Blue("[INFO] ")
	fmt.Printf(format+"\n", args...)
}

func Debug(format string, args ...any) {
	color.Green("[DEBUG] ")
	fmt.Printf(format+"\n", args...)
}

func Error(format string, args ...any) {
	color.Red("[ERROR]")
	fmt.Printf(format+"\n", args...)
}
