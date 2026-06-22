package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/frevens/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
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
	"explore": {
		name:        "explore",
		description: "Explore an area location",
		callback:    commandExplore,
	},
	"catch": {
		name:        "catch",
		description: "Attempt to catch a pokemon",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "Inspect a caught pokemon",
		callback:    commandInspect,
	},
	"pokedex": {
		name:        "pokedex",
		description: "List your caught pokemon",
		callback:    commandPokedex,
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
	pokedex     map[string]Pokemon
}

type ExploreResponse struct {
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEncounter struct {
	Pokemon PokemonReference `json:"pokemon"`
}

type PokemonReference struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Pokemon struct {
	Name           string        `json:"name"`
	BaseExperience int           `json:"base_experience"`
	Height         int           `json:"height"`
	Weight         int           `json:"weight"`
	Stats          []PokemonStat `json:"stats"`
	Types          []PokemonType `json:"types"`
}
type PokemonStat struct {
	BaseStat int              `json:"base_stat"`
	Stat     NamedAPIResource `json:"stat"`
}

type PokemonType struct {
	Type NamedAPIResource `json:"type"`
}

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // Nunca se ejecuta, pero satisface al compilador.
}

func commandHelp(cfg *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("map: Displays the next 20 locations")
	fmt.Println("mapb: Displays the previous 20 locations")
	fmt.Println("explore: Explores an area location")
	fmt.Println("catch: Attempt to catch a pokemon")
	fmt.Println("inspect: Inspect a caught pokemon")
	fmt.Println("pokedex: List your caught pokemon")

	return nil
}
func commandMap(cfg *config, args []string) error {
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

func commandMapb(cfg *config, args []string) error {
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

func commandExplore(cfg *config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("please provide a location area")
	}
	area := args[1]

	url := "https://pokeapi.co/api/v2/location-area/" + area + "/"

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

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("location area not found")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		cfg.cache.Add(url, body)
		data = body
	}
	var response ExploreResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", area)
	fmt.Println("Found Pokemon:")
	for _, p := range response.PokemonEncounters {
		fmt.Printf(" - %s\n", p.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing pokemon name")
	}
	pokemonName := args[1]

	url := "https://pokeapi.co/api/v2/pokemon/" + pokemonName + "/"

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	var data []byte
	if cached, ok := cfg.cache.Get(url); ok {
		data = cached

	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("pokemon not found")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		cfg.cache.Add(url, body)
		data = body
	}
	var catchingPokemon Pokemon
	if err := json.Unmarshal(data, &catchingPokemon); err != nil {
		return err
	}
	chance := 300 - catchingPokemon.BaseExperience
	if chance < 20 {
		chance = 20
	}
	roll := rand.Intn(300)
	if roll < chance {
		cfg.pokedex[catchingPokemon.Name] = catchingPokemon
		fmt.Printf("%s was caught!\n", catchingPokemon.Name)
		fmt.Println("You may now inspect it with the inspect command.")

	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}
	return nil
}

func commandInspect(cfg *config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing pokemon name")
	}

	name := args[1]

	pokemon, ok := cfg.pokedex[name]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, args []string) error {
	fmt.Println("Your Pokedex:")

	if len(cfg.pokedex) == 0 {
		fmt.Println(" (empty)")
		return nil
	}

	for name := range cfg.pokedex {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := config{
		cache:   pokecache.NewCache(5 * time.Minute),
		pokedex: make(map[string]Pokemon),
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

		if err := cmd.callback(&cfg, palabras); err != nil {
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
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
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
