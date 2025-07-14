package nixsearch

// Deduplicate filters a slice of [Package] objects to remove any duplicates
// that have the same ID (attr name). Notably, this does not sort the input list
// of packages, so if there are multiple entries with the same ID, the result
// will just have whichever one appears first in the list.
//
// This is needed because the [ElasticSearchClient]'s `Search` method can
// sometimes read results from multiple different elasticsearch indexes, which
// means the same package can appear more than once in the list of results.
// Although the duplicate entries may have different versions, metadata, etc.
// this deduplication filter just returns whichever one shows up first.
func Deduplicate(packages []Package) []Package {
	seen := map[string]struct{}{}
	var deduped []Package
	for _, pkg := range packages {
		id := pkg.ID()
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, pkg)
	}
	return deduped
}
