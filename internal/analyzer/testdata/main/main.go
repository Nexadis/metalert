package main

import (
	"os"
)

func some() {
	os.Exit(1)
}

func main() {
	// формулируем ожидания: анализатор должен находить ошибку,
	// описанную в комментарии want
	os.Exit(0) // want "don't use exit in main"
}
