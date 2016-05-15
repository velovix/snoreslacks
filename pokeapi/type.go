package pokeapi

// Type describes a PokeAPI type.
type Type struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Accuracy     int    `json:"accuracy"`
	EffectChance int    `json:"effect_chance"`
	PP           int    `json:"pp"`
	Power        int    `json:"power"`
	DamageClass  struct {
		Name string `json:"name"`
	} `json:"damage_class"`
	EffectEntries struct {
		Effect string `json:"effect"`
	} `json:"effect_entries"`
	Type struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`
}
