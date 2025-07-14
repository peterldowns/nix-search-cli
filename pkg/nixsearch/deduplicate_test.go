package nixsearch

import (
	"testing"

	"github.com/peterldowns/testy/check"
)

func TestDeduplicateRegularPackages(t *testing.T) {
	t.Parallel()

	pkgGo := Package{
		AttrName: "go",
		Version:  "1.24.4",
	}
	pkgGoDuplicate := Package{
		AttrName: "go",
		Version:  "1.99.9", // higher version intentionally
	}
	check.Equal(t, pkgGo.ID(), pkgGoDuplicate.ID())

	// The duplicate is removed from the final list. Whichever one is seen first
	// is preserved.
	check.Equal(t, []Package{pkgGo}, Deduplicate([]Package{
		pkgGo, pkgGoDuplicate,
	}))
	check.Equal(t, []Package{pkgGoDuplicate}, Deduplicate([]Package{
		pkgGoDuplicate, pkgGo,
	}))
	// Any number of duplicates are removed.
	check.Equal(t, []Package{pkgGo}, Deduplicate([]Package{
		pkgGo, pkgGo, pkgGoDuplicate, pkgGo, pkgGoDuplicate,
	}))
}

func TestDeduplicateFlakes(t *testing.T) {
	t.Parallel()

	pkgFlake := Package{
		AttrName:  "nix-search",
		Version:   "0.0.1", // fake
		FlakeName: "nix-search-cli",
		FlakeResolved: FlakeResolved{
			Type:  "github",
			Owner: "peterldowns",
			Repo:  "nix-search-cli",
		},
	}
	pkgFlakeDuplicate := pkgFlake
	pkgFlakeDuplicate.Version = "9.9.9"
	check.Equal(t, pkgFlake.ID(), pkgFlakeDuplicate.ID())

	// The duplicate is removed from the final list. Whichever one is seen first
	// is preserved.
	check.Equal(t, []Package{pkgFlake}, Deduplicate([]Package{
		pkgFlake, pkgFlakeDuplicate,
	}))
	check.Equal(t, []Package{pkgFlakeDuplicate}, Deduplicate([]Package{
		pkgFlakeDuplicate, pkgFlake,
	}))
	// Any number of duplicates are removed.
	check.Equal(t, []Package{pkgFlake}, Deduplicate([]Package{
		pkgFlake, pkgFlake, pkgFlakeDuplicate, pkgFlake, pkgFlakeDuplicate,
	}))
}

func TestDeduplicateFlakeAttrnames(t *testing.T) {
	pkgFlake := Package{
		AttrName:  "nix-search",
		Version:   "0.0.1", // fake
		FlakeName: "nix-search-cli",
		FlakeResolved: FlakeResolved{
			Type:  "github",
			Owner: "peterldowns",
			Repo:  "nix-search-cli",
		},
	}
	notADuplicate := pkgFlake
	notADuplicate.FlakeResolved = FlakeResolved{
		Type:  "github",
		Owner: "someotheraccount",
		Repo:  "adifferentrepo",
	}

	// Although these flakes provide the same package (nix-search), they're
	// different results and will not be deduplicated.
	check.NotEqual(t, pkgFlake.ID(), notADuplicate.ID())
	pkgs := []Package{pkgFlake, notADuplicate}
	check.Equal(t, pkgs, Deduplicate(pkgs))
}
