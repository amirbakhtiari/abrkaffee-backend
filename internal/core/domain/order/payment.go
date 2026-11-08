package order

import "time"

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "CASH"
	PaymentMethodPOS    PaymentMethod = "POS"
	PaymentMethodOnline PaymentMethod = "ONLINE"
)

type PaymentTransactionStatus string

const (
	PaymentTxPending    PaymentTransactionStatus = "PENDING"
	PaymentTxSuccessful PaymentTransactionStatus = "SUCCESSFUL"
	PaymentTxFailed     PaymentTransactionStatus = "FAILED"
	PaymentTxRefunded   PaymentTransactionStatus = "REFUNDED"
)

type Payment struct {
	id                   uint64
	orderID              uint64
	amount               int64
	paymentMethod        PaymentMethod
	status               PaymentTransactionStatus
	transactionReference *string
	paidAt               *time.Time
	createdAt            time.Time
}

func NewPayment(orderID uint64, amount int64, method PaymentMethod, ref *string) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidPaymentAmount
	}
	now := time.Now()
	return &Payment{
		orderID:              orderID,
		amount:               amount,
		paymentMethod:        method,
		status:               PaymentTxSuccessful,
		transactionReference: ref,
		paidAt:               &now,
		createdAt:            now,
	}, nil
}

func (p *Payment) ID() uint64                       { return p.id }
func (p *Payment) OrderID() uint64                  { return p.orderID }
func (p *Payment) Amount() int64                    { return p.amount }
func (p *Payment) PaymentMethod() PaymentMethod     { return p.paymentMethod }
func (p *Payment) Status() PaymentTransactionStatus { return p.status }
func (p *Payment) TransactionReference() *string    { return p.transactionReference }
func (p *Payment) PaidAt() *time.Time               { return p.paidAt }

type StatusHistory struct {
	ID         uint64
	OrderID    uint64
	FromStatus *Status
	ToStatus   Status
	ChangedBy  *uint64
	Note       *string
	CreatedAt  time.Time
}
