package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

// An infinite price could be pasted in a link ("…price:to]=Inf"). It cannot
// be stored as JSON, so one such filter broke the export of the whole chat.
func TestValidateRejectsUnusableNumbers(t *testing.T) {
	inf := math.Inf(1)
	bad := map[string]Search{
		"infinite price":   {Query: "x", MaxPrice: inf},
		"NaN price":        {Query: "x", MinPrice: math.NaN()},
		"absurd price":     {Query: "x", MaxPrice: 1e300},
		"infinite bound":   {Query: "x", Ranges: map[string]FilterRange{"filter_float_year": {To: inf}}},
		"NaN bound":        {Query: "x", Ranges: map[string]FilterRange{"filter_float_year": {From: math.NaN(), To: 5}}},
		"negative bound":   {Query: "x", Ranges: map[string]FilterRange{"filter_float_year": {From: -1, To: 5}}},
		"infinite minimum": {Query: "x", MinPrice: math.Inf(-1)},
	}
	for name, s := range bad {
		if err := s.Validate(); !errors.Is(err, ErrBadNumber) {
			t.Errorf("%s: err = %v, want ErrBadNumber", name, err)
		}
	}

	ok := Search{Query: "x", MinPrice: 100, MaxPrice: 5e6, Ranges: map[string]FilterRange{"filter_float_year": {From: 2010, To: 2020}}}
	if err := ok.Validate(); err != nil {
		t.Errorf("an ordinary filter was refused: %v", err)
	}
	if _, err := json.Marshal(ok); err != nil {
		t.Errorf("a valid filter must be exportable: %v", err)
	}
}

// A file sent to the bot is somebody's input. Names that the screens print
// as they are must not be able to outgrow a Telegram message: a filter whose
// card cannot be sent cannot be deleted either.
func TestNormalizeBoundsWhatAFileCanCarry(t *testing.T) {
	long := strings.Repeat("я", 5000)
	s := Search{
		Query: "x", CityName: long, DistrictName: long, CatName: long,
		BrandParam: "filter_enum_brand", BrandValue: long, BrandName: long, BrandNames: []string{long, long},
		QueryVariants: []string{long},
		Filters:       map[string][]string{"filter_enum_color": {long}},
	}
	s.Normalize()

	runes := func(v string) int { return len([]rune(v)) }
	for name, v := range map[string]string{
		"city": s.CityName, "district": s.DistrictName, "category": s.CatName,
		"brand": s.BrandName, "brand value": s.BrandValue, "brand names": s.BrandNames[0],
	} {
		if runes(v) > MaxLabelRunes+1 {
			t.Errorf("%s is %d runes long, the limit is %d", name, runes(v), MaxLabelRunes)
		}
	}
	if n := runes(s.QueryVariants[0]); n > MaxQueryRunes {
		t.Errorf("a query variant is %d runes long", n)
	}
	if n := runes(s.Filters["filter_enum_color"][0]); n > MaxValueRunes {
		t.Errorf("a filter value is %d runes long", n)
	}

	many := Search{Query: "x", Filters: map[string][]string{}}
	for i := 0; i <= MaxFilterKeys; i++ {
		many.Filters[fmt.Sprintf("filter_enum_k%d", i)] = []string{"v"}
	}
	many.Normalize()
	if err := many.Validate(); !errors.Is(err, ErrTooLarge) {
		t.Errorf("%d extra parameters: err = %v, want ErrTooLarge", len(many.Filters), err)
	}
}

// A price beyond int64 turned into MinInt64, whose negation is itself: the
// formatter recursed until the stack overflowed, which no recover can catch.
func TestAbsurdPriceIsFormattedNotFatal(t *testing.T) {
	for _, value := range []string{"1e30", "-1e30", "9223372036854775808"} {
		o := Offer{Params: []Param{{Key: "price", Type: "price",
			Value: json.RawMessage(`{"value":` + value + `,"type":"price","currency":"UAH"}`)}}}
		if got := o.GetPrice(); got == "" {
			t.Errorf("price %s formatted to nothing", value)
		}
	}
	if got := formatIntWithSpaces(math.MinInt64); !strings.HasPrefix(got, "-9 223 372") {
		t.Errorf("formatIntWithSpaces(MinInt64) = %q", got)
	}
	if roundPrice(math.NaN()) != 0 {
		t.Error("NaN must round to 0")
	}
}
