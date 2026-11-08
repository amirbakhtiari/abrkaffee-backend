package order

type Item struct {
	id                      uint64
	orderID                 uint64
	productID               uint64
	categoryID              uint64
	priceAnnouncementItemID *uint64
	productName             string
	productCode             string
	categoryTitle           string
	unitPrice               int64
	quantity                int
	totalPrice              int64
	notes                   *string
}

func NewItem(
	productID uint64,
	categoryID uint64,
	priceAnnouncementItemID *uint64,
	productName string,
	productCode string,
	categoryTitle string,
	unitPrice int64,
	quantity int,
	notes *string,
) (Item, error) {
	if quantity <= 0 {
		return Item{}, ErrInvalidQuantity
	}
	if unitPrice < 0 {
		return Item{}, ErrInvalidUnitPrice
	}

	return Item{
		productID:               productID,
		categoryID:              categoryID,
		priceAnnouncementItemID: priceAnnouncementItemID,
		productName:             productName,
		productCode:             productCode,
		categoryTitle:           categoryTitle,
		unitPrice:               unitPrice,
		quantity:                quantity,
		totalPrice:              unitPrice * int64(quantity),
		notes:                   notes,
	}, nil
}

func ReconstructItem(
	id uint64,
	orderID uint64,
	productID uint64,
	categoryID uint64,
	priceAnnouncementItemID *uint64,
	productName string,
	productCode string,
	categoryTitle string,
	unitPrice int64,
	quantity int,
	totalPrice int64,
	notes *string,
) Item {
	return Item{
		id:                      id,
		orderID:                 orderID,
		productID:               productID,
		categoryID:              categoryID,
		priceAnnouncementItemID: priceAnnouncementItemID,
		productName:             productName,
		productCode:             productCode,
		categoryTitle:           categoryTitle,
		unitPrice:               unitPrice,
		quantity:                quantity,
		totalPrice:              totalPrice,
		notes:                   notes,
	}
}

func (i Item) ID() uint64                       { return i.id }
func (i Item) OrderID() uint64                  { return i.orderID }
func (i Item) ProductID() uint64                { return i.productID }
func (i Item) CategoryID() uint64               { return i.categoryID }
func (i Item) PriceAnnouncementItemID() *uint64 { return i.priceAnnouncementItemID }
func (i Item) ProductName() string              { return i.productName }
func (i Item) ProductCode() string              { return i.productCode }
func (i Item) CategoryTitle() string            { return i.categoryTitle }
func (i Item) UnitPrice() int64                 { return i.unitPrice }
func (i Item) Quantity() int                    { return i.quantity }
func (i Item) TotalPrice() int64                { return i.totalPrice }
func (i Item) Notes() *string                   { return i.notes }
