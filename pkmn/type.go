package pkmn

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

// Initialize the Values of each type.
func init() {
	// Fill out each type's defense modifiers
	Normal.Name = "normal"
	Normal.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 2.0},
		TypeMod{T: Ghost, Val: 0.0}}
	Fighting.Name = "fighting"
	Fighting.Mods = []TypeMod{
		TypeMod{T: Bug, Val: 0.5},
		TypeMod{T: Dark, Val: 0.5},
		TypeMod{T: Rock, Val: 0.5},
		TypeMod{T: Fairy, Val: 2.0},
		TypeMod{T: Flying, Val: 2.0},
		TypeMod{T: Psychic, Val: 2.0}}
	Flying.Name = "flying"
	Flying.Mods = []TypeMod{
		TypeMod{T: Bug, Val: 0.5},
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Electricity, Val: 2.0},
		TypeMod{T: Ice, Val: 2.0},
		TypeMod{T: Rock, Val: 2.0},
		TypeMod{T: Ground, Val: 0.0}}
	Poison.Name = "poison"
	Poison.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Poison, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Ground, Val: 2.0},
		TypeMod{T: Bug, Val: 2.0},
		TypeMod{T: Psychic, Val: 2.0}}
	Ground.Name = "ground"
	Ground.Mods = []TypeMod{
		TypeMod{T: Poison, Val: 0.5},
		TypeMod{T: Rock, Val: 0.5},
		TypeMod{T: Grass, Val: 2.0},
		TypeMod{T: Ice, Val: 2.0},
		TypeMod{T: Water, Val: 2.0},
		TypeMod{T: Electricity, Val: 0.0}}
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
	Bug.Name = "bug"
	Bug.Mods = []TypeMod{
		TypeMod{T: Fighting, Val: 0.5},
		TypeMod{T: Grass, Val: 0.5},
		TypeMod{T: Ground, Val: 0.5},
		TypeMod{T: Fire, Val: 2.0},
		TypeMod{T: Flying, Val: 2.0},
		TypeMod{T: Rock, Val: 2.0}}
	Ghost.Name = "ghost"
	Steel.Name = "steel"
	Fire.Name = "fire"
	Water.Name = "water"
	Grass.Name = "grass"
	Electricity.Name = "elecricity"
	Psychic.Name = "psychic"
	Ice.Name = "ice"
	Dragon.Name = "dragon"
	Dark.Name = "dark"
	Fairy.Name = "fairy"
}
