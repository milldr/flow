package workspace

import (
	"fmt"
	"math/rand/v2"
)

var adjectives = []string{
	"calm", "bold", "cool", "fast", "keen",
	"warm", "blue", "gold", "iron", "dark",
	"wild", "free", "deep", "pure", "fair",
	"soft", "true", "open", "wise", "safe",
	"brave", "crisp", "deft", "eager", "firm",
	"glad", "hale", "just", "kind", "lean",
	"mild", "neat", "pale", "quick", "rare",
	"slim", "tame", "vast", "zeal", "apt",
	"airy", "avid", "brisk", "civic", "clear",
	"dense", "dry", "even", "fine", "flat",
	"fresh", "grand", "green", "half", "hardy",
	"icy", "jade", "lush", "main", "mere",
	"next", "noble", "odd", "peak", "plain",
	"prime", "quiet", "rapid", "rich", "rigid",
	"rough", "round", "royal", "sharp", "short",
	"sleek", "solid", "spare", "steep", "still",
	"stout", "sunny", "swift", "tall", "taut",
	"thick", "thin", "tidy", "tight", "trim",
	"upper", "vivid", "warm", "whole", "wide",
	"young", "zinc", "amber", "ashen", "blunt",
	"coral", "dusky", "dusty", "faint", "fleet",
}

var nouns = []string{
	"brook", "cliff", "delta", "flame", "grove",
	"haven", "ridge", "spark", "stone", "trail",
	"creek", "drift", "field", "frost", "maple",
	"ocean", "pearl", "river", "shore", "cedar",
	"basin", "bluff", "brine", "cairn", "chase",
	"cloud", "coast", "copse", "crest", "dune",
	"ember", "falls", "fjord", "glade", "gorge",
	"heath", "knoll", "ledge", "marsh", "mesa",
	"mound", "oasis", "orbit", "patch", "plume",
	"point", "pond", "prairie", "quartz", "rapid",
	"reach", "reef", "ridge", "sage", "shoal",
	"slate", "slope", "spire", "steppe", "surge",
	"terra", "thaw", "thorn", "tide", "vale",
	"vault", "verge", "weld", "wharf", "yield",
	"alley", "arch", "bank", "bench", "birch",
	"bloom", "bower", "briar", "camp", "cove",
	"crane", "crown", "dale", "den", "dew",
	"dock", "elm", "fern", "flint", "forge",
	"gate", "glen", "gulch", "haze", "helm",
	"holly", "ivy", "lake", "lane", "lark",
}

// GenerateID returns a random adjective-noun workspace ID.
func GenerateID() string {
	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]
	return adj + "-" + noun
}

// GenerateUniqueID returns a unique ID that doesn't collide with any existing IDs.
// After 10 attempts, it appends a random numeric suffix.
func GenerateUniqueID(existingIDs []string) string {
	existing := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}

	for range 10 {
		id := GenerateID()
		if _, taken := existing[id]; !taken {
			return id
		}
	}

	// Fallback: append random suffix
	id := GenerateID()
	return fmt.Sprintf("%s-%d", id, rand.IntN(9000)+1000)
}
