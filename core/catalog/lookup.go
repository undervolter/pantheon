package catalog

import "sync"

// RealEstateCategoryID — корень раздела «Нерухомість» в дереве OLX.
const RealEstateCategoryID = 1

// realEstateType — значение category.type, которым OLX помечает лоты недвижимости.
const realEstateType = "real_estate"

// IsReservedFilterKey сообщает, что параметр выдачи обслуживается отдельным
// шагом (цена, тип продавца, сортировка) и не показывается среди доп. фильтров.
func IsReservedFilterKey(key string) bool {
	switch key {
	case "filter_float_price", "filter_enum_price", "owner_type", "sort_by", "price":
		return true
	}
	return false
}

// CategoryInSubtree сообщает, что категория id лежит в поддереве root
// (включая сам корень). Неизвестная категория ни в какое поддерево не входит.
func CategoryInSubtree(id, root int) bool {
	if id <= 0 || root <= 0 {
		return false
	}
	// Дерево неглубокое; ограничение — защита от цикла в данных.
	for depth := 0; depth < 16; depth++ {
		if id == root {
			return true
		}
		c := FindCategory(id)
		if c == nil || c.ParentID == 0 {
			return false
		}
		id = c.ParentID
	}
	return false
}

// IsRealEstate сообщает, что категория относится к недвижимости. Бот этот
// раздел не обслуживает: такие категории не предлагаются и не принимаются.
func IsRealEstate(categoryID int) bool {
	return CategoryInSubtree(categoryID, RealEstateCategoryID)
}

// IsRealEstateOffer сообщает, что лот — недвижимость: по типу категории из
// ответа OLX либо по месту категории в дереве. Тип надёжнее: он есть и у
// листовых категорий, которых может не быть в локальном дереве.
func IsRealEstateOffer(categoryID int, categoryType string) bool {
	return categoryType == realEstateType || IsRealEstate(categoryID)
}

// VisibleTopCategories возвращает разделы верхнего уровня, которые бот
// предлагает на выбор.
func VisibleTopCategories() []CategoryItem {
	out := make([]CategoryItem, 0, len(TopCategories))
	for _, c := range TopCategories {
		if !IsRealEstate(c.ID) {
			out = append(out, c)
		}
	}
	return out
}

var brandParams = sync.OnceValue(func() map[string]bool {
	out := make(map[string]bool, len(CategoryBrands))
	for _, info := range CategoryBrands {
		if info.ParamName != "" {
			out[info.ParamName] = true
		}
	}
	return out
})

// IsBrandParam сообщает, что параметр — справочник брендов хотя бы одной
// категории. Только такой параметр может стать «брендом» фильтра.
func IsBrandParam(name string) bool {
	return brandParams()[name]
}

// FindCategoryInPath resolves one segment of a category path such as
// /moda-i-stil/muzhskaya-odezhda/shorty/. Slugs repeat all over the tree
// ("shorty" names six categories, "drugoe" more than a hundred), so a segment
// is looked up under the category the previous segment named. Only the first
// segment is searched everywhere, top level first.
func FindCategoryInPath(parent *CategoryItem, slug string) *CategoryItem {
	if slug == "" {
		return nil
	}
	if parent == nil {
		for i := range TopCategories {
			if TopCategories[i].Slug == slug {
				return &TopCategories[i]
			}
		}
		return FindCategoryBySlug(slug)
	}
	for i := range parent.Children {
		if parent.Children[i].Slug == slug {
			return &parent.Children[i]
		}
	}
	for i := range parent.Children {
		if found := findCategoryBySlugRec(&parent.Children[i], slug); found != nil {
			return found
		}
	}
	return nil
}
