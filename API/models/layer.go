package models

const (
	LayerFilters       = "filters"
	LayerFiltersRemove = LayerFilters + "-"
)

// Edits object's "filters" list by removing filters in "filters-"
func removeFromFilters(object map[string]any) {
	filters, filtersPresent := object[LayerFilters].(map[string]any)
	if !filtersPresent || filters == nil {
		filters = map[string]any{}
	}

	minusFilter, minusFilterPresent := object[LayerFiltersRemove]
	if minusFilterPresent {
		delete(filters, minusFilter.(string))
		delete(object, LayerFiltersRemove)
	}
}
