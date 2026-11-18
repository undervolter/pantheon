//go:build live

// Checks against the real OLX API. They are read-only and are not part of the
// normal test run:
//
//	go test -tags live -run Live -v ./core/checker/
//
// They pin down the facts about the API that the pipeline relies on, so that
// a change on the OLX side shows up here and not as silently missing alerts.
package checker

import (
	"context"
	"testing"
	"time"

	"pantheon/core/catalog"
	"pantheon/core/domain"
	"pantheon/core/olx"
)

func liveFetch(t *testing.T, s domain.Search) []domain.Offer {
	t.Helper()
	client, err := olx.NewOLXClient()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	offers, err := client.FetchPage(ctx, s, 50, 0)
	if err != nil {
		t.Fatalf("OLX: %v", err)
	}
	if len(offers) == 0 {
		t.Skip("OLX returned nothing for this search")
	}
	return offers
}

// The free-only mode asks OLX for listings marked «Безкоштовно», in a
// category and across the whole site.
func TestLiveFreeOnly(t *testing.T) {
	for name, s := range map[string]domain.Search{
		"in a category": {ID: "a", Category: 44, PriceMode: domain.PriceFree},
		"everywhere":    {ID: "b", PriceMode: domain.PriceFree},
	} {
		t.Run(name, func(t *testing.T) {
			offers := liveFetch(t, s)
			free, other, realEstate := 0, 0, 0
			for i := range offers {
				o := &offers[i]
				if catalog.IsRealEstateOffer(o.Category.ID, o.Category.Type) {
					realEstate++
					continue
				}
				if o.IsFree() {
					free++
				} else {
					other++
					t.Errorf("listing %d is not free: %s", o.ID, o.GetPrice())
				}
			}
			t.Logf("%d listings: %d free, %d not free, %d real estate", len(offers), free, other, realEstate)

			// And the pipeline lets every free one through as far as price goes.
			rules := Rules{Now: time.Now(), StartedAt: time.Now().Add(-24 * time.Hour)}
			for i := range offers {
				if v, detail := Evaluate(&offers[i], s, rules); v == SkipPrice {
					t.Errorf("listing %d was refused by its price: %s", offers[i].ID, detail)
				}
			}
		})
	}
}

// A search without a category returns real estate mixed with everything else;
// none of it may get past the pipeline.
func TestLiveRealEstateIsDropped(t *testing.T) {
	s := domain.Search{ID: "c", Query: "квартира"}
	offers := liveFetch(t, s)
	rules := Rules{Now: time.Now(), StartedAt: time.Now().Add(-24 * time.Hour)}

	dropped, passed := 0, 0
	for i := range offers {
		o := &offers[i]
		v, _ := Evaluate(o, s, rules)
		isRealEstate := o.Category.Type == "real_estate" || catalog.IsRealEstate(o.Category.ID)
		switch {
		case isRealEstate && v != SkipRealEstate:
			t.Errorf("real-estate listing %d (category %d, type %q) got verdict %q", o.ID, o.Category.ID, o.Category.Type, v.Reason())
		case v == SkipRealEstate:
			dropped++
		case v == Match:
			passed++
		}
	}
	t.Logf("%d listings: %d real estate dropped, %d matched", len(offers), dropped, passed)
	if dropped == 0 {
		t.Log("no real estate in this page; the check proved nothing this time")
	}
}
