package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands = map[string]cliCommand{
	"help": {
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	},
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // Nunca se ejecuta, pero satisface al compilador.
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("help: Displays a help message")
	fmt.Println("exit: Exit the Pokedex")
	return nil
}
func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Mostrar el prompt
		fmt.Print("Pokedex > ")

		// Leer una línea
		if !scanner.Scan() {
			break
		}

		// Limpiar el input
		palabras := cleanInput(scanner.Text())

		// Ignorar líneas vacías
		if len(palabras) == 0 {
			continue
		}

		// Buscar el comando
		comando := palabras[0]
		cmd, ok := commands[comando]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		// Ejecutar el callback
		if err := cmd.callback(); err != nil {
			fmt.Println(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}

func cleanInput(text string) []string {
	text = strings.ToLower(strings.TrimSpace(text))

	if text == "" {
		return []string{}
	}

	var palabras []string
	var actual strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			actual.WriteRune(r)
		} else if unicode.IsSpace(r) {
			if actual.Len() > 0 {
				palabras = append(palabras, actual.String())
				actual.Reset()
			}
		}
	}

	if actual.Len() > 0 {
		palabras = append(palabras, actual.String())
	}

	return palabras
}
