package pokeapi

// BaseStat describes a PokeAPI base stat.
type BaseStat struct {
	BaseStat *int `json:"base_stat"`
	Effort   *int `json:"effort"`
	Stat     *struct {
		Name *string `json:"name"`
		URL  *string `json:"url"`
	} `json:"stat"`
}
