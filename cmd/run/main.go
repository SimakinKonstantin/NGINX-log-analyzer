package main

import (
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/session"
)

func main() {
	err := session.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
