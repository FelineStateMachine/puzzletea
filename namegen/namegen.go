// Package namegen generates random human-readable names for saved games.
package namegen

import "math/rand/v2"

var adjectives = []string{
	"amber", "ancient", "aqua", "arctic", "autumn",
	"blazing", "bold", "bright", "bronze", "calm",
	"cedar", "celestial", "cherry", "chilly", "citrus",
	"clever", "cobalt", "copper", "coral", "cosmic",
	"crimson", "crystal", "daring", "dawn", "deep",
	"desert", "diamond", "drifting", "dusk", "dusty",
	"eager", "echo", "elegant", "ember", "emerald",
	"endless", "fading", "fierce", "fiery", "floral",
	"foggy", "forest", "frosty", "gentle", "gilded",
	"glacial", "gleaming", "glowing", "golden", "granite",
	"grassy", "hidden", "hollow", "humble", "hushed",
	"icy", "indigo", "iron", "ivory", "jade",
	"keen", "lavender", "lazy", "lemon", "lightning",
	"lofty", "lonely", "lunar", "marble", "meadow",
	"mighty", "misty", "molten", "mossy", "muted",
	"noble", "obsidian", "ocean", "onyx", "opal",
	"pale", "pearl", "pine", "plum", "polar",
	"prairie", "quiet", "radiant", "rapid", "raven",
	"roaming", "ruby", "rustic", "sable", "sage",
	"sandy", "scarlet", "serene", "shadow", "silver",
	"sleepy", "solar", "starry", "steel", "stone",
	"stormy", "sunlit", "swift", "tawny", "tidal",
	"timber", "topaz", "twilight", "velvet", "violet",
	"wandering", "whisper", "wild", "winter", "woven",
}

var nouns = []string{
	"acorn", "arch", "aurora", "badger", "basalt",
	"beacon", "birch", "bison", "blaze", "boulder",
	"bramble", "brook", "canyon", "cedar", "cliff",
	"cloud", "compass", "condor", "coral", "coyote",
	"crane", "creek", "crest", "crow", "delta",
	"dune", "eagle", "elk", "falcon", "fern",
	"fjord", "flame", "flint", "forge", "fossil",
	"fox", "frost", "geyser", "glacier", "grove",
	"harbor", "hawk", "heath", "heron", "hollow",
	"horizon", "island", "ivy", "jasper", "juniper",
	"lagoon", "lantern", "larch", "lava", "leaf",
	"ledge", "lichen", "lynx", "marsh", "mesa",
	"meteor", "mist", "moon", "moss", "moth",
	"mountain", "nebula", "newt", "oak", "orchid",
	"osprey", "otter", "owl", "peak", "pebble",
	"pine", "plover", "pond", "quail", "quartz",
	"rain", "raptor", "reef", "ridge", "river",
	"robin", "sage", "seal", "sequoia", "shell",
	"shore", "sparrow", "spruce", "star", "stone",
	"storm", "summit", "thistle", "thorn", "tide",
	"timber", "trail", "trout", "vale", "valley",
	"viper", "vista", "willow", "wolf", "wren",
}

// Generate returns a random "adjective-noun" name.
func Generate() string {
	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]
	return adj + "-" + noun
}

// GenerateSeeded returns a deterministic "adjective-noun" name using the
// provided RNG. Two calls with identically-seeded RNGs produce the same name.
func GenerateSeeded(rng *rand.Rand) string {
	adj := adjectives[rng.IntN(len(adjectives))]
	noun := nouns[rng.IntN(len(nouns))]
	return adj + "-" + noun
}
