package domain

import "sort"

func Merge(events []Event) []Event {
	out := append([]Event(nil), events...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].LeaseID == out[j].LeaseID {
			return out[i].Version > out[j].Version
		}
		return out[i].LeaseID < out[j].LeaseID
	})
	seen := map[string]bool{}
	uniq := out[:0]
	for _, e := range out {
		if !seen[e.LeaseID] {
			uniq = append(uniq, e)
			seen[e.LeaseID] = true
		}
	}
	return uniq
}
