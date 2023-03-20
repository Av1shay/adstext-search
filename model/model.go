package model

type PublisherType string

const (
	PublisherTypeReseller            PublisherType = "RESELLER"
	PublisherTypeDirectPublisherType PublisherType = "DIRECT"
)

type AdstxtLine struct {
	Host          string
	SellerID      string
	PublisherType PublisherType
	PublisherID   string
}

func (a *AdstxtLine) Reset() {
	a.Host = ""
	a.SellerID = ""
	a.PublisherType = ""
	a.PublisherID = ""
}
