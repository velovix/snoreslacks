package app

// typeMod is a single type modifier. It contains a defense modifier and the
// type that the defense modifier applies to.
type typeMod struct {
	t   pkmnType
	val float64
}

// pkmnType is a single type which contains a slice of defense modifiers for
// every type where the modifier would not equal 1.0.
type pkmnType struct {
	Name string
	mods []typeMod
}

// A lookup table from PokeAPI type names to actual types.
var types map[string]pkmnType

// Initialize the values of each type.
func init() {
	types = make(map[string]pkmnType)

	var normal, fighting, flying, poison, ground, rock, bug,
		ghost, steel, fire, water, grass, electricity, psychic,
		ice, dragon, dark, fairy pkmnType

	// Fill out each type's defense modifiers
	normal.Name = "normal"
	normal.mods = []typeMod{
		typeMod{t: fighting, val: 2.0},
		typeMod{t: ghost, val: 0.0}}
	fighting.Name = "fighting"
	fighting.mods = []typeMod{
		typeMod{t: bug, val: 0.5},
		typeMod{t: dark, val: 0.5},
		typeMod{t: rock, val: 0.5},
		typeMod{t: fairy, val: 2.0},
		typeMod{t: flying, val: 2.0},
		typeMod{t: psychic, val: 2.0}}
	flying.Name = "flying"
	flying.mods = []typeMod{
		typeMod{t: bug, val: 0.5},
		typeMod{t: fighting, val: 0.5},
		typeMod{t: grass, val: 0.5},
		typeMod{t: electricity, val: 2.0},
		typeMod{t: ice, val: 2.0},
		typeMod{t: rock, val: 2.0},
		typeMod{t: ground, val: 0.0}}
	poison.Name = "poison"
	poison.mods = []typeMod{
		typeMod{t: fighting, val: 0.5},
		typeMod{t: poison, val: 0.5},
		typeMod{t: grass, val: 0.5},
		typeMod{t: ground, val: 2.0},
		typeMod{t: bug, val: 2.0},
		typeMod{t: psychic, val: 2.0}}
	ground.Name = "ground"
	ground.mods = []typeMod{
		typeMod{t: poison, val: 0.5},
		typeMod{t: rock, val: 0.5},
		typeMod{t: grass, val: 2.0},
		typeMod{t: ice, val: 2.0},
		typeMod{t: water, val: 2.0},
		typeMod{t: electricity, val: 0.0}}
	rock.Name = "rock"
	rock.mods = []typeMod{
		typeMod{t: fire, val: 0.5},
		typeMod{t: flying, val: 0.5},
		typeMod{t: normal, val: 0.5},
		typeMod{t: poison, val: 0.5},
		typeMod{t: fighting, val: 2.0},
		typeMod{t: grass, val: 2.0},
		typeMod{t: ground, val: 2.0},
		typeMod{t: steel, val: 2.0},
		typeMod{t: water, val: 2.0}}
	bug.Name = "bug"
	bug.mods = []typeMod{
		typeMod{t: fighting, val: 0.5},
		typeMod{t: grass, val: 0.5},
		typeMod{t: ground, val: 0.5},
		typeMod{t: fire, val: 2.0},
		typeMod{t: flying, val: 2.0},
		typeMod{t: rock, val: 2.0}}
	ghost.Name = "ghost"
	steel.Name = "steel"
	fire.Name = "fire"
	water.Name = "water"
	grass.Name = "grass"
	electricity.Name = "electricity"
	psychic.Name = "psychic"
	ice.Name = "ice"
	dragon.Name = "dragon"
	dark.Name = "dark"
	fairy.Name = "fairy"

	// Assign the types to their PokeAPI names for lookup when loading Pokmeon
	types["normal"] = normal
	types["fighting"] = fighting
	types["flying"] = flying
	types["poison"] = poison
	types["ground"] = ground
	types["rock"] = rock
	types["bug"] = bug
	types["ghost"] = ghost
	types["steel"] = steel
	types["fire"] = fire
	types["water"] = water
	types["grass"] = grass
	types["electricity"] = electricity
	types["psychic"] = psychic
	types["ice"] = ice
	types["dragon"] = dragon
	types["dark"] = dark
	types["fairy"] = fairy
}
