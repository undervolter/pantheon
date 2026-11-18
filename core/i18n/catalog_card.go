package i18n

// Strings of the alert card.
func init() {
	register(map[string]map[Lang]string{
		"card.new":            {RU: "🔥 <b>Новое объявление!</b>", UK: "🔥 <b>Нове оголошення!</b>", EN: "🔥 <b>New listing!</b>"},
		"card.untitled":       {RU: "Без названия", UK: "Без назви", EN: "Untitled"},
		"card.no_description": {RU: "Без описания", UK: "Без опису", EN: "No description"},
		"card.price":          {RU: "Продаётся за <b>%s</b>", UK: "Продається за <b>%s</b>", EN: "Listed for <b>%s</b>"},
		"card.free":           {RU: "🆓 <b>Отдают бесплатно</b>", UK: "🆓 <b>Віддають безкоштовно</b>", EN: "🆓 <b>Given away for free</b>"},
		"card.exchange":       {RU: "🔄 <b>Обмен</b>", UK: "🔄 <b>Обмін</b>", EN: "🔄 <b>Exchange</b>"},
		"card.price_arranged": {RU: "Цена: <b>договорная</b>", UK: "Ціна: <b>договірна</b>", EN: "Price: <b>negotiable</b>"},
		"card.price_none":     {RU: "Цена не указана", UK: "Ціну не вказано", EN: "No price given"},
		"card.published":      {RU: "🕒 <b>Опубликовано в</b> %s", UK: "🕒 <b>Опубліковано о</b> %s", EN: "🕒 <b>Published at</b> %s"},
		"card.contact_chat":   {RU: "💬 Только сообщения в OLX", UK: "💬 Лише повідомлення в OLX", EN: "💬 OLX chat only"},
		"card.contact_phone":  {RU: "💬 Сообщения в OLX + номер по запросу", UK: "💬 Повідомлення в OLX + номер на запит", EN: "💬 OLX chat + phone on request"},
		"card.filter":         {RU: "🎯 <b>Фильтр:</b> %s", UK: "🎯 <b>Фільтр:</b> %s", EN: "🎯 <b>Filter:</b> %s"},
		"card.filter_footer":  {RU: "Фильтр: %s", UK: "Фільтр: %s", EN: "Filter: %s"},
		"card.params":         {RU: "Характеристики (%d)", UK: "Характеристики (%d)", EN: "Details (%d)"},
		"card.params_title":   {RU: "📋 <b>Характеристики:</b>", UK: "📋 <b>Характеристики:</b>", EN: "📋 <b>Details:</b>"},
		"card.link":           {RU: "Перейти на лот", UK: "Перейти до лота", EN: "Open listing"},
		"card.test_filter":    {RU: "Тест", UK: "Тест", EN: "Test"},

		"btn.open": {RU: "Открыть ↗", UK: "Відкрити ↗", EN: "Open ↗"},

		"price.uah":             {RU: "грн.", UK: "грн.", EN: "UAH"},
		"price.negotiable_note": {RU: "договорная", UK: "договірна", EN: "negotiable"},
		"price.negotiable":      {RU: "Договорная", UK: "Договірна", EN: "Negotiable"},
		"price.free":            {RU: "Бесплатно", UK: "Безкоштовно", EN: "Free"},
		"price.exchange":        {RU: "Обмен", UK: "Обмін", EN: "Exchange"},
		"price.none":            {RU: "Не указана", UK: "Не вказана", EN: "Not specified"},
	})
}
