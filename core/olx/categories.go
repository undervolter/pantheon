package olx

import "strings"

type CategoryItem struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Slug     string         `json:"slug"`
	Icon     string         `json:"icon,omitempty"`
	ParentID int            `json:"parent_id,omitempty"`
	Children []CategoryItem `json:"children,omitempty"`
}

type CityItem struct {
	CityID   int    `json:"city_id"`
	RegionID int    `json:"region_id"`
	Name     string `json:"name"`
}

var TopCities = []CityItem{
	{CityID: 0, RegionID: 0, Name: "🇺🇦 Вся Україна"},
	{CityID: 268, RegionID: 25, Name: "📍 Київ"},
	{CityID: 176, RegionID: 5, Name: "📍 Львів"},
	{CityID: 62, RegionID: 9, Name: "📍 Одеса"},
	{CityID: 121, RegionID: 21, Name: "📍 Дніпро"},
	{CityID: 280, RegionID: 8, Name: "📍 Харків"},
	{CityID: 194, RegionID: 17, Name: "📍 Запоріжжя"},
	{CityID: 313, RegionID: 24, Name: "📍 Вінниця"},
	{CityID: 284, RegionID: 15, Name: "📍 Полтава"},
	{CityID: 132, RegionID: 13, Name: "📍 Івано-Франківськ"},
	{CityID: 369, RegionID: 18, Name: "📍 Чернівці"},
	{CityID: 145, RegionID: 12, Name: "📍 Черкаси"},
	{CityID: 324, RegionID: 6, Name: "📍 Житомир"},
	{CityID: 106, RegionID: 19, Name: "📍 Миколаїв"},
	{CityID: 101, RegionID: 20, Name: "📍 Хмельницький"},
	{CityID: 124, RegionID: 14, Name: "📍 Рівне"},
	{CityID: 362, RegionID: 11, Name: "📍 Тернопіль"},
	{CityID: 72, RegionID: 22, Name: "📍 Луцьк"},
	{CityID: 247, RegionID: 10, Name: "📍 Ужгород"},
}

// Полное дерево категорий OLX: сгенерировано из __PRERENDERED_STATE__.
var TopCategories = []CategoryItem{
	{
		ID: 36, Name: "Дитячий світ", Slug: "detskiy-mir", Icon: "🧸", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 68, Name: "Дитячі коляски", Slug: "detskie-kolyaski", ParentID: 36,
			},
			{
				ID: 69, Name: "Дитячі меблі", Slug: "detskaya-mebel", ParentID: 36,
			},
			{
				ID: 71, Name: "Інші дитячі товари", Slug: "prochie-detskie-tovary", ParentID: 36,
			},
			{
				ID: 538, Name: "Дитячі автокрісла", Slug: "detskie-avtokresla", ParentID: 36,
				Children: []CategoryItem{
					{
						ID: 3903, Name: "Автокрісла", Slug: "avtokrisla", ParentID: 538,
					},
					{
						ID: 3904, Name: "Бази для автокрісел", Slug: "bazy-dlia-avtokrisel", ParentID: 538,
					},
					{
						ID: 3905, Name: "Аксесуари для автокрісел", Slug: "aksesuary-dlia-avtokrisel", ParentID: 538,
					},
				},
			},
			{
				ID: 539, Name: "Дитячий одяг", Slug: "detskaya-odezhda", ParentID: 36,
				Children: []CategoryItem{
					{
						ID: 70, Name: "Одяг для дівчаток", Slug: "odezhda-dlya-devochek", ParentID: 539,
					},
					{
						ID: 540, Name: "Одяг для новонароджених", Slug: "odezhda-dlya-novorozhdennyh", ParentID: 539,
					},
					{
						ID: 541, Name: "Одяг для хлопчиків", Slug: "odezhda-dlya-malchikov", ParentID: 539,
					},
				},
			},
			{
				ID: 542, Name: "Дитяче взуття", Slug: "detskaya-obuv", ParentID: 36,
			},
			{
				ID: 543, Name: "Іграшки", Slug: "igrushki", ParentID: 36,
				Children: []CategoryItem{
					{
						ID: 3564, Name: "Конструктори", Slug: "konstruktory", ParentID: 543,
					},
					{
						ID: 3565, Name: "Ляльки та пупси", Slug: "kukly-pupsy", ParentID: 543,
					},
					{
						ID: 3566, Name: "Аксесуари для ляльок та пупсів", Slug: "aksessuary-dlya-kukol-pupsov", ParentID: 543,
					},
					{
						ID: 3567, Name: "Розвиваючі іграшки", Slug: "razvivayushchie-igrushki", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3581, Name: "Мозаїки", Slug: "mozaiki", ParentID: 3567,
							},
							{
								ID: 3583, Name: "Іграшки сортери", Slug: "igrushki-sortery", ParentID: 3567,
							},
							{
								ID: 3585, Name: "Інтерактивні іграшки", Slug: "interaktivnye-igrushki", ParentID: 3567,
							},
							{
								ID: 3586, Name: "Бізіборди та куби", Slug: "bizibordy-kuby", ParentID: 3567,
							},
							{
								ID: 3587, Name: "Ігрові центри", Slug: "igrovye-tsentry", ParentID: 3567,
							},
							{
								ID: 3588, Name: "Дзиги", Slug: "yula", ParentID: 3567,
							},
							{
								ID: 3589, Name: "Іграшки неваляшки", Slug: "igrushki-nevalyashki", ParentID: 3567,
							},
							{
								ID: 3590, Name: "Іграшки каталки", Slug: "igrushki-katalki", ParentID: 3567,
							},
							{
								ID: 3591, Name: "Музичні інструменти", Slug: "muzykalnye-instrumenty", ParentID: 3567,
							},
							{
								ID: 3592, Name: "Розвиваючі килимки", Slug: "razvivayushchie-kovriki", ParentID: 3567,
							},
							{
								ID: 3593, Name: "Інші розвиваючі ігри", Slug: "drugoe", ParentID: 3567,
							},
						},
					},
					{
						ID: 3568, Name: "М'які іграшки", Slug: "myagkie-igrushki", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3582, Name: "М'які тварини", Slug: "myagkie-zhivotnye", ParentID: 3568,
							},
							{
								ID: 3584, Name: "Іграшки-подушки", Slug: "igrushki-podushki", ParentID: 3568,
							},
							{
								ID: 3670, Name: "Інші м'які іграшки", Slug: "drugoe", ParentID: 3568,
							},
						},
					},
					{
						ID: 3569, Name: "Настільні ігри та пазли", Slug: "nastolnye-igry-pazly", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3645, Name: "Пазли", Slug: "pazly", ParentID: 3569,
							},
							{
								ID: 3646, Name: "3D-пазли", Slug: "three-d-pazly", ParentID: 3569,
							},
							{
								ID: 3647, Name: "Дитячі головоломки", Slug: "detskie-golovolomki", ParentID: 3569,
							},
							{
								ID: 3648, Name: "Дитячі настільні ігри", Slug: "detskie-nastolnye-igry", ParentID: 3569,
							},
							{
								ID: 3649, Name: "Інші настільні ігри", Slug: "drugoe", ParentID: 3569,
							},
						},
					},
					{
						ID: 3570, Name: "Брязкальця та підвіски", Slug: "pogremushki-podveski", ParentID: 543,
					},
					{
						ID: 3571, Name: "Іграшки для купання", Slug: "igrushki-dlya-kupaniya", ParentID: 543,
					},
					{
						ID: 3572, Name: "Іграшкова зброя", Slug: "igrushechnoe-oruzhie", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3650, Name: "Пістолети", Slug: "pistolety", ParentID: 3572,
							},
							{
								ID: 3652, Name: "Автомати", Slug: "avtomaty", ParentID: 3572,
							},
							{
								ID: 3654, Name: "Бластери", Slug: "blastery", ParentID: 3572,
							},
							{
								ID: 3655, Name: "Мечі та щити", Slug: "mechi-shchity", ParentID: 3572,
							},
							{
								ID: 3656, Name: "Луки та арбалети", Slug: "luki-arbalety", ParentID: 3572,
							},
							{
								ID: 3657, Name: "Водяна зброя", Slug: "vodyanoe-oruzhie", ParentID: 3572,
							},
							{
								ID: 3658, Name: "Інша іграшкова зброя", Slug: "drugoe", ParentID: 3572,
							},
						},
					},
					{
						ID: 3573, Name: "Ігрові фігурки", Slug: "igrovye-figurki", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3651, Name: "Фігурки персонажів", Slug: "figurki-personazhey", ParentID: 3573,
							},
							{
								ID: 3653, Name: "Фігурки тварин", Slug: "figurki-zhivotnyh", ParentID: 3573,
							},
							{
								ID: 3672, Name: "Інші фігурки", Slug: "drugoe", ParentID: 3573,
							},
						},
					},
					{
						ID: 3574, Name: "Іграшковий транспорт", Slug: "igrushechniy-transport", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3659, Name: "Залізниці та потяги", Slug: "zheleznye-dorogi-poezda", ParentID: 3574,
							},
							{
								ID: 3661, Name: "Машинки", Slug: "mashinki", ParentID: 3574,
							},
							{
								ID: 3663, Name: "Мотоцикли", Slug: "mototsikly", ParentID: 3574,
							},
							{
								ID: 3665, Name: "Танки", Slug: "tanki", ParentID: 3574,
							},
							{
								ID: 3666, Name: "Гелікоптери", Slug: "vertolety", ParentID: 3574,
							},
							{
								ID: 3667, Name: "Трактори", Slug: "traktory", ParentID: 3574,
							},
							{
								ID: 3668, Name: "Літаки", Slug: "samolety", ParentID: 3574,
							},
							{
								ID: 3669, Name: "Інший іграшковий транспорт", Slug: "drugoe", ParentID: 3574,
							},
						},
					},
					{
						ID: 3575, Name: "Набори для творчості", Slug: "nabory-dlya-tvorchestva", ParentID: 543,
					},
					{
						ID: 3576, Name: "Ігрові набори", Slug: "igrovye-nabory", ParentID: 543,
						Children: []CategoryItem{
							{
								ID: 3660, Name: "Косметика для дівчаток", Slug: "kosmetika-dlya-devochek", ParentID: 3576,
							},
							{
								ID: 3662, Name: "Набори професій", Slug: "nabory-professiy", ParentID: 3576,
							},
							{
								ID: 3664, Name: "Дитячі кухні", Slug: "detskie-kuhni", ParentID: 3576,
							},
							{
								ID: 3673, Name: "Інші ігрові набори", Slug: "drugoe", ParentID: 3576,
							},
						},
					},
					{
						ID: 3577, Name: "Іграшки для пісочниці", Slug: "igrushki-dlya-pesochnitsy", ParentID: 543,
					},
					{
						ID: 3578, Name: "Дитячі м'ячі", Slug: "detskie-myachi", ParentID: 543,
					},
					{
						ID: 3579, Name: "Дитячі маски", Slug: "detskie-maski", ParentID: 543,
					},
					{
						ID: 3580, Name: "Інші іграшки для дітей", Slug: "drugoe", ParentID: 543,
					},
				},
			},
			{
				ID: 544, Name: "Годування", Slug: "kormlenie", ParentID: 36,
			},
			{
				ID: 545, Name: "Товари для школярів", Slug: "tovary-dlya-shkolnikov", ParentID: 36,
			},
			{
				ID: 889, Name: "Дитячий транспорт", Slug: "detskiy-transport", ParentID: 36,
			},
		},
	},
	{
		ID: 1, Name: "Нерухомість", Slug: "nedvizhimost", Icon: "🏠", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 206, Name: "Будинки", Slug: "doma", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 330, Name: "Довгострокова оренда будинків", Slug: "arenda-domov", ParentID: 206,
					},
					{
						ID: 1602, Name: "Продаж будинків", Slug: "prodazha-domov", ParentID: 206,
					},
				},
			},
			{
				ID: 1522, Name: "Нерухомість за кордоном", Slug: "nedvizhimost-za-rubezhom", ParentID: 1,
			},
			{
				ID: 1606, Name: "Земля", Slug: "zemlya", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 20, Name: "Оренда землі", Slug: "arenda-zemli", ParentID: 1606,
					},
					{
						ID: 1608, Name: "Продаж землі", Slug: "prodazha-zemli", ParentID: 1606,
					},
				},
			},
			{
				ID: 1610, Name: "Комерційна нерухомість", Slug: "kommercheskaya-nedvizhimost", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 1612, Name: "Продаж комерційної нерухомості", Slug: "prodazha-kommercheskoy-nedvizhimosti", ParentID: 1610,
						Children: []CategoryItem{
							{
								ID: 3679, Name: "Продаж торгових приміщень", Slug: "torgovye-pomeshcheniya", ParentID: 1612,
							},
							{
								ID: 3680, Name: "Продаж МАФів", Slug: "mafy", ParentID: 1612,
							},
							{
								ID: 3681, Name: "Продаж офісів", Slug: "ofisy", ParentID: 1612,
							},
							{
								ID: 3682, Name: "Продаж складів, ангарів", Slug: "sklady-angary", ParentID: 1612,
							},
							{
								ID: 3683, Name: "Продаж кав'ярень, кафе, ресторанів", Slug: "kafe-restorany-kofeyni", ParentID: 1612,
							},
							{
								ID: 3684, Name: "Продаж готелів, баз відпочинку", Slug: "oteli-bazy-otdyha", ParentID: 1612,
							},
							{
								ID: 3685, Name: "Продаж СТО, автомийок, АЗС", Slug: "sto-avtomoyki-azs", ParentID: 1612,
							},
							{
								ID: 3686, Name: "Продаж салонів краси, клінік", Slug: "salony-kabinety-krasoty", ParentID: 1612,
							},
							{
								ID: 3687, Name: "Продаж виробничих приміщень", Slug: "proizvodstvennye-pomeshcheniya", ParentID: 1612,
							},
							{
								ID: 3688, Name: "Продаж приміщень вільного призначення", Slug: "pomeshcheniya-svobodnogo-naznacheniya", ParentID: 1612,
							},
							{
								ID: 3689, Name: "Продаж інших приміщень", Slug: "drugoe", ParentID: 1612,
							},
						},
					},
					{
						ID: 1614, Name: "Оренда комерційної нерухомості", Slug: "arenda-kommercheskoy-nedvizhimosti", ParentID: 1610,
						Children: []CategoryItem{
							{
								ID: 3690, Name: "Оренда офісів", Slug: "ofisy", ParentID: 1614,
							},
							{
								ID: 3691, Name: "Оренда виробничих приміщень, ферм", Slug: "proizvodstvennye-pomeshcheniya-fermy", ParentID: 1614,
							},
							{
								ID: 3692, Name: "Оренда торгових приміщень", Slug: "torgovye-pomeshcheniya", ParentID: 1614,
							},
							{
								ID: 3693, Name: "Оренда складів, ангарів", Slug: "sklady-angary", ParentID: 1614,
							},
							{
								ID: 3694, Name: "Оренда кафе, ресторанів, кав'ярень", Slug: "kafe-restorany-kofeyni", ParentID: 1614,
							},
							{
								ID: 3695, Name: "Оренда МАФів", Slug: "mafy", ParentID: 1614,
							},
							{
								ID: 3696, Name: "Оренда СТО, автомийок, АЗС", Slug: "sto-avtomoyki-azs", ParentID: 1614,
							},
							{
								ID: 3697, Name: "Оренда салонів, кабінетів краси", Slug: "salony-kabinety-krasoty", ParentID: 1614,
							},
							{
								ID: 3698, Name: "Оренда залів, спортивних приміщень", Slug: "zaly-sportivnye-pomeshcheniya", ParentID: 1614,
							},
							{
								ID: 3699, Name: "Оренда медичних приміщень", Slug: "meditsinskie-pomeshcheniya", ParentID: 1614,
							},
							{
								ID: 3700, Name: "Оренда приміщення вільного призначення", Slug: "pomeshcheniya-svobodnogo-naznacheniya", ParentID: 1614,
							},
							{
								ID: 3701, Name: "Оренда інших приміщень", Slug: "drugoe", ParentID: 1614,
							},
						},
					},
					{
						ID: 3298, Name: "Коворкінги", Slug: "kovorkingi", ParentID: 1610,
					},
				},
			},
			{
				ID: 1616, Name: "Гаражі, парковки", Slug: "garazhy-parkovki", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 21, Name: "Продаж гаражів, парковок", Slug: "prodazha-garazhey-parkovok", ParentID: 1616,
					},
					{
						ID: 28, Name: "Оренда гаражів, парковок", Slug: "arenda-garazhey-parkovok", ParentID: 1616,
					},
				},
			},
			{
				ID: 1754, Name: "Кімнати", Slug: "komnaty", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 1755, Name: "Продаж кімнат", Slug: "prodazha-komnat", ParentID: 1754,
					},
					{
						ID: 1756, Name: "Довгострокова оренда кімнат", Slug: "dolgosrochnaya-arenda-komnat", ParentID: 1754,
					},
				},
			},
			{
				ID: 1757, Name: "Квартири", Slug: "kvartiry", ParentID: 1,
				Children: []CategoryItem{
					{
						ID: 1758, Name: "Продаж квартир", Slug: "prodazha-kvartir", ParentID: 1757,
					},
					{
						ID: 1760, Name: "Довгострокова оренда квартир", Slug: "dolgosrochnaya-arenda-kvartir", ParentID: 1757,
					},
				},
			},
		},
	},
	{
		ID: 1532, Name: "Авто", Slug: "transport", Icon: "🚗", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 108, Name: "Легкові автомобілі", Slug: "legkovye-avtomobili", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 963, Name: "Acura", Slug: "acura", ParentID: 108,
					},
					{
						ID: 965, Name: "Alfa Romeo", Slug: "alfa-romeo", ParentID: 108,
					},
					{
						ID: 969, Name: "Audi", Slug: "audi", ParentID: 108,
					},
					{
						ID: 973, Name: "BMW", Slug: "bmw", ParentID: 108,
					},
					{
						ID: 981, Name: "Chery", Slug: "chery", ParentID: 108,
					},
					{
						ID: 983, Name: "Chevrolet", Slug: "chevrolet", ParentID: 108,
					},
					{
						ID: 985, Name: "Chrysler", Slug: "chrysler", ParentID: 108,
					},
					{
						ID: 987, Name: "Citroen", Slug: "citroen", ParentID: 108,
					},
					{
						ID: 989, Name: "Dacia", Slug: "dacia", ParentID: 108,
					},
					{
						ID: 991, Name: "Daewoo", Slug: "daewoo", ParentID: 108,
					},
					{
						ID: 997, Name: "Dodge", Slug: "dodge", ParentID: 108,
					},
					{
						ID: 1007, Name: "Fiat", Slug: "fiat", ParentID: 108,
					},
					{
						ID: 1009, Name: "Ford", Slug: "ford", ParentID: 108,
					},
					{
						ID: 1011, Name: "Geely", Slug: "geely", ParentID: 108,
					},
					{
						ID: 1015, Name: "Great Wall", Slug: "great-wall", ParentID: 108,
					},
					{
						ID: 1019, Name: "Honda", Slug: "honda", ParentID: 108,
					},
					{
						ID: 1023, Name: "Hyundai", Slug: "hyundai", ParentID: 108,
					},
					{
						ID: 1035, Name: "Jeep", Slug: "jeep", ParentID: 108,
					},
					{
						ID: 1037, Name: "Kia", Slug: "kia", ParentID: 108,
					},
					{
						ID: 1043, Name: "Land Rover", Slug: "land-rover", ParentID: 108,
					},
					{
						ID: 1045, Name: "Lexus", Slug: "lexus", ParentID: 108,
					},
					{
						ID: 1053, Name: "Mazda", Slug: "mazda", ParentID: 108,
					},
					{
						ID: 1055, Name: "Mercedes-Benz", Slug: "mercedes-benz", ParentID: 108,
					},
					{
						ID: 1063, Name: "Mitsubishi", Slug: "mitsubishi", ParentID: 108,
					},
					{
						ID: 1065, Name: "Nissan", Slug: "nissan", ParentID: 108,
					},
					{
						ID: 1067, Name: "Oldsmobile", Slug: "oldsmobile", ParentID: 108,
					},
					{
						ID: 1069, Name: "Opel", Slug: "opel", ParentID: 108,
					},
					{
						ID: 1071, Name: "Peugeot", Slug: "peugeot", ParentID: 108,
					},
					{
						ID: 1077, Name: "Porsche", Slug: "porsche", ParentID: 108,
					},
					{
						ID: 1081, Name: "Renault", Slug: "renault", ParentID: 108,
					},
					{
						ID: 1093, Name: "Seat", Slug: "seat", ParentID: 108,
					},
					{
						ID: 1095, Name: "Skoda", Slug: "skoda", ParentID: 108,
					},
					{
						ID: 1097, Name: "Smart", Slug: "smart", ParentID: 108,
					},
					{
						ID: 1101, Name: "SsangYong", Slug: "ssangyong", ParentID: 108,
					},
					{
						ID: 1103, Name: "Subaru", Slug: "subaru", ParentID: 108,
					},
					{
						ID: 1105, Name: "Suzuki", Slug: "suzuki", ParentID: 108,
					},
					{
						ID: 1107, Name: "Toyota", Slug: "toyota", ParentID: 108,
					},
					{
						ID: 1109, Name: "Volkswagen", Slug: "volkswagen", ParentID: 108,
					},
					{
						ID: 1111, Name: "Volvo", Slug: "volvo", ParentID: 108,
					},
					{
						ID: 1115, Name: "ВАЗ", Slug: "vaz", ParentID: 108,
					},
					{
						ID: 1117, Name: "ГАЗ", Slug: "gaz", ParentID: 108,
					},
					{
						ID: 1119, Name: "ЗАЗ", Slug: "zaz", ParentID: 108,
					},
					{
						ID: 1121, Name: "ИЖ", Slug: "izh", ParentID: 108,
					},
					{
						ID: 1123, Name: "Москвич / АЗЛК", Slug: "moskvich-azlk", ParentID: 108,
					},
					{
						ID: 1131, Name: "УАЗ", Slug: "uaz", ParentID: 108,
					},
					{
						ID: 1133, Name: "Інші", Slug: "drugie", ParentID: 108,
					},
					{
						ID: 1361, Name: "Roewe", Slug: "roewe", ParentID: 108,
					},
					{
						ID: 1413, Name: "Brilliance", Slug: "brilliance", ParentID: 108,
					},
					{
						ID: 1534, Name: "Infiniti", Slug: "infiniti", ParentID: 108,
					},
					{
						ID: 1536, Name: "ЛуАЗ", Slug: "luaz", ParentID: 108,
					},
					{
						ID: 1538, Name: "Jaguar", Slug: "jaguar", ParentID: 108,
					},
					{
						ID: 1540, Name: "MINI", Slug: "mini", ParentID: 108,
					},
					{
						ID: 1542, Name: "Rover", Slug: "rover", ParentID: 108,
					},
					{
						ID: 1544, Name: "Bentley", Slug: "bentley", ParentID: 108,
					},
					{
						ID: 1546, Name: "Cadillac", Slug: "cadillac", ParentID: 108,
					},
					{
						ID: 1548, Name: "Lincoln", Slug: "lincoln", ParentID: 108,
					},
					{
						ID: 1550, Name: "Iveco", Slug: "iveco", ParentID: 108,
					},
					{
						ID: 1552, Name: "Lancia", Slug: "lancia", ParentID: 108,
					},
					{
						ID: 1554, Name: "Saab", Slug: "saab", ParentID: 108,
					},
					{
						ID: 1556, Name: "Tesla", Slug: "tesla", ParentID: 108,
					},
					{
						ID: 1558, Name: "BYD", Slug: "byd", ParentID: 108,
					},
					{
						ID: 1560, Name: "Daihatsu", Slug: "daihatsu", ParentID: 108,
					},
					{
						ID: 1562, Name: "Ravon", Slug: "ravon", ParentID: 108,
					},
					{
						ID: 1564, Name: "Lifan", Slug: "lifan", ParentID: 108,
					},
					{
						ID: 1566, Name: "MG", Slug: "mg", ParentID: 108,
					},
					{
						ID: 1568, Name: "Богдан", Slug: "bogdan", ParentID: 108,
					},
					{
						ID: 1570, Name: "Isuzu", Slug: "isuzu", ParentID: 108,
					},
					{
						ID: 1572, Name: "Samand", Slug: "samand", ParentID: 108,
					},
					{
						ID: 1574, Name: "Dadi", Slug: "dadi", ParentID: 108,
					},
					{
						ID: 1576, Name: "Maserati", Slug: "maserati", ParentID: 108,
					},
					{
						ID: 1578, Name: "Wartburg", Slug: "wartburg", ParentID: 108,
					},
					{
						ID: 1580, Name: "РАФ", Slug: "raf", ParentID: 108,
					},
					{
						ID: 1582, Name: "Hummer", Slug: "hummer", ParentID: 108,
					},
					{
						ID: 1584, Name: "Pontiac", Slug: "pontiac", ParentID: 108,
					},
					{
						ID: 1586, Name: "JAC", Slug: "jac", ParentID: 108,
					},
					{
						ID: 1588, Name: "Ferrari", Slug: "ferrari", ParentID: 108,
					},
					{
						ID: 1590, Name: "ZX", Slug: "zx", ParentID: 108,
					},
					{
						ID: 1592, Name: "Chana", Slug: "chana", ParentID: 108,
					},
					{
						ID: 1594, Name: "GMC", Slug: "gmc", ParentID: 108,
					},
					{
						ID: 1762, Name: "Aston Martin", Slug: "aston-martin", ParentID: 108,
					},
					{
						ID: 1763, Name: "BMW-Alpina", Slug: "bmw-alpina", ParentID: 108,
					},
					{
						ID: 1764, Name: "Buick", Slug: "buick", ParentID: 108,
					},
					{
						ID: 1765, Name: "FAW", Slug: "faw", ParentID: 108,
					},
					{
						ID: 1766, Name: "Groz", Slug: "groz", ParentID: 108,
					},
					{
						ID: 1767, Name: "Lamborghini", Slug: "lamborghini", ParentID: 108,
					},
					{
						ID: 1768, Name: "Lotus", Slug: "lotus", ParentID: 108,
					},
					{
						ID: 1769, Name: "Maybach", Slug: "maybach", ParentID: 108,
					},
					{
						ID: 1770, Name: "McLaren", Slug: "mclaren", ParentID: 108,
					},
					{
						ID: 1771, Name: "Mercury", Slug: "mercury", ParentID: 108,
					},
					{
						ID: 1772, Name: "Proton", Slug: "proton", ParentID: 108,
					},
					{
						ID: 1773, Name: "Ram", Slug: "ram", ParentID: 108,
					},
					{
						ID: 1774, Name: "Rolls-Royce", Slug: "rolls-royce", ParentID: 108,
					},
					{
						ID: 1775, Name: "Samsung", Slug: "samsung", ParentID: 108,
					},
					{
						ID: 1776, Name: "Shelby", Slug: "shelby", ParentID: 108,
					},
					{
						ID: 1777, Name: "TATA", Slug: "tata", ParentID: 108,
					},
					{
						ID: 3706, Name: "Polestar", Slug: "polestar", ParentID: 108,
					},
				},
			},
			{
				ID: 109, Name: "Мото", Slug: "moto", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 264, Name: "Мотоцикли", Slug: "mototsikly", ParentID: 109,
					},
					{
						ID: 265, Name: "Мопеди / скутери", Slug: "mopedy-skutery", ParentID: 109,
					},
					{
						ID: 266, Name: "Квадроцикли", Slug: "kvadrotsikly", ParentID: 109,
					},
					{
						ID: 326, Name: "Мото - інше", Slug: "moto-prochee", ParentID: 109,
					},
					{
						ID: 1936, Name: "Максі-скутер", Slug: "maksi-skuter", ParentID: 109,
					},
					{
						ID: 1937, Name: "Снігохід", Slug: "snihokhid", ParentID: 109,
					},
					{
						ID: 1938, Name: "Баггі", Slug: "bahhi", ParentID: 109,
					},
					{
						ID: 1939, Name: "Картинг", Slug: "kartynh", ParentID: 109,
					},
					{
						ID: 1940, Name: "Електросамокати / гіроскутери", Slug: "elektrosamokaty-hiroskutery", ParentID: 109,
					},
					{
						ID: 1941, Name: "Електроскутери", Slug: "elektroskutery", ParentID: 109,
					},
					{
						ID: 1942, Name: "Електромотоцикли", Slug: "elektromotosykly", ParentID: 109,
					},
					{
						ID: 1943, Name: "Електроквадрацикли", Slug: "elektrokvadratsykly", ParentID: 109,
					},
				},
			},
			{
				ID: 110, Name: "Інший транспорт", Slug: "drugoy-transport", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 2680, Name: "Літаки", Slug: "samolety", ParentID: 110,
					},
					{
						ID: 2681, Name: "Гелікоптери", Slug: "vertolety", ParentID: 110,
					},
					{
						ID: 2682, Name: "Дельтаплани", Slug: "deltaplany", ParentID: 110,
					},
					{
						ID: 2683, Name: "Карети", Slug: "karety", ParentID: 110,
					},
					{
						ID: 2684, Name: "Залізничні тепловози", Slug: "zhd-teplovozy", ParentID: 110,
					},
					{
						ID: 2685, Name: "Залізничні вагони", Slug: "zhd-vagony", ParentID: 110,
					},
					{
						ID: 2686, Name: "Інші види транспорту", Slug: "drugiye-vidy-transporta", ParentID: 110,
					},
				},
			},
			{
				ID: 222, Name: "Вантажні автомобілі", Slug: "gruzovye-avtomobili", ParentID: 1532,
			},
			{
				ID: 225, Name: "Автобуси", Slug: "avtobusy", ParentID: 1532,
			},
			{
				ID: 253, Name: "Водний транспорт ", Slug: "vodnyy-transport", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 1985, Name: "Яхта", Slug: "yakhta", ParentID: 253,
					},
					{
						ID: 1986, Name: "Катер", Slug: "kater", ParentID: 253,
					},
					{
						ID: 1987, Name: "Моторний човен", Slug: "motornyi-choven", ParentID: 253,
					},
					{
						ID: 1988, Name: "Катамаран", Slug: "katamaran", ParentID: 253,
					},
					{
						ID: 1989, Name: "Гребний човен", Slug: "hrebnyi-choven", ParentID: 253,
					},
					{
						ID: 1990, Name: "Гідроцикл", Slug: "hidrotsykl", ParentID: 253,
					},
					{
						ID: 1991, Name: "Каяк", Slug: "kaiak", ParentID: 253,
					},
					{
						ID: 1992, Name: "Інший водний транспорт", Slug: "inshyi-vodnyi-transport", ParentID: 253,
					},
				},
			},
			{
				ID: 261, Name: "Причепи / будинки на колесах", Slug: "pritsepy-doma-na-kolesah", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 1993, Name: "Легковий причіп", Slug: "lehkovyi-prychip", ParentID: 261,
					},
					{
						ID: 1994, Name: "Лафет", Slug: "lafet", ParentID: 261,
					},
					{
						ID: 1995, Name: "Напівпричіп", Slug: "napivprychip", ParentID: 261,
					},
					{
						ID: 1996, Name: "Бортовий напівпричіп", Slug: "bortovyi-napivprychip", ParentID: 261,
					},
					{
						ID: 1997, Name: "Вантажний причіп", Slug: "vantazhnyi-prychip", ParentID: 261,
					},
					{
						ID: 1998, Name: "Причіп контейнеровоз", Slug: "prychip-konteinerovoz", ParentID: 261,
					},
					{
						ID: 1999, Name: "Причіп зерновоз", Slug: "prychip-zernovoz", ParentID: 261,
					},
					{
						ID: 2000, Name: "Самоскид напівпричіп", Slug: "samoskyd-napivprychip", ParentID: 261,
					},
					{
						ID: 2001, Name: "Трал", Slug: "tral", ParentID: 261,
					},
					{
						ID: 2002, Name: "Торговий причіп", Slug: "torhovyi-prychip", ParentID: 261,
					},
					{
						ID: 2003, Name: "Тракторний причіп", Slug: "traktornyi-prychip", ParentID: 261,
					},
					{
						ID: 2004, Name: "Причіп фургон", Slug: "prychip-furhon", ParentID: 261,
					},
					{
						ID: 2005, Name: "Причепи цистерни", Slug: "prychepy-tsysterny", ParentID: 261,
					},
					{
						ID: 2006, Name: "Автодім", Slug: "avtodim", ParentID: 261,
					},
					{
						ID: 2007, Name: "Кемпінг", Slug: "kempinh", ParentID: 261,
					},
					{
						ID: 2008, Name: "Причіп дача", Slug: "prychip-dacha", ParentID: 261,
					},
					{
						ID: 2009, Name: "Інші причепи", Slug: "inshi-prychepy", ParentID: 261,
					},
				},
			},
			{
				ID: 273, Name: "Спецтехніка", Slug: "spetstehnika", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 293, Name: "Бульдозери / трактори", Slug: "buldozery-traktory", ParentID: 273,
					},
					{
						ID: 295, Name: "Комунальна техніка", Slug: "kommunalnaya-tehnika", ParentID: 273,
					},
					{
						ID: 297, Name: "Навантажувачі", Slug: "pogruzchiki", ParentID: 273,
					},
					{
						ID: 298, Name: "Самоскиди", Slug: "samosvaly", ParentID: 273,
					},
					{
						ID: 299, Name: "Будівельна техніка", Slug: "stroitelnaya-tehnika", ParentID: 273,
					},
					{
						ID: 301, Name: "Екскаватори", Slug: "ekskavatory", ParentID: 273,
					},
					{
						ID: 302, Name: "Обладнання для спецтехніки", Slug: "oborudovanie-dlya-spetstehniki", ParentID: 273,
					},
					{
						ID: 303, Name: "Інша спецтехніка", Slug: "prochaya-spetstehnika", ParentID: 273,
					},
					{
						ID: 1965, Name: "Автовишки", Slug: "avtovyshky", ParentID: 273,
					},
					{
						ID: 1966, Name: "Автокрани та маніпулятори", Slug: "avtokrany-ta-manipuliatory", ParentID: 273,
					},
					{
						ID: 1967, Name: "Автогрейдер", Slug: "avtohreider", ParentID: 273,
					},
					{
						ID: 1968, Name: "Асфальтоукладач", Slug: "asfaltoukladach", ParentID: 273,
					},
					{
						ID: 1969, Name: "Дорожній каток", Slug: "dorozhnii-katok", ParentID: 273,
					},
					{
						ID: 1970, Name: "Всюдиходи", Slug: "vsiudykhody", ParentID: 273,
					},
					{
						ID: 1971, Name: "Мінінавантажувач", Slug: "mininavantazhuvach", ParentID: 273,
					},
					{
						ID: 1972, Name: "Бетонозмішувачі (міксери)", Slug: "betonozmishuvachi", ParentID: 273,
					},
					{
						ID: 1973, Name: "Бетононасоси", Slug: "betononasosy", ParentID: 273,
					},
					{
						ID: 1974, Name: "Бурова установка", Slug: "burova-ustanovka", ParentID: 273,
					},
					{
						ID: 1975, Name: "Сміттєвоз", Slug: "smittievoz", ParentID: 273,
					},
					{
						ID: 1976, Name: "Лісозаготівельна техніка", Slug: "lisozahotivelna-tekhnika", ParentID: 273,
					},
					{
						ID: 1977, Name: "Машина асенізатор (вакуумна)", Slug: "mashyna-asenizator", ParentID: 273,
					},
				},
			},
			{
				ID: 304, Name: "Сільгосптехніка", Slug: "selhoztehnika", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 1978, Name: "Трактор сільськогосподарський", Slug: "traktor-silskohospodarskyi", ParentID: 304,
					},
					{
						ID: 1979, Name: "Мотоблок", Slug: "motoblok", ParentID: 304,
					},
					{
						ID: 1980, Name: "Мотокультиватор", Slug: "motokultyvator", ParentID: 304,
					},
					{
						ID: 1981, Name: "Міні-трактор", Slug: "mini-traktor", ParentID: 304,
					},
					{
						ID: 1982, Name: "Комбайн", Slug: "kombain", ParentID: 304,
					},
					{
						ID: 1983, Name: "Навісне та додаткове обладнання", Slug: "navisne-obladnannia", ParentID: 304,
					},
					{
						ID: 1984, Name: "Інша сільгосптехніка", Slug: "insha-silhosptekhnika", ParentID: 304,
					},
				},
			},
			{
				ID: 1701, Name: "Автомобілі з Польщі", Slug: "avtomobili-iz-polshi", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 1702, Name: "Alfa Romeo", Slug: "alfa-romeo", ParentID: 1701,
					},
					{
						ID: 1703, Name: "Audi", Slug: "audi", ParentID: 1701,
					},
					{
						ID: 1704, Name: "BMW", Slug: "bmw", ParentID: 1701,
					},
					{
						ID: 1705, Name: "Chevrolet", Slug: "chevrolet", ParentID: 1701,
					},
					{
						ID: 1706, Name: "Chrysler", Slug: "chrysler", ParentID: 1701,
					},
					{
						ID: 1707, Name: "Citroen", Slug: "citroen", ParentID: 1701,
					},
					{
						ID: 1708, Name: "Daewoo", Slug: "daewoo", ParentID: 1701,
					},
					{
						ID: 1709, Name: "Fiat", Slug: "fiat", ParentID: 1701,
					},
					{
						ID: 1710, Name: "Ford", Slug: "ford", ParentID: 1701,
					},
					{
						ID: 1711, Name: "Honda", Slug: "honda", ParentID: 1701,
					},
					{
						ID: 1712, Name: "Hyundai", Slug: "hyundai", ParentID: 1701,
					},
					{
						ID: 1713, Name: "Kia", Slug: "kia", ParentID: 1701,
					},
					{
						ID: 1714, Name: "Lancia", Slug: "lancia", ParentID: 1701,
					},
					{
						ID: 1715, Name: "Mazda", Slug: "mazda", ParentID: 1701,
					},
					{
						ID: 1716, Name: "Mercedes-Benz", Slug: "mercedes-benz", ParentID: 1701,
					},
					{
						ID: 1717, Name: "Mitsubishi", Slug: "mitsubishi", ParentID: 1701,
					},
					{
						ID: 1718, Name: "Nissan", Slug: "nissan", ParentID: 1701,
					},
					{
						ID: 1719, Name: "Opel", Slug: "opel", ParentID: 1701,
					},
					{
						ID: 1720, Name: "Peugeot", Slug: "peugeot", ParentID: 1701,
					},
					{
						ID: 1721, Name: "Renault", Slug: "renault", ParentID: 1701,
					},
					{
						ID: 1722, Name: "Saab", Slug: "saab", ParentID: 1701,
					},
					{
						ID: 1723, Name: "Seat", Slug: "seat", ParentID: 1701,
					},
					{
						ID: 1724, Name: "Skoda", Slug: "skoda", ParentID: 1701,
					},
					{
						ID: 1725, Name: "Subaru", Slug: "subaru", ParentID: 1701,
					},
					{
						ID: 1726, Name: "Suzuki", Slug: "suzuki", ParentID: 1701,
					},
					{
						ID: 1727, Name: "Toyota", Slug: "toyota", ParentID: 1701,
					},
					{
						ID: 1728, Name: "Volkswagen", Slug: "volkswagen", ParentID: 1701,
					},
					{
						ID: 1729, Name: "Volvo", Slug: "volvo", ParentID: 1701,
					},
					{
						ID: 1730, Name: "Pozostałe osobowe", Slug: "pozostale", ParentID: 1701,
					},
					{
						ID: 1731, Name: "Rover", Slug: "rover", ParentID: 1701,
					},
					{
						ID: 1732, Name: "Land Rover", Slug: "landrover", ParentID: 1701,
					},
					{
						ID: 1733, Name: "Jeep", Slug: "jeep", ParentID: 1701,
					},
					{
						ID: 1734, Name: "Polonez", Slug: "polonez", ParentID: 1701,
					},
					{
						ID: 1735, Name: "Smart", Slug: "smart", ParentID: 1701,
					},
					{
						ID: 1736, Name: "Mini", Slug: "mini", ParentID: 1701,
					},
					{
						ID: 1737, Name: "Porsche", Slug: "porsche", ParentID: 1701,
					},
					{
						ID: 1738, Name: "Dodge", Slug: "dodge", ParentID: 1701,
					},
					{
						ID: 1739, Name: "Jaguar", Slug: "jaguar", ParentID: 1701,
					},
					{
						ID: 1740, Name: "Dacia", Slug: "dacia", ParentID: 1701,
					},
					{
						ID: 1741, Name: "Daihatsu", Slug: "daihatsu", ParentID: 1701,
					},
					{
						ID: 1742, Name: "Lexus", Slug: "lexus", ParentID: 1701,
					},
					{
						ID: 1743, Name: "Cadillac", Slug: "cadillac", ParentID: 1701,
					},
					{
						ID: 1744, Name: "Infiniti", Slug: "infiniti", ParentID: 1701,
					},
					{
						ID: 1745, Name: "Aixam", Slug: "aixam", ParentID: 1701,
					},
					{
						ID: 1746, Name: "SsangYong", Slug: "ssangyong", ParentID: 1701,
					},
				},
			},
			{
				ID: 1748, Name: "Вантажівки та спецтехніка з Польщі", Slug: "gruzoviki-i-spetstehnika-iz-polshi", ParentID: 1532,
				Children: []CategoryItem{
					{
						ID: 1749, Name: "Вантажні автомобілі з Польщі", Slug: "gruzovye-avtomobili-iz-polshi", ParentID: 1748,
					},
					{
						ID: 1750, Name: "Причепи з Польщі", Slug: "pritsepy-iz-polshi", ParentID: 1748,
					},
					{
						ID: 1751, Name: "Сільгосптехніка з Польщі", Slug: "selhoztehnika-iz-polshi", ParentID: 1748,
					},
					{
						ID: 1752, Name: "Будівельна техніка з Польщі", Slug: "stroitelnaya-tehnika-iz-polshi", ParentID: 1748,
					},
				},
			},
		},
	},
	{
		ID: 3, Name: "Запчастини", Slug: "zapchasti-dlya-transporta", Icon: "🛠", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 270, Name: "Шини, диски і колеса", Slug: "shiny-diski-i-kolesa", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 1463, Name: "Диски", Slug: "diski", ParentID: 270,
						Children: []CategoryItem{
							{
								ID: 3917, Name: "Для автомобілів", Slug: "dlya-avtomobiley", ParentID: 1463,
							},
							{
								ID: 3920, Name: "Для інших транспортних засобів", Slug: "dlya-drugikh-transportnykh-sredstv", ParentID: 1463,
							},
							{
								ID: 3923, Name: "Для вантажівок та фургонів", Slug: "dlya-gruzovikov-i-furgonov", ParentID: 1463,
							},
							{
								ID: 3925, Name: "Для мотоциклів та скутерів", Slug: "dlya-mototsiklov-i-skuterov", ParentID: 1463,
							},
						},
					},
					{
						ID: 1465, Name: "Колеса в зборі", Slug: "kolesa-v-sbore", ParentID: 270,
						Children: []CategoryItem{
							{
								ID: 3927, Name: "Запасні колеса", Slug: "zapasnye-kolesa", ParentID: 1465,
							},
							{
								ID: 3935, Name: "Для автомобілів", Slug: "dlya-avtomobiley", ParentID: 1465,
							},
							{
								ID: 3936, Name: "Для інших транспортних засобів", Slug: "dlya-drugikh-transportnykh-sredstv", ParentID: 1465,
							},
							{
								ID: 3937, Name: "Для вантажівок та фургонів", Slug: "dlya-gruzovikov-i-furgonov", ParentID: 1465,
							},
							{
								ID: 3938, Name: "Для мотоциклів та скутерів", Slug: "dlya-mototsiklov-i-skuterov", ParentID: 1465,
							},
						},
					},
					{
						ID: 3909, Name: "Ковпаки та аксесуари", Slug: "kolpaki-i-aksessuary", ParentID: 270,
					},
					{
						ID: 3912, Name: "Шини", Slug: "shiny", ParentID: 270,
						Children: []CategoryItem{
							{
								ID: 3939, Name: "Для інших транспортних засобів", Slug: "dlya-drugikh-transportnykh-sredstv", ParentID: 3912,
							},
							{
								ID: 3940, Name: "Для вантажівок та фургонів", Slug: "dlya-gruzovikov-i-furgonov", ParentID: 3912,
							},
							{
								ID: 3961, Name: "Автошини", Slug: "avtoshiny", ParentID: 3912,
							},
							{
								ID: 3962, Name: "Мотошини", Slug: "motoshiny", ParentID: 3912,
							},
						},
					},
				},
			},
			{
				ID: 316, Name: "Запчастини для іншої техніки", Slug: "prochie-zapchasti", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 2581, Name: "Для спецтехніки", Slug: "dlya-spetstekhniki", ParentID: 316,
					},
					{
						ID: 2582, Name: "Для водного транспорту", Slug: "dlya-vodnogo-transporta", ParentID: 316,
					},
					{
						ID: 2583, Name: "Для інших видів техніки", Slug: "dlya-drugikh-vidov-tekhniki", ParentID: 316,
					},
				},
			},
			{
				ID: 2017, Name: "Мотозапчастини та мотоаксесуари", Slug: "motozapchasti", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 2021, Name: "Амортизатори та підвіска", Slug: "amortizatory-i-podveska", ParentID: 2017,
					},
					{
						ID: 2023, Name: "Вихлопні системи, прямоструми, глушники", Slug: "vykhlopnyye-sistemy-pryamotoki-glushiteli", ParentID: 2017,
					},
					{
						ID: 2027, Name: "Освітлення та лампи", Slug: "osveshcheniye-i-lampy", ParentID: 2017,
					},
					{
						ID: 2029, Name: "Привід та ходова", Slug: "privod-i-khodovaya", ParentID: 2017,
					},
					{
						ID: 2031, Name: "Кермо та управління", Slug: "rul-i-upravleniye", ParentID: 2017,
					},
					{
						ID: 2035, Name: "Паливна система", Slug: "toplivnaya-sistema", ParentID: 2017,
					},
					{
						ID: 2036, Name: "Гальмівна система", Slug: "tormoznaya-sistema", ParentID: 2017,
					},
					{
						ID: 2038, Name: "Фільтри", Slug: "filtry", ParentID: 2017,
					},
					{
						ID: 2040, Name: "Електрика", Slug: "elektrika", ParentID: 2017,
					},
					{
						ID: 3911, Name: "Двигуни та комплектуючі", Slug: "dvigateli-i-komplektuyushchie", ParentID: 2017,
					},
					{
						ID: 3913, Name: "Кузовні елементи та пластик", Slug: "kuzovnyye-elementy-i-plastik", ParentID: 2017,
					},
					{
						ID: 3966, Name: "Мотоекіпірування", Slug: "motoekipirovka", ParentID: 2017,
						Children: []CategoryItem{
							{
								ID: 3994, Name: "Інше мотоекіпірування", Slug: "drugaya-motoekipirovka", ParentID: 3966,
							},
							{
								ID: 3995, Name: "Мотокомбінезони", Slug: "motokombinezony", ParentID: 3966,
							},
							{
								ID: 3996, Name: "Мотокуртки", Slug: "motokurtki", ParentID: 3966,
							},
							{
								ID: 3997, Name: "Мотовзуття", Slug: "motoobuv", ParentID: 3966,
							},
							{
								ID: 3998, Name: "Моторукавички", Slug: "motoperchatki", ParentID: 3966,
							},
							{
								ID: 3999, Name: "Мотошоломи", Slug: "motoshlemy", ParentID: 3966,
							},
							{
								ID: 4000, Name: "Мотоштани", Slug: "motoshtany", ParentID: 3966,
							},
							{
								ID: 4001, Name: "Мотозахист", Slug: "motozashchita", ParentID: 3966,
							},
							{
								ID: 4002, Name: "Підшоломники та балаклави", Slug: "podshlemniki-i-balaklavy", ParentID: 3966,
							},
						},
					},
					{
						ID: 4003, Name: "Інші мотозапчастини", Slug: "drugie-motozapchasti", ParentID: 2017,
					},
					{
						ID: 4014, Name: "Мотоаксесуари", Slug: "motoaksessuary", ParentID: 2017,
						Children: []CategoryItem{
							{
								ID: 4015, Name: "Чохли та накидки для мотоцикла", Slug: "chekhly-i-nakidki", ParentID: 4014,
							},
							{
								ID: 4016, Name: "Тримач для телефона, планшета", Slug: "derzhatel-telefona-plansheta", ParentID: 4014,
							},
							{
								ID: 4018, Name: "Інші мотоаксесуари", Slug: "drugie-motoaksessuary", ParentID: 4014,
							},
							{
								ID: 4020, Name: "Інструменти", Slug: "instrumenty", ParentID: 4014,
							},
							{
								ID: 4021, Name: "Мотосумки, кофри, рюкзаки", Slug: "motosumki-kofry-ryukzaki", ParentID: 4014,
							},
							{
								ID: 4022, Name: "Дзеркала на мотоцикл", Slug: "motozerkala", ParentID: 4014,
							},
							{
								ID: 4023, Name: "Наклейки на мотоцикл, скутер", Slug: "nakleyki-na-mototsikl-skuter", ParentID: 4014,
							},
							{
								ID: 4024, Name: "Обтічники та вітрове скло", Slug: "obtekateli-i-vetrovyye-stekla", ParentID: 4014,
							},
							{
								ID: 4025, Name: "Панельні прилади", Slug: "panelnyye-pribory", ParentID: 4014,
							},
							{
								ID: 4026, Name: "Переговорні пристрої", Slug: "peregovornyye-ustroystva", ParentID: 4014,
							},
							{
								ID: 4027, Name: "Підніжки", Slug: "podnozhki", ParentID: 4014,
							},
							{
								ID: 4028, Name: "Стайлінг", Slug: "stayling", ParentID: 4014,
							},
						},
					},
				},
			},
			{
				ID: 2154, Name: "Аксесуари для авто", Slug: "aksessuary-dlya-avto", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 2158, Name: "Автомобільні килимки", Slug: "avtokovriki", ParentID: 2154,
					},
					{
						ID: 2165, Name: "Багажники", Slug: "bagazhniki", ParentID: 2154,
					},
					{
						ID: 2170, Name: "Підставки та тримачі", Slug: "derzhateli-i-podstavki", ParentID: 2154,
					},
					{
						ID: 2171, Name: "Дефлектори", Slug: "deflektory", ParentID: 2154,
					},
					{
						ID: 2176, Name: "Пускові та зарядні пристрої", Slug: "puskovyye-i-zaryadnyye-ustroystva", ParentID: 2154,
					},
					{
						ID: 2179, Name: "Органайзери в багажник", Slug: "organayzery-avtomobilnyye", ParentID: 2154,
					},
					{
						ID: 2184, Name: "Вогнегасник, аптечка, аварійка", Slug: "ognetushitel-aptechka-avariyka", ParentID: 2154,
					},
					{
						ID: 2186, Name: "Фаркопи", Slug: "farkopy", ParentID: 2154,
					},
					{
						ID: 2187, Name: "Ланцюги, браслети протиковзання", Slug: "tsepi-braslety-protivoskolzheniya", ParentID: 2154,
					},
					{
						ID: 2188, Name: "Чохли, тенти та накидки", Slug: "chekhly-tenty-nakidki", ParentID: 2154,
					},
					{
						ID: 2290, Name: "Каністри та аксесуари", Slug: "kanistry", ParentID: 2154,
					},
					{
						ID: 3941, Name: "Автоелектроніка", Slug: "avtoelektronika", ParentID: 2154,
						Children: []CategoryItem{
							{
								ID: 3946, Name: "Інша автоелектроніка", Slug: "drugaya-avtoelektronika", ParentID: 3941,
							},
							{
								ID: 3967, Name: "Автозвук та мультимедіа", Slug: "avtozvuk-i-multimedia", ParentID: 3941,
							},
							{
								ID: 3968, Name: "GPS-навігатори та аксесуари", Slug: "gps-navigatory", ParentID: 3941,
							},
							{
								ID: 3972, Name: "Відеореєстратори", Slug: "videoregistratory", ParentID: 3941,
							},
						},
					},
					{
						ID: 3942, Name: "Підйомники та транспортери", Slug: "podyomniki-i-transportery", ParentID: 2154,
						Children: []CategoryItem{
							{
								ID: 3947, Name: "Інше для підйому авто", Slug: "drugoye-dlya-podyema-avto", ParentID: 3942,
							},
							{
								ID: 3969, Name: "Домкрати", Slug: "domkraty", ParentID: 3942,
							},
						},
					},
					{
						ID: 3943, Name: "Аксесуари для догляду за авто", Slug: "aksessuary-dlya-uhoda-za-avto", ParentID: 2154,
					},
					{
						ID: 3963, Name: "Емблеми та значки", Slug: "emblemy-i-znachki", ParentID: 2154,
					},
					{
						ID: 3964, Name: "Візуальний тюнінг", Slug: "vizualnyy-tyuning", ParentID: 2154,
					},
					{
						ID: 4004, Name: "Інші автоаксесуари", Slug: "drugie-avtoaksessuary", ParentID: 2154,
					},
				},
			},
			{
				ID: 2157, Name: "Транспорт на запчастини / авторозбірки", Slug: "transport-na-zapchasti-avtorazborki", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 3945, Name: "Легкові авто на запчастини", Slug: "legkovye-avto", ParentID: 2157,
					},
					{
						ID: 3958, Name: "Мотоцикли на запчастини", Slug: "motocikly", ParentID: 2157,
					},
					{
						ID: 3959, Name: "Вантажівки на запчастини", Slug: "gruzoviki", ParentID: 2157,
					},
					{
						ID: 3960, Name: "Сільськогосподарська, спецтехніка та інші транспортні засоби на запчастини", Slug: "specztehnika", ParentID: 2157,
					},
				},
			},
			{
				ID: 2357, Name: "Автозапчастини", Slug: "avtozapchasti", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 2358, Name: "Двигун та комплектуючі", Slug: "dvigatel", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2442, Name: "Блок циліндрів та його частини", Slug: "blok-tsilindrov-i-ego-chasti", ParentID: 2358,
							},
							{
								ID: 2443, Name: "Газорозподільний механізм (ГРМ)", Slug: "gazoraspredelitelnyy-mekhanizm-grm", ParentID: 2358,
							},
							{
								ID: 2444, Name: "Двигуни у зборі", Slug: "dvigateli-v-sbore", ParentID: 2358,
							},
							{
								ID: 2445, Name: "Кожухи та кришки двигуна", Slug: "kozhukhi-i-kryshki-dvigatelya", ParentID: 2358,
							},
							{
								ID: 2446, Name: "Кривошипно-шатунний механізм", Slug: "krivoshipno-shatunnyy-mekhanizm", ParentID: 2358,
							},
							{
								ID: 2448, Name: "Подушки двигуна", Slug: "podushki-dvigatelya", ParentID: 2358,
							},
							{
								ID: 2450, Name: "Поршнева група", Slug: "porshnevaya-gruppa", ParentID: 2358,
							},
							{
								ID: 2451, Name: "Прокладки та комплекти прокладок", Slug: "prokladki-i-komplekty-prokladok", ParentID: 2358,
							},
							{
								ID: 2455, Name: "Система змащування двигуна", Slug: "sistema-smazki-dvigatelya", ParentID: 2358,
							},
							{
								ID: 2456, Name: "Турбіни, турбокомпресори", Slug: "turbiny-turbokompressory", ParentID: 2358,
							},
							{
								ID: 2458, Name: "Інші запчастини двигуна", Slug: "drugiye-zapchasti-dvigatelya", ParentID: 2358,
							},
							{
								ID: 3955, Name: "Повітряні патрубки", Slug: "vozdushnyye-patrubki", ParentID: 2358,
							},
							{
								ID: 3973, Name: "Датчики температури оливи", Slug: "datchiki-temperatury-masla", ParentID: 2358,
							},
							{
								ID: 3974, Name: "Дросельні заслінки", Slug: "drosselnyye-zaslonki", ParentID: 2358,
							},
							{
								ID: 3975, Name: "Корпуси повітряного фільтра", Slug: "korpusa-vozdushnogo-filtra", ParentID: 2358,
							},
							{
								ID: 3976, Name: "Впускний та випускний колектори", Slug: "vpusknoy-i-vypusknoy-kollektory", ParentID: 2358,
							},
						},
					},
					{
						ID: 2359, Name: "Вихлопна система", Slug: "vykhlopnaya-sistema", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2382, Name: "Вихлопні колектори", Slug: "vypusknyye-kollektory", ParentID: 2359,
							},
							{
								ID: 2383, Name: "Вихлопні системи в зборі", Slug: "vykhlopnyye-sistemy-v-sbore", ParentID: 2359,
							},
							{
								ID: 2384, Name: "Глушники", Slug: "glushiteli", ParentID: 2359,
							},
							{
								ID: 2385, Name: "Гофри", Slug: "gofry", ParentID: 2359,
							},
							{
								ID: 2386, Name: "Датчики кисневі (лямбда-зонд)", Slug: "datchiki-kislorodnyye-lyambda-zond", ParentID: 2359,
							},
							{
								ID: 2388, Name: "Каталізатори", Slug: "katalizatory", ParentID: 2359,
							},
							{
								ID: 2389, Name: "Клапани EGR", Slug: "klapany-egr", ParentID: 2359,
							},
							{
								ID: 2393, Name: "Інші деталі вихлопної системи", Slug: "drugiye-zapchasti-vykhlopnoy-sistemy", ParentID: 2359,
							},
							{
								ID: 3993, Name: "Насадки на глушники", Slug: "nasadki-glushitelya", ParentID: 2359,
							},
						},
					},
					{
						ID: 2360, Name: "Трансмісія / КПП та АКПП", Slug: "transmissiya-kpp-i-akpp", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2459, Name: "Карданні вали", Slug: "valy-kardannyye", ParentID: 2360,
							},
							{
								ID: 2460, Name: "Приводні вали", Slug: "valy-privodnyye", ParentID: 2360,
							},
							{
								ID: 2465, Name: "Диференціали", Slug: "differentsialy-reduktory", ParentID: 2360,
							},
							{
								ID: 2466, Name: "Запчастини для коробок передач", Slug: "komponenty-kpp-akpp", ParentID: 2360,
							},
							{
								ID: 2468, Name: "КПП / АКПП", Slug: "kpp-akpp-v-sbore", ParentID: 2360,
							},
							{
								ID: 2472, Name: "Циліндри зчеплення", Slug: "tsilindry-stsepleniya", ParentID: 2360,
							},
							{
								ID: 2475, Name: "Інші запчастини трансмісії", Slug: "drugiye-zapchasti-transmissii", ParentID: 2360,
							},
							{
								ID: 3948, Name: "Зчеплення", Slug: "stseplenie", ParentID: 2360,
							},
							{
								ID: 3949, Name: "Півосі, шарніри, вали", Slug: "poluosi-sharniry-valy", ParentID: 2360,
							},
							{
								ID: 3990, Name: "Маховики", Slug: "makhoviki-i-ikh-sostavlyayushchiye", ParentID: 2360,
							},
						},
					},
					{
						ID: 2361, Name: "Кермове управління", Slug: "rulevoye-upravleniye", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2477, Name: "Амортизатори та стабілізатори рульового керування", Slug: "amortizatory-i-stabilizatory-rulevogo-upravleniya", ParentID: 2361,
							},
							{
								ID: 2479, Name: "Гідропідсилювач керма", Slug: "gidrousilitel-rulya", ParentID: 2361,
							},
							{
								ID: 2480, Name: "Наконечники кермові", Slug: "nakonechniki-rulevyye", ParentID: 2361,
							},
							{
								ID: 2482, Name: "Пильовики, відбійники", Slug: "pylniki-otboyniki", ParentID: 2361,
							},
							{
								ID: 2484, Name: "Кермові механізми", Slug: "rulevoy-mekhanizm", ParentID: 2361,
							},
							{
								ID: 2485, Name: "Кермові тяги", Slug: "rulevyye-tyagi", ParentID: 2361,
							},
							{
								ID: 2486, Name: "Інші деталі системи керування", Slug: "drugiye-zapchasti-rulevogo-upravleniya", ParentID: 2361,
							},
							{
								ID: 3950, Name: "Кермові колонки", Slug: "rulevye-kolonki", ParentID: 2361,
							},
						},
					},
					{
						ID: 2362, Name: "Ходова частина / підвіска", Slug: "khodovaya-chast-podveska", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2509, Name: "Амортизатори та елементи", Slug: "amortizatory", ParentID: 2362,
							},
							{
								ID: 2510, Name: "Балки підвіски", Slug: "balki-podveski", ParentID: 2362,
							},
							{
								ID: 2513, Name: "Пневморесори, пневмоподушки", Slug: "pnevmoressory-pnevmopodushki", ParentID: 2362,
							},
							{
								ID: 2514, Name: "Поворотні кулаки", Slug: "povorotnyye-kulaki", ParentID: 2362,
							},
							{
								ID: 2515, Name: "Проставки для збільшення кліренсу", Slug: "prostavki-dlya-uvelicheniya-klirensa", ParentID: 2362,
							},
							{
								ID: 2516, Name: "Пружини підвіски", Slug: "pruzhiny-podveski", ParentID: 2362,
							},
							{
								ID: 2517, Name: "Ресори та їхні елементи", Slug: "ressory-i-ikh-sostavlyayushchiye", ParentID: 2362,
							},
							{
								ID: 2518, Name: "Важелі, тяги", Slug: "rychagi-tyagi", ParentID: 2362,
							},
							{
								ID: 2520, Name: "Стійки стабілізаторів", Slug: "stoyki-stabilizatorov", ParentID: 2362,
							},
							{
								ID: 2521, Name: "Маточини, підшипники маточини", Slug: "stupitsa-podshipnik-stupitsy", ParentID: 2362,
							},
							{
								ID: 2523, Name: "Інші запчастини підвіски", Slug: "drugiye-zapchasti-podveski", ParentID: 2362,
							},
							{
								ID: 3952, Name: "Монтажні елементи підвіски", Slug: "montazhnyye-elementy-podveski", ParentID: 2362,
							},
						},
					},
					{
						ID: 2363, Name: "Паливна система", Slug: "toplivnaya-sistema", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2501, Name: "Паливні насоси", Slug: "toplivnyye-nasosy", ParentID: 2363,
							},
							{
								ID: 2502, Name: "Баки та бачки паливної системи", Slug: "baki-i-bachki-toplivnoy-sistemy", ParentID: 2363,
							},
							{
								ID: 2505, Name: "Карбюратори та комплектуючі", Slug: "karbyuratory-i-komplektuyushchiye", ParentID: 2363,
							},
							{
								ID: 2506, Name: "Розпилювачі та паливні форсунки", Slug: "raspyliteli-i-toplivnyye-forsunki", ParentID: 2363,
							},
							{
								ID: 2507, Name: "Регулятори тиску палива", Slug: "regulyatory-davleniya-topliva", ParentID: 2363,
							},
							{
								ID: 2508, Name: "Інші запчастини паливної системи", Slug: "drugiye-zapchasti-toplivnoy-sistemy", ParentID: 2363,
							},
							{
								ID: 3989, Name: "Датчики температури палива", Slug: "datchiki-temperatury-topliva", ParentID: 2363,
							},
							{
								ID: 3991, Name: "ГБО", Slug: "gbo", ParentID: 2363,
							},
						},
					},
					{
						ID: 2364, Name: "Гальмівна система", Slug: "tormoznaya-sistema", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2487, Name: "Барабани гальмівні", Slug: "barabany-tormoznyye", ParentID: 2364,
							},
							{
								ID: 2488, Name: "Деталі гальмівних систем ABS, EBD", Slug: "detali-tormoznykh-sistem-abs-ebd", ParentID: 2364,
							},
							{
								ID: 2493, Name: "Гальмівні диски", Slug: "tormoznyye-diski", ParentID: 2364,
							},
							{
								ID: 2497, Name: "Гальмівні циліндри", Slug: "tormoznyye-tsilindry", ParentID: 2364,
							},
							{
								ID: 2499, Name: "Підсилювачі гальм та їхні компоненти", Slug: "usiliteli-tormozov-i-ikh-komponenty", ParentID: 2364,
							},
							{
								ID: 2500, Name: "Інші запчастини гальмівної системи", Slug: "drugiye-zapchasti-tormoznoy-sistemy", ParentID: 2364,
							},
							{
								ID: 3953, Name: "Пневматичні гальма", Slug: "pnevmaticheskiye-tormoza", ParentID: 2364,
							},
							{
								ID: 3954, Name: "Гальмівні колодки", Slug: "tormoznyye-kolodki", ParentID: 2364,
							},
						},
					},
					{
						ID: 2365, Name: "Кузовні запчастини", Slug: "kuzovnyye-zapchasti", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2417, Name: "Багажники, задні панелі та їхні деталі", Slug: "bagazhniki-zadniye-paneli-i-ikh-detali", ParentID: 2365,
							},
							{
								ID: 2418, Name: "Бампери", Slug: "bampery", ParentID: 2365,
							},
							{
								ID: 2419, Name: "Підсилювачі, буфери бампера", Slug: "usiliteli-bampera-bufer-bampera", ParentID: 2365,
							},
							{
								ID: 2420, Name: "Двері", Slug: "dveri", ParentID: 2365,
							},
							{
								ID: 2421, Name: "Захист картера та КПП", Slug: "zashchita-kartera-i-kpp", ParentID: 2365,
							},
							{
								ID: 2422, Name: "Капоти та їхні частини", Slug: "kapoty-i-ikh-chasti", ParentID: 2365,
							},
							{
								ID: 2424, Name: "Крила авто та їхні частини", Slug: "krylia-i-ikh-chasti", ParentID: 2365,
							},
							{
								ID: 2425, Name: "Кузови автомобілів", Slug: "kuzova-legkovykh-avtomobiley", ParentID: 2365,
							},
							{
								ID: 2426, Name: "Дверні замки та ключі", Slug: "dvernyye-zamki-i-klyuchi", ParentID: 2365,
							},
							{
								ID: 2427, Name: "Автомобільні люки, дахи", Slug: "avtomobilnyye-lyuki", ParentID: 2365,
							},
							{
								ID: 2428, Name: "Лонжерони, підрамники", Slug: "lonzherony-podramniki", ParentID: 2365,
							},
							{
								ID: 2429, Name: "Лючки, кришки бензобака", Slug: "lyuchki-kryshki-benzobaka", ParentID: 2365,
							},
							{
								ID: 2430, Name: "Нижня частина кузова та рама автомобіля", Slug: "nizhnyaya-chast-kuzova-i-rama-avtomobilya", ParentID: 2365,
							},
							{
								ID: 2431, Name: "Інші кузовні запчастини", Slug: "drugiye-kuzovnyye-zapchasti", ParentID: 2365,
							},
							{
								ID: 2616, Name: "Решітки радіатора", Slug: "reshetka-radiatora", ParentID: 2365,
							},
							{
								ID: 2617, Name: "Передня панель, телевізори", Slug: "perednyaya-panel-televizor", ParentID: 2365,
							},
							{
								ID: 3981, Name: "Автоскло інших видів", Slug: "avtostekla-drugikh-vidov", ParentID: 2365,
							},
							{
								ID: 3982, Name: "Бризковики", Slug: "bryzgoviki", ParentID: 2365,
							},
							{
								ID: 3983, Name: "Скло дверей", Slug: "dvernyye-stekla", ParentID: 2365,
							},
							{
								ID: 3984, Name: "Бічне скло", Slug: "stekla-bokovyye", ParentID: 2365,
							},
							{
								ID: 3987, Name: "Скло вітрове (лобове)", Slug: "stekla-vetrovyye-lobovyye", ParentID: 2365,
							},
							{
								ID: 3988, Name: "Заднє скло", Slug: "zadniye-stekla", ParentID: 2365,
							},
						},
					},
					{
						ID: 2366, Name: "Деталі салону", Slug: "detali-salona", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2398, Name: "Автосидіння, крісла", Slug: "avtosideniya-kresla", ParentID: 2366,
							},
							{
								ID: 2399, Name: "Бардачки", Slug: "bardachki", ParentID: 2366,
							},
							{
								ID: 2400, Name: "Дверні карти, обшивка дверей", Slug: "dvernyye-karty-obshivka-dveri", ParentID: 2366,
							},
							{
								ID: 2401, Name: "Дзеркала", Slug: "zerkala", ParentID: 2366,
							},
							{
								ID: 2402, Name: "Задні полиці", Slug: "zadnyaya-polka", ParentID: 2366,
							},
							{
								ID: 2403, Name: "Кнопки / Перемикачі", Slug: "knopki-pereklyuchateli", ParentID: 2366,
							},
							{
								ID: 2404, Name: "Матеріал для підлоги", Slug: "material-dlya-pola", ParentID: 2366,
							},
							{
								ID: 2405, Name: "Попільнички", Slug: "pepelnitsa", ParentID: 2366,
							},
							{
								ID: 2406, Name: "Пластик салону", Slug: "plastik-salona", ParentID: 2366,
							},
							{
								ID: 2407, Name: "Підлокітники", Slug: "podlokotnik", ParentID: 2366,
							},
							{
								ID: 2409, Name: "Подушки безпеки, AIRBAG", Slug: "podushki-bezopasnosti-airbag", ParentID: 2366,
							},
							{
								ID: 2411, Name: "Ремені безпеки та натягувачі", Slug: "remni-bezopasnosti-i-natyazhitel", ParentID: 2366,
							},
							{
								ID: 2412, Name: "Кермо", Slug: "rul", ParentID: 2366,
							},
							{
								ID: 2413, Name: "Ручки дверей та салону", Slug: "ruchki-dverey-i-salona", ParentID: 2366,
							},
							{
								ID: 2414, Name: "Ручки та чохли на КПП", Slug: "ruchki-kpp", ParentID: 2366,
							},
							{
								ID: 2415, Name: "Торпедо, консолі", Slug: "torpedo-konsol", ParentID: 2366,
							},
							{
								ID: 2416, Name: "Інші деталі салону", Slug: "drugiye-detali-salona", ParentID: 2366,
							},
							{
								ID: 3951, Name: "Педалі", Slug: "pedali", ParentID: 2366,
							},
							{
								ID: 3970, Name: "Освітлення салону", Slug: "osveshcheniye-interyera", ParentID: 2366,
							},
							{
								ID: 3971, Name: "Панелі приладів", Slug: "paneli-priborov", ParentID: 2366,
							},
						},
					},
					{
						ID: 2367, Name: "Кліматична система", Slug: "konditsioner-otopitel-ventilyatsiya", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2432, Name: "Автомобільні кондиціонери у зборі", Slug: "avtomobilnyye-konditsionery-v-sbore", ParentID: 2367,
							},
							{
								ID: 2434, Name: "Блок управління кондиціонером", Slug: "bloki-upravleniya-konditsionerom", ParentID: 2367,
							},
							{
								ID: 2435, Name: "Вентилятори, системи кондиціювання та опалення", Slug: "ventilyatory-sistemy-konditsionirovaniya-i-otopleniya", ParentID: 2367,
							},
							{
								ID: 2436, Name: "Випарники", Slug: "ispariteli", ParentID: 2367,
							},
							{
								ID: 2437, Name: "Компресори", Slug: "kompressory", ParentID: 2367,
							},
							{
								ID: 2438, Name: "Конденсатори кондиціонера", Slug: "kondensatory-konditsionera", ParentID: 2367,
							},
							{
								ID: 2439, Name: "Радіатори обігрівачів", Slug: "radiatory-otopiteley", ParentID: 2367,
							},
							{
								ID: 2440, Name: "Трубки, шланги автокондиціонера", Slug: "trubki-shlangi-vozdukhovody", ParentID: 2367,
							},
							{
								ID: 2441, Name: "Інші запчастини кондиціонерів та опалення", Slug: "drugiye-zapchasti-konditsionerov-i-otopleniya", ParentID: 2367,
							},
						},
					},
					{
						ID: 2370, Name: "Система очищення вікон та фар", Slug: "sistema-ochistki-okon-i-far", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2537, Name: "Бачки та кришки омивачів", Slug: "bachki-i-kryshki-omyvateley", ParentID: 2370,
							},
							{
								ID: 2538, Name: "Моторчики двірників", Slug: "motorchiki-dvornikov", ParentID: 2370,
							},
							{
								ID: 2539, Name: "Насоси омивачів", Slug: "nasosy-omyvatelya", ParentID: 2370,
							},
							{
								ID: 2540, Name: "Гумки для склоочисників", Slug: "rezinki-dlya-stekloochistiteley", ParentID: 2370,
							},
							{
								ID: 2541, Name: "Форсунки омивачів", Slug: "forsunki-omyvatelya", ParentID: 2370,
							},
							{
								ID: 2542, Name: "Двірники", Slug: "dvorniki", ParentID: 2370,
							},
							{
								ID: 2543, Name: "Інші запчастини очищення вікон та фар", Slug: "drugiye-zapchasti-ochistki-okon-i-far", ParentID: 2370,
							},
						},
					},
					{
						ID: 2371, Name: "Фари та освітлення", Slug: "fary-i-osveshcheniye", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2544, Name: "Передні фари", Slug: "peredniye-fary", ParentID: 2371,
							},
							{
								ID: 2545, Name: "Задні фари", Slug: "zadniye-fary", ParentID: 2371,
							},
							{
								ID: 2546, Name: "Протитуманні фари", Slug: "protivotumannyye-fary", ParentID: 2371,
							},
							{
								ID: 2547, Name: "Денні ходові вогні", Slug: "dnevnyye-khodovyye-ogni", ParentID: 2371,
							},
							{
								ID: 2548, Name: "Ксенон", Slug: "ksenon", ParentID: 2371,
							},
							{
								ID: 2549, Name: "Автолампи", Slug: "lampochki", ParentID: 2371,
							},
							{
								ID: 2550, Name: "Відбивачі, корпуси фар", Slug: "otrazhateli-korpusa-far", ParentID: 2371,
							},
							{
								ID: 2551, Name: "Повторювачі поворотів", Slug: "povtoriteli-povorotov", ParentID: 2371,
							},
							{
								ID: 2553, Name: "Інші елементи освітлення", Slug: "drugiye-elementy-osveshcheniya", ParentID: 2371,
							},
						},
					},
					{
						ID: 2372, Name: "Електрообладнання та система запалювання", Slug: "elektrooborudovaniye", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2554, Name: "Акумулятори", Slug: "akkumulyatory", ParentID: 2372,
							},
							{
								ID: 2555, Name: "Контролери двигуна", Slug: "bloki-upravleniya", ParentID: 2372,
							},
							{
								ID: 2559, Name: "Запобіжники, переривники", Slug: "predokhraniteli-preryvateli", ParentID: 2372,
							},
							{
								ID: 2561, Name: "Реле", Slug: "rele", ParentID: 2372,
							},
							{
								ID: 2562, Name: "Звукові сигнали", Slug: "signaly-zvukovyye", ParentID: 2372,
							},
							{
								ID: 2563, Name: "Електропроводка та компоненти", Slug: "elektroprovodka-i-komponenty", ParentID: 2372,
							},
							{
								ID: 2564, Name: "Інше електрообладнання для авто", Slug: "drugoye-elektroborudovaniye", ParentID: 2372,
							},
							{
								ID: 3956, Name: "Електрообладнання двигуна", Slug: "elektrooborudovaniye-dvigatelya", ParentID: 2372,
							},
							{
								ID: 3957, Name: "Датчики для систем запалювання", Slug: "datchiki-dlya-sistem-zazhiganiya", ParentID: 2372,
							},
							{
								ID: 3977, Name: "Котушки запалювання", Slug: "katushki-zazhiganiya", ParentID: 2372,
							},
							{
								ID: 3978, Name: "Стартер, запчастини стартера", Slug: "starter-zapchasti-startera", ParentID: 2372,
							},
							{
								ID: 3979, Name: "Свічки запалювання/розжарювання", Slug: "svechi-zazhiganiya-nakalivaniya", ParentID: 2372,
							},
							{
								ID: 3980, Name: "Замки запалювання та ключ", Slug: "zamki-zazhiganiya-i-klyuchi", ParentID: 2372,
							},
						},
					},
					{
						ID: 2373, Name: "Фільтри", Slug: "filtry", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 2565, Name: "Повітряні фільтри", Slug: "vozdushnyye-filtry", ParentID: 2373,
							},
							{
								ID: 2566, Name: "Мастильні фільтри", Slug: "maslyanyye-filtry", ParentID: 2373,
							},
							{
								ID: 2567, Name: "Паливні фільтри", Slug: "toplivnyye-filtry", ParentID: 2373,
							},
							{
								ID: 2568, Name: "Фільтри салону", Slug: "filtry-salona", ParentID: 2373,
							},
							{
								ID: 2570, Name: "Інші фільтри для авто", Slug: "filtry-drugikh-tipov", ParentID: 2373,
							},
						},
					},
					{
						ID: 2376, Name: "Інші автозапчастин", Slug: "drugiye-vidy-zapchastey", ParentID: 2357,
					},
					{
						ID: 3944, Name: "Система охолодження двигуна", Slug: "sistema-okhlazhdeniya-dvigatelya", ParentID: 2357,
						Children: []CategoryItem{
							{
								ID: 3985, Name: "Інші деталі охолодження двигуна", Slug: "drugiye-zapchasti-dvigatelya", ParentID: 3944,
							},
							{
								ID: 3986, Name: "Інтеркулери", Slug: "interkulery", ParentID: 3944,
							},
							{
								ID: 3992, Name: "Радіатори автомобільні", Slug: "radiatory-avtomobilnyye", ParentID: 3944,
							},
						},
					},
					{
						ID: 3965, Name: "Автономні обігрівачі та кондиціонери", Slug: "avtonomnye-obogrevateli-i-konditsionery", ParentID: 2357,
					},
				},
			},
			{
				ID: 2571, Name: "Мастила та автохімія", Slug: "masla-i-avtokhimiya", ParentID: 3,
				Children: []CategoryItem{
					{
						ID: 2572, Name: "Моторні оливи", Slug: "motornyye-masla", ParentID: 2571,
					},
					{
						ID: 2576, Name: "Автохімія та автокосметика", Slug: "avtokosmetika-i-avtokhimiya", ParentID: 2571,
					},
					{
						ID: 3908, Name: "Інша автохімія та рідини", Slug: "drugie-masla-i-zhidkosti", ParentID: 2571,
					},
					{
						ID: 3910, Name: "Технічні рідини", Slug: "tekhnicheskiye-zhidkosti", ParentID: 2571,
						Children: []CategoryItem{
							{
								ID: 3915, Name: "Інші технічні рідини", Slug: "drugiye-vidy", ParentID: 3910,
							},
							{
								ID: 4017, Name: "Гідравлічні оливи", Slug: "gidravlicheskoye-maslo", ParentID: 3910,
							},
							{
								ID: 4019, Name: "Трансмісійні оливи", Slug: "transmissionnoye-maslo", ParentID: 3910,
							},
						},
					},
				},
			},
			{
				ID: 4012, Name: "Запчастини для вантажівок", Slug: "dlya-gruzovikov", ParentID: 3,
			},
			{
				ID: 4013, Name: "Запчастини для сільгосптехніки", Slug: "dlya-selkhoztekhniki", ParentID: 3,
			},
		},
	},
	{
		ID: 6, Name: "Робота", Slug: "rabota", Icon: "💼", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 134, Name: "Медицина / фармацевтика", Slug: "meditsina-farmatsiya", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1818, Name: "Медсестра", Slug: "medsestra", ParentID: 134,
					},
					{
						ID: 1820, Name: "Стоматолог", Slug: "stomatolog", ParentID: 134,
					},
					{
						ID: 1822, Name: "Фармацевт", Slug: "farmatsevt", ParentID: 134,
					},
					{
						ID: 1824, Name: "Інше", Slug: "drugoe", ParentID: 134,
					},
				},
			},
			{
				ID: 136, Name: "Часткова зайнятість", Slug: "chastichnaya-zanyatost", ParentID: 6,
			},
			{
				ID: 137, Name: "Краса / фітнес / спорт", Slug: "krasota-fitnes-sport", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1797, Name: "Перукар", Slug: "parikmaher", ParentID: 137,
					},
					{
						ID: 1800, Name: "Майстер манікюру", Slug: "master-manikyura", ParentID: 137,
					},
					{
						ID: 1802, Name: "Адміністратор", Slug: "administrator", ParentID: 137,
					},
					{
						ID: 1803, Name: "Масажист", Slug: "massazhist", ParentID: 137,
					},
					{
						ID: 1804, Name: "Косметолог", Slug: "kosmetolog", ParentID: 137,
					},
					{
						ID: 1805, Name: "Тренер", Slug: "trener", ParentID: 137,
					},
					{
						ID: 1807, Name: "Інше", Slug: "drugoe", ParentID: 137,
					},
				},
			},
			{
				ID: 140, Name: "Логістика / Склад / Доставка", Slug: "transport-logistika", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1911, Name: "Водій", Slug: "voditel", ParentID: 140,
					},
					{
						ID: 1912, Name: "Вантажник", Slug: "gruzchik", ParentID: 140,
					},
					{
						ID: 1913, Name: "Кур'єр", Slug: "kurer", ParentID: 140,
					},
					{
						ID: 1914, Name: "Комірник", Slug: "kladovshchik", ParentID: 140,
					},
					{
						ID: 1915, Name: "Диспетчер", Slug: "dispetcher", ParentID: 140,
					},
					{
						ID: 1916, Name: "Експедитор", Slug: "ekspeditor", ParentID: 140,
					},
					{
						ID: 1917, Name: "Далекобійник", Slug: "dalnoboyshchik", ParentID: 140,
					},
					{
						ID: 1918, Name: "Таксист", Slug: "taksist", ParentID: 140,
					},
					{
						ID: 1919, Name: "Інше", Slug: "drugoe", ParentID: 140,
					},
				},
			},
			{
				ID: 141, Name: "Освіта / переклад", Slug: "obrazovanie", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1814, Name: "Викладач / Вчитель", Slug: "prepodavatel-uchitel", ParentID: 141,
					},
					{
						ID: 1815, Name: "Вихователь", Slug: "vospitatel", ParentID: 141,
					},
					{
						ID: 1816, Name: "Репетитор", Slug: "repetitor", ParentID: 141,
					},
					{
						ID: 1817, Name: "Інше", Slug: "drugoe", ParentID: 141,
					},
				},
			},
			{
				ID: 143, Name: "Виробництво / робітничі спеціальності", Slug: "proizvodstvo-energetika", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1837, Name: "Електрик", Slug: "elektrik", ParentID: 143,
					},
					{
						ID: 1838, Name: "Столяр", Slug: "stolyar", ParentID: 143,
					},
					{
						ID: 1839, Name: "Швачка, кравець", Slug: "shveya-portnoy", ParentID: 143,
					},
					{
						ID: 1840, Name: "Токар", Slug: "tokar", ParentID: 143,
					},
					{
						ID: 1841, Name: "Машиніст екскаватора", Slug: "mashinist-ekskavatora", ParentID: 143,
					},
					{
						ID: 1842, Name: "Конструктор", Slug: "konstruktor", ParentID: 143,
					},
					{
						ID: 1843, Name: "Комплектувальник", Slug: "komplektovshchik", ParentID: 143,
					},
					{
						ID: 1844, Name: "Складальник меблів", Slug: "sborshchik-mebeli", ParentID: 143,
					},
					{
						ID: 1845, Name: "Різноробочий", Slug: "raznorabochiy", ParentID: 143,
					},
					{
						ID: 1846, Name: "Зварювальник", Slug: "svarshchik", ParentID: 143,
					},
					{
						ID: 1847, Name: "Складальник", Slug: "sborshchik", ParentID: 143,
					},
					{
						ID: 1848, Name: "Слюсар", Slug: "slesar", ParentID: 143,
					},
					{
						ID: 1849, Name: "Маляр", Slug: "malyar", ParentID: 143,
					},
					{
						ID: 1850, Name: "Пакувальник", Slug: "upakovshchik", ParentID: 143,
					},
					{
						ID: 1851, Name: "Монтажник", Slug: "montazhnik", ParentID: 143,
					},
					{
						ID: 1852, Name: "Механік", Slug: "mekhanik", ParentID: 143,
					},
					{
						ID: 1853, Name: "Підсобник", Slug: "podsobnik", ParentID: 143,
					},
					{
						ID: 1854, Name: "Оператор ЧПК", Slug: "operator-chpu", ParentID: 143,
					},
					{
						ID: 1855, Name: "Технолог", Slug: "tehnolog", ParentID: 143,
					},
					{
						ID: 1856, Name: "Інше", Slug: "drugoe", ParentID: 143,
					},
				},
			},
			{
				ID: 144, Name: "Роздрібна торгівля / продажі / закупки", Slug: "roznichnaya-torgovlya-prodazhi-zakupki", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1920, Name: "Мерчендайзер", Slug: "merchendayzer", ParentID: 144,
					},
					{
						ID: 1921, Name: "Менеджер з продажу", Slug: "menedzher-po-prodazham", ParentID: 144,
					},
					{
						ID: 1922, Name: "Продавець", Slug: "prodavets", ParentID: 144,
					},
					{
						ID: 1923, Name: "Супервайзер", Slug: "supervayzer", ParentID: 144,
					},
					{
						ID: 1924, Name: "Флорист", Slug: "florist", ParentID: 144,
					},
					{
						ID: 1925, Name: "Продавець-консультант", Slug: "prodavets-konsultant", ParentID: 144,
					},
					{
						ID: 1926, Name: "Торговий представник", Slug: "torgovyy-predstavitel", ParentID: 144,
					},
					{
						ID: 1927, Name: "Менеджер інтернет магазину", Slug: "menedzher-internet-magazina", ParentID: 144,
					},
					{
						ID: 1928, Name: "Касир", Slug: "kassir", ParentID: 144,
					},
					{
						ID: 1929, Name: "Менеджер по роботі з клієнтами", Slug: "menedzher-po-rabote-s-klientami", ParentID: 144,
					},
					{
						ID: 1930, Name: "Товарознавець", Slug: "tovaroved", ParentID: 144,
					},
					{
						ID: 1931, Name: "Керівник відділу продажів", Slug: "rukovoditel-otdela-prodazh", ParentID: 144,
					},
					{
						ID: 1932, Name: "Менеджер по закупці", Slug: "menedzher-po-zakupkam", ParentID: 144,
					},
					{
						ID: 1933, Name: "Адміністратор", Slug: "administrator", ParentID: 144,
					},
					{
						ID: 1934, Name: "Пекар", Slug: "pekar", ParentID: 144,
					},
					{
						ID: 1935, Name: "Інше", Slug: "drugoe", ParentID: 144,
					},
				},
			},
			{
				ID: 145, Name: "Реклама / дизайн / PR", Slug: "marketing-reklama-dizayn", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1830, Name: "Промоутер", Slug: "promouter", ParentID: 145,
					},
					{
						ID: 1831, Name: "Дизайнер", Slug: "dizayner", ParentID: 145,
					},
					{
						ID: 1832, Name: "Маркетолог", Slug: "marketolog", ParentID: 145,
					},
					{
						ID: 1833, Name: "SMM-менеджер", Slug: "smm-menedzher", ParentID: 145,
					},
					{
						ID: 1834, Name: "Контент-менеджер", Slug: "kontent-menedzher", ParentID: 145,
					},
					{
						ID: 1835, Name: "Копірайтер", Slug: "kopirayter", ParentID: 145,
					},
					{
						ID: 1836, Name: "Інше", Slug: "drugoe", ParentID: 145,
					},
				},
			},
			{
				ID: 147, Name: "IT / комп'ютери", Slug: "it-telekom-kompyutery", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1819, Name: "Оператор ПК", Slug: "operator-pk", ParentID: 147,
					},
					{
						ID: 1821, Name: "Майстер з ремонту", Slug: "master-po-remontu", ParentID: 147,
					},
					{
						ID: 1823, Name: "Системний адміністратор", Slug: "sistemnyy-administrator", ParentID: 147,
					},
					{
						ID: 1826, Name: "Інше", Slug: "drugoe", ParentID: 147,
					},
				},
			},
			{
				ID: 150, Name: "Будівництво / облицювальні роботи", Slug: "stroitelstvo", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1778, Name: "Різноробочий", Slug: "raznorabochiy", ParentID: 150,
					},
					{
						ID: 1779, Name: "Муляр", Slug: "kamenshchik", ParentID: 150,
					},
					{
						ID: 1780, Name: "Плиточник", Slug: "plitochnik", ParentID: 150,
					},
					{
						ID: 1781, Name: "Монтажник", Slug: "montazhnik", ParentID: 150,
					},
					{
						ID: 1782, Name: "Зварювальник", Slug: "svarshchik", ParentID: 150,
					},
					{
						ID: 1783, Name: "Будівельник", Slug: "stroitel", ParentID: 150,
					},
					{
						ID: 1784, Name: "Маляр", Slug: "malyar", ParentID: 150,
					},
					{
						ID: 1785, Name: "Виконроб", Slug: "prorab", ParentID: 150,
					},
					{
						ID: 1786, Name: "Сантехнік", Slug: "santekhnik", ParentID: 150,
					},
					{
						ID: 1787, Name: "Фасадник", Slug: "fasadchik", ParentID: 150,
					},
					{
						ID: 1788, Name: "Штукатур", Slug: "shtukatur", ParentID: 150,
					},
					{
						ID: 1789, Name: "Електрик", Slug: "elektrik", ParentID: 150,
					},
					{
						ID: 1790, Name: "Інше", Slug: "drugoe", ParentID: 150,
					},
				},
			},
			{
				ID: 151, Name: "Інші сфери занять", Slug: "drugie-sfery-zanyatiy", ParentID: 6,
			},
			{
				ID: 152, Name: "Охорона / безпека", Slug: "ohrana-bezopasnost", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1796, Name: "Охоронець", Slug: "ohrannik", ParentID: 152,
					},
					{
						ID: 1799, Name: "Сторож", Slug: "storozh", ParentID: 152,
					},
					{
						ID: 1801, Name: "Інше", Slug: "drygoe", ParentID: 152,
					},
				},
			},
			{
				ID: 154, Name: "Культура / мистецтво / розваги", Slug: "kultura-iskusstvo", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1889, Name: "Фотограф", Slug: "fotograf", ParentID: 154,
					},
					{
						ID: 1890, Name: "Модель", Slug: "model", ParentID: 154,
					},
					{
						ID: 1891, Name: "Аніматор", Slug: "animator", ParentID: 154,
					},
					{
						ID: 1892, Name: "Інше", Slug: "drugoe", ParentID: 154,
					},
				},
			},
			{
				ID: 183, Name: "Клінінг / Домашній персонал", Slug: "domashniy-personal", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1808, Name: "Доглядальниця", Slug: "sidelka", ParentID: 183,
					},
					{
						ID: 1809, Name: "Прибиральниця", Slug: "uborshchitsa", ParentID: 183,
					},
					{
						ID: 1810, Name: "Двірник", Slug: "dvornik", ParentID: 183,
					},
					{
						ID: 1811, Name: "Няня", Slug: "nyanya", ParentID: 183,
					},
					{
						ID: 1812, Name: "Домробітниця", Slug: "domrabotnitsa", ParentID: 183,
					},
					{
						ID: 1813, Name: "Інше", Slug: "drugoe", ParentID: 183,
					},
				},
			},
			{
				ID: 1159, Name: "Нерухомість", Slug: "nedvizhimost", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1791, Name: "Ріелтор", Slug: "rieltor", ParentID: 1159,
					},
					{
						ID: 1792, Name: "Менеджер з продажу", Slug: "menedzher-po-prodazham", ParentID: 1159,
					},
					{
						ID: 1793, Name: "Брокер", Slug: "broker", ParentID: 1159,
					},
					{
						ID: 1794, Name: "Інше", Slug: "drugoe", ParentID: 1159,
					},
				},
			},
			{
				ID: 1165, Name: "Початок кар'єри / Студенти", Slug: "nachalo-karery-studenty", ParentID: 6,
			},
			{
				ID: 1473, Name: "Адміністративний персонал / HR / Секретаріат", Slug: "upravleniye-personalom-hr", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1857, Name: "Адміністратор", Slug: "administrator", ParentID: 1473,
					},
					{
						ID: 1858, Name: "Секретар", Slug: "sekretar", ParentID: 1473,
					},
					{
						ID: 1859, Name: "Офіс менеджер", Slug: "ofis-menedzher", ParentID: 1473,
					},
					{
						ID: 1860, Name: "HR", Slug: "hr", ParentID: 1473,
					},
					{
						ID: 1861, Name: "Інше", Slug: "drugoe", ParentID: 1473,
					},
				},
			},
			{
				ID: 1475, Name: "Банки / фінанси / страхування / юриспруденція", Slug: "banki-finansy-strakhovaniye", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1795, Name: "Юрист", Slug: "yurist", ParentID: 1475,
					},
					{
						ID: 1798, Name: "Інше", Slug: "drugoe", ParentID: 1475,
					},
				},
			},
			{
				ID: 1477, Name: "Колл-центри / Телекомунікації", Slug: "telekommunikatsii-svyaz", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1827, Name: "Оператор колл-центру", Slug: "operator-koll-tsentra", ParentID: 1477,
					},
					{
						ID: 1828, Name: "Оператор чату", Slug: "operator-chata", ParentID: 1477,
					},
					{
						ID: 1829, Name: "Інше", Slug: "drugoe", ParentID: 1477,
					},
				},
			},
			{
				ID: 1479, Name: "Сільське і лісове господарство / агробізнес", Slug: "selskoye-khozyaystvo-agrobiznes-lesnoye-khozyaystvo", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1862, Name: "Агроном", Slug: "agronom", ParentID: 1479,
					},
					{
						ID: 1863, Name: "Тракторист", Slug: "traktorist", ParentID: 1479,
					},
					{
						ID: 1864, Name: "Пилорамщик", Slug: "piloramshchik", ParentID: 1479,
					},
					{
						ID: 1865, Name: "Інше", Slug: "drugoe", ParentID: 1479,
					},
				},
			},
			{
				ID: 1481, Name: "Робота за кордоном", Slug: "rabota-za-rubezhom", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1893, Name: "Різноробочий", Slug: "raznorabochiy", ParentID: 1481,
					},
					{
						ID: 1894, Name: "Кухар", Slug: "povar", ParentID: 1481,
					},
					{
						ID: 1895, Name: "Водій", Slug: "voditel", ParentID: 1481,
					},
					{
						ID: 1896, Name: "Няня", Slug: "nyanya", ParentID: 1481,
					},
					{
						ID: 1897, Name: "Доглядальниця", Slug: "sidelka", ParentID: 1481,
					},
					{
						ID: 1898, Name: "Зварювальник", Slug: "svarschik", ParentID: 1481,
					},
					{
						ID: 1899, Name: "Будівельник", Slug: "stroitel", ParentID: 1481,
					},
					{
						ID: 1900, Name: "Пакувальник", Slug: "upakovschik", ParentID: 1481,
					},
					{
						ID: 1901, Name: "Інше", Slug: "drugoe", ParentID: 1481,
					},
				},
			},
			{
				ID: 1866, Name: "Бухгалтерія", Slug: "buhgalteriya", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1871, Name: "Бухгалтер", Slug: "buhgalter", ParentID: 1866,
					},
					{
						ID: 1872, Name: "Аудитор", Slug: "auditor", ParentID: 1866,
					},
					{
						ID: 1873, Name: "Інше", Slug: "drugoe", ParentID: 1866,
					},
				},
			},
			{
				ID: 1874, Name: "Готельно-ресторанний бізнес / Туризм", Slug: "otelno-restorannyy-biznes-turizm", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1875, Name: "Адміністратор", Slug: "administrator", ParentID: 1874,
					},
					{
						ID: 1876, Name: "Покоївка", Slug: "gornichnaya", ParentID: 1874,
					},
					{
						ID: 1877, Name: "Менеджер", Slug: "menedzher", ParentID: 1874,
					},
					{
						ID: 1878, Name: "Бариста", Slug: "barista", ParentID: 1874,
					},
					{
						ID: 1879, Name: "Кухар", Slug: "povar", ParentID: 1874,
					},
					{
						ID: 1880, Name: "Бармен", Slug: "barmen", ParentID: 1874,
					},
					{
						ID: 1881, Name: "Офіціант", Slug: "ofitsiant", ParentID: 1874,
					},
					{
						ID: 1882, Name: "Кондитер", Slug: "konditer", ParentID: 1874,
					},
					{
						ID: 1883, Name: "Пекар", Slug: "pekar", ParentID: 1874,
					},
					{
						ID: 1884, Name: "Помічник кухаря", Slug: "pomoshchnik-povara", ParentID: 1874,
					},
					{
						ID: 1885, Name: "Сушист", Slug: "sushist", ParentID: 1874,
					},
					{
						ID: 1886, Name: "Посудомийниця", Slug: "posudomoyshchitsa", ParentID: 1874,
					},
					{
						ID: 1887, Name: "Піцайоло", Slug: "pitstsayolo", ParentID: 1874,
					},
					{
						ID: 1888, Name: "Інше", Slug: "drugoe", ParentID: 1874,
					},
				},
			},
			{
				ID: 1902, Name: "СТО / автомийки", Slug: "sto-avtomoyki", ParentID: 6,
				Children: []CategoryItem{
					{
						ID: 1903, Name: "Автослюсар", Slug: "avtoslesar", ParentID: 1902,
					},
					{
						ID: 1904, Name: "Мийник", Slug: "moyshchik", ParentID: 1902,
					},
					{
						ID: 1905, Name: "Механік", Slug: "mehanik", ParentID: 1902,
					},
					{
						ID: 1906, Name: "Автомаляр", Slug: "avtomalyar", ParentID: 1902,
					},
					{
						ID: 1907, Name: "Автоелектрик", Slug: "avtoelektrik", ParentID: 1902,
					},
					{
						ID: 1908, Name: "Рихтувальник", Slug: "rihtovshchik", ParentID: 1902,
					},
					{
						ID: 1909, Name: "Шиномонтажник", Slug: "shinomontazhnik", ParentID: 1902,
					},
					{
						ID: 1910, Name: "Інше", Slug: "drugoe", ParentID: 1902,
					},
				},
			},
			{
				ID: 3677, Name: "Служба в Силах оборони", Slug: "zsu", ParentID: 6,
			},
		},
	},
	{
		ID: 35, Name: "Тварини", Slug: "zhivotnye", Icon: "🐾", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 64, Name: "Собаки", Slug: "sobaki", ParentID: 35,
			},
			{
				ID: 65, Name: "Коти", Slug: "koshki", ParentID: 35,
			},
			{
				ID: 66, Name: "Інші тварини", Slug: "drugie-zhivotnye", ParentID: 35,
			},
			{
				ID: 67, Name: "Зоотовари", Slug: "tovary-dlya-zhivotnyh", ParentID: 35,
				Children: []CategoryItem{
					{
						ID: 3851, Name: "Товари для котів", Slug: "dlya-koshek", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3858, Name: "Корм для котів", Slug: "korm-i-lakomstva", ParentID: 3851,
							},
							{
								ID: 3859, Name: "Миски та поїлки для котів", Slug: "miski-i-poilki", ParentID: 3851,
							},
							{
								ID: 3860, Name: "Вітаміни та добавки для котів", Slug: "vitaminy-i-dobavki", ParentID: 3851,
							},
							{
								ID: 3861, Name: "Іграшки для котів", Slug: "igrushki", ParentID: 3851,
							},
							{
								ID: 3862, Name: "Догляд та гігієна для котів", Slug: "uhod-i-gigiena", ParentID: 3851,
							},
							{
								ID: 3863, Name: "Наповнювачі для котів", Slug: "napolniteli", ParentID: 3851,
							},
							{
								ID: 3864, Name: "Лотки для котів та аксесуари", Slug: "lotki-i-aksessuary", ParentID: 3851,
							},
							{
								ID: 3865, Name: "Дряпки для котів (кігтеточки)", Slug: "kogtetochki", ParentID: 3851,
							},
							{
								ID: 3866, Name: "Лежанки для котів", Slug: "lezhanki", ParentID: 3851,
							},
							{
								ID: 3867, Name: "Переноски для котів", Slug: "perenoski", ParentID: 3851,
							},
							{
								ID: 3868, Name: "Одяг для котів та аксесуари", Slug: "odezhda-i-amunitsiya", ParentID: 3851,
							},
							{
								ID: 3869, Name: "Інші товари для котів", Slug: "prochie-tovary", ParentID: 3851,
							},
						},
					},
					{
						ID: 3852, Name: "Товари для собак", Slug: "dlya-sobak", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3870, Name: "Корм для собак", Slug: "korm-i-lakomstva", ParentID: 3852,
							},
							{
								ID: 3871, Name: "Вітаміни та добавки для собак", Slug: "vitaminy-i-dobavki", ParentID: 3852,
							},
							{
								ID: 3872, Name: "Миски та поїлки для собак", Slug: "miski-i-poilki", ParentID: 3852,
							},
							{
								ID: 3873, Name: "Іграшки для собак", Slug: "igrushki", ParentID: 3852,
							},
							{
								ID: 3874, Name: "Засоби гігієни та грумінгу", Slug: "sredstva-gigieny-i-gruming", ParentID: 3852,
							},
							{
								ID: 3875, Name: "Одяг для собак", Slug: "odezhda", ParentID: 3852,
							},
							{
								ID: 3876, Name: "Амуніція для собак", Slug: "amunitsiya", ParentID: 3852,
							},
							{
								ID: 3877, Name: "Лежанки для собак", Slug: "lezhanki", ParentID: 3852,
							},
							{
								ID: 3878, Name: "Переноски для собак", Slug: "perenoski", ParentID: 3852,
							},
							{
								ID: 3879, Name: "Вольєри та будки для собак", Slug: "volery-i-budki", ParentID: 3852,
							},
							{
								ID: 3880, Name: "Пелюшки та туалети для собак", Slug: "pelenki-i-tualety", ParentID: 3852,
							},
							{
								ID: 3881, Name: "Інші товари для собак", Slug: "prochie-tovary", ParentID: 3852,
							},
						},
					},
					{
						ID: 3853, Name: "Товари для птахів", Slug: "dlya-ptits", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3882, Name: "Корм для птахів та добавки", Slug: "korm-i-dobavki", ParentID: 3853,
							},
							{
								ID: 3883, Name: "Клітки та переноски для птахів", Slug: "kletki-i-perenoski", ParentID: 3853,
							},
							{
								ID: 3884, Name: "Аксесуари для кліток птахів", Slug: "igrushki-aksessuary-dlya-kletok", ParentID: 3853,
							},
							{
								ID: 3885, Name: "Інші товари для птахів", Slug: "prochie-tovary", ParentID: 3853,
							},
						},
					},
					{
						ID: 3854, Name: "Товари для гризунів", Slug: "dlya-gryzunov", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3886, Name: "Корм та добавки для гризунів", Slug: "korm-i-dobavki", ParentID: 3854,
							},
							{
								ID: 3887, Name: "Годівниці, поїлки для гризунів", Slug: "kormushki-i-poilki", ParentID: 3854,
							},
							{
								ID: 3888, Name: "Клітки і переноски для гризунів", Slug: "kletki-i-perenoski", ParentID: 3854,
							},
							{
								ID: 3889, Name: "Іграшки й аксесуари для гризунів", Slug: "igrushki-i-aksessuary-dlya-kletok", ParentID: 3854,
							},
							{
								ID: 3890, Name: "Наповнювачі та догляд для гризунів", Slug: "napolniteli-i-uhod", ParentID: 3854,
							},
							{
								ID: 3891, Name: "Інші товари для гризунів", Slug: "prochie-tovary", ParentID: 3854,
							},
						},
					},
					{
						ID: 3855, Name: "Товари для рептилій", Slug: "dlya-reptiliy", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3892, Name: "Корм та добавки для рептилій", Slug: "korm-i-dobavki", ParentID: 3855,
							},
							{
								ID: 3893, Name: "Тераріуми та фаунаріуми", Slug: "terrariumy-i-faunariumy", ParentID: 3855,
							},
							{
								ID: 3894, Name: "Обладнання та аксесуари для тераріуму", Slug: "oborudovanie-i-aksessuary", ParentID: 3855,
							},
							{
								ID: 3895, Name: "Інші товари для рептилій", Slug: "prochie-tovary", ParentID: 3855,
							},
						},
					},
					{
						ID: 3856, Name: "Товари для тваринництва", Slug: "dlya-zhivotnovodstva", ParentID: 67,
						Children: []CategoryItem{
							{
								ID: 3896, Name: "Комбікорм та добавки", Slug: "kombikorm-i-dobavki", ParentID: 3856,
							},
							{
								ID: 3897, Name: "Інкубатори та обладнання", Slug: "inkubatory-i-oborudovanie", ParentID: 3856,
							},
							{
								ID: 3898, Name: "Інвентар для тваринництва", Slug: "inventar", ParentID: 3856,
							},
							{
								ID: 3899, Name: "Товари для бджільництва", Slug: "tovary-dlya-pchelovodstva", ParentID: 3856,
							},
							{
								ID: 3900, Name: "Інші товари для тваринництва", Slug: "prochee-dlya-zhivotnovodstva", ParentID: 3856,
							},
						},
					},
					{
						ID: 3857, Name: "Інші зоотовари", Slug: "prochee", ParentID: 67,
					},
				},
			},
			{
				ID: 120, Name: "Бюро знахідок", Slug: "byuro-nahodok", ParentID: 35,
			},
			{
				ID: 247, Name: "В'язка", Slug: "vyazka", ParentID: 35,
			},
			{
				ID: 365, Name: "Гризуни", Slug: "gryzuny", ParentID: 35,
			},
			{
				ID: 367, Name: "Акваріумістика", Slug: "akvariumnye-rybki", ParentID: 35,
			},
			{
				ID: 368, Name: "Сільгосп тварини", Slug: "selskohozyaystvennye-zhivotnye", ParentID: 35,
			},
			{
				ID: 1520, Name: "Безкоштовно (тварини і в'язка)", Slug: "besplatno-zhivotnye-i-vyazka", ParentID: 35,
				Children: []CategoryItem{
					{
						ID: 4005, Name: "Собаки (безкоштовно віддам)", Slug: "sobaki-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4006, Name: "Коти (безкоштовно віддам)", Slug: "koshki-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4007, Name: "Акваріумістика (безкоштовно віддам)", Slug: "akvariumistika-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4008, Name: "Гризуни (безкоштовно віддам)", Slug: "gryzuny-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4009, Name: "Сільгосп тварини (безкоштовно віддам)", Slug: "selskohozyaystvennye-zhivotnye-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4010, Name: "Інші тварини (безкоштовно віддам)", Slug: "drugie-zhivotnye-otdam-besplatno", ParentID: 1520,
					},
					{
						ID: 4011, Name: "В'язка (безкоштовно)", Slug: "vyazka-besplatno", ParentID: 1520,
					},
				},
			},
		},
	},
	{
		ID: 899, Name: "Дім і сад", Slug: "dom-i-sad", Icon: "🏡", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 48, Name: "Кімнатні рослини", Slug: "komnatnye-rasteniya", ParentID: 899,
			},
			{
				ID: 57, Name: "Меблі", Slug: "mebel", ParentID: 899,
				Children: []CategoryItem{
					{
						ID: 107, Name: "Меблі на замовлення", Slug: "mebel-na-zakaz", ParentID: 57,
					},
					{
						ID: 511, Name: "Меблі для вітальні", Slug: "mebel-dlya-gostinoy", ParentID: 57,
					},
					{
						ID: 515, Name: "Меблі для спальні", Slug: "mebel-dlya-spalni", ParentID: 57,
					},
					{
						ID: 522, Name: "Меблі для передпокою", Slug: "mebel-dlya-prihozhey", ParentID: 57,
					},
					{
						ID: 523, Name: "Кухонні меблі", Slug: "kuhonnaya-mebel", ParentID: 57,
					},
					{
						ID: 524, Name: "Меблі для ванної кімнати", Slug: "mebel-dlya-vannoy-komnaty", ParentID: 57,
					},
					{
						ID: 525, Name: "Офісні меблі", Slug: "ofisnaya-mebel", ParentID: 57,
					},
					{
						ID: 3902, Name: "Садові меблі", Slug: "sadovaya-mebel", ParentID: 57,
					},
				},
			},
			{
				ID: 60, Name: "Предмети інтер'єру", Slug: "predmety-interera", ParentID: 899,
				Children: []CategoryItem{
					{
						ID: 61, Name: "Освітлення", Slug: "svetilniki", ParentID: 60,
					},
					{
						ID: 529, Name: "Текстиль", Slug: "tekstil", ParentID: 60,
					},
					{
						ID: 533, Name: "Декор вікон", Slug: "dekor-okon", ParentID: 60,
					},
				},
			},
			{
				ID: 63, Name: "Інші товари для дому", Slug: "prochie-tovary-dlya-doma", ParentID: 899,
			},
			{
				ID: 93, Name: "Продукти харчування / напої", Slug: "produkty-pitaniya-napitki", ParentID: 899,
			},
			{
				ID: 537, Name: "Господарський інвентар / побутова хімія", Slug: "hozyaystvennyy-inventar", ParentID: 899,
			},
			{
				ID: 611, Name: "Садовий інвентар", Slug: "sadovyy-inventar", ParentID: 899,
			},
			{
				ID: 632, Name: "Канцтовари / витратні матеріали", Slug: "kantstovary-rashodnye-materialy", ParentID: 899,
			},
			{
				ID: 641, Name: "Будівництво / ремонт", Slug: "stroitelstvo-remont", ParentID: 899,
				Children: []CategoryItem{
					{
						ID: 642, Name: "Сантехніка", Slug: "santehnika", ParentID: 641,
					},
					{
						ID: 643, Name: "Вентиляція", Slug: "ventilyatsiya", ParentID: 641,
					},
					{
						ID: 644, Name: "Опалення", Slug: "otoplenie", ParentID: 641,
					},
					{
						ID: 645, Name: "Електрика", Slug: "elektrika", ParentID: 641,
					},
					{
						ID: 646, Name: "Пиломатеріали", Slug: "pilomaterialy", ParentID: 641,
					},
					{
						ID: 647, Name: "Оздоблювальні та облицювальні матеріали", Slug: "otdelochnye-i-oblitsovochnye-materialy", ParentID: 641,
					},
					{
						ID: 648, Name: "Вікна / двері / скло / дзеркала", Slug: "okna-steklo-zerkala", ParentID: 641,
					},
					{
						ID: 649, Name: "Лакофарбові матеріали", Slug: "lakokrasochnye-materialy", ParentID: 641,
					},
					{
						ID: 650, Name: "Металопрокат / арматура", Slug: "metalloprokat-armatura", ParentID: 641,
					},
					{
						ID: 651, Name: "Елементи кріплення", Slug: "elementy-krepezha", ParentID: 641,
					},
					{
						ID: 652, Name: "Цегла / бетон / піноблоки", Slug: "kirpich-beton-penobloki", ParentID: 641,
					},
					{
						ID: 653, Name: "Інші будматеріали", Slug: "prochie-stroymaterialy", ParentID: 641,
					},
					{
						ID: 3703, Name: "Шпалери", Slug: "oboi", ParentID: 641,
					},
					{
						ID: 3704, Name: "Фурнітура (віконна, дверна)", Slug: "okonnaya-dvernaya-furnitura", ParentID: 641,
					},
					{
						ID: 3705, Name: "Автоматика для воріт", Slug: "avtomatika-dlya-vorot", ParentID: 641,
					},
				},
			},
			{
				ID: 654, Name: "Інструменти", Slug: "instrumenty", ParentID: 899,
				Children: []CategoryItem{
					{
						ID: 655, Name: "Ручний інструмент", Slug: "ruchnoy-instrument", ParentID: 654,
					},
					{
						ID: 656, Name: "Бензоінструмент", Slug: "benzoinstrument", ParentID: 654,
					},
					{
						ID: 657, Name: "Електроінструмент", Slug: "elektroinstrument", ParentID: 654,
					},
					{
						ID: 658, Name: "Пневмоінструмент", Slug: "pnevmoinstrument", ParentID: 654,
					},
					{
						ID: 659, Name: "Інший інструмент", Slug: "prochiy-instrument", ParentID: 654,
					},
					{
						ID: 3285, Name: "Генератори", Slug: "generatory", ParentID: 654,
					},
				},
			},
			{
				ID: 667, Name: "Посуд / кухонне приладдя", Slug: "posuda-kuhonnaya-utvar", ParentID: 899,
			},
			{
				ID: 1433, Name: "Сад / город", Slug: "sad-ogorod", ParentID: 899,
			},
		},
	},
	{
		ID: 37, Name: "Електроніка", Slug: "elektronika", Icon: "📱", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 38, Name: "Комп'ютери та комплектуючі", Slug: "kompyutery-i-komplektuyuschie", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 78, Name: "Настільні комп'ютери", Slug: "nastolnye-kompyutery", ParentID: 38,
					},
					{
						ID: 83, Name: "Монітори", Slug: "monitory", ParentID: 38,
					},
					{
						ID: 84, Name: "Інше", Slug: "drugoe", ParentID: 38,
					},
					{
						ID: 453, Name: "Сервери", Slug: "servery", ParentID: 38,
					},
					{
						ID: 458, Name: "Комплектуючі та запчастини", Slug: "komplektuyuschie-i-aksesuary", ParentID: 38,
					},
					{
						ID: 466, Name: "Периферійні пристрої", Slug: "periferiynye-ustroystva", ParentID: 38,
					},
					{
						ID: 476, Name: "Зовнішні накопичувачі", Slug: "vneshnie-nakopiteli", ParentID: 38,
					},
					{
						ID: 480, Name: "Витратні матеріали", Slug: "rashodnye-materialy", ParentID: 38,
					},
					{
						ID: 3825, Name: "Мережеве обладнання", Slug: "setevoye-oborudovaniye", ParentID: 38,
						Children: []CategoryItem{
							{
								ID: 3829, Name: "Роутери (маршрутизатори)", Slug: "routery", ParentID: 3825,
							},
							{
								ID: 3833, Name: "Комутатори", Slug: "kommutatory", ParentID: 3825,
							},
							{
								ID: 3836, Name: "Мережеві адаптери", Slug: "setevye-adaptery", ParentID: 3825,
							},
							{
								ID: 3839, Name: "Ретранслятори Wi-Fi", Slug: "retranslyatory-wi-fi", ParentID: 3825,
							},
							{
								ID: 3842, Name: "Інше мережеве обладнання", Slug: "drugoe", ParentID: 3825,
							},
						},
					},
					{
						ID: 3826, Name: "Оргтехніка", Slug: "orgtehnika", ParentID: 38,
						Children: []CategoryItem{
							{
								ID: 3830, Name: "Принтери / БФП", Slug: "printery-mfu", ParentID: 3826,
							},
							{
								ID: 3834, Name: "Сканери", Slug: "skanery", ParentID: 3826,
							},
							{
								ID: 3837, Name: "Копіри", Slug: "kopiry", ParentID: 3826,
							},
							{
								ID: 3840, Name: "3D принтери та аксесуари", Slug: "3d-printery-i-aksessuary", ParentID: 3826,
							},
							{
								ID: 3843, Name: "Матеріали для 3D-друку", Slug: "materialy-3d", ParentID: 3826,
							},
							{
								ID: 3844, Name: "Ламінатори / біндери / шредери", Slug: "laminatory-bindery-shredery", ParentID: 3826,
							},
							{
								ID: 3845, Name: "Картриджі / тонери", Slug: "kartridzhi-tonery", ParentID: 3826,
							},
							{
								ID: 3850, Name: "Інша оргтехніка", Slug: "drugoe", ParentID: 3826,
							},
						},
					},
				},
			},
			{
				ID: 44, Name: "Телефони та аксесуари", Slug: "telefony-i-aksesuary", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 85, Name: "Смартфони / мобільні телефони", Slug: "mobilnye-telefony-smartfony", ParentID: 44,
					},
					{
						ID: 86, Name: "Сім-карти / тарифи / номери", Slug: "sim-karty-tarify-nomera", ParentID: 44,
					},
					{
						ID: 88, Name: "Рації та інші телефони", Slug: "ratsiyi-i-prochie-telefony", ParentID: 44,
					},
					{
						ID: 442, Name: "Аксесуари для телефонів", Slug: "aksessuary-dlya-telefonov", ParentID: 44,
						Children: []CategoryItem{
							{
								ID: 3272, Name: "PowerBank та акумулятори", Slug: "akkumulyatory-i-portativnye-batarei", ParentID: 442,
							},
							{
								ID: 3273, Name: "Зарядні пристрої та док-станції", Slug: "zaryadnye-ustroystva-i-dok-stantsii", ParentID: 442,
							},
							{
								ID: 3274, Name: "Гарнітури", Slug: "garnitury", ParentID: 442,
							},
							{
								ID: 3276, Name: "Чохли", Slug: "chehly", ParentID: 442,
							},
							{
								ID: 3277, Name: "Data-кабелі / адаптери", Slug: "data-kabeli", ParentID: 442,
							},
							{
								ID: 3278, Name: "Карти пам'яті", Slug: "karty-pamyati", ParentID: 442,
							},
							{
								ID: 3279, Name: "Захисні плівки / скло", Slug: "zaschitnye-plenki-i-stekla", ParentID: 442,
							},
							{
								ID: 3280, Name: "Штативи / селфі-палиці", Slug: "monopody-dlya-selfi", ParentID: 442,
							},
							{
								ID: 3281, Name: "Окуляри віртуальної реальності", Slug: "ochki-virtualnoy-realnosti", ParentID: 442,
							},
							{
								ID: 3284, Name: "Інші аксесуари для телефонів", Slug: "prochie-aksessuary-dlya-telefonov", ParentID: 442,
							},
							{
								ID: 3783, Name: "Стабілізатори для телефону", Slug: "stabilizatory", ParentID: 442,
							},
						},
					},
					{
						ID: 450, Name: "Стаціонарні телефони", Slug: "statsionarnye-telefony", ParentID: 44,
					},
					{
						ID: 1482, Name: "Запчастини для телефонів", Slug: "zapchasti-dlya-telefonov", ParentID: 44,
						Children: []CategoryItem{
							{
								ID: 3747, Name: "Дисплеї / екрани", Slug: "displei-ekrany", ParentID: 1482,
							},
							{
								ID: 3748, Name: "Акумулятори", Slug: "akkumulyatory", ParentID: 1482,
							},
							{
								ID: 3749, Name: "Роз'єми", Slug: "razemy", ParentID: 1482,
							},
							{
								ID: 3750, Name: "Задні кришки", Slug: "zadnie-kryshki", ParentID: 1482,
							},
							{
								ID: 3751, Name: "Корпуси / панелі", Slug: "korpusa-paneli", ParentID: 1482,
							},
							{
								ID: 3752, Name: "Камери", Slug: "kamery", ParentID: 1482,
							},
							{
								ID: 3753, Name: "Шлейфи", Slug: "shleyfy", ParentID: 1482,
							},
							{
								ID: 3754, Name: "Динаміки / мікрофони", Slug: "dinamiki-mikrofony", ParentID: 1482,
							},
							{
								ID: 3755, Name: "Інші запчастини для телефонів", Slug: "drugoe", ParentID: 1482,
							},
						},
					},
					{
						ID: 3805, Name: "Смарт годинники / фітнес браслети", Slug: "smart-chasy-fitnes-braslety", ParentID: 44,
						Children: []CategoryItem{
							{
								ID: 3806, Name: "Смарт-годинники", Slug: "smart-chasy", ParentID: 3805,
							},
							{
								ID: 3807, Name: "Фітнес-браслети / трекери", Slug: "fitnes-braslety-trekery", ParentID: 3805,
							},
							{
								ID: 3808, Name: "Аксесуари та запчастини для смарт-годинників", Slug: "aksessuary-i-zapchasti", ParentID: 3805,
							},
						},
					},
				},
			},
			{
				ID: 55, Name: "Ігри та ігрові приставки", Slug: "igry-i-igrovye-pristavki", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 101, Name: "Приставки", Slug: "pristavki", ParentID: 55,
					},
					{
						ID: 102, Name: "Ігри для приставок", Slug: "igry-dlya-pristavok", ParentID: 55,
					},
					{
						ID: 103, Name: "Ігрові акаунти", Slug: "geroi-igr", ParentID: 55,
					},
					{
						ID: 484, Name: "Аксесуари для приставок", Slug: "aksessuary", ParentID: 55,
						Children: []CategoryItem{
							{
								ID: 3822, Name: "Джойстики, геймпади", Slug: "dzhoystiki-geympady", ParentID: 484,
							},
							{
								ID: 3823, Name: "Ігрові маніпулятори", Slug: "igrovye-manipulyatory", ParentID: 484,
							},
							{
								ID: 3824, Name: "Інші ігрові аксесуари", Slug: "drugoe", ParentID: 484,
							},
						},
					},
					{
						ID: 1161, Name: "Ігри для PC", Slug: "igry-dlya-pc", ParentID: 55,
					},
					{
						ID: 3815, Name: "Магазин предметів інвентарю", Slug: "magazin-predmetov-inventarya", ParentID: 55,
					},
					{
						ID: 3817, Name: "VR окуляри", Slug: "vr-ochki", ParentID: 55,
					},
				},
			},
			{
				ID: 369, Name: "Техніка для кухні", Slug: "tehnika-dlya-kuhni", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 370, Name: "Холодильники", Slug: "holodilniki", ParentID: 369,
					},
					{
						ID: 371, Name: "Мікрохвильові печі", Slug: "mikrovolnovye-pechi", ParentID: 369,
					},
					{
						ID: 372, Name: "Плити / варильні поверхні", Slug: "plity-pechi", ParentID: 369,
					},
					{
						ID: 380, Name: "Посудомийні машини", Slug: "posudomoechnye-mashiny", ParentID: 369,
					},
					{
						ID: 381, Name: "Витяжки", Slug: "vytyazhki", ParentID: 369,
					},
					{
						ID: 382, Name: "Інша техніка для кухні", Slug: "prochaya-tehnika-dlya-kuhni", ParentID: 369,
					},
					{
						ID: 3785, Name: "Морозильні камери", Slug: "morozilnye-kamery", ParentID: 369,
					},
					{
						ID: 3786, Name: "Дрібна побутова техніка", Slug: "melkaya-bytovaya-tehnika", ParentID: 369,
						Children: []CategoryItem{
							{
								ID: 3787, Name: "Кавоварки / кавомолки", Slug: "kofevarki-kofemolki", ParentID: 3786,
							},
							{
								ID: 3788, Name: "Електрочайники", Slug: "elektrochayniki", ParentID: 3786,
							},
							{
								ID: 3789, Name: "Хлібопічки", Slug: "hlebopechki", ParentID: 3786,
							},
							{
								ID: 3790, Name: "Соковижималки", Slug: "sokovyzhimalki", ParentID: 3786,
							},
							{
								ID: 3791, Name: "Блендери", Slug: "blendery", ParentID: 3786,
							},
							{
								ID: 3792, Name: "Вафельниці / бутербродниці", Slug: "vafelnitsy-buterbrodnitsy", ParentID: 3786,
							},
							{
								ID: 3793, Name: "Тостери", Slug: "tostery", ParentID: 3786,
							},
							{
								ID: 3794, Name: "Фритюрниці / аерогрилі", Slug: "frityurnitsy-aerogrili", ParentID: 3786,
							},
							{
								ID: 3795, Name: "Міксери", Slug: "miksery", ParentID: 3786,
							},
							{
								ID: 3796, Name: "Електром'ясорубки", Slug: "elektromyasorubki", ParentID: 3786,
							},
							{
								ID: 3797, Name: "Грилі / електробарбекю", Slug: "grili-elektrobarbekyu", ParentID: 3786,
							},
							{
								ID: 3798, Name: "Сушарки (дегідратори)", Slug: "sushilki", ParentID: 3786,
							},
							{
								ID: 3799, Name: "Кулери для води", Slug: "kulery-dlya-vody", ParentID: 3786,
							},
							{
								ID: 3800, Name: "Кухонні комбайни / подрібнювачі", Slug: "kuhonnye-kombayny-izmelchiteli", ParentID: 3786,
							},
							{
								ID: 3801, Name: "Кавомашини", Slug: "kofemashiny", ParentID: 3786,
							},
							{
								ID: 3802, Name: "Мультиварки / пароварки", Slug: "multivarki-parovarki", ParentID: 3786,
							},
							{
								ID: 3803, Name: "Вакуумні пакувальники", Slug: "vakuumnye-upakovshchiki", ParentID: 3786,
							},
							{
								ID: 3804, Name: "Інша дрібна побутова техніка", Slug: "drugoe", ParentID: 3786,
							},
						},
					},
					{
						ID: 3827, Name: "Запчастини для кухонної техніки", Slug: "zapchasti", ParentID: 369,
						Children: []CategoryItem{
							{
								ID: 3831, Name: "Запчастини для холодильників", Slug: "zapchasti-holodilnika", ParentID: 3827,
							},
							{
								ID: 3846, Name: "Запчастини для іншої кухонної техніки", Slug: "drugoe", ParentID: 3827,
							},
						},
					},
				},
			},
			{
				ID: 383, Name: "Техніка для дому", Slug: "tehnika-dlya-doma", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 76, Name: "Інша техніка для дому", Slug: "prochaya-tehnika-dlya-doma", ParentID: 383,
					},
					{
						ID: 384, Name: "Пилососи", Slug: "pylesosy", ParentID: 383,
					},
					{
						ID: 385, Name: "Праски", Slug: "utyugi", ParentID: 383,
					},
					{
						ID: 386, Name: "Пральні машини", Slug: "stiralnye-mashiny", ParentID: 383,
					},
					{
						ID: 387, Name: "Швейні машини та оверлоки", Slug: "shveynye-mashiny-i-overloki", ParentID: 383,
					},
					{
						ID: 388, Name: "В'язальні машини", Slug: "vyazalnye-mashiny", ParentID: 383,
					},
					{
						ID: 389, Name: "Фільтри для води", Slug: "filtry-dlya-vody", ParentID: 383,
					},
					{
						ID: 3286, Name: "Зарядні станції", Slug: "zaryadnyye-stantsii", ParentID: 383,
					},
					{
						ID: 3287, Name: "Перетворювачі напруги", Slug: "preobrazovateli-napryazheniya", ParentID: 383,
					},
					{
						ID: 3288, Name: "Безперебійники живлення", Slug: "bespereboyniki-pitaniya", ParentID: 383,
					},
					{
						ID: 3728, Name: "Сушильні машини", Slug: "sushilnye-mashiny", ParentID: 383,
					},
					{
						ID: 3828, Name: "Аксесуари та запчастини", Slug: "aksessuary-i-zapchasti", ParentID: 383,
						Children: []CategoryItem{
							{
								ID: 3832, Name: "Аксесуари та запчастини для пилососів", Slug: "dlya-pylesosov", ParentID: 3828,
							},
							{
								ID: 3835, Name: "Аксесуари та запчастини для пральних машин", Slug: "dlya-stiralnyh-mashin", ParentID: 3828,
							},
							{
								ID: 3838, Name: "Аксесуари та запчастини для сушильних машин", Slug: "dlya-sushilnyh-mashin", ParentID: 3828,
							},
							{
								ID: 3841, Name: "Аксесуари та запчастини для прасок", Slug: "dlya-utyugov", ParentID: 3828,
							},
							{
								ID: 3847, Name: "Інше", Slug: "drugoe", ParentID: 3828,
							},
						},
					},
				},
			},
			{
				ID: 390, Name: "Індивідуальний догляд", Slug: "individualnyy-uhod", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 391, Name: "Бритви", Slug: "britvy-epilyatory-mashinki-dlya-strizhki", ParentID: 390,
					},
					{
						ID: 392, Name: "Фени для волосся", Slug: "feny-ukladka-volos", ParentID: 390,
					},
					{
						ID: 393, Name: "Ваги підлогові", Slug: "vesy", ParentID: 390,
					},
					{
						ID: 394, Name: "Інша техніка для індивідуального догляду", Slug: "prochaya-tehnika-dlya-individualnogo-uhoda", ParentID: 390,
					},
					{
						ID: 3756, Name: "Машинки для стрижки", Slug: "mashinki-dlya-strizhki", ParentID: 390,
					},
					{
						ID: 3757, Name: "Епілятори", Slug: "epilyatory", ParentID: 390,
					},
					{
						ID: 3758, Name: "Тримери", Slug: "trimmery", ParentID: 390,
					},
					{
						ID: 3759, Name: "Щипці / плойки для волосся", Slug: "shchiptsy-ployki-dlya-volos", ParentID: 390,
					},
					{
						ID: 3760, Name: "Стайлери", Slug: "staylery", ParentID: 390,
					},
					{
						ID: 3761, Name: "Масажери", Slug: "massazhery", ParentID: 390,
					},
					{
						ID: 3762, Name: "Зубні щітки / іригатори", Slug: "zubnye-shchetki-irrigatory", ParentID: 390,
					},
					{
						ID: 3763, Name: "Девайси для обличчя", Slug: "devaysy-dlya-litsa", ParentID: 390,
					},
					{
						ID: 3764, Name: "Небулайзери / інгалятори", Slug: "nebulayzery-ingalyatory", ParentID: 390,
					},
				},
			},
			{
				ID: 395, Name: "Кліматичне обладнання", Slug: "klimaticheskoe-oborudovanie", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 3809, Name: "Кондиціонери", Slug: "kondicioneri", ParentID: 395,
					},
					{
						ID: 3810, Name: "Обігрівачі", Slug: "obogrevateli", ParentID: 395,
					},
					{
						ID: 3811, Name: "Тепловентилятори", Slug: "teploventilyatory-teplovye-pushki", ParentID: 395,
					},
					{
						ID: 3812, Name: "Вентилятори", Slug: "ventilyatory", ParentID: 395,
					},
					{
						ID: 3813, Name: "Очищувачі та озонатори повітря", Slug: "ochistiteli-ozonatory-vozduha", ParentID: 395,
					},
					{
						ID: 3814, Name: "Зволожувачі та осушувачі повітря", Slug: "uvlazhniteli-osushyteli-vozduha", ParentID: 395,
					},
					{
						ID: 3816, Name: "Метеостанції", Slug: "meteostancii", ParentID: 395,
					},
					{
						ID: 3818, Name: "Інше кліматичне обладнання", Slug: "prochee-klimaticheskoe-oborudovanie", ParentID: 395,
					},
				},
			},
			{
				ID: 401, Name: "Аксесуари й комплектуючі", Slug: "aksessuary-i-komplektuyuschie", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 3716, Name: "Адаптери / конектори", Slug: "adaptery-konnektory", ParentID: 401,
					},
					{
						ID: 3717, Name: "Батарейки / акумулятори", Slug: "batareyki-akkumulyatory", ParentID: 401,
					},
					{
						ID: 3718, Name: "Діоди", Slug: "diody", ParentID: 401,
					},
					{
						ID: 3719, Name: "Конденсатори", Slug: "kondensatory", ParentID: 401,
					},
					{
						ID: 3720, Name: "Плати", Slug: "platy", ParentID: 401,
					},
					{
						ID: 3721, Name: "Резистори", Slug: "rezistory", ParentID: 401,
					},
					{
						ID: 3722, Name: "Реле / регулятори / стабілізатори", Slug: "rele-regulyatory-stabilizatory", ParentID: 401,
					},
					{
						ID: 3723, Name: "Транзистори", Slug: "tranzistory", ParentID: 401,
					},
					{
						ID: 3724, Name: "Трансформатори", Slug: "transformatory", ParentID: 401,
					},
					{
						ID: 3725, Name: "Електродвигуни", Slug: "elektrodvigateli-motory", ParentID: 401,
					},
					{
						ID: 3726, Name: "Інші комплектуючі", Slug: "drugoe", ParentID: 401,
					},
				},
			},
			{
				ID: 402, Name: "Аудіотехніка", Slug: "audiotehnika", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 72, Name: "Mp3 плеєри", Slug: "mp3-pleery", ParentID: 402,
					},
					{
						ID: 404, Name: "Навушники та аксесуари", Slug: "naushniki", ParentID: 402,
						Children: []CategoryItem{
							{
								ID: 3848, Name: "Навушники", Slug: "naushniki", ParentID: 404,
							},
							{
								ID: 3849, Name: "Аксесуари для навушників", Slug: "aksessuary-dlya-naushnikov", ParentID: 404,
							},
						},
					},
					{
						ID: 405, Name: "Портативна акустика", Slug: "portativnaya-akustika", ParentID: 402,
					},
					{
						ID: 409, Name: "Магнітоли та магнітофони", Slug: "magnitoly", ParentID: 402,
					},
					{
						ID: 410, Name: "Радіоприймачі", Slug: "radiopriemniki", ParentID: 402,
					},
					{
						ID: 413, Name: "Музичні центри", Slug: "muzykalnye-tsentry", ParentID: 402,
					},
					{
						ID: 414, Name: "Акустичні системи", Slug: "akusticheskie-sistemy", ParentID: 402,
					},
					{
						ID: 415, Name: "Підсилювачі та ресивери", Slug: "usiliteli-resivery", ParentID: 402,
					},
					{
						ID: 416, Name: "Cd / md / вінілові програвачі", Slug: "cd-md-vinilovye-proigryvateli", ParentID: 402,
					},
					{
						ID: 419, Name: "Інша аудіотехніка", Slug: "prochaya-audiotehnika", ParentID: 402,
					},
					{
						ID: 3819, Name: "Диктофони", Slug: "diktofony", ParentID: 402,
					},
					{
						ID: 3820, Name: "Мікрофони", Slug: "mikrofony", ParentID: 402,
					},
					{
						ID: 3821, Name: "Комплектуючі для аудіотехніки", Slug: "aksessuary-komplektuyushie-dlya-audiotehniki", ParentID: 402,
					},
				},
			},
			{
				ID: 420, Name: "Тв / відеотехніка", Slug: "tv-videotehnika", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 75, Name: "Телевізори", Slug: "televizory", ParentID: 420,
					},
					{
						ID: 421, Name: "Медіа програвачі", Slug: "dvd-pleery", ParentID: 420,
					},
					{
						ID: 424, Name: "Аксесуари для ТВ/Відеотехніки", Slug: "aksessuary-dlya-tv-videotehniki", ParentID: 420,
					},
					{
						ID: 426, Name: "Супутникове ТВ", Slug: "sputnikovoe-tv", ParentID: 420,
					},
					{
						ID: 428, Name: "Інша ТВ / відеотехніка", Slug: "prochaya-tv-videotehnika", ParentID: 420,
					},
					{
						ID: 1518, Name: "Проектори", Slug: "proektory", ParentID: 420,
					},
					{
						ID: 3746, Name: "Системи відеоспостереження", Slug: "sistemy-videonablyudeniya", ParentID: 420,
					},
				},
			},
			{
				ID: 429, Name: "Фото / відео", Slug: "foto-video", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 73, Name: "Цифрові фотоапарати", Slug: "tsifrovye-fotoapparaty", ParentID: 429,
					},
					{
						ID: 430, Name: "Плівкові фотоапарати", Slug: "plenochnye-fotoapparaty", ParentID: 429,
					},
					{
						ID: 431, Name: "Відеокамери", Slug: "videokamery", ParentID: 429,
					},
					{
						ID: 432, Name: "Об'єктиви", Slug: "obektivy", ParentID: 429,
					},
					{
						ID: 433, Name: "Штативи / моноподи", Slug: "shtativy-monopody", ParentID: 429,
					},
					{
						ID: 434, Name: "Фотоспалахи", Slug: "fotovspyshki", ParentID: 429,
					},
					{
						ID: 435, Name: "Аксесуари для фото / відеокамер", Slug: "aksessuary-dlya-foto-videokamer", ParentID: 429,
					},
					{
						ID: 441, Name: "Телескопи / біноклі", Slug: "teleskopy-binokli", ParentID: 429,
					},
					{
						ID: 1500, Name: "Екшн-камери", Slug: "ekshn-kamery", ParentID: 429,
					},
					{
						ID: 3727, Name: "Студійне обладнання", Slug: "studiynoe-oborudovanie", ParentID: 429,
						Children: []CategoryItem{
							{
								ID: 3741, Name: "Світло", Slug: "svet", ParentID: 3727,
							},
							{
								ID: 3742, Name: "Відбивачі", Slug: "otrazhateli", ParentID: 3727,
							},
							{
								ID: 3743, Name: "Фотофон", Slug: "fotofon", ParentID: 3727,
							},
							{
								ID: 3744, Name: "Підставки / кріплення / тримачі", Slug: "podstavki-krepleniya-derzhateli", ParentID: 3727,
							},
							{
								ID: 3745, Name: "Інша студійна техніка", Slug: "drugoe", ParentID: 3727,
							},
						},
					},
				},
			},
			{
				ID: 1431, Name: "Інша електроніка", Slug: "prochaja-electronika", ParentID: 37,
			},
			{
				ID: 1502, Name: "Ноутбуки та аксесуари", Slug: "noutbuki-i-aksesuary", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 80, Name: "Ноутбуки", Slug: "noutbuki", ParentID: 1502,
					},
					{
						ID: 1504, Name: "Запчастини для ноутбуків", Slug: "zapchasti-dlya-noutbukov", ParentID: 1502,
						Children: []CategoryItem{
							{
								ID: 3770, Name: "Акумулятори для ноутбуків", Slug: "akkumulyatory", ParentID: 1504,
							},
							{
								ID: 3772, Name: "Зарядні пристрої", Slug: "zaryadnye-ustroystva", ParentID: 1504,
							},
							{
								ID: 3774, Name: "Матриці для ноутбуків", Slug: "matritsy", ParentID: 1504,
							},
							{
								ID: 3776, Name: "Клавіатури для ноутбуків", Slug: "klaviatury", ParentID: 1504,
							},
							{
								ID: 3778, Name: "Кулери / системи охолодження", Slug: "kulery-sistemy-ohlazhdeniya", ParentID: 1504,
							},
							{
								ID: 3784, Name: "Інші запчастини для ноутбуків", Slug: "drugoe", ParentID: 1504,
							},
						},
					},
					{
						ID: 1506, Name: "Аксесуари для ноутбуків", Slug: "aksessuary-dlya-noutbukov", ParentID: 1502,
						Children: []CategoryItem{
							{
								ID: 3769, Name: "USB Hubs", Slug: "usb-hubs", ParentID: 1506,
							},
							{
								ID: 3771, Name: "Wi-Fi адаптери", Slug: "wi-fi-adaptery", ParentID: 1506,
							},
							{
								ID: 3773, Name: "Док-станції", Slug: "dok-stantsii", ParentID: 1506,
							},
							{
								ID: 3775, Name: "Килимки для миші", Slug: "kovriki-dlya-myshi", ParentID: 1506,
							},
							{
								ID: 3777, Name: "Підставки / стійки", Slug: "podstavki-stoyki", ParentID: 1506,
							},
							{
								ID: 3779, Name: "Чохли / накладки для ноутбуків", Slug: "chehly-nakladki-dlya-noutbukov", ParentID: 1506,
							},
							{
								ID: 3780, Name: "Сумки та рюкзаки для ноутбуків", Slug: "sumki-i-ryukzaki-dlya-noutbukov", ParentID: 1506,
							},
							{
								ID: 3781, Name: "Засоби для чищення", Slug: "chistyashchie-sredstva", ParentID: 1506,
							},
							{
								ID: 3782, Name: "Інші комп'ютерні аксесуари", Slug: "drugoe", ParentID: 1506,
							},
						},
					},
				},
			},
			{
				ID: 1508, Name: "Планшети / ел. книги та аксесуари", Slug: "planshety-el-knigi-i-aksessuary", ParentID: 37,
				Children: []CategoryItem{
					{
						ID: 81, Name: "Планшетні комп'ютери", Slug: "planshetnye-kompyutery", ParentID: 1508,
						Children: []CategoryItem{
							{
								ID: 3729, Name: "Acer", Slug: "acer", ParentID: 81,
							},
							{
								ID: 3730, Name: "Asus", Slug: "asus", ParentID: 81,
							},
							{
								ID: 3731, Name: "Apple", Slug: "ipad", ParentID: 81,
							},
							{
								ID: 3732, Name: "Blackview", Slug: "blackview", ParentID: 81,
							},
							{
								ID: 3733, Name: "Huawei", Slug: "huawei", ParentID: 81,
							},
							{
								ID: 3734, Name: "Lenovo", Slug: "lenovo", ParentID: 81,
							},
							{
								ID: 3735, Name: "Microsoft", Slug: "microsoft", ParentID: 81,
							},
							{
								ID: 3736, Name: "Pixus", Slug: "pixus", ParentID: 81,
							},
							{
								ID: 3737, Name: "Samsung", Slug: "samsung", ParentID: 81,
							},
							{
								ID: 3738, Name: "Teclast", Slug: "teclast", ParentID: 81,
							},
							{
								ID: 3739, Name: "Xiaomi", Slug: "xiaomi", ParentID: 81,
							},
							{
								ID: 3740, Name: "Інше", Slug: "drugoe", ParentID: 81,
							},
						},
					},
					{
						ID: 1510, Name: "Електронні книги", Slug: "elektronnye-knigi", ParentID: 1508,
					},
					{
						ID: 1512, Name: "Запчастини для планшетів / ел. книг", Slug: "zapchasti-dlya-planshetov-el-knig", ParentID: 1508,
					},
					{
						ID: 1514, Name: "Аксесуари для планшетів / ел. книг", Slug: "aksessuary-dlya-planshetov-el-knig", ParentID: 1508,
						Children: []CategoryItem{
							{
								ID: 3765, Name: "Клавіатури", Slug: "klaviatura", ParentID: 1514,
							},
							{
								ID: 3766, Name: "Стілуси", Slug: "stilus", ParentID: 1514,
							},
							{
								ID: 3767, Name: "Чохли", Slug: "chehly", ParentID: 1514,
							},
							{
								ID: 3768, Name: "Інші аксесуари для планшетів", Slug: "drugoe", ParentID: 1514,
							},
						},
					},
					{
						ID: 1516, Name: "Графічні планшети", Slug: "graficheskiye-planshety", ParentID: 1508,
					},
				},
			},
		},
	},
	{
		ID: 7, Name: "Бізнес та послуги", Slug: "uslugi", Icon: "🛠", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 42, Name: "Продаж бізнесу", Slug: "prodazha-biznesa", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3532, Name: "Магазини", Slug: "magaziny", ParentID: 42,
					},
					{
						ID: 3535, Name: "Акаунти в соц мережах", Slug: "akkaunty-v-sots-setyah", ParentID: 42,
					},
					{
						ID: 3538, Name: "Підприємства, компанії", Slug: "predpriyatiya-kompanii", ParentID: 42,
					},
					{
						ID: 3540, Name: "Кафе, ресторани, кав'ярні", Slug: "kafe-restorany-kofeyni", ParentID: 42,
					},
					{
						ID: 3542, Name: "Франшизи", Slug: "franshizy", ParentID: 42,
					},
					{
						ID: 3543, Name: "Інтернет-магазини", Slug: "internet-magaziny", ParentID: 42,
					},
					{
						ID: 3544, Name: "Інші види бізнесу", Slug: "drugie-vidy-biznesa", ParentID: 42,
					},
				},
			},
			{
				ID: 47, Name: "Продаж обладнання для бізнесу", Slug: "oborudovanie", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3555, Name: "Торгово-виставкове обладнання", Slug: "torgovo-vystavochnoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3556, Name: "Холодильне обладнання", Slug: "holodilnoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3557, Name: "Розрахунково-касове обладнання", Slug: "raschetno-kassovoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3558, Name: "Промислове обладнання", Slug: "promyshlennoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3559, Name: "Обладнання для ресторанів і кафе", Slug: "oborudovanie-dlya-restoranov-kafe", ParentID: 47,
					},
					{
						ID: 3560, Name: "Лабораторне обладнання", Slug: "laboratornoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3561, Name: "Медичне обладнання", Slug: "meditsinskoe-oborudovanie", ParentID: 47,
					},
					{
						ID: 3562, Name: "Обладнання для салонів краси", Slug: "oborudovanie-dlya-salonov-krasoty", ParentID: 47,
					},
					{
						ID: 3563, Name: "Інше обладнання", Slug: "drugoe", ParentID: 47,
					},
				},
			},
			{
				ID: 112, Name: "Авто / мото послуги", Slug: "avto-moto-uslugi", ParentID: 7,
			},
			{
				ID: 186, Name: "Інші послуги", Slug: "prochie-uslugi", ParentID: 7,
			},
			{
				ID: 188, Name: "Будівництво та ремонт", Slug: "stroitelstvo-otdelka-remont", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 822, Name: "Будівельні послуги", Slug: "stroitelnye-uslugi", ParentID: 188,
					},
					{
						ID: 823, Name: "Проєктування та дизайн", Slug: "dizayn-arhitektura", ParentID: 188,
					},
					{
						ID: 824, Name: "Оздоблення, ремонт", Slug: "otdelka-remont", ParentID: 188,
					},
					{
						ID: 825, Name: "Монтаж вікон, дверей, балконів", Slug: "okna-dveri-balkony", ParentID: 188,
					},
					{
						ID: 826, Name: "Сантехніка та комунікації", Slug: "santehnika-kommunikatsii", ParentID: 188,
					},
					{
						ID: 827, Name: "Вентиляція та кондиціонування", Slug: "ventilyatsiya-konditsionirovanie", ParentID: 188,
					},
					{
						ID: 828, Name: "Електромонтажні послуги", Slug: "elektrika", ParentID: 188,
					},
					{
						ID: 829, Name: "Готові конструкції", Slug: "gotovye-konstruktsii", ParentID: 188,
					},
					{
						ID: 3517, Name: "Будівництво під ключ", Slug: "stroitelstvo-pod-klyuch", ParentID: 188,
					},
					{
						ID: 3518, Name: "Ремонт під ключ", Slug: "remont-pod-klyuch", ParentID: 188,
					},
					{
						ID: 3519, Name: "Встановлення та ремонт систем безпеки", Slug: "ustanovka-sistem-bezopasnosti", ParentID: 188,
					},
					{
						ID: 3520, Name: "Демонтажні роботи", Slug: "demontazhnye-raboty", ParentID: 188,
					},
					{
						ID: 3521, Name: "Ландшафтні та земляні роботи", Slug: "landshaftnye-zemlyanye-raboty", ParentID: 188,
					},
					{
						ID: 3522, Name: "Буріння свердловин", Slug: "burenie-skvazhin", ParentID: 188,
					},
					{
						ID: 3523, Name: "Інші послуги з будівництва", Slug: "drugoe", ParentID: 188,
					},
					{
						ID: 3524, Name: "Зварювальні роботи", Slug: "svarochnye-raboty", ParentID: 188,
					},
				},
			},
			{
				ID: 190, Name: "Туризм / іміграція", Slug: "turizm-immigratsiya", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3530, Name: "Тури та екскурсії", Slug: "tury-ekskursii", ParentID: 190,
					},
					{
						ID: 3533, Name: "Оформлення віз", Slug: "oformlenie-viz", ParentID: 190,
					},
					{
						ID: 3536, Name: "Інші послуги туризму", Slug: "drugoe", ParentID: 190,
					},
				},
			},
			{
				ID: 192, Name: "Перевезення та послуги спецтехніки", Slug: "perevozki-arenda-transporta", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3257, Name: "Інші транспортні послуги", Slug: "drugiye-transportnye-uslugi", ParentID: 192,
					},
					{
						ID: 3260, Name: "Вантажоперевезення", Slug: "gruzoperevozki", ParentID: 192,
					},
					{
						ID: 3261, Name: "Пасажирські перевезення", Slug: "passazhirskie-perevozki", ParentID: 192,
					},
					{
						ID: 3457, Name: "Послуги маніпулятора", Slug: "uslugi-manipulyatora", ParentID: 192,
					},
					{
						ID: 3501, Name: "Послуги евакуатора, автовозу", Slug: "evakuatory-avtovozy", ParentID: 192,
					},
					{
						ID: 3502, Name: "Послуги вантажників", Slug: "gruzchiki", ParentID: 192,
					},
					{
						ID: 3503, Name: "Послуги водіїв", Slug: "uslugi-voditeley", ParentID: 192,
					},
				},
			},
			{
				ID: 193, Name: "Краса / здоров'я", Slug: "krasota-zdorove", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 851, Name: "Перукарські послуги", Slug: "strizhki-naraschivanie-volos", ParentID: 193,
					},
					{
						ID: 852, Name: "Манікюр, педикюр, нарощування", Slug: "manikyur-naraschivanie-nogtey", ParentID: 193,
					},
					{
						ID: 853, Name: "Вії та брови", Slug: "makiyazh-kosmetologiya-naraschivanie-resnits", ParentID: 193,
					},
					{
						ID: 854, Name: "Медичні послуги", Slug: "meditsina", ParentID: 193,
					},
					{
						ID: 855, Name: "Масаж", Slug: "massazh", ParentID: 193,
					},
					{
						ID: 856, Name: "Інші б'юті-послуги", Slug: "krasota-zdorove-prochee", ParentID: 193,
					},
					{
						ID: 1175, Name: "Послуги психолога", Slug: "psihologiya", ParentID: 193,
					},
					{
						ID: 3458, Name: "Тату, татуаж, пірсинг", Slug: "tatu-tatuazh-pirsing", ParentID: 193,
					},
					{
						ID: 3459, Name: "Візаж та макіяж", Slug: "vizazh-makiyazh", ParentID: 193,
					},
					{
						ID: 3460, Name: "Косметологічні послуги", Slug: "kosmetologicheskie-uslugi", ParentID: 193,
					},
					{
						ID: 3461, Name: "Видалення волосся", Slug: "udalenie-volos", ParentID: 193,
					},
					{
						ID: 3462, Name: "Лазні, сауни та СПА", Slug: "bani-sauny-spa", ParentID: 193,
					},
					{
						ID: 3463, Name: "Моделі для б'юті-послуг", Slug: "modeli-dlya-byuti-uslug", ParentID: 193,
					},
				},
			},
			{
				ID: 195, Name: "Послуги для тварин", Slug: "uslugi-dlya-zhivotnyh", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3531, Name: "Дресирування, кінологи, вигул", Slug: "dressirovka-kinologi-vygul", ParentID: 195,
					},
					{
						ID: 3534, Name: "Грумінг, стрижки, салони", Slug: "gruming-strizhki-salony", ParentID: 195,
					},
					{
						ID: 3537, Name: "Перетримка тварин і готелі", Slug: "perederzhka-zhivotnyh-gostinitsy", ParentID: 195,
					},
					{
						ID: 3539, Name: "Ветеринарна допомога", Slug: "veterinarnaya-pomoshch", ParentID: 195,
					},
					{
						ID: 3541, Name: "Інші послуги для тварин", Slug: "drugoe_2", ParentID: 195,
					},
				},
			},
			{
				ID: 196, Name: "Фото та відеозйомка", Slug: "razvlechenie-foto-video", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3497, Name: "Фотозйомка та ретуш фото", Slug: "fotosemka-retush-foto", ParentID: 196,
					},
					{
						ID: 3498, Name: "Оцифрування та реставрація", Slug: "otsifrovka-restavratsiya", ParentID: 196,
					},
					{
						ID: 3499, Name: "Відеозйомка та монтаж відео", Slug: "videosemka-montazh-video", ParentID: 196,
					},
					{
						ID: 3500, Name: "Інші фото та відео послуги", Slug: "drugoe", ParentID: 196,
					},
				},
			},
			{
				ID: 203, Name: "Послуги освіти та спорту", Slug: "obrazovanie", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3490, Name: "Автоінструктори", Slug: "avtoinstruktory", ParentID: 203,
					},
					{
						ID: 3491, Name: "Дитячий розвиток, логопеди", Slug: "detskoe-razvitie-logopedy", ParentID: 203,
					},
					{
						ID: 3492, Name: "Репетитори", Slug: "repetitory", ParentID: 203,
					},
					{
						ID: 3493, Name: "Курси та тренінги", Slug: "kursy-treningi", ParentID: 203,
					},
					{
						ID: 3494, Name: "Викладачі музики та вокалу", Slug: "prepodavateli-muzyki-vokala", ParentID: 203,
					},
					{
						ID: 3495, Name: "Спорт і танці", Slug: "sport-tantsy", ParentID: 203,
					},
					{
						ID: 3496, Name: "Інші послуги освіти", Slug: "drugoe", ParentID: 203,
					},
				},
			},
			{
				ID: 204, Name: "Догляд за дітьми та літніми людьми", Slug: "nyani-sidelki", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3464, Name: "Няня", Slug: "nyanya", ParentID: 204,
					},
					{
						ID: 3465, Name: "Доглядальниця", Slug: "sidelka", ParentID: 204,
					},
				},
			},
			{
				ID: 726, Name: "Сировина / матеріали", Slug: "syre-materialy", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3525, Name: "Паливо", Slug: "toplivo", ParentID: 726,
					},
					{
						ID: 3526, Name: "Дрова, вугілля, пелети", Slug: "drova-ugol-pellety", ParentID: 726,
					},
					{
						ID: 3527, Name: "Текстильні матеріали", Slug: "tekstilnye-materialy", ParentID: 726,
					},
					{
						ID: 3528, Name: "Гума, пластик, полімери", Slug: "rezina-plastik-polimery", ParentID: 726,
					},
					{
						ID: 3529, Name: "Інша сировина", Slug: "drugoe", ParentID: 726,
					},
				},
			},
			{
				ID: 836, Name: "Ремонт та обслуговування техніки", Slug: "remont-i-obsluzhivanie-tehniki", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 87, Name: "Телефони і смартфони", Slug: "remont-proshivka-telefonov-i-smartfonov", ParentID: 836,
					},
					{
						ID: 104, Name: "Ігрові приставки", Slug: "remont-proshivka-igrovyh-pristavok", ParentID: 836,
					},
					{
						ID: 1484, Name: "Кліматична техніка", Slug: "remont-i-obsluzhivanie-klimatycheskoi-tekhniki", ParentID: 836,
					},
					{
						ID: 1486, Name: "Комп'ютерна техніка", Slug: "remont-i-obsluzhivanie-kompyuterov", ParentID: 836,
					},
					{
						ID: 1488, Name: "Фото і відеоапаратура", Slug: "remont-i-obsluzhivanie-foto-i-videoapparatury", ParentID: 836,
					},
					{
						ID: 1490, Name: "Тв і відеотехніка", Slug: "remont-i-obsluzhivanie-tv-i-videotekhniki", ParentID: 836,
					},
					{
						ID: 1492, Name: "Аудіотехніка", Slug: "remont-i-obsluzhivanie-audiotekhniki", ParentID: 836,
					},
					{
						ID: 1494, Name: "Побутова техніка", Slug: "remont-i-obsluzhivanie-bytovoi-tekhniki", ParentID: 836,
					},
					{
						ID: 1496, Name: "Інша техніка", Slug: "remont-i-obsluzhivanie-prochei-tekhniki", ParentID: 836,
					},
					{
						ID: 3514, Name: "Оргтехніка, заправка картриджів", Slug: "orgtehnika-zapravka-kartridzhey", ParentID: 836,
					},
					{
						ID: 3515, Name: "Електротранспорт", Slug: "elektrotransport", ParentID: 836,
					},
					{
						ID: 3516, Name: "Електро та бензоінструмент", Slug: "elektro-benzoinstrument", ParentID: 836,
					},
				},
			},
			{
				ID: 3466, Name: "Побутові послуги", Slug: "bytovye-uslugi", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3479, Name: "Хімчистка та прання", Slug: "himchistka-stirka", ParentID: 3466,
					},
					{
						ID: 3480, Name: "Пошиття та ремонт одягу", Slug: "poshiv-remont-odezhdy", ParentID: 3466,
					},
					{
						ID: 3481, Name: "Ремонт взуття та сумок", Slug: "remont-obuvi-sumok", ParentID: 3466,
					},
					{
						ID: 3482, Name: "Дезінфекція та дезінсекція", Slug: "dezinfektsiya-dezinsektsiya", ParentID: 3466,
					},
					{
						ID: 3483, Name: "Аварійне відкриття замків", Slug: "avariynoe-vskrytie-zamkov", ParentID: 3466,
					},
					{
						ID: 3484, Name: "Виготовлення ключів", Slug: "izgotovlenie-klyuchey", ParentID: 3466,
					},
					{
						ID: 3485, Name: "Послуги домробітниці", Slug: "domrabotnitsy", ParentID: 3466,
					},
					{
						ID: 3486, Name: "Послуги асенізатора", Slug: "uslugi-assenizatora", ParentID: 3466,
					},
					{
						ID: 3487, Name: "Вивезення сміття", Slug: "vyvoz-musora", ParentID: 3466,
					},
					{
						ID: 3488, Name: "Послуги садівника, догляд за територією", Slug: "uslugi-sadovnika", ParentID: 3466,
					},
					{
						ID: 3489, Name: "Інші побутові послуги", Slug: "drugoe", ParentID: 3466,
					},
				},
			},
			{
				ID: 3467, Name: "Клінінг", Slug: "klining", ParentID: 7,
			},
			{
				ID: 3504, Name: "Організація свят", Slug: "organizatsiya-prazdnikov", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3506, Name: "Аніматори та ростові ляльки", Slug: "animatory-rostovye-kukly", ParentID: 3504,
					},
					{
						ID: 3508, Name: "Кейтеринг, кухарі", Slug: "keytering-povary", ParentID: 3504,
					},
					{
						ID: 3510, Name: "Ведучі, музиканти, тамада", Slug: "vedushchie-muzykanty-tamada", ParentID: 3504,
					},
					{
						ID: 3511, Name: "Оформлення кульками", Slug: "oformlenie-sharami", ParentID: 3504,
					},
					{
						ID: 3512, Name: "Десерти та випічка на замовлення", Slug: "deserty-vupechka-na-zakaz", ParentID: 3504,
					},
					{
						ID: 3513, Name: "Інші послуги для свят", Slug: "drugoe", ParentID: 3504,
					},
				},
			},
			{
				ID: 3505, Name: "Ритуальні послуги", Slug: "ritualnye-uslugi", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3507, Name: "Виготовлення пам'ятників", Slug: "izgotovlenie-pamyatnikov", ParentID: 3505,
					},
					{
						ID: 3509, Name: "Інші ритуальні послуги", Slug: "inshe", ParentID: 3505,
					},
				},
			},
			{
				ID: 3545, Name: "Прийом вторсировини", Slug: "priem-vtorsyrya", ParentID: 7,
			},
			{
				ID: 3546, Name: "Ділові послуги", Slug: "delovye-uslugi", ParentID: 7,
				Children: []CategoryItem{
					{
						ID: 3547, Name: "Бухгалтерські послуги", Slug: "buhgalterskie", ParentID: 3546,
					},
					{
						ID: 3548, Name: "Створення сайтів", Slug: "sozdanie-prodvizhenie-saytov", ParentID: 3546,
					},
					{
						ID: 3549, Name: "Юридичні послуги", Slug: "yuridicheskie", ParentID: 3546,
					},
					{
						ID: 3550, Name: "Поліграфія, реклама, друк", Slug: "poligrafiya-reklama-pechat", ParentID: 3546,
					},
					{
						ID: 3551, Name: "Фінансові послуги, партнерство", Slug: "finansovye-uslugi-partnerstvo", ParentID: 3546,
					},
					{
						ID: 3552, Name: "Маркетингові послуги", Slug: "marketingovye-uslugi", ParentID: 3546,
					},
					{
						ID: 3553, Name: "Послуги перекладачів, набір тексту", Slug: "perevodchiki-nabor-teksta", ParentID: 3546,
					},
					{
						ID: 3554, Name: "Інші ділові послуги", Slug: "drugoe", ParentID: 3546,
					},
				},
			},
		},
	},
	{
		ID: 3709, Name: "Житло подобово", Slug: "zhytlo-podobovo", Icon: "🏨", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 3710, Name: "Будинки подобово, погодинно", Slug: "podobovo-pohodynno-budynky", ParentID: 3709,
			},
			{
				ID: 3711, Name: "Квартири подобово, погодинно", Slug: "podobovo-pohodynno-kvartyry", ParentID: 3709,
			},
			{
				ID: 3712, Name: "Кімнати подобово, погодинно", Slug: "podobovo-pohodynno-kimnaty", ParentID: 3709,
			},
			{
				ID: 3713, Name: "Готелі, бази відпочинку", Slug: "podobovo-pohodynno-hoteli", ParentID: 3709,
			},
			{
				ID: 3714, Name: "Хостели, койко-місця", Slug: "podobovo-pohodynno-khostely", ParentID: 3709,
			},
			{
				ID: 3715, Name: "Пропозиції Туроператорів", Slug: "propozytsii-turoperatoriv", ParentID: 3709,
			},
		},
	},
	{
		ID: 3428, Name: "Оренда та прокат", Slug: "arenda-prokat", Icon: "🎒", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 3429, Name: "Оренда транспорту та спецтехніки", Slug: "transport-spetstehnika", ParentID: 3428,
				Children: []CategoryItem{
					{
						ID: 3440, Name: "Авто", Slug: "avto", ParentID: 3429,
					},
					{
						ID: 3441, Name: "Автобус, мікроавтобус", Slug: "avtobusy-mikroavtobusy", ParentID: 3429,
					},
					{
						ID: 3442, Name: "Причеп", Slug: "pritsepy-lafety", ParentID: 3429,
					},
					{
						ID: 3443, Name: "Фура, трал, довгомір", Slug: "fury-traly-dlinnomery", ParentID: 3429,
					},
					{
						ID: 3444, Name: "Самоскид, зерновоз", Slug: "samosvaly-zernovozy", ParentID: 3429,
					},
					{
						ID: 3445, Name: "Навантажувач", Slug: "pogruzchiky", ParentID: 3429,
					},
					{
						ID: 3446, Name: "Кран", Slug: "krany", ParentID: 3429,
					},
					{
						ID: 3447, Name: "Автовишка", Slug: "avtovyshki", ParentID: 3429,
					},
					{
						ID: 3448, Name: "Екскаватор, бульдозер", Slug: "ekskavatory-buldozery", ParentID: 3429,
					},
					{
						ID: 3449, Name: "Трактор, комбайн", Slug: "traktory-kombayny", ParentID: 3429,
					},
					{
						ID: 3450, Name: "Інший транспорт", Slug: "drugoe", ParentID: 3429,
					},
				},
			},
			{
				ID: 3430, Name: "Прокат велосипедів і мото", Slug: "velosipedy-moto", ParentID: 3428,
			},
			{
				ID: 3431, Name: "Оренда обладнання", Slug: "oborudovanie", ParentID: 3428,
				Children: []CategoryItem{
					{
						ID: 3451, Name: "Оренда генераторів", Slug: "generatory", ParentID: 3431,
					},
					{
						ID: 3452, Name: "Оренда опалубки", Slug: "opalubka", ParentID: 3431,
					},
					{
						ID: 3453, Name: "Оренда риштувань", Slug: "lesa", ParentID: 3431,
					},
					{
						ID: 3454, Name: "Оренда бетонозмішувачів та віброплит", Slug: "betonomeshalki-vibroplity", ParentID: 3431,
					},
					{
						ID: 3455, Name: "Оренда біотуалетів і побутівок", Slug: "biotualety-bytovki", ParentID: 3431,
					},
					{
						ID: 3456, Name: "Інше обладнання", Slug: "drugoe", ParentID: 3431,
					},
				},
			},
			{
				ID: 3432, Name: "Прокат інструментів", Slug: "instrumenty", ParentID: 3428,
			},
			{
				ID: 3433, Name: "Прокат товарів мед призначення", Slug: "tovary-med-naznacheniya", ParentID: 3428,
			},
			{
				ID: 3434, Name: "Прокат техніки та електроніки", Slug: "tehnika-elektronika", ParentID: 3428,
			},
			{
				ID: 3435, Name: "Прокат товарів для заходів", Slug: "tovary-dlya-meropriyatiy", ParentID: 3428,
			},
			{
				ID: 3436, Name: "Прокат спорт і туристичних товарів", Slug: "sport-turisticheskie-tovary", ParentID: 3428,
			},
			{
				ID: 3437, Name: "Прокат одягу та аксесуарів", Slug: "odezhda-aksessuary", ParentID: 3428,
			},
			{
				ID: 3438, Name: "Прокат дитячого одягу та товарів", Slug: "detskaya-odezhda-tovary", ParentID: 3428,
			},
			{
				ID: 3439, Name: "Інші товари на прокат", Slug: "drugoe", ParentID: 3428,
			},
		},
	},
	{
		ID: 891, Name: "Мода і стиль", Slug: "moda-i-stil", Icon: "👗", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 52, Name: "Краса / здоров'я", Slug: "krasota-zdorove", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 568, Name: "Парфумерія", Slug: "parfyumeriya", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 2728, Name: "Духи (парфуми)", Slug: "duhi", ParentID: 568,
							},
							{
								ID: 2729, Name: "Парфумована вода", Slug: "parfyumirovannaya-voda", ParentID: 568,
							},
							{
								ID: 2730, Name: "Парфумовані олії і масляні духи", Slug: "maslo-maslyanye-duhi", ParentID: 568,
							},
							{
								ID: 2731, Name: "Одеколон", Slug: "odekolon", ParentID: 568,
							},
							{
								ID: 2733, Name: "Набори", Slug: "nabory", ParentID: 568,
							},
							{
								ID: 2734, Name: "Інше", Slug: "drugoe", ParentID: 568,
							},
							{
								ID: 2735, Name: "Туалетна вода", Slug: "tualetnaya-voda", ParentID: 568,
							},
						},
					},
					{
						ID: 570, Name: "Обладнання перукарень / салонів краси", Slug: "oborudovanie-parikmaherskih-salonov-krasoty", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3000, Name: "Для манікюру і педикюру", Slug: "dlya-manikyura-pedikyura", ParentID: 570,
							},
							{
								ID: 3001, Name: "Для перукарні", Slug: "dlya-parikmaherskoy", ParentID: 570,
							},
							{
								ID: 3002, Name: "Для масажних салонів", Slug: "dlya-massazhnyh-salonov", ParentID: 570,
							},
							{
								ID: 3003, Name: "Для солярію", Slug: "dlya-solyariya", ParentID: 570,
							},
							{
								ID: 3004, Name: "Меблі для салонів краси", Slug: "mebel-dlya-salonov-krasoty", ParentID: 570,
							},
							{
								ID: 3005, Name: "Кільцеві лампи", Slug: "koltsevye-lampy", ParentID: 570,
							},
							{
								ID: 3006, Name: "Інше", Slug: "drugoe", ParentID: 570,
							},
						},
					},
					{
						ID: 571, Name: "Товари для реабілітації та осіб з інвалідністю", Slug: "tovary-dlya-invalidov", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 2830, Name: "Коляски", Slug: "kolyaski", ParentID: 571,
							},
							{
								ID: 2831, Name: "Тростини", Slug: "trosti", ParentID: 571,
							},
							{
								ID: 2832, Name: "Милиці", Slug: "kostyli", ParentID: 571,
							},
							{
								ID: 2833, Name: "Ходунки", Slug: "hodunki", ParentID: 571,
							},
							{
								ID: 2834, Name: "Туалетні стільці", Slug: "tualetnye-stulya", ParentID: 571,
							},
							{
								ID: 2835, Name: "Матраци від пролежнів", Slug: "matrasy-ot-prolezhney", ParentID: 571,
							},
							{
								ID: 2836, Name: "Інше", Slug: "drugoe", ParentID: 571,
							},
						},
					},
					{
						ID: 572, Name: "Інші товари для краси і здоров'я", Slug: "prochie-tovary-dlya-krasoty-i-zdorovya", ParentID: 52,
					},
					{
						ID: 3007, Name: "Догляд за обличчям", Slug: "uhod-za-litsom", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3019, Name: "Креми для обличчя", Slug: "kremy", ParentID: 3007,
							},
							{
								ID: 3031, Name: "Маски для обличчя", Slug: "maski", ParentID: 3007,
							},
							{
								ID: 3043, Name: "Олія для обличчя", Slug: "masla", ParentID: 3007,
							},
							{
								ID: 3055, Name: "Скраб для обличчя", Slug: "skrab", ParentID: 3007,
							},
							{
								ID: 3067, Name: "Спрей для обличчя", Slug: "sprey", ParentID: 3007,
							},
							{
								ID: 3078, Name: "Пілінг для обличчя", Slug: "piling", ParentID: 3007,
							},
							{
								ID: 3087, Name: "Засоби для очищення обличчя", Slug: "sredstva-dlya-ochishcheniya-litsa", ParentID: 3007,
							},
							{
								ID: 3095, Name: "Сонцезахисні засоби", Slug: "solntsezashchitnye-sredstva", ParentID: 3007,
							},
							{
								ID: 3101, Name: "Тоніки для обличчя", Slug: "toniki", ParentID: 3007,
							},
							{
								ID: 3106, Name: "Міцелярна вода", Slug: "mitsellyarnaya-voda", ParentID: 3007,
							},
							{
								ID: 3111, Name: "Сироватки для обличчя", Slug: "syvorotki", ParentID: 3007,
							},
							{
								ID: 3115, Name: "Догляд за очима", Slug: "uhod-dlya-glaz", ParentID: 3007,
							},
							{
								ID: 3128, Name: "Інше", Slug: "drugoe", ParentID: 3007,
							},
						},
					},
					{
						ID: 3008, Name: "Макіяж / декоративна косметика", Slug: "makiyazh", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3020, Name: "Туш", Slug: "tush", ParentID: 3008,
							},
							{
								ID: 3032, Name: "Тіні", Slug: "teni", ParentID: 3008,
							},
							{
								ID: 3044, Name: "Контурні олівці", Slug: "konturnye-karandashi", ParentID: 3008,
							},
							{
								ID: 3056, Name: "Підводка", Slug: "podvodka", ParentID: 3008,
							},
							{
								ID: 3068, Name: "Коректори", Slug: "korrektory", ParentID: 3008,
							},
							{
								ID: 3079, Name: "Накладні вії", Slug: "nakladnye-resnitsy", ParentID: 3008,
							},
							{
								ID: 3088, Name: "Губна помада", Slug: "gubnaya-pomada", ParentID: 3008,
							},
							{
								ID: 3096, Name: "Блиск для губ", Slug: "blesk-dlya-gub", ParentID: 3008,
							},
							{
								ID: 3102, Name: "Тінт для губ", Slug: "tint-dlya-gub", ParentID: 3008,
							},
							{
								ID: 3107, Name: "Бальзами для губ", Slug: "balzamy-dlya-gub", ParentID: 3008,
							},
							{
								ID: 3112, Name: "Косметика для брів", Slug: "kosmetika-dlya-brovey", ParentID: 3008,
							},
							{
								ID: 3116, Name: "Тональний крем", Slug: "tonalniy-krem", ParentID: 3008,
							},
							{
								ID: 3119, Name: "Пудри", Slug: "pudry", ParentID: 3008,
							},
							{
								ID: 3121, Name: "Хайлайтер", Slug: "haylayter", ParentID: 3008,
							},
							{
								ID: 3122, Name: "Рум'яна", Slug: "rumyana", ParentID: 3008,
							},
							{
								ID: 3123, Name: "Бронзери", Slug: "bronzery", ParentID: 3008,
							},
							{
								ID: 3124, Name: "BB-креми і CC-креми", Slug: "bb-cc-kremy", ParentID: 3008,
							},
							{
								ID: 3125, Name: "Набори", Slug: "nabory", ParentID: 3008,
							},
							{
								ID: 3126, Name: "Бази / праймери під макіяж", Slug: "bazy-praymery", ParentID: 3008,
							},
							{
								ID: 3127, Name: "Пензлі", Slug: "kisti", ParentID: 3008,
							},
							{
								ID: 3132, Name: "Інше", Slug: "drugoe", ParentID: 3008,
							},
						},
					},
					{
						ID: 3009, Name: "Косметика для волосся", Slug: "kosmetika-dlya-volos", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3021, Name: "Шампуні", Slug: "shampuni", ParentID: 3009,
							},
							{
								ID: 3045, Name: "Кондиціонери для волосся", Slug: "konditsionery", ParentID: 3009,
							},
							{
								ID: 3057, Name: "Бальзами для волосся", Slug: "balzamy", ParentID: 3009,
							},
							{
								ID: 3069, Name: "Лосьйони для волосся", Slug: "losony", ParentID: 3009,
							},
							{
								ID: 3080, Name: "Олія для волосся", Slug: "maslo", ParentID: 3009,
							},
							{
								ID: 3089, Name: "Лаки для волосся", Slug: "laki", ParentID: 3009,
							},
							{
								ID: 3097, Name: "Мус для волосся", Slug: "muss", ParentID: 3009,
							},
							{
								ID: 3103, Name: "Засоби від лупи", Slug: "sredstva-ot-perhoti", ParentID: 3009,
							},
							{
								ID: 3108, Name: "Засоби від випадіння волосся", Slug: "sredstva-ot-vypadeniya", ParentID: 3009,
							},
							{
								ID: 3113, Name: "Термозахист", Slug: "termozashchita", ParentID: 3009,
							},
							{
								ID: 3117, Name: "Канекалони", Slug: "kanekalony", ParentID: 3009,
							},
							{
								ID: 3130, Name: "Маски для волосся", Slug: "maski", ParentID: 3009,
							},
							{
								ID: 3133, Name: "Інше", Slug: "drugoe", ParentID: 3009,
							},
						},
					},
					{
						ID: 3010, Name: "Фарбування", Slug: "okrashivanie", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3022, Name: "Фарба для волосся", Slug: "kraska", ParentID: 3010,
							},
							{
								ID: 3034, Name: "Тонуючі засоби", Slug: "toniruyushchie-sredstva", ParentID: 3010,
							},
							{
								ID: 3046, Name: "Окислювач для волосся", Slug: "okislitel", ParentID: 3010,
							},
							{
								ID: 3058, Name: "Знебарвлення та освітлення", Slug: "obestsvechivanie-osvetlenie", ParentID: 3010,
							},
							{
								ID: 3070, Name: "Хна", Slug: "hna", ParentID: 3010,
							},
							{
								ID: 3134, Name: "Інше", Slug: "drugoe", ParentID: 3010,
							},
						},
					},
					{
						ID: 3011, Name: "Косметика для тіла", Slug: "kosmetika-dlya-tela", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3023, Name: "Дезодоранти і антиперспіранти", Slug: "dezodoranty-antiperspiranty", ParentID: 3011,
							},
							{
								ID: 3035, Name: "Гелі для душу", Slug: "geli", ParentID: 3011,
							},
							{
								ID: 3047, Name: "Мило", Slug: "mylo", ParentID: 3011,
							},
							{
								ID: 3059, Name: "Креми для тіла", Slug: "krema-dlya-tela", ParentID: 3011,
							},
							{
								ID: 3071, Name: "Скраби для тіла", Slug: "skraby", ParentID: 3011,
							},
							{
								ID: 3081, Name: "Сіль для ванни", Slug: "sol", ParentID: 3011,
							},
							{
								ID: 3090, Name: "Піна для ванни", Slug: "pena", ParentID: 3011,
							},
							{
								ID: 3098, Name: "Засоби для засмаги", Slug: "sredstva-dlya-zagara", ParentID: 3011,
							},
							{
								ID: 3104, Name: "Косметика для рук", Slug: "kosmetika-dlya-ruk", ParentID: 3011,
							},
							{
								ID: 3109, Name: "Крем для рук", Slug: "krem-dlya-ruk", ParentID: 3011,
							},
							{
								ID: 3114, Name: "Косметика для ніг", Slug: "kosmetika-dlya-nog", ParentID: 3011,
							},
							{
								ID: 3118, Name: "Крем для ніг", Slug: "krem-dlya-nog", ParentID: 3011,
							},
							{
								ID: 3120, Name: "Дезодорант для ніг", Slug: "dezodorant-dlya-nog", ParentID: 3011,
							},
							{
								ID: 3131, Name: "Масла для тіла", Slug: "masla", ParentID: 3011,
							},
							{
								ID: 3135, Name: "Інше", Slug: "drugoe", ParentID: 3011,
							},
						},
					},
					{
						ID: 3012, Name: "Для ротової порожнини", Slug: "dlya-polosti-rta", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3024, Name: "Зубні пасти", Slug: "pasty", ParentID: 3012,
							},
							{
								ID: 3036, Name: "Зубні щітки", Slug: "shchetki", ParentID: 3012,
							},
							{
								ID: 3048, Name: "Зубний порошок", Slug: "poroshok", ParentID: 3012,
							},
							{
								ID: 3060, Name: "Ополіскувачі для порожнини рота, зубів, ясен", Slug: "opolaskivateli", ParentID: 3012,
							},
							{
								ID: 3072, Name: "Дорожні набори", Slug: "dorozhnye-nabory", ParentID: 3012,
							},
							{
								ID: 3136, Name: "Інше", Slug: "drugoe", ParentID: 3012,
							},
						},
					},
					{
						ID: 3013, Name: "Для нігтів", Slug: "dlya-nogtey", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3025, Name: "Гель-лаки", Slug: "gel-laki", ParentID: 3013,
							},
							{
								ID: 3037, Name: "Рідина для зняття лаку", Slug: "zhidkost-dlya-snyatiya-laka", ParentID: 3013,
							},
							{
								ID: 3049, Name: "Декор для нігтів", Slug: "dekor", ParentID: 3013,
							},
							{
								ID: 3061, Name: "Догляд за кутикулою", Slug: "uhod-za-kutikuloy", ParentID: 3013,
							},
							{
								ID: 3073, Name: "Накладні нігті", Slug: "nakladnye-nogti", ParentID: 3013,
							},
							{
								ID: 3082, Name: "Інструменти для манікюру та педикюру", Slug: "instrumenty-dlya-manikyura-pedikyura", ParentID: 3013,
							},
							{
								ID: 3091, Name: "Стартовий набір для манікюру", Slug: "startoviy-nabor-dlya-manikyura", ParentID: 3013,
							},
							{
								ID: 3129, Name: "Лаки для нігтів", Slug: "laki", ParentID: 3013,
							},
							{
								ID: 3137, Name: "Інше", Slug: "drugoe", ParentID: 3013,
							},
						},
					},
					{
						ID: 3014, Name: "Для гоління", Slug: "dlya-britya", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3026, Name: "Косметика для гоління", Slug: "kosmetika", ParentID: 3014,
							},
							{
								ID: 3038, Name: "Станок для гоління", Slug: "stanok", ParentID: 3014,
							},
							{
								ID: 3050, Name: "Леза для гоління", Slug: "lezviya", ParentID: 3014,
							},
							{
								ID: 3062, Name: "Косметика після гоління", Slug: "kosmetika-posle-britya", ParentID: 3014,
							},
							{
								ID: 3074, Name: "Крем для депіляції", Slug: "krem-dlya-depilyatsii", ParentID: 3014,
							},
							{
								ID: 3083, Name: "Віск для депіляції", Slug: "vosk-dlya-depilyatsii", ParentID: 3014,
							},
							{
								ID: 3092, Name: "Паста для шугарінгу", Slug: "pasta-dlya-shugaringa", ParentID: 3014,
							},
							{
								ID: 3138, Name: "Інше", Slug: "drugoe", ParentID: 3014,
							},
						},
					},
					{
						ID: 3015, Name: "Гігієнічні засоби", Slug: "gigienicheskie-sredstva", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3027, Name: "Прокладки", Slug: "prokladki", ParentID: 3015,
							},
							{
								ID: 3039, Name: "Тампони", Slug: "tampony", ParentID: 3015,
							},
							{
								ID: 3051, Name: "Антисептики", Slug: "antiseptiki", ParentID: 3015,
							},
							{
								ID: 3063, Name: "Підгузки для дорослих", Slug: "podguzniki-dlya-vzroslyh", ParentID: 3015,
							},
							{
								ID: 3075, Name: "Менструальні чаші", Slug: "menstrualnye-chashi", ParentID: 3015,
							},
							{
								ID: 3084, Name: "Засоби для інтимної гігієни", Slug: "sredstva-dlya-intimnoy-gigieny", ParentID: 3015,
							},
							{
								ID: 3139, Name: "Інше", Slug: "drugoe", ParentID: 3015,
							},
						},
					},
					{
						ID: 3016, Name: "Ортопедичні товари", Slug: "ortopedicheskie-tovary", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3028, Name: "Пристосування для стоп", Slug: "prisposobleniya-dlya-stopy", ParentID: 3016,
							},
							{
								ID: 3040, Name: "Фіксатори", Slug: "fiksatory", ParentID: 3016,
							},
							{
								ID: 3052, Name: "Корсети, коректори постави", Slug: "korsety-korrektory-osanki", ParentID: 3016,
							},
							{
								ID: 3064, Name: "Наколінники", Slug: "nakolenniki", ParentID: 3016,
							},
							{
								ID: 3076, Name: "Бандажі та пояси", Slug: "bandazhi-poyasa", ParentID: 3016,
							},
							{
								ID: 3085, Name: "Гіпс, бинти, тейпи", Slug: "gips-binty-teypy", ParentID: 3016,
							},
							{
								ID: 3093, Name: "Ортези", Slug: "ortezy", ParentID: 3016,
							},
							{
								ID: 3099, Name: "Інші ортопедичні товари", Slug: "drugie-ortopedicheskie-tovary", ParentID: 3016,
							},
						},
					},
					{
						ID: 3017, Name: "Медичне приладдя", Slug: "meditsinskie-prinadlezhnosti", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3029, Name: "Одноразові шапочки", Slug: "odnorazovye-shapochki", ParentID: 3017,
							},
							{
								ID: 3041, Name: "Медичні рукавиці", Slug: "meditsinskie-perchatki", ParentID: 3017,
							},
							{
								ID: 3053, Name: "Перев'язувальні матеріали", Slug: "perevyazochnye-materialy", ParentID: 3017,
							},
							{
								ID: 3065, Name: "Медичні маски і респіратори", Slug: "meditsinskie-maski-respiratory", ParentID: 3017,
							},
							{
								ID: 3140, Name: "Інше", Slug: "drugoe", ParentID: 3017,
							},
						},
					},
					{
						ID: 3018, Name: "Косметичні інструменти та аксесуари", Slug: "kosmeticheskie-instrumenty-aksessuary", ParentID: 52,
						Children: []CategoryItem{
							{
								ID: 3030, Name: "Дзеркала для макіяжу", Slug: "zerkala-dlya-makiyazha", ParentID: 3018,
							},
							{
								ID: 3042, Name: "Пінцети та щипці", Slug: "pintsety-shchiptsy", ParentID: 3018,
							},
							{
								ID: 3054, Name: "Гребінці", Slug: "rascheski", ParentID: 3018,
							},
							{
								ID: 3066, Name: "Спонжі", Slug: "sponzhi", ParentID: 3018,
							},
							{
								ID: 3077, Name: "Інструменти для чистки обличчя", Slug: "instrumenty-dlya-chistki-litsa", ParentID: 3018,
							},
							{
								ID: 3086, Name: "Органайзери для косметики", Slug: "organayzery-dlya-kosmetiki", ParentID: 3018,
							},
							{
								ID: 3094, Name: "Бігуді", Slug: "bigudi", ParentID: 3018,
							},
							{
								ID: 3100, Name: "Мочалки для душу", Slug: "mochalki-dlya-dusha", ParentID: 3018,
							},
							{
								ID: 3105, Name: "Щітки для сухого масажу", Slug: "shchetki-dlya-suhogo-massazha", ParentID: 3018,
							},
							{
								ID: 3110, Name: "Інше", Slug: "drugoe", ParentID: 3018,
							},
						},
					},
				},
			},
			{
				ID: 97, Name: "Для весілля", Slug: "dlya-svadby", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2584, Name: "Весільні сукні", Slug: "platya", ParentID: 97,
					},
					{
						ID: 2585, Name: "Весільні костюми", Slug: "kostyumy", ParentID: 97,
					},
					{
						ID: 2586, Name: "Сукні для подружок нареченої", Slug: "plate-podruzhki-nevesty", ParentID: 97,
					},
					{
						ID: 2587, Name: "Весільні аксесуари для нареченої", Slug: "aksessuary", ParentID: 97,
						Children: []CategoryItem{
							{
								ID: 2590, Name: "Підв'язки для нареченої", Slug: "podvyazki-dlya-nevesty", ParentID: 2587,
							},
							{
								ID: 2592, Name: "Шубки", Slug: "shubki", ParentID: 2587,
							},
							{
								ID: 2594, Name: "Накидки", Slug: "nakidki", ParentID: 2587,
							},
							{
								ID: 2596, Name: "Болеро", Slug: "bolero", ParentID: 2587,
							},
							{
								ID: 2598, Name: "Під'юбники", Slug: "podyubniki", ParentID: 2587,
							},
							{
								ID: 2600, Name: "Віяла і сумочки", Slug: "veera-sumochki", ParentID: 2587,
							},
							{
								ID: 2602, Name: "Парасолі", Slug: "zonty", ParentID: 2587,
							},
							{
								ID: 2604, Name: "Фата", Slug: "fata", ParentID: 2587,
							},
							{
								ID: 2606, Name: "Весільні рукавички і пояси", Slug: "perchatki-poyasa", ParentID: 2587,
							},
							{
								ID: 2608, Name: "Весільні ювелірні вироби та прикраси", Slug: "yuvelirnye-izdeliya-ukrasheniya", ParentID: 2587,
							},
						},
					},
					{
						ID: 2588, Name: "Весільна білизна", Slug: "svadebnoe-bele", ParentID: 97,
					},
					{
						ID: 2589, Name: "Весільні аксесуари (декор)", Slug: "dekor", ParentID: 97,
						Children: []CategoryItem{
							{
								ID: 2591, Name: "Бутоньєрки", Slug: "butonerki", ParentID: 2589,
							},
							{
								ID: 2593, Name: "Подушки для обручок", Slug: "podushki-dlya-kolets", ParentID: 2589,
							},
							{
								ID: 2595, Name: "Прикраси на машину", Slug: "ukrasheniya-na-mashinu", ParentID: 2589,
							},
							{
								ID: 2597, Name: "Весільна арка", Slug: "svadebnaya-arka", ParentID: 2589,
							},
							{
								ID: 2599, Name: "Весільні келихи", Slug: "svadebnye-bokaly", ParentID: 2589,
							},
							{
								ID: 2601, Name: "Скриньки для грошей", Slug: "shkatulki-dlya-deneg", ParentID: 2589,
							},
							{
								ID: 2603, Name: "Весільні букети", Slug: "svadebnye-bukety", ParentID: 2589,
							},
							{
								ID: 2605, Name: "Запрошення на весілля", Slug: "priglasheniya-na-svadbu", ParentID: 2589,
							},
							{
								ID: 2607, Name: "Весільний рушник", Slug: "svadebniy-rushnik", ParentID: 2589,
							},
						},
					},
					{
						ID: 2609, Name: "Весільне взуття", Slug: "obuv", ParentID: 97,
						Children: []CategoryItem{
							{
								ID: 2611, Name: "Туфлі", Slug: "tufli", ParentID: 2609,
							},
							{
								ID: 2612, Name: "Босоніжки", Slug: "bosonozhki", ParentID: 2609,
							},
							{
								ID: 2613, Name: "Балетки", Slug: "baletki", ParentID: 2609,
							},
							{
								ID: 2614, Name: "Ботильйони", Slug: "botilony", ParentID: 2609,
							},
							{
								ID: 2615, Name: "Черевики і кросівки", Slug: "botinki-krossovki", ParentID: 2609,
							},
						},
					},
					{
						ID: 2610, Name: "Інше", Slug: "drugoe", ParentID: 97,
					},
				},
			},
			{
				ID: 98, Name: "Аксесуари", Slug: "aksessuary", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 552, Name: "Сумки", Slug: "sumki", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3153, Name: "Аксесуари для сумок", Slug: "aksessuary", ParentID: 552,
							},
							{
								ID: 3163, Name: "Шкіряні сумки", Slug: "kozhanye", ParentID: 552,
							},
							{
								ID: 3173, Name: "Клатчі", Slug: "klatchi", ParentID: 552,
							},
							{
								ID: 3181, Name: "Пляжні сумки", Slug: "plyazhnye", ParentID: 552,
							},
							{
								ID: 3188, Name: "Поясні сумки", Slug: "poyasnye", ParentID: 552,
							},
							{
								ID: 3195, Name: "Спортивні сумки", Slug: "sportivnye", ParentID: 552,
							},
							{
								ID: 3199, Name: "Сумки через плече", Slug: "cherez-plecho", ParentID: 552,
							},
							{
								ID: 3202, Name: "Портфелі", Slug: "portfeli", ParentID: 552,
							},
							{
								ID: 3204, Name: "Шоппери", Slug: "shoppery", ParentID: 552,
							},
							{
								ID: 3206, Name: "Барсетки", Slug: "barsetki", ParentID: 552,
							},
							{
								ID: 3208, Name: "Валізи", Slug: "chemodany", ParentID: 552,
							},
							{
								ID: 3210, Name: "Дорожні сумки", Slug: "dorozhnye", ParentID: 552,
							},
							{
								ID: 3213, Name: "Інше", Slug: "drugoe", ParentID: 552,
							},
						},
					},
					{
						ID: 553, Name: "Біжутерія", Slug: "bizhuteriya", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3154, Name: "Каблучки", Slug: "koltsa", ParentID: 553,
							},
							{
								ID: 3164, Name: "Сережки", Slug: "sergi", ParentID: 553,
							},
							{
								ID: 3174, Name: "Браслети", Slug: "braslety", ParentID: 553,
							},
							{
								ID: 3182, Name: "Кулони / підвіски", Slug: "kulony-podveski", ParentID: 553,
							},
							{
								ID: 3189, Name: "Брошки", Slug: "broshi", ParentID: 553,
							},
							{
								ID: 3196, Name: "Ланцюжки", Slug: "tsepochki", ParentID: 553,
							},
							{
								ID: 3200, Name: "Чокери", Slug: "choker", ParentID: 553,
							},
							{
								ID: 3203, Name: "Запонки", Slug: "zaponki", ParentID: 553,
							},
							{
								ID: 3205, Name: "Кліпси", Slug: "klipsy", ParentID: 553,
							},
							{
								ID: 3207, Name: "Шпильки", Slug: "bulavki", ParentID: 553,
							},
							{
								ID: 3209, Name: "Бісер / фурнітура", Slug: "biser-furnitura", ParentID: 553,
							},
							{
								ID: 3211, Name: "Кольє", Slug: "kole", ParentID: 553,
							},
							{
								ID: 3212, Name: "Набори біжутерії", Slug: "nabory-bizhuterii", ParentID: 553,
							},
							{
								ID: 3214, Name: "Інше", Slug: "drugoe", ParentID: 553,
							},
						},
					},
					{
						ID: 557, Name: "Ювелірні вироби", Slug: "yuvelirnye-izdeliya", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3242, Name: "Каблучки", Slug: "koltsa", ParentID: 557,
							},
							{
								ID: 3243, Name: "Сережки", Slug: "sergi", ParentID: 557,
							},
							{
								ID: 3244, Name: "Браслети", Slug: "braslety", ParentID: 557,
							},
							{
								ID: 3245, Name: "Підвіски", Slug: "podveski", ParentID: 557,
							},
							{
								ID: 3246, Name: "Брошки", Slug: "broshi", ParentID: 557,
							},
							{
								ID: 3247, Name: "Ланцюжки", Slug: "tsepochki", ParentID: 557,
							},
							{
								ID: 3248, Name: "Прикраси для пірсингу", Slug: "ukrasheniya-dlya-pirsinga", ParentID: 557,
							},
							{
								ID: 3249, Name: "Кольє", Slug: "kole", ParentID: 557,
							},
							{
								ID: 3250, Name: "Шпильки", Slug: "bulavki", ParentID: 557,
							},
							{
								ID: 3251, Name: "Запонки", Slug: "zaponki", ParentID: 557,
							},
							{
								ID: 3252, Name: "Шарми", Slug: "sharmy", ParentID: 557,
							},
							{
								ID: 3253, Name: "Набори", Slug: "nabory", ParentID: 557,
							},
							{
								ID: 3254, Name: "Інше", Slug: "drugoe", ParentID: 557,
							},
						},
					},
					{
						ID: 893, Name: "Інші аксесуари", Slug: "drugie-aksessuary", ParentID: 98,
					},
					{
						ID: 3141, Name: "Скриньки для прикрас", Slug: "shkatulki-dlya-ukrasheniy", ParentID: 98,
					},
					{
						ID: 3142, Name: "Рюкзаки міські та спортивні", Slug: "ryukzaki", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3155, Name: "Рюкзаки для ноутбука", Slug: "dlya-noutbuka", ParentID: 3142,
							},
							{
								ID: 3165, Name: "Рюкзаки для подорожей", Slug: "dlya-puteshestviy", ParentID: 3142,
							},
							{
								ID: 3175, Name: "Міські рюкзаки", Slug: "gorodskie", ParentID: 3142,
							},
							{
								ID: 3215, Name: "Інше", Slug: "drugoe", ParentID: 3142,
							},
						},
					},
					{
						ID: 3143, Name: "Гаманці, косметички, візитниці", Slug: "koshelki-kosmetichki-vizitnitsy", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3156, Name: "Візитниці та картхолдери", Slug: "vizitnitsy-kartholdery", ParentID: 3143,
							},
							{
								ID: 3166, Name: "Гаманці", Slug: "koshelki", ParentID: 3143,
							},
							{
								ID: 3176, Name: "Косметички", Slug: "kosmetichki", ParentID: 3143,
							},
							{
								ID: 3183, Name: "Чохли для ключів", Slug: "chehly-dlya-klyuchey", ParentID: 3143,
							},
							{
								ID: 3190, Name: "Затискачі для грошей", Slug: "zazhimy-dlya-deneg", ParentID: 3143,
							},
							{
								ID: 3216, Name: "Інше", Slug: "drugoe", ParentID: 3143,
							},
						},
					},
					{
						ID: 3144, Name: "Парасолі", Slug: "zonty", ParentID: 98,
					},
					{
						ID: 3145, Name: "Аксесуари для волосся", Slug: "dlya-volos", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3157, Name: "Пов'язки", Slug: "povyazki", ParentID: 3145,
							},
							{
								ID: 3167, Name: "Гумки", Slug: "rezinki", ParentID: 3145,
							},
							{
								ID: 3177, Name: "Заколки", Slug: "zakolki", ParentID: 3145,
							},
							{
								ID: 3184, Name: "Обручі", Slug: "obruchi", ParentID: 3145,
							},
							{
								ID: 3191, Name: "Шпильки", Slug: "shpilki", ParentID: 3145,
							},
							{
								ID: 3217, Name: "Інше", Slug: "drugoe", ParentID: 3145,
							},
						},
					},
					{
						ID: 3146, Name: "Ремені та пояси", Slug: "remni-poyasa", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3158, Name: "Ремені", Slug: "remni", ParentID: 3146,
							},
							{
								ID: 3168, Name: "Підтяжки", Slug: "podtyazhki", ParentID: 3146,
							},
							{
								ID: 3178, Name: "Пояси", Slug: "poyasa", ParentID: 3146,
							},
							{
								ID: 3185, Name: "Шнурки", Slug: "shnurki", ParentID: 3146,
							},
							{
								ID: 3192, Name: "Портупеї", Slug: "portupei", ParentID: 3146,
							},
							{
								ID: 3218, Name: "Інше", Slug: "drugoe", ParentID: 3146,
							},
						},
					},
					{
						ID: 3147, Name: "Шарфи і рукавички", Slug: "sharfy-perchatki", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3159, Name: "Носові хустинки", Slug: "nosovye-platki", ParentID: 3147,
							},
							{
								ID: 3169, Name: "Палантини і шалі", Slug: "palantiny-shali", ParentID: 3147,
							},
							{
								ID: 3179, Name: "Хустки", Slug: "platki", ParentID: 3147,
							},
							{
								ID: 3186, Name: "Снуди", Slug: "snudy", ParentID: 3147,
							},
							{
								ID: 3193, Name: "Шарфи", Slug: "sharfy", ParentID: 3147,
							},
							{
								ID: 3197, Name: "Рукавички", Slug: "perchatki", ParentID: 3147,
							},
							{
								ID: 3201, Name: "Рукавиці", Slug: "varezhki", ParentID: 3147,
							},
							{
								ID: 3219, Name: "Інше", Slug: "drugoe", ParentID: 3147,
							},
						},
					},
					{
						ID: 3148, Name: "Окуляри", Slug: "ochki", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3160, Name: "Оправи", Slug: "opravy", ParentID: 3148,
							},
							{
								ID: 3170, Name: "Сонцезахисні окуляри", Slug: "solntsezashchitnye", ParentID: 3148,
							},
							{
								ID: 3180, Name: "Окуляри для зору", Slug: "dlya-zreniya", ParentID: 3148,
							},
							{
								ID: 3187, Name: "Комп'ютерні окуляри", Slug: "kompyuternye", ParentID: 3148,
							},
							{
								ID: 3194, Name: "Контактні лінзи", Slug: "kontaktnye-linzy", ParentID: 3148,
							},
							{
								ID: 3198, Name: "Чохли для окулярів", Slug: "chehly", ParentID: 3148,
							},
							{
								ID: 3220, Name: "Інше", Slug: "drugoe", ParentID: 3148,
							},
							{
								ID: 3222, Name: "Аксесуари для окулярів і лінз", Slug: "aksessuary", ParentID: 3148,
							},
						},
					},
					{
						ID: 3149, Name: "Краватки і метелики", Slug: "galstuki-babochki", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3161, Name: "Краватки", Slug: "galstuki", ParentID: 3149,
							},
							{
								ID: 3171, Name: "Метелики", Slug: "babochki", ParentID: 3149,
							},
						},
					},
					{
						ID: 3150, Name: "Аксесуари для взуття", Slug: "dlya-obuvi", ParentID: 98,
						Children: []CategoryItem{
							{
								ID: 3162, Name: "Устілки для взуття", Slug: "stelki", ParentID: 3150,
							},
							{
								ID: 3172, Name: "Креми, щітки та інші аксесуари", Slug: "kremy-shchetki-aksessuary", ParentID: 3150,
							},
							{
								ID: 3221, Name: "Шнурки для взуття", Slug: "shnurki", ParentID: 3150,
							},
						},
					},
					{
						ID: 3151, Name: "Ремінці та браслети для годинника", Slug: "remeshki-braslety", ParentID: 98,
					},
					{
						ID: 3152, Name: "Маски для сну", Slug: "maski-dlya-sna", ParentID: 98,
					},
				},
			},
			{
				ID: 556, Name: "Наручні годинники", Slug: "naruchnye-chasy", ParentID: 891,
			},
			{
				ID: 897, Name: "Подарунки", Slug: "podarki", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2338, Name: "Подарунки для чоловіків", Slug: "dlya-muzhchin", ParentID: 897,
					},
					{
						ID: 2339, Name: "Подарунки для жінок", Slug: "dlya-zhenshchin", ParentID: 897,
					},
					{
						ID: 2340, Name: "Подарунки для дітей", Slug: "dlya-detey", ParentID: 897,
					},
					{
						ID: 2341, Name: "Подарунки на день народження", Slug: "den-rozhdeniya", ParentID: 897,
					},
					{
						ID: 2342, Name: "Подарунки на весілля", Slug: "svadba", ParentID: 897,
					},
					{
						ID: 2343, Name: "Подарунки на 8 березня", Slug: "8-marta", ParentID: 897,
					},
					{
						ID: 2344, Name: "Подарунки на Новий рік", Slug: "noviy-god", ParentID: 897,
					},
					{
						ID: 2345, Name: "Подарунки колегам", Slug: "kollegam", ParentID: 897,
					},
					{
						ID: 2346, Name: "Смачні подарунки", Slug: "vkusnye", ParentID: 897,
					},
					{
						ID: 2347, Name: "Подарункові коробки та пакування", Slug: "korobki-upakovki", ParentID: 897,
					},
					{
						ID: 2348, Name: "Подарунки ручної роботи", Slug: "ruchnoy-raboty", ParentID: 897,
						Children: []CategoryItem{
							{
								ID: 2352, Name: "Мило ручної роботи", Slug: "mylo", ParentID: 2348,
							},
							{
								ID: 2353, Name: "Свічки ручної роботи", Slug: "svechi", ParentID: 2348,
							},
							{
								ID: 2354, Name: "Іграшки ручної роботи", Slug: "igrushki", ParentID: 2348,
							},
							{
								ID: 2355, Name: "Інші подарунки ручної роботи", Slug: "drugoe", ParentID: 2348,
							},
						},
					},
					{
						ID: 2349, Name: "Листівки та конверти", Slug: "otkrytki-konverty", ParentID: 897,
					},
					{
						ID: 2350, Name: "Букети", Slug: "bukety", ParentID: 897,
					},
					{
						ID: 2351, Name: "Повітряні кульки та композиції з них", Slug: "vozdushnye-shary-i-kompozicii", ParentID: 897,
					},
					{
						ID: 2356, Name: "Інші подарунки", Slug: "drugoe", ParentID: 897,
					},
				},
			},
			{
				ID: 1427, Name: "Мода різне", Slug: "moda-raznoe", ParentID: 891,
			},
			{
				ID: 2010, Name: "Чоловіча білизна та плавки", Slug: "muzhskoe-bele-plavki", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2110, Name: "Труси", Slug: "trusy", ParentID: 2010,
					},
					{
						ID: 2111, Name: "Термобілизна", Slug: "termobele", ParentID: 2010,
					},
					{
						ID: 2112, Name: "Шкарпетки", Slug: "noski", ParentID: 2010,
					},
					{
						ID: 2113, Name: "Плавки", Slug: "plavki", ParentID: 2010,
					},
					{
						ID: 2293, Name: "Інша чоловіча білизна", Slug: "drugoe", ParentID: 2010,
					},
				},
			},
			{
				ID: 2011, Name: "Жіноча білизна та купальники", Slug: "zhenskoe-bele-kupalniki", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2296, Name: "Пляжні сукні та туніки", Slug: "plyazhnye-platya-tuniki", ParentID: 2011,
					},
					{
						ID: 2297, Name: "Парео, накидки", Slug: "pareo-nakidki", ParentID: 2011,
					},
					{
						ID: 2298, Name: "Пляжні шорти", Slug: "plyazhnye-shorty", ParentID: 2011,
					},
					{
						ID: 2299, Name: "Купальники", Slug: "kupalniki", ParentID: 2011,
						Children: []CategoryItem{
							{
								ID: 2311, Name: "Роздільні купальники", Slug: "razdelnye", ParentID: 2299,
							},
							{
								ID: 2315, Name: "Суцільні купальники", Slug: "slitnye", ParentID: 2299,
							},
							{
								ID: 2319, Name: "Спортивні купальники", Slug: "sportivnye", ParentID: 2299,
							},
							{
								ID: 2323, Name: "Інші купальники", Slug: "drugoe", ParentID: 2299,
							},
						},
					},
					{
						ID: 2300, Name: "Труси", Slug: "trusy", ParentID: 2011,
					},
					{
						ID: 2301, Name: "Бюстгалтери", Slug: "byustgaltery", ParentID: 2011,
						Children: []CategoryItem{
							{
								ID: 2312, Name: "Балконет", Slug: "balkonet", ParentID: 2301,
							},
							{
								ID: 2316, Name: "Бандо (без бретелів)", Slug: "bando-bez-breteley", ParentID: 2301,
							},
							{
								ID: 2320, Name: "Бралетт (без кісточок)", Slug: "bralett-bez-kostochek", ParentID: 2301,
							},
							{
								ID: 2324, Name: "Бюстгалтери з Push up", Slug: "push-up", ParentID: 2301,
							},
							{
								ID: 2327, Name: "Безшовні бюстгалтери", Slug: "besshovnye", ParentID: 2301,
							},
							{
								ID: 2329, Name: "М'які бюстгалтери", Slug: "myagkie", ParentID: 2301,
							},
							{
								ID: 2331, Name: "Бюстгалтери топи", Slug: "topy", ParentID: 2301,
							},
							{
								ID: 2334, Name: "Інші бюстгалтери", Slug: "drugoe", ParentID: 2301,
							},
						},
					},
					{
						ID: 2302, Name: "Комплекти спідньої білизни", Slug: "komplekty-nizhnego-belya", ParentID: 2011,
					},
					{
						ID: 2303, Name: "Корегуюча білизна", Slug: "korrektiruyushchee-bele", ParentID: 2011,
						Children: []CategoryItem{
							{
								ID: 2313, Name: "Корегуючі шорти і панталони", Slug: "shorty-pantalony", ParentID: 2303,
							},
							{
								ID: 2317, Name: "Корегуючі пояси", Slug: "poyasa", ParentID: 2303,
							},
							{
								ID: 2321, Name: "Корегуючі корсети", Slug: "korsety", ParentID: 2303,
							},
							{
								ID: 2325, Name: "Корегуючі боді", Slug: "bodi", ParentID: 2303,
							},
							{
								ID: 2333, Name: "Корегуючі труси", Slug: "trusy", ParentID: 2303,
							},
							{
								ID: 2335, Name: "Інша корегуюча білизна", Slug: "drugoe", ParentID: 2303,
							},
						},
					},
					{
						ID: 2304, Name: "Термобілизна", Slug: "termobele", ParentID: 2011,
					},
					{
						ID: 2305, Name: "Еротична білизна", Slug: "eroticheskoe-bele", ParentID: 2011,
					},
					{
						ID: 2306, Name: "Комбінації", Slug: "kombinatsii", ParentID: 2011,
					},
					{
						ID: 2307, Name: "Топи бюстьє", Slug: "topy-byuste", ParentID: 2011,
					},
					{
						ID: 2308, Name: "Пояси і підв'язки для панчох", Slug: "poyasa-podvyazki", ParentID: 2011,
					},
					{
						ID: 2309, Name: "Комбідреси", Slug: "kombidressy", ParentID: 2011,
					},
					{
						ID: 2310, Name: "Панчохи / шкарпетки / колготки", Slug: "chulki-noski-kolgotki", ParentID: 2011,
						Children: []CategoryItem{
							{
								ID: 2314, Name: "Шкарпетки", Slug: "noski", ParentID: 2310,
							},
							{
								ID: 2318, Name: "Панчохи", Slug: "chulki", ParentID: 2310,
							},
							{
								ID: 2322, Name: "Колготки", Slug: "kolgotki", ParentID: 2310,
							},
							{
								ID: 2326, Name: "Гетри та гольфи", Slug: "getry-golfy", ParentID: 2310,
							},
							{
								ID: 2328, Name: "Сліди", Slug: "sledy", ParentID: 2310,
							},
							{
								ID: 2330, Name: "Компресійні панчохи", Slug: "kompressionnye-chulki", ParentID: 2310,
							},
							{
								ID: 2336, Name: "Інше", Slug: "drugoe", ParentID: 2310,
							},
						},
					},
					{
						ID: 2332, Name: "Корсети", Slug: "korsety", ParentID: 2011,
					},
					{
						ID: 2337, Name: "Інша жіноча білизна", Slug: "drugoe", ParentID: 2011,
					},
				},
			},
			{
				ID: 2012, Name: "Одяг для вагітних", Slug: "odezhda-dlya-beremennyh", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2114, Name: "Футболки та майки для вагітних та годуючих", Slug: "futbolki-mayki", ParentID: 2012,
					},
					{
						ID: 2115, Name: "Блузки та туніки для вагітних та годуючих", Slug: "bluzki-tuniki", ParentID: 2012,
					},
					{
						ID: 2116, Name: "Плаття для вагітних та годуючих", Slug: "platya", ParentID: 2012,
					},
					{
						ID: 2117, Name: "Спідниці для вагітних", Slug: "yubki", ParentID: 2012,
					},
					{
						ID: 2118, Name: "Купальники для вагітних", Slug: "kupalniki", ParentID: 2012,
					},
					{
						ID: 2119, Name: "Нижня білизна для вагітних та годуючих", Slug: "nizhnee-bele", ParentID: 2012,
					},
					{
						ID: 2120, Name: "Шорти для вагітних", Slug: "shorty", ParentID: 2012,
					},
					{
						ID: 2121, Name: "Комбінезони для вагітних та годуючих", Slug: "kombinezony", ParentID: 2012,
					},
					{
						ID: 2122, Name: "Штани і легінси для вагітних", Slug: "shtany-legginsy", ParentID: 2012,
					},
					{
						ID: 2123, Name: "Джинси для вагітних", Slug: "dzhinsy", ParentID: 2012,
					},
					{
						ID: 2124, Name: "Колготки для вагітних", Slug: "kolgotki", ParentID: 2012,
					},
					{
						ID: 2125, Name: "Верхній одяг для вагітних", Slug: "verhnyaya-odezhda", ParentID: 2012,
					},
					{
						ID: 2126, Name: "Одяг для сну і дому для вагітних та годуючих", Slug: "odezhda-dlya-sna-doma", ParentID: 2012,
					},
					{
						ID: 2127, Name: "Светри і толстовки для вагітних та годуючих", Slug: "svitery-tolstovki", ParentID: 2012,
					},
					{
						ID: 2128, Name: "Костюми для вагітних та годуючих", Slug: "kostyumy", ParentID: 2012,
					},
					{
						ID: 2129, Name: "Сарафани для вагітних та годуючих", Slug: "sarafany", ParentID: 2012,
					},
					{
						ID: 2130, Name: "Пояси для вагітних та годуючих", Slug: "poyasa", ParentID: 2012,
					},
					{
						ID: 2294, Name: "Інший одяг для вагітних", Slug: "drugoe", ParentID: 2012,
					},
				},
			},
			{
				ID: 2013, Name: "Головні убори", Slug: "golovnye-ubory", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2131, Name: "Бейсболки та кепки", Slug: "beysbolki-kepki", ParentID: 2013,
					},
					{
						ID: 2132, Name: "Балаклави", Slug: "balaklavy", ParentID: 2013,
					},
					{
						ID: 2135, Name: "Комплекти", Slug: "komplekty", ParentID: 2013,
					},
					{
						ID: 2138, Name: "Панами", Slug: "panamy", ParentID: 2013,
					},
					{
						ID: 2141, Name: "Шапки", Slug: "shapki", ParentID: 2013,
					},
					{
						ID: 2144, Name: "Капелюхи", Slug: "shlyapy", ParentID: 2013,
					},
					{
						ID: 2147, Name: "Берети", Slug: "berety", ParentID: 2013,
					},
					{
						ID: 2150, Name: "Бандани, бафи", Slug: "bandany-baffy", ParentID: 2013,
					},
					{
						ID: 2151, Name: "Пов'язки та хутряні навушники", Slug: "povyazki-teplye-naushniki", ParentID: 2013,
					},
					{
						ID: 2295, Name: "Інші головні убори", Slug: "drugoe", ParentID: 2013,
					},
				},
			},
			{
				ID: 2014, Name: "Чоловіче взуття", Slug: "muzhskaya-obuv", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2687, Name: "Кросівки", Slug: "krossovki", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2696, Name: "Кросівки для бігу", Slug: "beg", ParentID: 2687,
							},
							{
								ID: 2702, Name: "Кросівки для міста", Slug: "gorod", ParentID: 2687,
							},
							{
								ID: 2708, Name: "Кросівки для залу і тренувань", Slug: "zal-trenirovki", ParentID: 2687,
							},
							{
								ID: 2714, Name: "Трекінгові кросівки", Slug: "trekingovye", ParentID: 2687,
							},
							{
								ID: 2719, Name: "Інші кросівки", Slug: "drugoe", ParentID: 2687,
							},
						},
					},
					{
						ID: 2688, Name: "Кеди", Slug: "kedy", ParentID: 2014,
					},
					{
						ID: 2689, Name: "Черевики", Slug: "botinki", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2697, Name: "Челсі", Slug: "chelsi", ParentID: 2689,
							},
							{
								ID: 2703, Name: "Берці", Slug: "bertsy", ParentID: 2689,
							},
							{
								ID: 2709, Name: "Напівчеревики", Slug: "polubotinki", ParentID: 2689,
							},
							{
								ID: 2721, Name: "Трекінгові черевики", Slug: "trekingovye", ParentID: 2689,
							},
							{
								ID: 2722, Name: "Інші черевики", Slug: "drugoe", ParentID: 2689,
							},
						},
					},
					{
						ID: 2690, Name: "Сандалі і шльопанці", Slug: "sandali-shlepantsy", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2698, Name: "Сандалі", Slug: "sandali", ParentID: 2690,
							},
							{
								ID: 2704, Name: "Шльопанці", Slug: "shlepantsy", ParentID: 2690,
							},
							{
								ID: 2710, Name: "В'єтнамки", Slug: "vetnamki", ParentID: 2690,
							},
							{
								ID: 2715, Name: "Аквашузи (коралоходи)", Slug: "akvashuzy", ParentID: 2690,
							},
							{
								ID: 2723, Name: "Інше", Slug: "drugoe", ParentID: 2690,
							},
						},
					},
					{
						ID: 2691, Name: "Чоботи", Slug: "sapogi", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2699, Name: "Мунбути", Slug: "lunohody", ParentID: 2691,
							},
							{
								ID: 2705, Name: "Уггі та унти чоловічі", Slug: "uggi-unty", ParentID: 2691,
							},
							{
								ID: 2711, Name: "Дутики", Slug: "dutiki", ParentID: 2691,
							},
							{
								ID: 2716, Name: "Гумові чоботи", Slug: "rezinovye-sapogi", ParentID: 2691,
							},
							{
								ID: 2724, Name: "Інші чоботи", Slug: "drugoe", ParentID: 2691,
							},
						},
					},
					{
						ID: 2692, Name: "Домашні капці", Slug: "domashnie-tapochki", ParentID: 2014,
					},
					{
						ID: 2693, Name: "Мокасини та сліпони", Slug: "mokasiny-slipony", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2700, Name: "Топсайдери", Slug: "topsaydery", ParentID: 2693,
							},
							{
								ID: 2706, Name: "Мокасини", Slug: "mokasiny", ParentID: 2693,
							},
							{
								ID: 2712, Name: "Сліпони", Slug: "slipony", ParentID: 2693,
							},
							{
								ID: 2717, Name: "Еспадрилі", Slug: "espadrili", ParentID: 2693,
							},
							{
								ID: 2725, Name: "Інше", Slug: "drugoe", ParentID: 2693,
							},
						},
					},
					{
						ID: 2694, Name: "Ортопедичне взуття", Slug: "ortopedicheskaya-obuv", ParentID: 2014,
					},
					{
						ID: 2695, Name: "Туфлі", Slug: "tufli", ParentID: 2014,
						Children: []CategoryItem{
							{
								ID: 2701, Name: "Монки", Slug: "monki", ParentID: 2695,
							},
							{
								ID: 2707, Name: "Оксфорди", Slug: "oksfordy", ParentID: 2695,
							},
							{
								ID: 2713, Name: "Дербі", Slug: "derbi", ParentID: 2695,
							},
							{
								ID: 2718, Name: "Лофери", Slug: "lofery", ParentID: 2695,
							},
							{
								ID: 2720, Name: "Броги", Slug: "brogi", ParentID: 2695,
							},
							{
								ID: 2726, Name: "Інші туфлі", Slug: "drugoe", ParentID: 2695,
							},
						},
					},
					{
						ID: 2727, Name: "Інше чоловіче взуття", Slug: "drugoe", ParentID: 2014,
					},
				},
			},
			{
				ID: 2015, Name: "Жіноче взуття", Slug: "zhenskaya-obuv", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2618, Name: "Черевики та напівчеревики", Slug: "botinki-polubotinki", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2629, Name: "Ботильйони", Slug: "botilony", ParentID: 2618,
							},
							{
								ID: 2635, Name: "Челсі", Slug: "chelsi", ParentID: 2618,
							},
							{
								ID: 2641, Name: "Хайтопи", Slug: "haytopy", ParentID: 2618,
							},
							{
								ID: 2647, Name: "Берці", Slug: "bertsy", ParentID: 2618,
							},
							{
								ID: 2653, Name: "Броги", Slug: "brogi", ParentID: 2618,
							},
							{
								ID: 2658, Name: "Хайкери", Slug: "haykery", ParentID: 2618,
							},
							{
								ID: 2661, Name: "Дезерти", Slug: "dezerty", ParentID: 2618,
							},
							{
								ID: 2663, Name: "Козаки", Slug: "kazaki", ParentID: 2618,
							},
							{
								ID: 2672, Name: "Інші черевики та напівчеревики", Slug: "drugoe", ParentID: 2618,
							},
						},
					},
					{
						ID: 2619, Name: "Туфлі", Slug: "tufli", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2630, Name: "Балетки", Slug: "baletki", ParentID: 2619,
							},
							{
								ID: 2636, Name: "Дербі", Slug: "derbi", ParentID: 2619,
							},
							{
								ID: 2642, Name: "Лофери", Slug: "lofery", ParentID: 2619,
							},
							{
								ID: 2648, Name: "Туфлі-човники", Slug: "lodochki", ParentID: 2619,
							},
							{
								ID: 2654, Name: "Оксфорди", Slug: "oksfordy", ParentID: 2619,
							},
							{
								ID: 2671, Name: "Інші туфлі", Slug: "drugoe", ParentID: 2619,
							},
						},
					},
					{
						ID: 2620, Name: "Чоботи, напівчоботи", Slug: "sapogi-polusapozhki", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2631, Name: "Дутики", Slug: "dutiki", ParentID: 2620,
							},
							{
								ID: 2637, Name: "Ботфорти", Slug: "botforty", ParentID: 2620,
							},
							{
								ID: 2643, Name: "Місяцеходи", Slug: "lunohody", ParentID: 2620,
							},
							{
								ID: 2649, Name: "Гумові чоботи", Slug: "rezinovye-sapogi", ParentID: 2620,
							},
							{
								ID: 2655, Name: "Унти", Slug: "unty", ParentID: 2620,
							},
							{
								ID: 2670, Name: "Інше", Slug: "drugoe", ParentID: 2620,
							},
						},
					},
					{
						ID: 2621, Name: "Уггі", Slug: "uggi", ParentID: 2015,
					},
					{
						ID: 2622, Name: "Мокасини та сліпони", Slug: "mokasiny-slipony", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2632, Name: "Мокасини", Slug: "mokasiny", ParentID: 2622,
							},
							{
								ID: 2638, Name: "Топсайдери", Slug: "topsaydery", ParentID: 2622,
							},
							{
								ID: 2644, Name: "Сліпони", Slug: "slipony", ParentID: 2622,
							},
							{
								ID: 2650, Name: "Еспадрилі", Slug: "espadrili", ParentID: 2622,
							},
							{
								ID: 2669, Name: "Інше", Slug: "drugoe", ParentID: 2622,
							},
						},
					},
					{
						ID: 2623, Name: "Кросівки", Slug: "krossovki", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2633, Name: "Снікерси", Slug: "snikersy", ParentID: 2623,
							},
							{
								ID: 2645, Name: "Кросівки-шкарпетки", Slug: "noski", ParentID: 2623,
							},
							{
								ID: 2651, Name: "Кросівки для бігу", Slug: "beg", ParentID: 2623,
							},
							{
								ID: 2656, Name: "Кросівки для фітнесу та тренувань", Slug: "fitnes-trenirovki", ParentID: 2623,
							},
							{
								ID: 2659, Name: "Трекінгові кросівки", Slug: "trekingovye", ParentID: 2623,
							},
							{
								ID: 2666, Name: "Кросівки-хайтопи", Slug: "haytopy", ParentID: 2623,
							},
							{
								ID: 2667, Name: "Інші кросівки", Slug: "drugoe", ParentID: 2623,
							},
						},
					},
					{
						ID: 2624, Name: "Кеди", Slug: "kedy", ParentID: 2015,
					},
					{
						ID: 2625, Name: "Домашні капці", Slug: "domashnie-tapochki", ParentID: 2015,
					},
					{
						ID: 2626, Name: "Ортопедичне взуття", Slug: "ortopedicheskaya-obuv", ParentID: 2015,
					},
					{
						ID: 2627, Name: "Спортивне взуття", Slug: "sportivnaya-obuv", ParentID: 2015,
					},
					{
						ID: 2628, Name: "Босоніжки, сандалі, в'єтнамки", Slug: "bosonozhki-sandali-vetnamki", ParentID: 2015,
						Children: []CategoryItem{
							{
								ID: 2634, Name: "Аквашузи (коралоходи)", Slug: "akvashuzy", ParentID: 2628,
							},
							{
								ID: 2640, Name: "Босоніжки", Slug: "bosonozhki", ParentID: 2628,
							},
							{
								ID: 2646, Name: "Сандалі", Slug: "sandali", ParentID: 2628,
							},
							{
								ID: 2652, Name: "Шльопанці", Slug: "shlepantsy", ParentID: 2628,
							},
							{
								ID: 2657, Name: "Сланці", Slug: "slantsy", ParentID: 2628,
							},
							{
								ID: 2660, Name: "В'єтнамки", Slug: "vetnamki", ParentID: 2628,
							},
							{
								ID: 2662, Name: "Сабо", Slug: "sabo", ParentID: 2628,
							},
							{
								ID: 2664, Name: "Мюлі", Slug: "myuli", ParentID: 2628,
							},
							{
								ID: 2668, Name: "Інше", Slug: "drugoe", ParentID: 2628,
							},
						},
					},
					{
						ID: 2665, Name: "Інше жіноче взуття", Slug: "drugoe", ParentID: 2015,
					},
				},
			},
			{
				ID: 2016, Name: "Жіночий одяг", Slug: "zhenskaya-odezhda", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2837, Name: "Боді", Slug: "bodi", ParentID: 2016,
					},
					{
						ID: 2838, Name: "Майки та футболки", Slug: "mayki-futbolki", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2853, Name: "Футболки", Slug: "futbolki", ParentID: 2838,
							},
							{
								ID: 2867, Name: "Майки", Slug: "mayki", ParentID: 2838,
							},
							{
								ID: 2879, Name: "Топи", Slug: "topy", ParentID: 2838,
							},
							{
								ID: 2889, Name: "Поло", Slug: "polo", ParentID: 2838,
							},
							{
								ID: 2899, Name: "Лонгслів", Slug: "longsliv", ParentID: 2838,
							},
							{
								ID: 2990, Name: "Інше", Slug: "drugoe", ParentID: 2838,
							},
						},
					},
					{
						ID: 2839, Name: "Блузи і сорочки", Slug: "bluzy-rubashki", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2854, Name: "Блузи", Slug: "bluzy", ParentID: 2839,
							},
							{
								ID: 2868, Name: "Сорочки", Slug: "rubashki", ParentID: 2839,
							},
							{
								ID: 2991, Name: "Інше", Slug: "drugoe", ParentID: 2839,
							},
						},
					},
					{
						ID: 2840, Name: "Светри, кардигани, худі", Slug: "svitery-kardigany-hudi", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2855, Name: "Светри", Slug: "svitery", ParentID: 2840,
							},
							{
								ID: 2869, Name: "Кардигани", Slug: "kardigany", ParentID: 2840,
							},
							{
								ID: 2880, Name: "Худі", Slug: "hudi", ParentID: 2840,
							},
							{
								ID: 2890, Name: "Світшоти", Slug: "svitshoty", ParentID: 2840,
							},
							{
								ID: 2900, Name: "Толстовки", Slug: "tolstovki", ParentID: 2840,
							},
							{
								ID: 2908, Name: "Гольфи / водолазки", Slug: "golfy-vodolazki", ParentID: 2840,
							},
							{
								ID: 2915, Name: "Кофтинки", Slug: "koftochki", ParentID: 2840,
							},
							{
								ID: 2922, Name: "Джемпери", Slug: "dzhempery", ParentID: 2840,
							},
							{
								ID: 2929, Name: "Накидки", Slug: "nakidki", ParentID: 2840,
							},
							{
								ID: 2935, Name: "Пончо", Slug: "poncho", ParentID: 2840,
							},
							{
								ID: 2993, Name: "Інше", Slug: "drugoe", ParentID: 2840,
							},
						},
					},
					{
						ID: 2841, Name: "Плаття", Slug: "platya", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2856, Name: "Вечірні сукні", Slug: "vechernie", ParentID: 2841,
							},
							{
								ID: 2870, Name: "Туніки", Slug: "tuniki", ParentID: 2841,
							},
							{
								ID: 2881, Name: "Сарафани", Slug: "sarafany", ParentID: 2841,
							},
							{
								ID: 2891, Name: "Повсякденні сукні", Slug: "povsednevnye", ParentID: 2841,
							},
							{
								ID: 2901, Name: "Коктейльні сукні", Slug: "kokteylnye", ParentID: 2841,
							},
							{
								ID: 2909, Name: "Ділові сукні", Slug: "delovye", ParentID: 2841,
							},
							{
								ID: 2916, Name: "Випускні сукні", Slug: "vypusknye", ParentID: 2841,
							},
							{
								ID: 2923, Name: "Сукня-піджак", Slug: "plate-pidzhak", ParentID: 2841,
							},
							{
								ID: 2930, Name: "Сукні у білизняному стилі", Slug: "v-belevom-stile", ParentID: 2841,
							},
							{
								ID: 2992, Name: "Інші плаття", Slug: "drugoe", ParentID: 2841,
							},
						},
					},
					{
						ID: 2842, Name: "Спідниці", Slug: "yubki", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2857, Name: "Джинсові спідниці", Slug: "dzhinsovye", ParentID: 2842,
							},
							{
								ID: 2871, Name: "Спідниці олівець", Slug: "karandash", ParentID: 2842,
							},
							{
								ID: 2882, Name: "Шкіряні спідниці", Slug: "kozhanye", ParentID: 2842,
							},
							{
								ID: 2892, Name: "Спідниці довгі", Slug: "v-pol", ParentID: 2842,
							},
							{
								ID: 2902, Name: "Плісировані спідниці", Slug: "plissirovannye", ParentID: 2842,
							},
							{
								ID: 2910, Name: "Теплі спідниці", Slug: "teplye", ParentID: 2842,
							},
							{
								ID: 2917, Name: "Міні-спідниці", Slug: "mini", ParentID: 2842,
							},
							{
								ID: 2924, Name: "Замшеві спідниці", Slug: "zamshevye", ParentID: 2842,
							},
							{
								ID: 2931, Name: "Спідниці міді", Slug: "midi", ParentID: 2842,
							},
							{
								ID: 2936, Name: "Фатінові спідниці", Slug: "fatinovye", ParentID: 2842,
							},
							{
								ID: 2940, Name: "Спідниці-шорти", Slug: "shorty", ParentID: 2842,
							},
							{
								ID: 2943, Name: "Спідниці на запах", Slug: "na-zapah", ParentID: 2842,
							},
							{
								ID: 2945, Name: "Спідниці з екошкіри", Slug: "iz-ekokozhi", ParentID: 2842,
							},
							{
								ID: 2994, Name: "Інші спідниці", Slug: "drugoe", ParentID: 2842,
							},
						},
					},
					{
						ID: 2843, Name: "Верхній одяг", Slug: "verhnyaya-odezhda", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2858, Name: "Плащі й тренчі", Slug: "plashchi-trenchi", ParentID: 2843,
							},
							{
								ID: 2872, Name: "Дощовики", Slug: "dozhdeviki", ParentID: 2843,
							},
							{
								ID: 2883, Name: "Пальто", Slug: "palto", ParentID: 2843,
							},
							{
								ID: 2893, Name: "Вітровки", Slug: "vetrovki", ParentID: 2843,
							},
							{
								ID: 2903, Name: "Шкіряні куртки", Slug: "kozhanye-kurtki", ParentID: 2843,
							},
							{
								ID: 2911, Name: "Бомбери", Slug: "bombery", ParentID: 2843,
							},
							{
								ID: 2918, Name: "Джинсові куртки", Slug: "dzhinsovye-kurtki", ParentID: 2843,
							},
							{
								ID: 2925, Name: "Парки", Slug: "parki", ParentID: 2843,
							},
							{
								ID: 2932, Name: "Пуховики і зимові куртки", Slug: "puhoviki-zimnie-kurtki", ParentID: 2843,
							},
							{
								ID: 2937, Name: "Шуби, кожухи", Slug: "shuby-polushubki", ParentID: 2843,
							},
							{
								ID: 2941, Name: "Дублянки", Slug: "dublenki", ParentID: 2843,
							},
							{
								ID: 2944, Name: "Жилетки", Slug: "zhiletki", ParentID: 2843,
							},
							{
								ID: 2946, Name: "Болеро", Slug: "bolero", ParentID: 2843,
							},
							{
								ID: 2947, Name: "Екошуби", Slug: "ekoshuby", ParentID: 2843,
							},
							{
								ID: 2948, Name: "Анораки", Slug: "anoraki", ParentID: 2843,
							},
							{
								ID: 2949, Name: "Косухи", Slug: "kosuhi", ParentID: 2843,
							},
							{
								ID: 2950, Name: "Куртки-сорочки", Slug: "kurtki-rubashki", ParentID: 2843,
							},
							{
								ID: 2995, Name: "Інший верхній одяг", Slug: "drugoe", ParentID: 2843,
							},
						},
					},
					{
						ID: 2844, Name: "Джинси", Slug: "dzhinsy", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2859, Name: "Джинси-мом", Slug: "mom", ParentID: 2844,
							},
							{
								ID: 2873, Name: "Джеггінси", Slug: "dzhegginsy", ParentID: 2844,
							},
							{
								ID: 2884, Name: "Джинси-бойфренди", Slug: "boifrendy", ParentID: 2844,
							},
							{
								ID: 2894, Name: "Прямі джинси", Slug: "pryamye", ParentID: 2844,
							},
							{
								ID: 2904, Name: "Вузькі джинси / skinny", Slug: "uzkie-skinny", ParentID: 2844,
							},
							{
								ID: 2912, Name: "Високі джинси / джинси з високою талією", Slug: "vysokie", ParentID: 2844,
							},
							{
								ID: 2919, Name: "Джинси утеплені", Slug: "uteplennye", ParentID: 2844,
							},
							{
								ID: 2926, Name: "Рвані джинси", Slug: "rvanye", ParentID: 2844,
							},
							{
								ID: 2933, Name: "Джинси кльош", Slug: "klesh", ParentID: 2844,
							},
							{
								ID: 2938, Name: "Джинси кюлоти", Slug: "kyuloty", ParentID: 2844,
							},
							{
								ID: 2942, Name: "Широкі джинси", Slug: "shirokie", ParentID: 2844,
							},
							{
								ID: 2996, Name: "Інші джинси", Slug: "drugoe", ParentID: 2844,
							},
						},
					},
					{
						ID: 2845, Name: "Брюки", Slug: "bryuki", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2860, Name: "Класичні брюки", Slug: "klassicheskie", ParentID: 2845,
							},
							{
								ID: 2874, Name: "Штани чінос", Slug: "chinos", ParentID: 2845,
							},
							{
								ID: 2885, Name: "Штани карго", Slug: "kargo", ParentID: 2845,
							},
							{
								ID: 2895, Name: "Лляні штани", Slug: "lnyanye", ParentID: 2845,
							},
							{
								ID: 2905, Name: "Вельветові штани", Slug: "velvetovye", ParentID: 2845,
							},
							{
								ID: 2913, Name: "Лосини і легінси", Slug: "losiny-legginsy", ParentID: 2845,
							},
							{
								ID: 2920, Name: "Брюки з високою талією", Slug: "s-zavyshennoj-taliey", ParentID: 2845,
							},
							{
								ID: 2927, Name: "Укорочені брюки", Slug: "ukorochennye", ParentID: 2845,
							},
							{
								ID: 2964, Name: "Шкіряні брюки", Slug: "kozhanye", ParentID: 2845,
							},
							{
								ID: 2965, Name: "Широкі брюки", Slug: "shirokie", ParentID: 2845,
							},
							{
								ID: 2966, Name: "Брюки кльош", Slug: "klesh", ParentID: 2845,
							},
							{
								ID: 2967, Name: "Брюки-кюлоти", Slug: "kyuloty", ParentID: 2845,
							},
							{
								ID: 2968, Name: "Інші брюки", Slug: "drugoe", ParentID: 2845,
							},
						},
					},
					{
						ID: 2846, Name: "Комбінезони", Slug: "kombinezony", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2861, Name: "Літні комбінезони", Slug: "letnie", ParentID: 2846,
							},
							{
								ID: 2875, Name: "Зимові комбінезони", Slug: "zimnie", ParentID: 2846,
							},
							{
								ID: 2886, Name: "Комбінезони з брюками", Slug: "s-bryukami", ParentID: 2846,
							},
							{
								ID: 2896, Name: "Комбінезони з шортами", Slug: "s-shortami", ParentID: 2846,
							},
							{
								ID: 2962, Name: "Джинсові комбінезони", Slug: "dzhinsovye", ParentID: 2846,
							},
							{
								ID: 2963, Name: "Вечірні комбінезони", Slug: "vechernie", ParentID: 2846,
							},
							{
								ID: 2997, Name: "Інші комбінезони", Slug: "drugoe", ParentID: 2846,
							},
						},
					},
					{
						ID: 2847, Name: "Костюми", Slug: "kostyumy", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2862, Name: "Костюми зі спідницею", Slug: "s-yubkoy", ParentID: 2847,
							},
							{
								ID: 2960, Name: "Костюми з брюками", Slug: "s-bryukami", ParentID: 2847,
							},
							{
								ID: 2961, Name: "Костюми з шортами", Slug: "s-shortami", ParentID: 2847,
							},
							{
								ID: 2998, Name: "Інше", Slug: "drugoe", ParentID: 2847,
							},
						},
					},
					{
						ID: 2848, Name: "Жіночі піджаки та жакети", Slug: "pidzhaki-zhakety", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2863, Name: "Піджаки оверсайз", Slug: "oversayz", ParentID: 2848,
							},
							{
								ID: 2969, Name: "Джинсові піджаки", Slug: "dzhinsovye", ParentID: 2848,
							},
							{
								ID: 2970, Name: "Шкіряні піджаки", Slug: "kozhanye", ParentID: 2848,
							},
							{
								ID: 2971, Name: "Інші піджаки та жакети", Slug: "drugoe", ParentID: 2848,
							},
						},
					},
					{
						ID: 2849, Name: "Домашній одяг та одяг для сну", Slug: "domashnyaya-dlya-sna", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2864, Name: "Піжами", Slug: "pizhamy", ParentID: 2849,
							},
							{
								ID: 2876, Name: "Халати", Slug: "halaty", ParentID: 2849,
							},
							{
								ID: 2887, Name: "Кігурумі", Slug: "kigurumi", ParentID: 2849,
							},
							{
								ID: 2897, Name: "Нічні сорочки", Slug: "nochnye-rubashki", ParentID: 2849,
							},
							{
								ID: 2906, Name: "Пеньюари", Slug: "penyuary", ParentID: 2849,
							},
							{
								ID: 2972, Name: "Домашні костюми", Slug: "kostyumy", ParentID: 2849,
							},
							{
								ID: 2973, Name: "Домашні сукні", Slug: "platya", ParentID: 2849,
							},
							{
								ID: 2974, Name: "Домашні штани та брюки", Slug: "shtany-bryuki", ParentID: 2849,
							},
							{
								ID: 2999, Name: "Інше", Slug: "drugoe", ParentID: 2849,
							},
						},
					},
					{
						ID: 2850, Name: "Спортивний одяг", Slug: "sportivnaya-odezhda", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2865, Name: "Лижні костюми", Slug: "lyzhnye-kostyumy", ParentID: 2850,
							},
							{
								ID: 2877, Name: "Спортивна білизна", Slug: "bele", ParentID: 2850,
							},
							{
								ID: 2888, Name: "Спортивні лосини", Slug: "losiny", ParentID: 2850,
							},
							{
								ID: 2898, Name: "Джоггери", Slug: "dzhoggery", ParentID: 2850,
							},
							{
								ID: 2907, Name: "Спортивні капрі та бриджі", Slug: "kapri-bridzhi", ParentID: 2850,
							},
							{
								ID: 2914, Name: "Лижні куртки", Slug: "lyzhnye-kurtki", ParentID: 2850,
							},
							{
								ID: 2921, Name: "Лижні комбінезони", Slug: "lyzhnye-kombinezony", ParentID: 2850,
							},
							{
								ID: 2928, Name: "Лижні штани та брюки", Slug: "lyzhnye-shtany-bryuki", ParentID: 2850,
							},
							{
								ID: 2934, Name: "Штани для сноубордів", Slug: "shtany-dlya-snoubordov", ParentID: 2850,
							},
							{
								ID: 2939, Name: "Куртки для сноубордів", Slug: "kurtki-dlya-snoubordov", ParentID: 2850,
							},
							{
								ID: 2975, Name: "Спортивні костюми", Slug: "kostyumy", ParentID: 2850,
							},
							{
								ID: 2976, Name: "Спортивні футболки", Slug: "futbolki", ParentID: 2850,
							},
							{
								ID: 2977, Name: "Спортивні майки", Slug: "mayki", ParentID: 2850,
							},
							{
								ID: 2978, Name: "Спортивні топи", Slug: "topy", ParentID: 2850,
							},
							{
								ID: 2979, Name: "Спортивні спідниці", Slug: "yubki", ParentID: 2850,
							},
							{
								ID: 2980, Name: "Спортивні сукні", Slug: "platya", ParentID: 2850,
							},
							{
								ID: 2981, Name: "Спортивні шорти", Slug: "shorty", ParentID: 2850,
							},
							{
								ID: 2982, Name: "Спортивні комбінезони", Slug: "kombinezony", ParentID: 2850,
							},
							{
								ID: 2983, Name: "Спортивні кофти", Slug: "kofty", ParentID: 2850,
							},
							{
								ID: 2984, Name: "Спортивні куртки і вітровки", Slug: "kurtki-vetrovki", ParentID: 2850,
							},
							{
								ID: 2985, Name: "Спортивні штани та брюки", Slug: "shtany-bryuki", ParentID: 2850,
							},
							{
								ID: 2986, Name: "Інший спортивний одяг", Slug: "drugoe", ParentID: 2850,
							},
						},
					},
					{
						ID: 2851, Name: "Вишиванки", Slug: "vyshivanki", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2866, Name: "Національні костюми", Slug: "natsionalnye-kostyumy", ParentID: 2851,
							},
							{
								ID: 2878, Name: "Сорочки та блузки вишиванки", Slug: "rubashki-bluzki", ParentID: 2851,
							},
							{
								ID: 2987, Name: "Сукні вишиванки", Slug: "platya", ParentID: 2851,
							},
							{
								ID: 2988, Name: "Футболки вишиванки", Slug: "futbolki", ParentID: 2851,
							},
							{
								ID: 2989, Name: "Інші вишиванки", Slug: "drugoe", ParentID: 2851,
							},
						},
					},
					{
						ID: 2852, Name: "Інший жіночий одяг", Slug: "drugoe", ParentID: 2016,
					},
					{
						ID: 2951, Name: "Шорти", Slug: "shorty", ParentID: 2016,
						Children: []CategoryItem{
							{
								ID: 2952, Name: "Джинсові шорти", Slug: "dzhinsovye", ParentID: 2951,
							},
							{
								ID: 2953, Name: "Шкіряні шорти", Slug: "kozhanye", ParentID: 2951,
							},
							{
								ID: 2954, Name: "Трикотажні шорти", Slug: "trikotazhnye", ParentID: 2951,
							},
							{
								ID: 2955, Name: "Шорти з високою талією", Slug: "vysokie", ParentID: 2951,
							},
							{
								ID: 2956, Name: "Бриджі та капрі", Slug: "bridzhi-kapri", ParentID: 2951,
							},
							{
								ID: 2957, Name: "Шорти-бермуди", Slug: "bermudy", ParentID: 2951,
							},
							{
								ID: 2958, Name: "Велосипедки", Slug: "velosipedki", ParentID: 2951,
							},
							{
								ID: 2959, Name: "Інші шорти", Slug: "drugoe", ParentID: 2951,
							},
						},
					},
				},
			},
			{
				ID: 2107, Name: "Чоловічий одяг", Slug: "muzhskaya-odezhda", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 2736, Name: "Футболки та майки", Slug: "futbolki-mayki", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2747, Name: "Футболки", Slug: "futbolki", ParentID: 2736,
							},
							{
								ID: 2757, Name: "Лонгслів", Slug: "longsliv", ParentID: 2736,
							},
							{
								ID: 2767, Name: "Майки", Slug: "mayki", ParentID: 2736,
							},
							{
								ID: 2776, Name: "Поло", Slug: "polo", ParentID: 2736,
							},
							{
								ID: 2783, Name: "Інше", Slug: "drugoe", ParentID: 2736,
							},
						},
					},
					{
						ID: 2737, Name: "Сорочки і теніски", Slug: "rubashki-tenniski", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2748, Name: "Сорочки", Slug: "rubashki", ParentID: 2737,
							},
							{
								ID: 2758, Name: "Вишиванки", Slug: "vyshivanki", ParentID: 2737,
							},
							{
								ID: 2768, Name: "Теніски", Slug: "tenniski", ParentID: 2737,
							},
							{
								ID: 2810, Name: "Інше", Slug: "drugoe", ParentID: 2737,
							},
						},
					},
					{
						ID: 2738, Name: "Кофти й светри", Slug: "kofty-svitery", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2749, Name: "Худі", Slug: "hudi", ParentID: 2738,
							},
							{
								ID: 2759, Name: "Світшот", Slug: "svitshot", ParentID: 2738,
							},
							{
								ID: 2769, Name: "Толстовки", Slug: "tolstovka", ParentID: 2738,
							},
							{
								ID: 2777, Name: "Мантія", Slug: "mantiya", ParentID: 2738,
							},
							{
								ID: 2784, Name: "Олімпійка", Slug: "olimpiyka", ParentID: 2738,
							},
							{
								ID: 2789, Name: "Гольф / водолазка", Slug: "golf-vodolazka", ParentID: 2738,
							},
							{
								ID: 2793, Name: "Светр", Slug: "sviter", ParentID: 2738,
							},
							{
								ID: 2796, Name: "Кардиган", Slug: "kardigan", ParentID: 2738,
							},
							{
								ID: 2799, Name: "Светр реглан", Slug: "sviter-reglan", ParentID: 2738,
							},
							{
								ID: 2802, Name: "Кофти", Slug: "kofty", ParentID: 2738,
							},
							{
								ID: 2818, Name: "Інше", Slug: "drugoe", ParentID: 2738,
							},
						},
					},
					{
						ID: 2739, Name: "Верхній одяг", Slug: "verhnyaya-odezhda", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2750, Name: "Дощовики", Slug: "dozhdeviki", ParentID: 2739,
							},
							{
								ID: 2760, Name: "Пальто", Slug: "palto", ParentID: 2739,
							},
							{
								ID: 2770, Name: "Вітровки", Slug: "vetrovki", ParentID: 2739,
							},
							{
								ID: 2778, Name: "Шкіряні куртки", Slug: "kozhanye-kurtki", ParentID: 2739,
							},
							{
								ID: 2785, Name: "Бомбери", Slug: "bombery", ParentID: 2739,
							},
							{
								ID: 2790, Name: "Джинсові куртки", Slug: "dzhinsovye-kurtki", ParentID: 2739,
							},
							{
								ID: 2794, Name: "Парки", Slug: "parki", ParentID: 2739,
							},
							{
								ID: 2797, Name: "Пуховики та зимові куртки", Slug: "puhoviki-zimnie-kurtki", ParentID: 2739,
							},
							{
								ID: 2800, Name: "Шуби та дублянки", Slug: "shuby-dublenki", ParentID: 2739,
							},
							{
								ID: 2803, Name: "Жилети", Slug: "zhilety", ParentID: 2739,
							},
							{
								ID: 2805, Name: "Анорак", Slug: "anorak", ParentID: 2739,
							},
							{
								ID: 2807, Name: "Косуха", Slug: "kosuha", ParentID: 2739,
							},
							{
								ID: 2821, Name: "Інший верхній одяг", Slug: "drugoe", ParentID: 2739,
							},
							{
								ID: 2829, Name: "Плащі", Slug: "plashchi", ParentID: 2739,
							},
						},
					},
					{
						ID: 2740, Name: "Шорти", Slug: "shorty", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2751, Name: "Джинсові шорти", Slug: "dzhinsovye", ParentID: 2740,
							},
							{
								ID: 2761, Name: "Бриджі", Slug: "bridzhi", ParentID: 2740,
							},
							{
								ID: 2771, Name: "Карго", Slug: "kargo", ParentID: 2740,
							},
							{
								ID: 2779, Name: "Класичні шорти", Slug: "klassicheskie", ParentID: 2740,
							},
							{
								ID: 2786, Name: "Шорти бермуди", Slug: "bermudy", ParentID: 2740,
							},
							{
								ID: 2791, Name: "Пляжні шорти", Slug: "plyazhnye", ParentID: 2740,
							},
							{
								ID: 2822, Name: "Інші шорти", Slug: "drugoe", ParentID: 2740,
							},
						},
					},
					{
						ID: 2741, Name: "Джинси", Slug: "dzhinsy", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2752, Name: "Завужені джинси (Skinny)", Slug: "zauzhennye-skinni", ParentID: 2741,
							},
							{
								ID: 2762, Name: "Прямі джинси", Slug: "pryamye", ParentID: 2741,
							},
							{
								ID: 2772, Name: "Широкі джинси", Slug: "shirokie", ParentID: 2741,
							},
							{
								ID: 2780, Name: "Джинсові джогери", Slug: "dzhoggery", ParentID: 2741,
							},
							{
								ID: 2812, Name: "Класичні джинси", Slug: "klassicheskie", ParentID: 2741,
							},
							{
								ID: 2823, Name: "Інші джинси", Slug: "drugoe", ParentID: 2741,
							},
						},
					},
					{
						ID: 2742, Name: "Штани", Slug: "bryuki", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2753, Name: "Завужені штани", Slug: "zauzhennye", ParentID: 2742,
							},
							{
								ID: 2763, Name: "Чіноси", Slug: "chinosy", ParentID: 2742,
							},
							{
								ID: 2809, Name: "Джогери", Slug: "dzhoggery", ParentID: 2742,
							},
							{
								ID: 2811, Name: "Широкі штани", Slug: "shirokie", ParentID: 2742,
							},
							{
								ID: 2813, Name: "Карго", Slug: "kargo", ParentID: 2742,
							},
							{
								ID: 2817, Name: "Класичні штани", Slug: "klassicheskie", ParentID: 2742,
							},
							{
								ID: 2824, Name: "Інші штани", Slug: "drugoe", ParentID: 2742,
							},
						},
					},
					{
						ID: 2743, Name: "Костюми та піджаки", Slug: "kostyumy-pidzhaki", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2754, Name: "Чоловічі піджаки", Slug: "pidzhaki", ParentID: 2743,
							},
							{
								ID: 2764, Name: "Смокінги", Slug: "smokingi", ParentID: 2743,
							},
							{
								ID: 2773, Name: "Жилетки", Slug: "zhiletki", ParentID: 2743,
							},
							{
								ID: 2781, Name: "Блейзери", Slug: "bleyzery", ParentID: 2743,
							},
							{
								ID: 2787, Name: "Класичні костюми", Slug: "klassicheskie-kostyumy", ParentID: 2743,
							},
							{
								ID: 2825, Name: "Інше", Slug: "drugoe", ParentID: 2743,
							},
						},
					},
					{
						ID: 2744, Name: "Комбінезони", Slug: "kombinezony", ParentID: 2107,
					},
					{
						ID: 2745, Name: "Одяг для дому та сну", Slug: "odezhda-dlya-doma-sna", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2755, Name: "Домашній одяг", Slug: "odezhda", ParentID: 2745,
							},
							{
								ID: 2765, Name: "Піжами", Slug: "pizhamy", ParentID: 2745,
							},
							{
								ID: 2774, Name: "Халати", Slug: "halaty", ParentID: 2745,
							},
							{
								ID: 2826, Name: "Інше", Slug: "drugoe", ParentID: 2745,
							},
						},
					},
					{
						ID: 2746, Name: "Спортивний одяг", Slug: "sportivnaya-odezhda", ParentID: 2107,
						Children: []CategoryItem{
							{
								ID: 2756, Name: "Спортивні костюми", Slug: "kostyumy", ParentID: 2746,
							},
							{
								ID: 2766, Name: "Спортивні шкарпетки, гетри", Slug: "noski-getry", ParentID: 2746,
							},
							{
								ID: 2775, Name: "Спортивні куртки", Slug: "kurtki", ParentID: 2746,
							},
							{
								ID: 2782, Name: "Тайтси і легінси", Slug: "taytsy-legginsy", ParentID: 2746,
							},
							{
								ID: 2788, Name: "Лижні куртки", Slug: "lyzhnye-kurtki", ParentID: 2746,
							},
							{
								ID: 2792, Name: "Лижні костюми", Slug: "lyzhnye-kostyumy", ParentID: 2746,
							},
							{
								ID: 2795, Name: "Лижні штани", Slug: "lyzhnye-shtany", ParentID: 2746,
							},
							{
								ID: 2798, Name: "Лижні комбінезони", Slug: "lyzhnye-kombinezony", ParentID: 2746,
							},
							{
								ID: 2801, Name: "Куртки для сноубордів", Slug: "kurtki-dlya-snoubordov", ParentID: 2746,
							},
							{
								ID: 2804, Name: "Штани для сноубордів", Slug: "shtany-dlya-snoubordov", ParentID: 2746,
							},
							{
								ID: 2806, Name: "Одяг для тренажерного залу", Slug: "dlya-trenazhernogo-zala", ParentID: 2746,
							},
							{
								ID: 2808, Name: "Одяг для бігу", Slug: "dlya-bega", ParentID: 2746,
							},
							{
								ID: 2814, Name: "Спортивні футболки та майки", Slug: "futbolki-mayki", ParentID: 2746,
							},
							{
								ID: 2815, Name: "Спортивні шорти", Slug: "shorty", ParentID: 2746,
							},
							{
								ID: 2816, Name: "Спортивні кофти та светри", Slug: "kofty-svitery", ParentID: 2746,
							},
							{
								ID: 2819, Name: "Спортивні штани", Slug: "shtany", ParentID: 2746,
							},
							{
								ID: 2820, Name: "Джогери", Slug: "dzhoggery", ParentID: 2746,
							},
							{
								ID: 2827, Name: "Інший спортивний одяг", Slug: "drugoe", ParentID: 2746,
							},
						},
					},
					{
						ID: 2828, Name: "Інший чоловічий одяг", Slug: "drugoe", ParentID: 2107,
					},
				},
			},
			{
				ID: 3223, Name: "Спецодяг, спецвзуття та аксесуари", Slug: "spetsodezhda-i-spetsobuv", ParentID: 891,
				Children: []CategoryItem{
					{
						ID: 3224, Name: "Робочий спецодяг", Slug: "spetsodezhda", ParentID: 3223,
						Children: []CategoryItem{
							{
								ID: 3231, Name: "Робочі штани", Slug: "bryuki", ParentID: 3224,
							},
							{
								ID: 3234, Name: "Робочі куртки", Slug: "kurtki", ParentID: 3224,
							},
							{
								ID: 3237, Name: "Робочі комбінезони", Slug: "kombinezony", ParentID: 3224,
							},
							{
								ID: 3239, Name: "Робочі костюми", Slug: "kostyumy", ParentID: 3224,
							},
						},
					},
					{
						ID: 3225, Name: "Медичний одяг", Slug: "meditsinskaya-odezhda", ParentID: 3223,
						Children: []CategoryItem{
							{
								ID: 3232, Name: "Медичні халати", Slug: "halaty", ParentID: 3225,
							},
							{
								ID: 3235, Name: "Медичні костюми", Slug: "kostyumy_5", ParentID: 3225,
							},
						},
					},
					{
						ID: 3226, Name: "Світловідбиваючі жилети", Slug: "svetootrazhayushchie-zhilety", ParentID: 3223,
					},
					{
						ID: 3227, Name: "Одноразовий одяг", Slug: "odnorazovaya-odezhda", ParentID: 3223,
					},
					{
						ID: 3228, Name: "Дощовики", Slug: "dozhdeviki", ParentID: 3223,
					},
					{
						ID: 3229, Name: "Засоби захисту органів дихання", Slug: "sredstva-zashchity-organov-dyhaniya", ParentID: 3223,
					},
					{
						ID: 3230, Name: "Робоче взуття", Slug: "rabochaya-obuv", ParentID: 3223,
						Children: []CategoryItem{
							{
								ID: 3233, Name: "Робочі черевики", Slug: "botinki", ParentID: 3230,
							},
							{
								ID: 3236, Name: "Робочі чоботи", Slug: "sapogi", ParentID: 3230,
							},
							{
								ID: 3238, Name: "Робочі кросівки", Slug: "krossovki", ParentID: 3230,
							},
							{
								ID: 3241, Name: "Інше робоче взуття", Slug: "drugoe", ParentID: 3230,
							},
						},
					},
					{
						ID: 3240, Name: "Інше", Slug: "drugoe", ParentID: 3223,
					},
				},
			},
		},
	},
	{
		ID: 903, Name: "Хобі, відпочинок і спорт", Slug: "hobbi-otdyh-i-sport", Icon: "⚽", ParentID: 0,
		Children: []CategoryItem{
			{
				ID: 45, Name: "Музичні інструменти", Slug: "muzykalnye-instrumenty", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 90, Name: "Піаніно / фортепіано / роялі", Slug: "pianino-fortepiano-royali", ParentID: 45,
					},
					{
						ID: 489, Name: "Акустичні гітари", Slug: "akusticheskie-gitary", ParentID: 45,
					},
					{
						ID: 490, Name: "Електрогітари", Slug: "elektrogitary", ParentID: 45,
					},
					{
						ID: 491, Name: "Бас-гітари", Slug: "bas-gitary", ParentID: 45,
					},
					{
						ID: 493, Name: "Комбопідсилювачі", Slug: "kombousiliteli", ParentID: 45,
					},
					{
						ID: 495, Name: "Синтезатори", Slug: "sintezatory", ParentID: 45,
					},
					{
						ID: 497, Name: "Духові інструменти", Slug: "duhovye-instrumenty", ParentID: 45,
					},
					{
						ID: 498, Name: "Ударні інструменти", Slug: "udarnye-instrumenty", ParentID: 45,
					},
					{
						ID: 499, Name: "Студійне обладнання", Slug: "studiynoe-oborudovanie", ParentID: 45,
					},
					{
						ID: 509, Name: "Аксесуари для музичних інструментів", Slug: "aksessuary-dlya-muzykalnyh-instrumentov", ParentID: 45,
					},
					{
						ID: 1429, Name: "Інше", Slug: "prochee", ParentID: 45,
					},
				},
			},
			{
				ID: 49, Name: "Книги / журнали", Slug: "knigi-zhurnaly", ParentID: 903,
			},
			{
				ID: 53, Name: "Антикваріат / колекції", Slug: "antikvariat-kollektsii", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 622, Name: "Антикварні меблі", Slug: "antikvarnaya-mebel", ParentID: 53,
					},
					{
						ID: 623, Name: "Букіністика", Slug: "bukinistika", ParentID: 53,
					},
					{
						ID: 624, Name: "Живопис", Slug: "zhivopis", ParentID: 53,
					},
					{
						ID: 625, Name: "Предмети мистецтва", Slug: "predmety-iskusstva", ParentID: 53,
					},
					{
						ID: 626, Name: "Колекціонування", Slug: "kollektsionirovanie", ParentID: 53,
					},
					{
						ID: 631, Name: "Витвори майстрів / рукоділля", Slug: "podelki-rukodelie", ParentID: 53,
					},
				},
			},
			{
				ID: 54, Name: "Квитки", Slug: "bilety", ParentID: 903,
			},
			{
				ID: 115, Name: "Пошук попутників", Slug: "poisk-poputchikov", ParentID: 903,
			},
			{
				ID: 118, Name: "Пошук гуртів / музикантів", Slug: "poisk-grupp-muzykantov", ParentID: 903,
			},
			{
				ID: 485, Name: "CD / DVD / Платівки", Slug: "cd-dvd-plastinki", ParentID: 903,
			},
			{
				ID: 573, Name: "Спорт / відпочинок", Slug: "sport-otdyh", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 40, Name: "Лижі / сноуборди", Slug: "lyzhi-snoubordy", ParentID: 573,
					},
					{
						ID: 582, Name: "Ковзани", Slug: "konki", ParentID: 573,
					},
					{
						ID: 583, Name: "Роликові ковзани", Slug: "rolikovye-konki", ParentID: 573,
					},
					{
						ID: 584, Name: "Атлетика / фітнес", Slug: "atletika-fitnes", ParentID: 573,
					},
					{
						ID: 590, Name: "Туризм", Slug: "turizm", ParentID: 573,
						Children: []CategoryItem{
							{
								ID: 3290, Name: "Ліхтарики", Slug: "fonariki", ParentID: 590,
							},
							{
								ID: 3291, Name: "Газові пальники (портативні газові плити)", Slug: "gazovyye-gorelki", ParentID: 590,
							},
							{
								ID: 3292, Name: "Газові балони", Slug: "gazovyye-ballony", ParentID: 590,
							},
							{
								ID: 3293, Name: "GPS навігатори", Slug: "gps-navigatory", ParentID: 590,
							},
							{
								ID: 3294, Name: "Палатки", Slug: "palatki", ParentID: 590,
							},
							{
								ID: 3295, Name: "Спальні мішки", Slug: "spalnye-meshki", ParentID: 590,
							},
							{
								ID: 3296, Name: "Рюкзаки", Slug: "ryukzaki", ParentID: 590,
							},
							{
								ID: 3297, Name: "Інші туристичні товари", Slug: "prochie-turisticheskie-tovary", ParentID: 590,
							},
						},
					},
					{
						ID: 596, Name: "Полювання / риболовля", Slug: "ohota-rybalka", ParentID: 573,
					},
					{
						ID: 597, Name: "Ігри з ракеткою", Slug: "igry-s-raketkoy", ParentID: 573,
					},
					{
						ID: 602, Name: "Водні види спорту", Slug: "vodnye-vidy-sporta", ParentID: 573,
					},
					{
						ID: 603, Name: "Футбол", Slug: "futbol", ParentID: 573,
					},
					{
						ID: 604, Name: "Хокей", Slug: "hokkey", ParentID: 573,
					},
					{
						ID: 605, Name: "Єдиноборства / бокс", Slug: "edinoborstva-boks", ParentID: 573,
					},
					{
						ID: 607, Name: "Настільні ігри", Slug: "nastolnye-igry", ParentID: 573,
					},
					{
						ID: 608, Name: "Інші види спорту", Slug: "prochie-vidy-sporta", ParentID: 573,
					},
				},
			},
			{
				ID: 1425, Name: "Інше", Slug: "drugoe", ParentID: 903,
			},
			{
				ID: 3352, Name: "Мілітарія", Slug: "militariya", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 3353, Name: "Військовий одяг", Slug: "voennaya-odezhda", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3360, Name: "Камуфляжні костюми", Slug: "kamuflyazhnye-kostyumy", ParentID: 3353,
							},
							{
								ID: 3361, Name: "Штани", Slug: "shtany", ParentID: 3353,
							},
							{
								ID: 3362, Name: "Футболки", Slug: "futbolki", ParentID: 3353,
							},
							{
								ID: 3363, Name: "Сорочки", Slug: "rubashki", ParentID: 3353,
							},
							{
								ID: 3364, Name: "Кофти", Slug: "kofty", ParentID: 3353,
							},
							{
								ID: 3365, Name: "Куртки", Slug: "kurtki", ParentID: 3353,
							},
							{
								ID: 3366, Name: "Ремені", Slug: "remni", ParentID: 3353,
							},
							{
								ID: 3367, Name: "Рукавички", Slug: "perchatki", ParentID: 3353,
							},
							{
								ID: 3368, Name: "Шкарпетки", Slug: "noski", ParentID: 3353,
							},
							{
								ID: 3369, Name: "Балаклави та підшоломники", Slug: "balaklavy-i-podshlemniki", ParentID: 3353,
							},
							{
								ID: 3370, Name: "Кепки та бейсболки", Slug: "kepki-i-beysbolki", ParentID: 3353,
							},
							{
								ID: 3371, Name: "Панами", Slug: "panamy", ParentID: 3353,
							},
							{
								ID: 3372, Name: "Шапки", Slug: "shapki", ParentID: 3353,
							},
							{
								ID: 3373, Name: "Інший одяг", Slug: "drugaya-odezhda", ParentID: 3353,
							},
						},
					},
					{
						ID: 3354, Name: "Військове взуття", Slug: "voennaya-obuv", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3374, Name: "Берці", Slug: "bertsy", ParentID: 3354,
							},
							{
								ID: 3375, Name: "Кросівки", Slug: "krossovki", ParentID: 3354,
							},
							{
								ID: 3376, Name: "Черевики та напівчеревики", Slug: "botinki-i-polubotinki", ParentID: 3354,
							},
							{
								ID: 3377, Name: "Інше взуття", Slug: "drugaya-obuv", ParentID: 3354,
							},
						},
					},
					{
						ID: 3355, Name: "Рюкзаки та сумки", Slug: "ryukzaki-i-sumki", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3378, Name: "Рюкзаки", Slug: "ryukzaki", ParentID: 3355,
							},
							{
								ID: 3379, Name: "Сумки", Slug: "sumki", ParentID: 3355,
							},
							{
								ID: 3380, Name: "Баули", Slug: "bauly", ParentID: 3355,
							},
							{
								ID: 3381, Name: "Інше", Slug: "drugoe", ParentID: 3355,
							},
						},
					},
					{
						ID: 3356, Name: "Розвантажувальні системи", Slug: "razgruzochnye-sistemy", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3382, Name: "Жилети розвантажувальні", Slug: "zhilety-razgruzochnye", ParentID: 3356,
							},
							{
								ID: 3383, Name: "Тактичні пояси", Slug: "takticheskie-poyasa", ParentID: 3356,
							},
							{
								ID: 3384, Name: "Ремінно-плечові системи (РПС)", Slug: "remenno-plechevye-sistemy-rps", ParentID: 3356,
							},
							{
								ID: 3385, Name: "Інше", Slug: "drugoe", ParentID: 3356,
							},
						},
					},
					{
						ID: 3357, Name: "Підсумки", Slug: "podsumki", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3394, Name: "Медичні підсумки", Slug: "meditsinskie", ParentID: 3357,
							},
							{
								ID: 3395, Name: "Підсумки для гранат", Slug: "dlya-granat", ParentID: 3357,
							},
							{
								ID: 3396, Name: "Підсумки для магазинів штурмових гвинтівок", Slug: "dlya-magazinov", ParentID: 3357,
							},
							{
								ID: 3397, Name: "Сумки скидання магазинів", Slug: "sbrosa-magazinov", ParentID: 3357,
							},
							{
								ID: 3398, Name: "Спеціальні та інші підсумки", Slug: "drugie-podsumki", ParentID: 3357,
							},
						},
					},
					{
						ID: 3358, Name: "Спорядження", Slug: "snaryazhenie", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3399, Name: "Засоби маскування", Slug: "sredstva-maskirovki", ParentID: 3358,
							},
							{
								ID: 3400, Name: "Тепловізори", Slug: "teplovizory", ParentID: 3358,
							},
							{
								ID: 3401, Name: "Нашивки, шеврони, жетони", Slug: "nashivki-shevrony-zhetony", ParentID: 3358,
							},
							{
								ID: 3402, Name: "Кобури", Slug: "kobury", ParentID: 3358,
							},
							{
								ID: 3403, Name: "Ремені для зброї", Slug: "remni-dlya-oruzhiya", ParentID: 3358,
							},
							{
								ID: 3404, Name: "Пояси-патронташі", Slug: "poyasa-patrontashi", ParentID: 3358,
							},
							{
								ID: 3405, Name: "Тактичні навушники", Slug: "takticheskie-naushniki", ParentID: 3358,
							},
							{
								ID: 3406, Name: "Наколінники та налокітники", Slug: "nakolenniki-i-nalokotniki", ParentID: 3358,
							},
							{
								ID: 3407, Name: "Тактичні окуляри", Slug: "takticheskie-ochki", ParentID: 3358,
							},
							{
								ID: 3408, Name: "Інше спорядження", Slug: "drugoe-snaryazhenie", ParentID: 3358,
							},
						},
					},
					{
						ID: 3359, Name: "Збройові аксесуари", Slug: "oruzhejnye-aksessuary", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3409, Name: "Коліматори", Slug: "kollimatory", ParentID: 3359,
							},
							{
								ID: 3410, Name: "Оптичні приціли", Slug: "opticheskie-pritsely", ParentID: 3359,
							},
							{
								ID: 3411, Name: "Цівки, накладки, кріплення", Slug: "tsevya-nakladki-krepleniya", ParentID: 3359,
							},
							{
								ID: 3412, Name: "Полум'ягасники, ДТК, глушники", Slug: "plamegasiteli-dtk-glushiteli", ParentID: 3359,
							},
							{
								ID: 3413, Name: "Приклади, рукоятки, сошки", Slug: "priklady-rukoyatki-soshki", ParentID: 3359,
							},
							{
								ID: 3414, Name: "Інше", Slug: "drugoe", ParentID: 3359,
							},
						},
					},
					{
						ID: 3386, Name: "Медицина", Slug: "meditsina", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3388, Name: "Аптечки та медичні набори", Slug: "aptechki-i-meditsinskie-nabory", ParentID: 3386,
							},
							{
								ID: 3390, Name: "Джгути, турнікети", Slug: "zhguty-i-turnikety", ParentID: 3386,
							},
							{
								ID: 3392, Name: "Iнше", Slug: "drugoe", ParentID: 3386,
							},
						},
					},
					{
						ID: 3387, Name: "Захист", Slug: "zashchita", ParentID: 3352,
						Children: []CategoryItem{
							{
								ID: 3389, Name: "Бронепластини", Slug: "broneplastiny", ParentID: 3387,
							},
							{
								ID: 3391, Name: "Бронежилети", Slug: "bronezhilety", ParentID: 3387,
							},
							{
								ID: 3393, Name: "Каски та шоломи", Slug: "kaski-i-shlemy", ParentID: 3387,
							},
						},
					},
				},
			},
			{
				ID: 3415, Name: "Вело", Slug: "velo", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 3416, Name: "Велосипеди", Slug: "velosipedy", ParentID: 3415,
					},
					{
						ID: 3417, Name: "Велозахист", Slug: "velozashchita", ParentID: 3415,
					},
					{
						ID: 3418, Name: "Велоаксесуари", Slug: "veloaksessuary", ParentID: 3415,
					},
					{
						ID: 3419, Name: "Велозапчастини", Slug: "velozapchasti", ParentID: 3415,
					},
					{
						ID: 3420, Name: "Електровелосипеди", Slug: "elektrovelosipedy", ParentID: 3415,
					},
				},
			},
			{
				ID: 3421, Name: "Квадрокоптери та аксесуари", Slug: "kvadrokoptery-i-aksessuary", ParentID: 903,
				Children: []CategoryItem{
					{
						ID: 3426, Name: "Квадрокоптери", Slug: "kvadrokoptery", ParentID: 3421,
					},
					{
						ID: 3427, Name: "Аксесуари для квадрокоптерів", Slug: "instrumenty-i-aksessuary", ParentID: 3421,
					},
				},
			},
		},
	},
	{
		ID: 1151, Name: "Віддам безкоштовно", Slug: "otdam-darom", Icon: "🎁", ParentID: 0,
	},
	{
		ID: 1153, Name: "Обмін", Slug: "obmen-barter", Icon: "🔄", ParentID: 0,
	},
	{
		ID: 3901, Name: "До речі, відомі речі", Slug: "vidomi-rechi", Icon: "⭐", ParentID: 0,
	},
}

func FindCategory(id int) *CategoryItem {
	for i := range TopCategories {
		if found := findCategoryRec(&TopCategories[i], id); found != nil {
			return found
		}
	}
	return nil
}

func findCategoryRec(c *CategoryItem, id int) *CategoryItem {
	if c == nil {
		return nil
	}
	if c.ID == id {
		return c
	}
	for i := range c.Children {
		if found := findCategoryRec(&c.Children[i], id); found != nil {
			return found
		}
	}
	return nil
}

func FindCategoryBySlug(slug string) *CategoryItem {
	if slug == "" {
		return nil
	}
	for i := range TopCategories {
		if found := findCategoryBySlugRec(&TopCategories[i], slug); found != nil {
			return found
		}
	}
	return nil
}

func findCategoryBySlugRec(c *CategoryItem, slug string) *CategoryItem {
	if c == nil {
		return nil
	}
	if c.Slug == slug {
		return c
	}
	for i := range c.Children {
		if found := findCategoryBySlugRec(&c.Children[i], slug); found != nil {
			return found
		}
	}
	return nil
}

// CategoryBreadcrumb возвращает путь вида "Електроніка → Фото/Відео → Цифрові фотоапарати".
func CategoryBreadcrumb(id int) string {
	var names []string
	cur := FindCategory(id)
	for cur != nil {
		names = append([]string{cur.Name}, names...)
		if cur.ParentID == 0 {
			break
		}
		cur = FindCategory(cur.ParentID)
	}
	return strings.Join(names, " → ")
}

func FindCity(cityID int) *CityItem {
	for i := range TopCities {
		if TopCities[i].CityID == cityID {
			return &TopCities[i]
		}
	}
	return nil
}
