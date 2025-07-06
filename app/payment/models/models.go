package models

type PaymentAuditLog struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	UserID     int64     `json:"user_id"`
	PaymentID  int64     `json:"payment_id"`
	ExternalID string    `json:"external_id"`
	Event      string    `json:"event"` // save payment, create invoice, save payment anomaly
	Actor      string    `json:"actor"` // order, xendit, scheduler
	CreateTime time.Time `json:"create_time"`
}

type FailedEvents struct {
	ID         int       `json:"id"`
	OrderID    int64     `json:"order_id"`
	ExternalID string    `json:"external_id"`
	FailedType int       `json:"failed_type"`
	Status     int       `json:"status"`
	Notes      string    `json:"notes"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type OrderCreatedEvent struct {
	OrderID         int64   `json:"order_id"`
	UserID          int64   `json:"user_id"`
	TotalAmount     float64 `json:"total_amount"`
	PaymentMethod   string  `json:"payment_method"`
	ShippingAddress string  `json:"shipping_address"`
}

type Payment struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	ExternalID  string    `json:"external_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	ExpiredTime time.Time `json:"expired_time"`
}

type PaymentRequests struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	UserID     int64     `json:"user_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	RetryCount int       `json:"retry_count"`
	Notes      string    `json:"notes"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type PaymentStatusUpdateEvent struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type FailedPaymentList struct {
	TotalFailedPayment int `json:"total_failed_payments"`
	PaymentList        []PaymentRequests
}

type PaymentAnomaly struct {
	ID         int    `json:"id"`
	OrderID    int64  `json:"order_id"`
	ExternalID string `json:"external_id"`
	// 1: Anomaly Amount
	AnomalyType int    `json:"anomaly_type"`
	Notes       string `json:"notes"`
	// 1: Success
	// 99: Need to check
	Status     int       `json:"status"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type XenditWebhookPayload struct {
	ExternalID string  `json:"external_id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
}

type XenditInvoiceRequest struct {
	ExternalID  string  `json:"external_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	PayerEmail  string  `json:"payer_email"`
}

type XenditInvoiceResponse struct {
	ID         string    `json:"id"`
	ExpiryDate time.Time `json:"expiry_date"`
	InvoiceURL string    `json:"invoice_url"`
	Status     string    `json:"status"`
}