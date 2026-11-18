package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/i18n"
	"pantheon/core/logx"
	"pantheon/core/olx"
	"pantheon/core/telegram"
	"pantheon/core/textutil"
)

// The wizard builds one filter in five numbered steps:
//
//	1 category → 2 what to look for (brand, query, variants, exclusions,
//	extra filters) → 3 price → 4 seller → 5 location → review → launch
//
// Every screen is a view* function. A view that expects typed input arms the
// matching step itself, so the screen on display and the text the bot is
// waiting for can never disagree. A pressed button clears the step first.

const (
	wizardSteps     = 5
	brandsPerPage   = 10
	extrasPerPage   = 8
	valuesPerPage   = 8
	maxCityResults  = 8
	exclBackVariant = "v"
	exclBackReview  = "r"
)

var pricePresets = []float64{5000, 10000, 20000, 30000, 50000}

// wizardCallbacks are the buttons that belong to an unfinished filter. They
// only work on the message the wizard currently lives in.
var wizardCallbacks = map[string]bool{
	"add_start": true, "add_cat": true, "add_subcat": true, "add_cat_select": true, "add_cat_back": true,
	"add_brand_toggle": true, "add_brand_page": true, "add_brand_done": true, "add_brand_skip": true,
	"wiz_query": true, "add_query_skip": true,
	"filters_open": true, "filters_page": true, "filter_open": true, "filter_page": true,
	"filter_toggle": true, "filter_bool": true, "filter_input": true, "filter_clear": true, "extras_done": true,
	"variant_add": true, "variant_del": true, "variants_back": true, "variants_done": true,
	"excl_open": true, "excl_add": true, "excl_del": true, "excl_done": true,
	"wiz_price": true, "add_price": true, "add_price_free": true, "add_price_range": true,
	"wiz_owner": true, "add_owner": true,
	"wiz_where": true, "add_city": true, "add_city_manual": true, "add_city_pick": true, "add_region_pick": true,
	"add_city_whole": true, "add_district_pick": true,
	"add_finish": true,
}

// stepPage starts a wizard screen with its progress footer.
func (t *turn) stepPage(n int, text string) *page {
	p := promptPage(text)
	dots := strings.Repeat("●", n) + strings.Repeat("○", wizardSteps-n)
	p.note(t.tr("wz.progress", n, wizardSteps, dots))
	return p
}

// startWizard begins a new filter, dropping whatever was unfinished.
func (b *Bot) startWizard(t *turn) {
	t.resetState()
	t.wizard = true
	t.st.Active = true
	t.st.Brands = map[string]string{}
	// The wizard lives in the message the button was pressed on; for a typed
	// command this is 0 and gets set when the first screen is sent.
	t.st.MsgID = t.msgID
}

// ---------- step 1: category ----------

func (b *Bot) viewCategory(t *turn) (string, *markup) {
	var buttons []button
	for _, c := range catalog.VisibleTopCategories() {
		buttons = append(buttons, telegram.Btn(strings.TrimSpace(c.Icon+" "+c.Name), fmt.Sprintf("add_cat:%d", c.ID)))
	}
	rows := grid(buttons, 2)
	rows = append(rows,
		telegram.Row(telegram.Btn(t.tr("wz.all_categories"), "add_cat:0")),
		telegram.Row(btnCancel(t.lang)),
	)
	return t.stepPage(1, t.tr("wz.cat_prompt")).String(), telegram.Markup(rows...)
}

func (b *Bot) viewSubcategory(t *turn, parentID int) (string, *markup) {
	parent := catalog.FindCategory(parentID)
	if parent == nil || len(parent.Children) == 0 || catalog.IsRealEstate(parentID) {
		return b.viewCategory(t)
	}
	var buttons []button
	for _, sub := range parent.Children {
		buttons = append(buttons, telegram.Btn(textutil.TruncateRunes(sub.Name, 40), fmt.Sprintf("add_subcat:%d", sub.ID)))
	}
	rows := grid(buttons, 2)
	rows = append(rows, telegram.Row(telegram.Btn(
		t.tr("wz.whole_section", textutil.TruncateRunes(parent.Name, 35)), fmt.Sprintf("add_cat_select:%d", parentID))))

	back := "add_start"
	if parent.ParentID > 0 {
		back = fmt.Sprintf("add_cat_back:%d", parent.ParentID)
	}
	rows = append(rows, telegram.Row(btnBack(t.lang, back), btnCancel(t.lang)))
	return t.stepPage(1, t.tr("wz.subcat_prompt", esc(catalog.CategoryBreadcrumb(parentID)))).String(), telegram.Markup(rows...)
}

// selectCategory fixes the category of the draft. Everything that depends on
// the category — the brand and the extra filters — starts over.
func (b *Bot) selectCategory(t *turn, catID int) {
	d := &t.st.Draft
	d.Category, d.CatName = 0, ""
	if c := catalog.FindCategory(catID); c != nil && !catalog.IsRealEstate(catID) {
		d.Category, d.CatName = c.ID, c.Name
	}
	d.BrandParam, d.BrandValue, d.BrandName, d.BrandValues, d.BrandNames = "", "", "", nil, nil
	d.Filters, d.Ranges = nil, nil
	t.st.Brands = map[string]string{}

	if info := catalog.GetCategoryBrands(d.Category); info != nil && len(info.Brands) > 0 {
		t.show(b.viewBrands(t, 0))
		return
	}
	t.show(b.viewQuery(t))
}

// ---------- step 2: brand ----------

func (b *Bot) viewBrands(t *turn, page int) (string, *markup) {
	catID := t.st.Draft.Category
	info := catalog.GetCategoryBrands(catID)
	if info == nil || len(info.Brands) == 0 {
		return b.viewQuery(t)
	}
	label := info.Label
	if label == "" {
		label = t.tr("wz.brand_default")
	}
	start, end, page, pages := pageBounds(len(info.Brands), brandsPerPage, page)

	var buttons []button
	for _, br := range info.Brands[start:end] {
		_, on := t.st.Brands[br.ID]
		buttons = append(buttons, toggleBtn(on, textutil.TruncateRunes(br.Name, 40), fmt.Sprintf("add_brand_toggle:%d:%d:%s", catID, page, br.ID)))
	}
	rows := grid(buttons, 2)
	if nav := pagerRow(page, pages, func(n int) string { return fmt.Sprintf("add_brand_page:%d:%d", catID, n) }); nav != nil {
		rows = append(rows, nav)
	}
	rows = append(rows, telegram.Row(
		telegram.Primary(t.tr("wz.brand_done", len(t.st.Brands)), fmt.Sprintf("add_brand_done:%d", catID)),
		telegram.Btn(t.tr("wz.brand_any"), fmt.Sprintf("add_brand_skip:%d", catID)),
	))
	rows = append(rows, telegram.Row(btnBack(t.lang, b.categoryBack(catID)), btnCancel(t.lang)))

	text := t.tr("wz.brand_prompt", esc(label), esc(t.st.Draft.CatName), len(t.st.Brands))
	return t.stepPage(2, text).String(), telegram.Markup(rows...)
}

// categoryBack is where "Back" leads from the first screen after a category
// was picked: to its siblings.
func (b *Bot) categoryBack(catID int) string {
	if c := catalog.FindCategory(catID); c != nil && c.ParentID > 0 {
		return fmt.Sprintf("add_cat_back:%d", c.ParentID)
	}
	return "add_start"
}

// applyBrands copies the multi-select into the draft, in catalog order.
func (b *Bot) applyBrands(t *turn) {
	d := &t.st.Draft
	d.BrandParam, d.BrandValue, d.BrandName, d.BrandValues, d.BrandNames = "", "", "", nil, nil
	info := catalog.GetCategoryBrands(d.Category)
	if info == nil {
		return
	}
	for _, br := range info.Brands {
		if _, ok := t.st.Brands[br.ID]; ok {
			d.BrandValues = append(d.BrandValues, br.ID)
			d.BrandNames = append(d.BrandNames, br.Name)
		}
	}
	if len(d.BrandValues) == 0 {
		return
	}
	d.BrandParam = info.ParamName
	if len(d.BrandValues) == 1 {
		d.BrandValue, d.BrandName = d.BrandValues[0], d.BrandNames[0]
	}
}

// ---------- step 2: query, variants, exclusions ----------

func (b *Bot) viewQuery(t *turn) (string, *markup) {
	t.st.Step = stepQuery
	d := t.st.Draft
	p := t.stepPage(2, t.tr("wz.query_prompt"))
	p.kv(t.tr("sum.category"), categoryText(t.lang, d))
	p.kv(t.tr("sum.brand"), esc(textutil.TruncateRunes(d.Brands(), 160)))

	var rows [][]button
	if d.Category > 0 && len(extraFilters(d.Category, d.BrandParam)) > 0 {
		rows = append(rows, telegram.Row(telegram.Btn(t.tr("wz.btn_extras", extrasSet(d)), "filters_open")))
	}
	rows = append(rows, telegram.Row(telegram.Btn(t.tr("wz.query_skip"), "add_query_skip")))

	back := b.categoryBack(d.Category)
	if info := catalog.GetCategoryBrands(d.Category); info != nil && len(info.Brands) > 0 {
		back = fmt.Sprintf("add_brand_page:%d:0", d.Category)
	}
	rows = append(rows, telegram.Row(btnBack(t.lang, back), btnCancel(t.lang)))
	return p.String(), telegram.Markup(rows...)
}

func chips(items []string) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, "<code>"+esc(it)+"</code>")
	}
	return strings.Join(parts, "  ")
}

func (b *Bot) viewVariants(t *turn) (string, *markup) {
	d := t.st.Draft
	p := t.stepPage(2, t.tr("wz.variants_prompt", esc(d.Query)))
	if len(d.QueryVariants) == 0 {
		p.text(t.tr("wz.variants_none"))
	} else {
		p.text(chips(d.QueryVariants))
	}

	var buttons []button
	for _, v := range d.QueryVariants {
		buttons = append(buttons, telegram.Btn("🗑 "+textutil.TruncateRunes(v, 28), "variant_del:"+itemTag(v)))
	}
	rows := grid(buttons, 2)
	rows = append(rows,
		telegram.Row(telegram.Btn(t.tr("wz.btn_variant_add"), "variant_add"), telegram.Btn(t.tr("wz.btn_exclude", len(d.ExcludeWords)), "excl_open:"+exclBackVariant)),
		telegram.Row(telegram.Primary(t.tr("btn.next"), "variants_done")),
		telegram.Row(btnBack(t.lang, "wiz_query"), btnCancel(t.lang)),
	)
	return p.String(), telegram.Markup(rows...)
}

func (b *Bot) viewVariantAsk(t *turn) (string, *markup) {
	t.st.Step = stepVariant
	return prompt(t.tr("wz.variant_ask")), telegram.Markup(telegram.Row(btnBack(t.lang, "variants_back")))
}

func (b *Bot) viewExclude(t *turn) (string, *markup) {
	d := t.st.Draft
	p := promptPage(t.tr("wz.excl_prompt"))
	if len(d.ExcludeWords) == 0 {
		p.text(t.tr("wz.excl_none"))
	} else {
		p.text(chips(d.ExcludeWords))
	}
	var buttons []button
	for _, w := range d.ExcludeWords {
		buttons = append(buttons, telegram.Btn("🗑 "+textutil.TruncateRunes(w, 28), "excl_del:"+itemTag(w)))
	}
	rows := grid(buttons, 2)
	rows = append(rows, telegram.Row(
		telegram.Btn(t.tr("wz.btn_excl_add"), "excl_add"),
		telegram.Primary(t.tr("btn.done"), "excl_done"),
	))
	return p.String(), telegram.Markup(rows...)
}

func (b *Bot) viewExcludeAsk(t *turn) (string, *markup) {
	t.st.Step = stepExclude
	return prompt(t.tr("wz.excl_ask")), telegram.Markup(telegram.Row(btnBack(t.lang, "excl_open:"+t.st.ExclBack)))
}

// ---------- step 2: extra (attribute) filters ----------

func (b *Bot) viewExtras(t *turn, page int) (string, *markup) {
	d := t.st.Draft
	defs := extraFilters(d.Category, d.BrandParam)
	if len(defs) == 0 {
		return prompt(t.tr("wz.extras_none")), telegram.Markup(telegram.Row(btnBack(t.lang, "wiz_query")))
	}
	start, end, page, pages := pageBounds(len(defs), extrasPerPage, page)
	t.st.FilterListPage = page

	var buttons []button
	for i := start; i < end; i++ {
		label := textutil.TruncateRunes(defs[i].Label, 26)
		n := extraSetCount(d, defs[i])
		if n > 1 {
			label += " · " + strconv.Itoa(n)
		}
		buttons = append(buttons, toggleBtn(n > 0, label, fmt.Sprintf("filter_open:%d:%d", i, page)))
	}
	rows := grid(buttons, 2)
	if nav := pagerRow(page, pages, func(n int) string { return fmt.Sprintf("filters_page:%d", n) }); nav != nil {
		rows = append(rows, nav)
	}
	rows = append(rows, telegram.Row(telegram.Primary(t.tr("btn.done"), "extras_done")))
	text := t.tr("wz.extras_prompt", esc(d.CatName), extrasSet(d), len(defs))
	return t.stepPage(2, text).String(), telegram.Markup(rows...)
}

func isBoolFilter(def catalog.FilterDef) bool {
	return strings.HasPrefix(def.Key, "filter_bool_") || (def.Type == "checkbox" && len(def.Values) == 0)
}

func isNumberFilter(def catalog.FilterDef) bool {
	return def.Type == "input" || strings.HasPrefix(def.Key, "filter_float_") || strings.HasPrefix(def.Key, "filter_int_")
}

func unitSuffix(unit string) string {
	if unit == "" {
		return ""
	}
	return " (" + esc(unit) + ")"
}

func (b *Bot) viewFilterValues(t *turn, idx, page int) (string, *markup) {
	d := t.st.Draft
	defs := extraFilters(d.Category, d.BrandParam)
	if idx < 0 || idx >= len(defs) {
		return b.viewExtras(t, t.st.FilterListPage)
	}
	def := defs[idx]
	t.st.FilterIdx = idx
	back := telegram.Row(btnBack(t.lang, fmt.Sprintf("filters_page:%d", t.st.FilterListPage)))

	if isBoolFilter(def) {
		on := len(d.Filters[def.Key]) > 0
		state, toggle := t.tr("common.off"), telegram.Success(t.tr("wz.btn_enable"), fmt.Sprintf("filter_bool:%d", idx))
		if on {
			state, toggle = t.tr("common.on"), telegram.Btn(t.tr("wz.btn_disable"), fmt.Sprintf("filter_bool:%d", idx))
		}
		return prompt(t.tr("wz.fv_bool", esc(def.Label), state)), telegram.Markup(telegram.Row(toggle), back)
	}

	if isNumberFilter(def) {
		current := t.tr("wz.fv_unset")
		if r, ok := d.Ranges[def.Key]; ok && (r.From > 0 || r.To > 0) {
			current = esc(filterRangeText(t.lang, r, def.Unit))
		}
		km := telegram.Markup(
			telegram.Row(
				telegram.Primary(t.tr("wz.btn_enter"), fmt.Sprintf("filter_input:%d", idx)),
				telegram.Btn(t.tr("wz.btn_reset"), fmt.Sprintf("filter_clear:%d", idx)),
			),
			back,
		)
		return prompt(t.tr("wz.fv_number", esc(def.Label), unitSuffix(def.Unit), current)), km
	}

	selected := d.Filters[def.Key]
	var chosen []string
	for _, v := range def.Values {
		if textutil.StringInSlice(selected, v.Value) {
			chosen = append(chosen, v.Label)
		}
	}
	current := t.tr("wz.fv_unset")
	if len(chosen) > 0 {
		current = esc(textutil.TruncateRunes(strings.Join(chosen, ", "), 120))
	}
	kind := t.tr("wz.fv_single")
	if def.Type == "checkboxes" {
		kind = t.tr("wz.fv_multi")
	}

	start, end, page, pages := pageBounds(len(def.Values), valuesPerPage, page)
	var buttons []button
	for i := start; i < end; i++ {
		v := def.Values[i]
		buttons = append(buttons, toggleBtn(textutil.StringInSlice(selected, v.Value),
			textutil.TruncateRunes(v.Label, 28), fmt.Sprintf("filter_toggle:%d:%d:%d", idx, i, page)))
	}
	rows := grid(buttons, 2)
	if nav := pagerRow(page, pages, func(n int) string { return fmt.Sprintf("filter_page:%d:%d", idx, n) }); nav != nil {
		rows = append(rows, nav)
	}
	rows = append(rows, telegram.Row(telegram.Btn(t.tr("wz.btn_reset"), fmt.Sprintf("filter_clear:%d", idx))), back)
	return prompt(t.tr("wz.fv_enum", esc(def.Label), kind, current)), telegram.Markup(rows...)
}

func (b *Bot) viewNumberAsk(t *turn, idx int) (string, *markup) {
	defs := extraFilters(t.st.Draft.Category, t.st.Draft.BrandParam)
	if idx < 0 || idx >= len(defs) {
		return b.viewExtras(t, t.st.FilterListPage)
	}
	t.st.Step = stepFilterValue
	t.st.FilterIdx = idx
	km := telegram.Markup(telegram.Row(btnBack(t.lang, fmt.Sprintf("filter_open:%d:%d", idx, t.st.FilterListPage))))
	return prompt(t.tr("wz.number_ask", esc(defs[idx].Label), unitSuffix(defs[idx].Unit))), km
}

// ---------- step 3: price ----------

func (b *Bot) viewPrice(t *turn) (string, *markup) {
	d := t.st.Draft
	var buttons []button
	for _, v := range pricePresets {
		on := !d.FreeOnly() && d.MinPrice == 0 && d.MaxPrice == v
		buttons = append(buttons, toggleBtn(on, t.tr("wz.price_upto", money(v)), fmt.Sprintf("add_price:%.0f", v)))
	}
	buttons = append(buttons, telegram.Btn(t.tr("wz.price_any"), "add_price:0"))
	rows := grid(buttons, 2)
	rows = append(rows, telegram.Row(
		toggleBtn(d.FreeOnly(), t.tr("wz.price_free"), "add_price_free"),
		telegram.Btn(t.tr("wz.price_range"), "add_price_range"),
	))
	back := "wiz_query"
	if d.Query != "" {
		back = "variants_back"
	}
	rows = append(rows, telegram.Row(btnBack(t.lang, back), btnCancel(t.lang)))
	return t.stepPage(3, t.tr("wz.price_prompt")).note(t.tr("wz.price_note")).String(), telegram.Markup(rows...)
}

func (b *Bot) viewPriceAsk(t *turn) (string, *markup) {
	t.st.Step = stepPrice
	return prompt(t.tr("wz.price_ask")), telegram.Markup(telegram.Row(btnBack(t.lang, "wiz_price")))
}

// ---------- step 4: seller ----------

func (b *Bot) viewOwner(t *turn) (string, *markup) {
	cur := t.st.Draft.OwnerType
	km := telegram.Markup(
		telegram.Row(toggleBtn(cur == "", t.tr("wz.owner_all"), "add_owner:all")),
		telegram.Row(
			toggleBtn(cur == "private", t.tr("wz.owner_private"), "add_owner:private"),
			toggleBtn(cur == "business", t.tr("wz.owner_business"), "add_owner:business"),
		),
		telegram.Row(btnBack(t.lang, "wiz_price"), btnCancel(t.lang)),
	)
	return t.stepPage(4, t.tr("wz.owner_prompt")).String(), km
}

// ---------- step 5: location ----------

// bareCity strips the icon the catalog keeps in front of a city name.
func bareCity(name string) string {
	return strings.TrimSpace(strings.TrimLeft(name, "📍🌍🇺🇦 "))
}

func (b *Bot) viewWhere(t *turn) (string, *markup) {
	rows := [][]button{telegram.Row(telegram.Btn(t.tr("wz.where_all"), "add_city:0:0"))}
	var buttons []button
	for _, c := range catalog.TopCities {
		if c.CityID == 0 {
			continue
		}
		buttons = append(buttons, telegram.Btn("📍 "+bareCity(c.Name), fmt.Sprintf("add_city:%d:%d", c.CityID, c.RegionID)))
	}
	rows = append(rows, grid(buttons, 2)...)
	rows = append(rows,
		telegram.Row(telegram.Btn(t.tr("wz.where_other"), "add_city_manual")),
		telegram.Row(btnBack(t.lang, "wiz_owner"), btnCancel(t.lang)),
	)
	return t.stepPage(5, t.tr("wz.where_prompt")).String(), telegram.Markup(rows...)
}

func (b *Bot) viewCityAsk(t *turn) (string, *markup) {
	t.st.Step = stepCity
	return prompt(t.tr("wz.city_ask")), telegram.Markup(telegram.Row(btnBack(t.lang, "wiz_where")))
}

func (b *Bot) viewCityResults(t *turn, query string, results []olx.GeoLocationSuggestion) (string, *markup) {
	// Another name can be typed straight away, without going back.
	t.st.Step = stepCity
	var rows [][]button
	seenRegions := map[int]bool{}
	addRegion := func(r olx.GeoRegion) {
		if r.ID > 0 && !seenRegions[r.ID] {
			seenRegions[r.ID] = true
			rows = append(rows, telegram.Row(telegram.Btn(
				t.tr("wz.region_whole", textutil.TruncateRunes(r.Name, 40)), fmt.Sprintf("add_region_pick:%d", r.ID))))
		}
	}
	for i, r := range results {
		if i >= maxCityResults {
			break
		}
		if r.City.ID == 0 {
			addRegion(r.Region)
			continue
		}
		label := r.City.Name
		if r.Region.Name != "" {
			label = fmt.Sprintf("%s (%s)", r.City.Name, r.Region.Name)
		}
		rows = append(rows, telegram.Row(telegram.Btn(
			"📍 "+textutil.TruncateRunes(label, 56), fmt.Sprintf("add_city_pick:%d:%d", r.City.ID, r.Region.ID))))
		addRegion(r.Region)
	}
	rows = append(rows, telegram.Row(btnBack(t.lang, "wiz_where"), btnCancel(t.lang)))
	return prompt(t.tr("wz.city_results", esc(query))), telegram.Markup(rows...)
}

func (b *Bot) viewDistricts(t *turn, cityName string, districts []olx.GeoDistrict, cityID, regionID int) (string, *markup) {
	rows := [][]button{telegram.Row(telegram.Btn(t.tr("wz.city_whole"), fmt.Sprintf("add_city_whole:%d:%d", cityID, regionID)))}
	var buttons []button
	for _, d := range districts {
		buttons = append(buttons, telegram.Btn(textutil.TruncateRunes(d.Name, 30), fmt.Sprintf("add_district_pick:%d:%d:%d", d.ID, cityID, regionID)))
	}
	rows = append(rows, grid(buttons, 2)...)
	rows = append(rows, telegram.Row(btnBack(t.lang, "wiz_where"), btnCancel(t.lang)))
	return t.stepPage(5, t.tr("wz.district_prompt", esc(cityName))).String(), telegram.Markup(rows...)
}

// applyCity puts a city into the draft and, when the city has districts,
// offers them; otherwise the wizard is done and goes to the review.
func (b *Bot) applyCity(t *turn, cityID, regionID int, name string) {
	d := &t.st.Draft
	d.CityID, d.RegionID, d.CityName = cityID, regionID, bareCity(name)
	d.DistrictID, d.DistrictName = 0, ""
	t.st.districts = nil

	if cityID > 0 && b.geo != nil {
		if districts, err := b.geo.FetchCityDistricts(cityID); err == nil && len(districts) > 0 {
			// Remembered so that picking one does not need a second request,
			// which could fail and leave the district without a name.
			t.st.districts = map[int]string{}
			for _, ds := range districts {
				t.st.districts[ds.ID] = ds.Name
			}
			t.show(b.viewDistricts(t, d.CityName, districts, cityID, regionID))
			return
		}
	}
	t.show(b.viewReview(t))
}

// ---------- review and launch ----------

func (b *Bot) viewReview(t *turn) (string, *markup) {
	d := t.st.Draft
	p := newPage(t.tr("wz.review_title"))
	p.kv(t.tr("sum.name"), esc(defaultFilterName(t.lang, d)))
	summary(p, t.lang, d)
	for _, n := range t.st.Notes {
		p.note(n)
	}
	p.note(t.tr("wz.review_note"))

	km := telegram.Markup(
		telegram.Row(telegram.Success(t.tr("wz.btn_launch"), "add_finish")),
		telegram.Row(telegram.Btn(t.tr("wz.btn_exclude", len(d.ExcludeWords)), "excl_open:"+exclBackReview)),
		telegram.Row(btnBack(t.lang, "wiz_where"), btnCancel(t.lang)),
	)
	return p.String(), km
}

// filterError explains to a user why a filter was refused.
func filterError(lang i18n.Lang, err error) string {
	var badKey *domain.BadKeyError
	switch {
	case errors.Is(err, domain.ErrNoCondition):
		return i18n.T(lang, "wz.err_no_condition")
	case errors.Is(err, domain.ErrRealEstate):
		return i18n.T(lang, "wz.err_real_estate")
	case errors.As(err, &badKey):
		return i18n.T(lang, "wz.err_bad_key", badKey.Key)
	}
	return i18n.T(lang, "wz.err_invalid")
}

// overLimit reports that the chat already has as many filters as allowed.
func (b *Bot) overLimit(t *turn, adding int) (int, bool) {
	limit := b.cfg.MaxFilters()
	if limit <= 0 {
		return 0, false
	}
	have, err := b.db.CountSearches(t.ctx, t.chatID)
	if err != nil {
		logx.Errorf("count filters: %v", err)
		return 0, false
	}
	return limit, have+adding > limit
}

// finishWizard stores the draft as a filter of this chat.
func (b *Bot) finishWizard(t *turn) {
	d := t.st.Draft.Clone()
	d.OwnerChatID = t.chatID
	d.Disabled, d.SnoozedUntil = false, 0
	d.Normalize()
	d.Name = defaultFilterName(t.lang, d)

	if err := d.Validate(); err != nil {
		t.alert(filterError(t.lang, err))
		return
	}
	if limit, over := b.overLimit(t, 1); over {
		t.alert(t.tr("wz.err_limit", limit))
		return
	}
	saved, err := b.db.AddSearch(t.ctx, d)
	if err != nil {
		t.saveFailed(err)
		return
	}
	t.resetState()
	t.toast(t.tr("wz.toast_launched"))
	t.show(b.viewFilter(t, saved, t.tr("wz.launched", esc(saved.Name))))
	if b.scan != nil {
		b.scan.TriggerCheck()
	}
}

// ---------- buttons ----------

func argInt(args []string, i int) int {
	if i >= len(args) {
		return 0
	}
	n, _ := strconv.Atoi(args[i])
	return n
}

func (b *Bot) wizardCallback(t *turn, name string, args []string) {
	d := &t.st.Draft
	switch name {
	case "add_start":
		t.show(b.viewCategory(t))

	case "add_cat":
		id := argInt(args, 0)
		if id == 0 {
			b.selectCategory(t, 0)
			return
		}
		c := catalog.FindCategory(id)
		if c == nil || catalog.IsRealEstate(id) {
			t.toast(t.tr("common.stale"))
			t.show(b.viewCategory(t))
			return
		}
		if len(c.Children) > 0 {
			t.show(b.viewSubcategory(t, id))
			return
		}
		b.selectCategory(t, id)

	case "add_subcat":
		id := argInt(args, 0)
		if c := catalog.FindCategory(id); c != nil && len(c.Children) > 0 && !catalog.IsRealEstate(id) {
			t.show(b.viewSubcategory(t, id))
			return
		}
		b.selectCategory(t, id)

	case "add_cat_select":
		b.selectCategory(t, argInt(args, 0))

	case "add_cat_back":
		if id := argInt(args, 0); id > 0 {
			t.show(b.viewSubcategory(t, id))
			return
		}
		t.show(b.viewCategory(t))

	case "add_brand_toggle":
		if len(args) == 3 && argInt(args, 0) == d.Category {
			brandID := args[2]
			if _, on := t.st.Brands[brandID]; on {
				delete(t.st.Brands, brandID)
			} else if br := catalog.FindBrand(d.Category, brandID); br != nil {
				t.st.Brands[brandID] = br.Name
			}
		}
		t.show(b.viewBrands(t, argInt(args, 1)))

	case "add_brand_page":
		t.show(b.viewBrands(t, argInt(args, 1)))

	case "add_brand_done":
		b.applyBrands(t)
		t.show(b.viewQuery(t))

	case "add_brand_skip":
		t.st.Brands = map[string]string{}
		b.applyBrands(t)
		t.show(b.viewQuery(t))

	case "wiz_query":
		t.show(b.viewQuery(t))

	case "add_query_skip":
		d.Query, d.QueryVariants = "", nil
		t.show(b.viewPrice(t))

	case "filters_open":
		t.show(b.viewExtras(t, 0))

	case "filters_page":
		t.show(b.viewExtras(t, argInt(args, 0)))

	case "extras_done":
		t.show(b.viewQuery(t))

	case "filter_open":
		if len(args) > 1 {
			t.st.FilterListPage = argInt(args, 1)
		}
		t.show(b.viewFilterValues(t, argInt(args, 0), 0))

	case "filter_page":
		t.show(b.viewFilterValues(t, argInt(args, 0), argInt(args, 1)))

	case "filter_toggle":
		idx, vIdx := argInt(args, 0), argInt(args, 1)
		defs := extraFilters(d.Category, d.BrandParam)
		if len(args) >= 2 && idx >= 0 && idx < len(defs) && vIdx >= 0 && vIdx < len(defs[idx].Values) {
			def, v := defs[idx], defs[idx].Values[vIdx].Value
			if d.Filters == nil {
				d.Filters = map[string][]string{}
			}
			cur := d.Filters[def.Key]
			switch {
			case textutil.StringInSlice(cur, v):
				d.Filters[def.Key] = removeValue(cur, v)
			case def.Type == "checkboxes":
				d.Filters[def.Key] = append(cur, v)
			default:
				d.Filters[def.Key] = []string{v}
			}
			if len(d.Filters[def.Key]) == 0 {
				delete(d.Filters, def.Key)
			}
		}
		t.show(b.viewFilterValues(t, idx, argInt(args, 2)))

	case "filter_bool":
		idx := argInt(args, 0)
		defs := extraFilters(d.Category, d.BrandParam)
		if len(args) >= 1 && idx >= 0 && idx < len(defs) {
			if d.Filters == nil {
				d.Filters = map[string][]string{}
			}
			if len(d.Filters[defs[idx].Key]) > 0 {
				delete(d.Filters, defs[idx].Key)
			} else {
				d.Filters[defs[idx].Key] = []string{"1"}
			}
		}
		t.show(b.viewFilterValues(t, idx, 0))

	case "filter_clear":
		idx := argInt(args, 0)
		defs := extraFilters(d.Category, d.BrandParam)
		if len(args) >= 1 && idx >= 0 && idx < len(defs) {
			delete(d.Filters, defs[idx].Key)
			delete(d.Ranges, defs[idx].Key)
		}
		t.show(b.viewFilterValues(t, idx, 0))

	case "filter_input":
		t.show(b.viewNumberAsk(t, argInt(args, 0)))

	case "variant_add":
		if len(d.QueryVariants) >= domain.MaxVariants {
			t.alert(t.tr("wz.variants_full", domain.MaxVariants))
			return
		}
		t.show(b.viewVariantAsk(t))

	case "variant_del":
		if len(args) > 0 {
			d.QueryVariants = removeTagged(d.QueryVariants, args[0])
		}
		t.show(b.viewVariants(t))

	case "variants_back":
		t.show(b.viewVariants(t))

	case "variants_done":
		t.show(b.viewPrice(t))

	case "excl_open":
		t.st.ExclBack = exclBackVariant
		if len(args) > 0 && args[0] == exclBackReview {
			t.st.ExclBack = exclBackReview
		}
		t.show(b.viewExclude(t))

	case "excl_add":
		if len(d.ExcludeWords) >= domain.MaxExcludeWords {
			t.alert(t.tr("wz.excl_full", domain.MaxExcludeWords))
			return
		}
		t.show(b.viewExcludeAsk(t))

	case "excl_del":
		if len(args) > 0 {
			d.ExcludeWords = removeTagged(d.ExcludeWords, args[0])
		}
		t.show(b.viewExclude(t))

	case "excl_done":
		if t.st.ExclBack == exclBackReview {
			t.show(b.viewReview(t))
			return
		}
		t.show(b.viewVariants(t))

	case "wiz_price":
		t.show(b.viewPrice(t))

	case "add_price":
		v, err := strconv.ParseFloat(firstArg(args), 64)
		if err != nil || !domain.ValidNumber(v) {
			v = 0
		}
		d.PriceMode, d.MinPrice, d.MaxPrice = domain.PriceByRange, 0, v
		t.show(b.viewOwner(t))

	case "add_price_free":
		d.PriceMode, d.MinPrice, d.MaxPrice = domain.PriceFree, 0, 0
		t.show(b.viewOwner(t))

	case "add_price_range":
		t.show(b.viewPriceAsk(t))

	case "wiz_owner":
		t.show(b.viewOwner(t))

	case "add_owner":
		switch firstArg(args) {
		case "private":
			d.OwnerType = "private"
		case "business":
			d.OwnerType = "business"
		default:
			d.OwnerType = ""
		}
		t.show(b.viewWhere(t))

	case "wiz_where":
		t.show(b.viewWhere(t))

	case "add_city":
		cityID, regionID := argInt(args, 0), argInt(args, 1)
		if cityID <= 0 {
			d.CityID, d.RegionID, d.CityName, d.DistrictID, d.DistrictName = 0, 0, "", 0, ""
			t.show(b.viewReview(t))
			return
		}
		b.applyCity(t, cityID, regionID, b.cityName(t, cityID, &regionID))

	case "add_city_manual":
		if b.geo != nil {
			// The fuzzy search needs the full city list; fetching it now keeps
			// the first misspelled name from waiting for ~25 requests.
			go b.geo.WarmGeoIndex()
		}
		t.show(b.viewCityAsk(t))

	case "add_city_pick":
		cityID, regionID := argInt(args, 0), argInt(args, 1)
		if cityID <= 0 {
			t.show(b.viewWhere(t))
			return
		}
		b.applyCity(t, cityID, regionID, b.cityName(t, cityID, &regionID))

	case "add_region_pick":
		regionID := argInt(args, 0)
		if regionID <= 0 {
			t.show(b.viewWhere(t))
			return
		}
		d.CityID, d.RegionID, d.DistrictID, d.DistrictName = 0, regionID, 0, ""
		d.CityName = ""
		if b.geo != nil {
			if region, err := b.geo.FetchRegion(regionID); err == nil && region != nil {
				d.CityName = region.Name
			}
		}
		t.show(b.viewReview(t))

	case "add_city_whole":
		d.DistrictID, d.DistrictName = 0, ""
		t.show(b.viewReview(t))

	case "add_district_pick":
		districtID, cityID := argInt(args, 0), argInt(args, 1)
		if len(args) == 3 && cityID == d.CityID && districtID > 0 {
			d.DistrictID = districtID
			d.DistrictName = t.st.districts[districtID]
		}
		t.show(b.viewReview(t))

	case "add_finish":
		b.finishWizard(t)
	}
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func removeValue(list []string, v string) []string {
	out := list[:0:0]
	for _, x := range list {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

// cityName resolves a city id to its name, filling in the region when the
// button did not carry one.
func (b *Bot) cityName(t *turn, cityID int, regionID *int) string {
	if c := catalog.FindCity(cityID); c != nil {
		if *regionID == 0 {
			*regionID = c.RegionID
		}
		return c.Name
	}
	if b.geo != nil {
		if city, err := b.geo.FetchCity(cityID); err == nil && city != nil {
			if *regionID == 0 {
				*regionID = city.RegionID
			}
			return city.Name
		}
	}
	return t.tr("sum.city_n", cityID)
}

// ---------- typed input ----------

// onStepText handles the text a wizard screen asked for.
func (b *Bot) onStepText(t *turn, text string) {
	t.wizard = true
	d := &t.st.Draft
	switch t.st.Step {
	case stepQuery:
		query := textutil.TruncateRunes(strings.Join(strings.Fields(text), " "), domain.MaxQueryRunes)
		d.Query = query
		d.QueryVariants = textutil.QueryVariants(query)
		t.st.Step = stepNone
		t.send(b.viewVariants(t))

	case stepVariant:
		norm := textutil.NormalizeQuery(text)
		switch {
		case norm == "":
			t.say("wz.variant_empty")
			return
		case textutil.StringInSlice(d.QueryVariants, norm):
			t.say("wz.variant_dup", esc(strings.TrimSpace(text)), esc(norm))
		case len(d.QueryVariants) >= domain.MaxVariants:
			t.say("wz.variants_full", domain.MaxVariants)
		default:
			d.QueryVariants = append(d.QueryVariants, norm)
		}
		t.st.Step = stepNone
		t.send(b.viewVariants(t))

	case stepExclude:
		for _, w := range strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
			w = strings.ToLower(strings.Join(strings.Fields(w), " "))
			if w == "" || textutil.StringInSlice(d.ExcludeWords, w) {
				continue
			}
			if len(d.ExcludeWords) >= domain.MaxExcludeWords {
				t.say("wz.excl_full", domain.MaxExcludeWords)
				break
			}
			d.ExcludeWords = append(d.ExcludeWords, textutil.TruncateRunes(w, domain.MaxExcludeRunes))
		}
		t.st.Step = stepNone
		t.send(b.viewExclude(t))

	case stepPrice:
		lo, hi, ok := parsePriceRange(text)
		if !ok || lo < 0 || hi < 0 {
			t.say("wz.price_bad")
			return
		}
		if lo == 0 && hi == 0 {
			// "0" asks for what costs nothing: the free-only mode.
			d.PriceMode, d.MinPrice, d.MaxPrice = domain.PriceFree, 0, 0
		} else {
			if hi > 0 && lo > hi {
				lo, hi = hi, lo
			}
			d.PriceMode, d.MinPrice, d.MaxPrice = domain.PriceByRange, lo, hi
		}
		t.st.Step = stepNone
		t.send(b.viewOwner(t))

	case stepFilterValue:
		defs := extraFilters(d.Category, d.BrandParam)
		idx := t.st.FilterIdx
		if idx < 0 || idx >= len(defs) {
			t.st.Step = stepNone
			t.send(b.viewExtras(t, 0))
			return
		}
		lo, hi, ok := parsePriceRange(text)
		if !ok || (lo <= 0 && hi <= 0) {
			t.say("wz.number_bad")
			return
		}
		if hi > 0 && lo > hi {
			lo, hi = hi, lo
		}
		if d.Ranges == nil {
			d.Ranges = map[string]domain.FilterRange{}
		}
		d.Ranges[defs[idx].Key] = domain.FilterRange{From: lo, To: hi}
		t.st.Step = stepNone
		t.send(b.viewFilterValues(t, idx, 0))

	case stepCity:
		query := strings.TrimSpace(text)
		if query == "" || b.geo == nil {
			return
		}
		results, err := b.geo.SearchLocations(query)
		if err != nil {
			t.say("wz.city_error", esc(err.Error()))
			return
		}
		if len(results) == 0 {
			t.say("wz.city_none")
			return
		}
		t.send(b.viewCityResults(t, query, results))
	}
}

// onURL turns a pasted OLX search link into a draft and shows the review.
func (b *Bot) onURL(t *turn, raw string) {
	parsed, err := olx.ParseSearchURL(raw)
	switch {
	case errors.Is(err, olx.ErrListingURL):
		t.say("wz.url_listing")
		return
	case err != nil:
		t.say("wz.url_bad")
		return
	}
	d := parsed.Search

	var notes []string
	if parsed.ExchangeDropped {
		notes = append(notes, t.tr("wz.url_exchange"))
	}
	if parsed.CurrencyDropped != "" {
		notes = append(notes, t.tr("wz.url_currency", esc(parsed.CurrencyDropped)))
	}
	// A city in the path ("/kiev/") is a slug; only the geo index knows it.
	if d.CityID == 0 && d.RegionID == 0 && len(parsed.Slugs) > 0 {
		resolved := false
		if b.geo != nil {
			for _, slug := range parsed.Slugs {
				city, region, ok := b.geo.ResolveSlug(slug)
				if !ok {
					continue
				}
				resolved = true
				if city.ID > 0 {
					d.CityID, d.CityName, d.RegionID = city.ID, city.Name, region.ID
				} else {
					d.RegionID, d.CityName = region.ID, region.Name
				}
				break
			}
		}
		if !resolved {
			notes = append(notes, t.tr("wz.url_city"))
		}
	} else if d.CityID > 0 && d.CityName == "" && b.geo != nil {
		if city, err := b.geo.FetchCity(d.CityID); err == nil && city != nil {
			d.CityName = city.Name
		}
	}

	d.Normalize()
	if err := d.Validate(); err != nil {
		_, _ = b.tg.SendHTML(t.ctx, t.chatID, "❌ "+esc(filterError(t.lang, err)), nil, false)
		return
	}

	t.resetState()
	t.wizard = true
	t.st.Active = true
	t.st.Draft = d
	t.st.Notes = notes
	t.st.Brands = map[string]string{}
	for i, id := range d.BrandValues {
		if i < len(d.BrandNames) {
			t.st.Brands[id] = d.BrandNames[i]
		}
	}
	view, km := b.viewReview(t)
	t.send(withNotice(t.tr("wz.url_ok"), view), km)
}
