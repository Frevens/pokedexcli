package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode"
	"time"
	"github.com/frevens/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
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
	"map": {
		name:        "map",
		description: "Displays the next 20 locations",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "displays the previous 20 locations",
		callback:    commandMapb,
	},
}

type LocationAreaResponse struct {
	Count int `json:"count"`

	Next *string `json:"next"`

	Previous *string `json:"previous"`

	Results []LocationArea `json:"results"`
}

type LocationArea struct {
	Name string `json:"name"`

	URL string `json:"url"`
}

type config struct {
	nextURL     *string
	previousURL *string
	cache       *pokecache.Cache
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // Nunca se ejecuta, pero satisface al compilador.
}

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("map: Displays the next 20 locations")
	fmt.Println("mapb: Displays the previous 20 locations")

	return nil
}
func commandMap(cfg *config) error {
	var url string
	if cfg.nextURL == nil {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = *cfg.nextURL
	}

	var data []byte
	if cached, ok := cfg.cache.Get(url); ok {
		data = cached
		fmt.Println("using cached data")
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cfg.cache.Add(url, body)
		data = body
	}

	var response LocationAreaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	cfg.nextURL = response.Next
	cfg.previousURL = response.Previous
	for _, area := range response.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapb(cfg *config) error {
	if cfg.previousURL == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	url := *cfg.previousURL

	var data []byte
	if cached, ok := cfg.cache.Get(url); ok {
		data = cached
		fmt.Println("using cached data")
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cfg.cache.Add(url, body)
		data = body
	}

	var response LocationAreaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	cfg.nextURL = response.Next
	cfg.previousURL = response.Previous
	for _, area := range response.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := config{
		cache: pokecache.NewCache(5 * time.Minute),
	}

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

		if err := cmd.callback(&cfg); err != nil {
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
