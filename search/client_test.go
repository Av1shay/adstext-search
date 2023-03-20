package search

import (
	"context"
	"github.com/Av1shay/adstext-search/model"
	"net/http"
	"testing"
	"time"
)

func TestClient_Do(t *testing.T) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	client := NewClient(httpClient)

	lines := []model.AdstxtLine{
		{Host: "indexexchange.com", SellerID: "193257", PublisherType: model.PublisherTypeDirectPublisherType},
		{Host: "openx.com", SellerID: "540922801", PublisherType: model.PublisherTypeDirectPublisherType},
		{Host: "rhythmone.com", SellerID: "1204009095", PublisherType: model.PublisherTypeDirectPublisherType, PublisherID: "a670c89d4a324e47"},
		{Host: "adform.com", SellerID: "1210", PublisherType: model.PublisherTypeReseller, PublisherID: "9f5210a2f0999e32"},
		{Host: "google.com", SellerID: "pub-2930805104418204", PublisherType: model.PublisherTypeReseller, PublisherID: "f08c47fec0942fa0 #APL"},
		{Host: "lijit.com", SellerID: "279713-eb", PublisherType: model.PublisherTypeDirectPublisherType, PublisherID: "fafdf38b16bf6b2b #SOVRN"},
	}

	res, err := client.Do(ctx, lines, []string{"nfl.com", "nascar.com"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("done", res)
}
