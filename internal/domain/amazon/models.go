package amazon

import "time"

type Purchase struct {
	ID                     int64
	OrderDate              string
	OrderID                string
	Title                  string
	Category               string
	ASINISBN               string
	PurchasePricePerUnit   float64
	Quantity               int
	ShipmentDate           string
	ShippingAddressName    string
	ShippingAddressStreet1 string
	ShippingAddressStreet2 string
	ShippingAddressCity    string
	ShippingAddressState   string
	ShippingAddressZip     string
	OrderStatus            string
	CarrierName            string
	ItemSubtotal           float64
	ItemSubtotalTax        float64
	ItemTotal              float64
}

type ImportResult struct {
	TotalRows int
	Inserted  int
	Skipped   int
	Errors    []ImportError
	Duration  time.Duration
}

type ImportError struct {
	Row  int
	Col  int
	Err  error
	Data map[string]string
}

func (e ImportError) Error() string {
	return e.Err.Error()
}

func (r ImportResult) IsSuccess() bool {
	return len(r.Errors) == 0
}
