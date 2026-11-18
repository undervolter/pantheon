package domain

// FilterRange — диапазон числового фильтра (например, год или пробег).
type FilterRange struct {
	From float64 `json:"from,omitempty"`
	To   float64 `json:"to,omitempty"`
}

// Search — один фильтр выдачи OLX.
type Search struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name"`                     // Название фильтра
	Query         string                 `json:"query"`                    // Поисковая фраза
	QueryVariants []string               `json:"query_variants,omitempty"` // Нормализованные варианты для пост-фильтра
	MinPrice      float64                `json:"min_price,omitempty"`      // Мин цена (0 если без ограничений)
	MaxPrice      float64                `json:"max_price,omitempty"`      // Макс цена (0 если без ограничений)
	CityID        int                    `json:"city_id,omitempty"`        // ID города (0 для всей Украины)
	CityName      string                 `json:"city_name,omitempty"`      // Название города
	RegionID      int                    `json:"region_id,omitempty"`      // ID области
	DistrictID    int                    `json:"district_id,omitempty"`    // ID района внутри города (0 — весь город)
	DistrictName  string                 `json:"district_name,omitempty"`  // Название района
	Category      int                    `json:"category_id,omitempty"`    // ID категории
	CatName       string                 `json:"category_name,omitempty"`  // Название категории
	BrandParam    string                 `json:"brand_param,omitempty"`    // Параметр бренда в API (например: filter_enum_mobile_phone_manufacturer)
	BrandValue    string                 `json:"brand_value,omitempty"`    // Значение бренда в API (например: 2065)
	BrandName     string                 `json:"brand_name,omitempty"`     // Отображаемое имя бренда (например: Apple)
	BrandValues   []string               `json:"brand_values,omitempty"`   // Мультивыбор брендов
	BrandNames    []string               `json:"brand_names,omitempty"`    // Имена выбранных брендов
	OwnerType     string                 `json:"owner_type,omitempty"`     // private | business | "" (все)
	Filters       map[string][]string    `json:"filters,omitempty"`        // enum/checkbox: ключ фильтра -> значения
	Ranges        map[string]FilterRange `json:"ranges,omitempty"`         // float/input: ключ -> диапазон
	Disabled      bool                   `json:"disabled,omitempty"`       // Поставлен на паузу
	CreatedAt     int64                  `json:"created_at,omitempty"`     // Unix timestamp создания фильтра
	OwnerChatID   string                 `json:"owner_chat_id,omitempty"`  // chat_id владельца: алерты только ему (пусто — общий фильтр)
}
