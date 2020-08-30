package main

import (
	"github.com/gorilla/mux"
	"github.com/mtslzr/pokeapi-go"
	_ "github.com/mtslzr/pokeapi-go"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

//templates
var tmpl = template.Must(template.ParseFiles("template/home.html"))
var detailedtmpl = template.Must(template.ParseFiles("template/detailedPokemon.html"))
var listtmp = template.Must(template.ParseFiles("template/listPokemon.html"))

//pokemon names
var pokemonNames = getNames()

//structs
type Pokemon struct {
	Name        string
	Abilities   []Pokemon_Abilty
	Stats       Stat
	Types       []Pokemon_Type
	ImageFront  string
	ImageBack   string
	ImageShiny  string
	Description string
}

type PokemonShort struct {
	Name        string
	Image       string
	Description string
}

type Stat struct {
	HP              int
	Attack          int
	Defense         int
	Special_Attack  int
	Special_Defense int
	Speed           int
}

type Stats []struct {
	BaseStat int `json:"base_stat"`
	Effort   int `json:"effort"`
	Stat     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"stat"`
}

type Pokemon_Abilty struct {
	Name string
}

type Pokemon_Type struct {
	Type string
}

type Types []struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`
}

type Abilities []struct {
	Ability struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"ability"`
	IsHidden bool `json:"is_hidden"`
	Slot     int  `json:"slot"`
}

//functions for templates and frontend
func homepage(writer http.ResponseWriter, request *http.Request) {
	if request.FormValue("pokemon_selected") != "" {
		name := request.FormValue("pokemon_selected")
		if Find(pokemonNames, name) {
			http.Redirect(writer, request, "/"+name, http.StatusFound)
		} else {
			http.Redirect(writer, request, request.Header.Get("Referer"), 302)
		}
	}
	_ = tmpl.Execute(writer, pokemonNames)
}

func detailedPokemon(writer http.ResponseWriter, request *http.Request) {
	//Fetch details about selected pokemon
	params := mux.Vars(request)
	detail, _ := pokeapi.Pokemon(params["string"])
	desc, _ := pokeapi.PokemonSpecies(params["string"])
	//make the smaller structs of the pokemon
	stats := Stat{}
	name := strings.Title(detail.Name)
	abilities := getAbilities(detail.Abilities)
	stats = getStat(detail.Stats)
	imageFront := detail.Sprites.FrontDefault
	imageBack := detail.Sprites.BackShiny
	imageShiny := detail.Sprites.FrontShiny
	types := getTypes(detail.Types)
	var description string
	if len(desc.FlavorTextEntries) > 0 {
		description = desc.FlavorTextEntries[0].FlavorText
	} else {
		description = "No Description from API Received"
	}
	//make the Pokemon struct using those from above
	pokemon := Pokemon{
		Name:        name,
		Abilities:   abilities,
		Stats:       stats,
		Types:       types,
		ImageFront:  imageFront,
		ImageBack:   imageBack,
		ImageShiny:  imageShiny,
		Description: description,
	}
	_ = detailedtmpl.Execute(writer, pokemon)
}

func listview(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	i, _ := strconv.Atoi(vars["id"])
	pokemonRange := make([]PokemonShort, 20)
	j := 0
	start := (i * 20) - 20
	end := start + 20
	if end >= 1048 {
		end = 1048
	}
	for start < end {
		pokemon := pokemonNames[start]
		detail, _ := pokeapi.Pokemon(pokemon)
		desc, _ := pokeapi.PokemonSpecies(pokemon)
		var description string
		if len(desc.FlavorTextEntries) > 0 {
			description = desc.FlavorTextEntries[0].FlavorText
		} else {
			description = "No Description from API Received"
		}
		tmp := PokemonShort{
			Name:        detail.Name,
			Image:       detail.Sprites.FrontDefault,
			Description: description,
		}
		pokemonRange[j] = tmp
		start++
		j++
	}
	_ = listtmp.Execute(writer, pokemonRange)
}

/*
The below functions will mess with the API and get us the desired results (hopefully)
*/

//this gets all the stats of the pokemon the user wants to look at
func getStat(Stats Stats) Stat {
	pokemonStat := Stat{
		HP:              Stats[0].BaseStat,
		Attack:          Stats[1].BaseStat,
		Defense:         Stats[2].BaseStat,
		Special_Attack:  Stats[3].BaseStat,
		Special_Defense: Stats[4].BaseStat,
		Speed:           Stats[5].BaseStat,
	}
	return pokemonStat
}

//this gets all the types the selected pokemon is of
func getTypes(types Types) []Pokemon_Type {
	pokemonTypes := make([]Pokemon_Type, 0)
	for i := range types {
		tmp := Pokemon_Type{Type: strings.Title(types[i].Type.Name)}
		pokemonTypes = append(pokemonTypes, tmp)
	}
	return pokemonTypes
}

//this gets all the abilities of the pokemon the user wants to look at
func getAbilities(abilities Abilities) []Pokemon_Abilty {
	pokemonAbilities := make([]Pokemon_Abilty, 0)
	for i := range abilities {
		tmp := Pokemon_Abilty{Name: strings.Title(abilities[i].Ability.Name)}
		pokemonAbilities = append(pokemonAbilities, tmp)
	}
	return pokemonAbilities
}
func Find(names []string, val string) bool {
	for _, item := range names {
		if item == val {
			return true
		}
	}
	return false
}

func getNames() []string {
	names := make([]string, 1048)
	l, _ := pokeapi.Resource("pokemon", 0, 1048)
	for i := range l.Results {
		names[i] = l.Results[i].Name
	}
	return names
}

//the main function that runs everything
func main() {
	r := mux.NewRouter()
	//Website pages to handle
	r.HandleFunc("/", homepage).Methods("POST", "GET")
	r.HandleFunc("/list/{id}", listview).Methods("POST", "GET")
	r.HandleFunc("/{string}", detailedPokemon).Methods("POST", "GET")
	//Handle Static Files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("template/static"))))

	http.Handle("/", r)
	_ = http.ListenAndServe(":80", r)
}
