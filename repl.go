package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// 1. Imprimir prompt sin salto de línea
		fmt.Print("Pokedex > ")

		// 2. Esperar input del usuario (bloqueante)
		if !scanner.Scan() {
			// Si hay error o EOF (Ctrl+D/Ctrl+Z), salir del loop
			break
		}

		input := scanner.Text()

		// 3. Limpiar el input
		palabras := cleanInput(input)

		// 4. Capturar la primera palabra y responder
		if len(palabras) > 0 {
			comando := palabras[0]
			fmt.Printf("Your command was: %s\n", comando)
		} else {
			// Opcional: manejar input vacío
			fmt.Println("Comando vacío.")
		}
	}
}

func cleanInput(text string) []string {
	// 1. Pasar a minúsculas y limpiar espacios extremos
	text = strings.ToLower(strings.TrimSpace(text))

	if text == "" {
		return []string{}
	}

	var palabras []string
	var actual strings.Builder

	// 2. Iterar rune por rune para manejar UTF-8 correctamente
	for _, r := range text {
		// Si es letra o número, lo agregamos a la palabra actual
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			actual.WriteRune(r)
		} else if unicode.IsSpace(r) {
			// Si es espacio y tenemos una palabra acumulada, la guardamos
			if actual.Len() > 0 {
				palabras = append(palabras, actual.String())
				actual.Reset()
			}
		}
		// Si es un signo de puntuación, simplemente lo ignoramos (no hacemos nada)
	}

	// 3. Agregar la última palabra si existe
	if actual.Len() > 0 {
		palabras = append(palabras, actual.String())
	}

	return palabras
}
