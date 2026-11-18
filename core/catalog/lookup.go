package catalog

// IsReservedFilterKey сообщает, что параметр выдачи обслуживается отдельным
// шагом (цена, тип продавца, сортировка) и не показывается среди доп. фильтров.
func IsReservedFilterKey(key string) bool {
	switch key {
	case "filter_float_price", "filter_enum_price", "owner_type", "sort_by", "price":
		return true
	}
	return false
}
