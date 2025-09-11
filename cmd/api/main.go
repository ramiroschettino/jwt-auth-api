package main

import (
	"fmt"
	"jwt-auth-api/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Printf("Configuración: %+v\n", cfg)
}
