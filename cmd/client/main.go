package main

import (
	"log"
)

func main() {
	err := executeCommand()
	if err != nil {
		log.Fatalf("Ошибка выполнения команды: %v", err)
	}
}
