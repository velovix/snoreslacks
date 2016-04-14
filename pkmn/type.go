package pkmn

import "strings"

// TypeMod is a single type modifier. It contains a defense modifier and the
// type that the defense modifier applies to.
type TypeMod struct {
	T   Type
	Val float64
}

// Type is a single type which contains a slIce of defense modifiers for
// every type where the modifier would not equal 1.0.
type Type struct {
	Name string
	Mods []TypeMod
}

var Normal, Fighting, Flying, Poison, Ground, Rock, Bug,
	Ghost, Steel, Fire, Water, Grass, Electricity, Psychic,
	Ice, Dragon, Dark, Fairy Type

var nameToType map[string]Type

func (t Type) Mod(attackType string) float64 {
	mod := 1.0
	for _, val := range t.Mods {
		if val.T.Name == attackType {
			mod *= val.Val
		}
	}

	return mod
}

// NameToType returns the Type object for a given type name, case
// insensitively, and false if the type could not be found.
func NameToType(name string) (Type, bool) {
	t, ok := nameToType[strings.ToLower(name)]
	return t, ok
}

// Initialize the Values of each type.
func init() {
	nameToType = make(map[string]Type)

	// Fill out each type's defense modifiers
	Normal.Name = "normal"
	Normal.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 2.0},
		TypeMod{T: Ghost, Val: 0.0}}
	nameToType[Normal.Name] = Normal
	Fighting.Name = "fighting"
	Fighting.Mods = []TypeMod{
		TypeMod{T: Bug, Val: 0.5},
		TypeMod{T: Dark, Val: 0.5},
		TypeMod{T: Rock, Val: 0.5},
		TypeMod{T: Fairy, Val: 2.0},
		TypeMod{T: Flying, Val: 2.0},
		TypeMod{T: Psychic, Val: 2.0}}
	nameToType[Fighting.Name] = Fighting
	Flying.Name = "flying"
	Flying.Mods = []TypeMod{
		TypeMod{T: Bug, Val: 0.5},
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Electricity, Val: 2.0},
		TypeMod{T: Ice, Val: 2.0},
		TypeMod{T: Rock, Val: 2.0},
		TypeMod{T: Ground, Val: 0.0}}
	nameToType[Flying.Name] = Flying
	Poison.Name = "poison"
	Poison.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Poison, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Ground, Val: 2.0},
		TypeMod{T: Bug, Val: 2.0},
		TypeMod{T: Psychic, Val: 2.0}}
	nameToType[Poison.Name] = Poison
	Ground.Name = "ground"
	Ground.Mods = []TypeMod{
		TypeMod{T: Poison, Val: 0.5},
		TypeMod{T: Rock, Val: 0.5},
		TypeMod{T: Grass, Val: 2.0},
		TypeMod{T: Ice, Val: 2.0},
		TypeMod{T: Water, Val: 2.0},
		TypeMod{T: Electricity, Val: 0.0}}
	nameToType[Ground.Name] = Ground
	Rock.Name = "rock"
	Rock.Mods = []TypeMod{
		TypeMod{T: Fire, Val: 0.5},
		TypeMod{T: Flying, Val: 0.5},
		TypeMod{T: Normal, Val: 0.5},
		TypeMod{T: Poison, Val: 0.5},
		TypeMod{T: Fighting, Val: 2.0},
		TypeMod{T: Grass, Val: 2.0},
		TypeMod{T: Ground, Val: 2.0},
		TypeMod{T: Steel, Val: 2.0},
		TypeMod{T: Water, Val: 2.0}}
	nameToType[Rock.Name] = Rock
	Bug.Name = "bug"
	Bug.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Ground, Val: 0.5},
		TypeMod{T: Fire, Val: 2.0},
		TypeMod{T: Flying, Val: 2.0},
		TypeMod{T: Rock, Val: 2.0}}
	nameToType[Bug.Name] = Bug
	Ghost.Name = "ghost"
	nameToType[Ghost.Name] = Ghost
	Steel.Name = "steel"
	nameToType[Steel.Name] = Steel
	Fire.Name = "fire"
	nameToType[Fire.Name] = Fire
	Water.Name = "water"
	nameToType[Water.Name] = Water
	Grass.Name = "grass"
	nameToType[Grass.Name] = Grass
	Electricity.Name = "elecricity"
	nameToType[Electricity.Name] = Electricity
	Psychic.Name = "psychic"
	nameToType[Psychic.Name] = Psychic
	Ice.Name = "ice"
	nameToType[Ice.Name] = Ice
	Dragon.Name = "dragon"
	nameToType[Dragon.Name] = Dragon
	Dark.Name = "dark"
	nameToType[Dark.Name] = Dark
	Fairy.Name = "fairy"
	nameToType[Fairy.Name] = Fairy
}
