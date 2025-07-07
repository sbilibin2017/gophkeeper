package models

// ---------------- SecretLoginPasswordDB ----------------

// SecretLoginPasswordDB represents a secret of type "login_password".
// It stores login credentials along with optional metadata.
type SecretLoginPasswordDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Login    string            `json:"login" db:"login"`         // Username or login
	Password string            `json:"password" db:"password"`   // Password
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// SecretLoginPasswordDBOption defines a functional option for configuring SecretLoginPasswordDB.
type SecretLoginPasswordDBOption func(*SecretLoginPasswordDB)

// NewSecretLoginPasswordDB creates a new SecretLoginPasswordDB configured with the given options.
func NewSecretLoginPasswordDB(opts ...SecretLoginPasswordDBOption) *SecretLoginPasswordDB {
	lp := &SecretLoginPasswordDB{
		Meta: make(map[string]string),
	}
	for _, opt := range opts {
		opt(lp)
	}
	return lp
}

// WithSecretLogin sets the Login field of SecretLoginPasswordDB.
func WithSecretLogin(login string) SecretLoginPasswordDBOption {
	return func(lp *SecretLoginPasswordDB) {
		lp.Login = login
	}
}

// WithSecretPassword sets the Password field of SecretLoginPasswordDB.
func WithSecretPassword(password string) SecretLoginPasswordDBOption {
	return func(lp *SecretLoginPasswordDB) {
		lp.Password = password
	}
}

// WithSecretLoginPasswordMeta sets the Meta map of SecretLoginPasswordDB.
func WithSecretLoginPasswordMeta(meta map[string]string) SecretLoginPasswordDBOption {
	return func(lp *SecretLoginPasswordDB) {
		lp.Meta = meta
	}
}

// WithSecretLoginPasswordSecretID sets the SecretID field of SecretLoginPasswordDB.
func WithSecretLoginPasswordSecretID(id string) SecretLoginPasswordDBOption {
	return func(lp *SecretLoginPasswordDB) {
		lp.SecretID = id
	}
}

// ---------------- SecretPayloadTextDB ----------------

// SecretPayloadTextDB represents a secret of type "text".
// It stores arbitrary text content along with optional metadata.
type SecretPayloadTextDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Content  string            `json:"content" db:"content"`     // Main text content
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// SecretPayloadTextDBOption defines a functional option for configuring SecretPayloadTextDB.
type SecretPayloadTextDBOption func(*SecretPayloadTextDB)

// NewSecretPayloadTextDB creates a new SecretPayloadTextDB configured with the given options.
func NewSecretPayloadTextDB(opts ...SecretPayloadTextDBOption) *SecretPayloadTextDB {
	pt := &SecretPayloadTextDB{
		Meta: make(map[string]string),
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

// WithSecretTextContent sets the Content field of SecretPayloadTextDB.
func WithSecretTextContent(content string) SecretPayloadTextDBOption {
	return func(pt *SecretPayloadTextDB) {
		pt.Content = content
	}
}

// WithSecretTextMeta sets the Meta map of SecretPayloadTextDB.
func WithSecretTextMeta(meta map[string]string) SecretPayloadTextDBOption {
	return func(pt *SecretPayloadTextDB) {
		pt.Meta = meta
	}
}

// WithSecretTextSecretID sets the SecretID field of SecretPayloadTextDB.
func WithSecretTextSecretID(id string) SecretPayloadTextDBOption {
	return func(pt *SecretPayloadTextDB) {
		pt.SecretID = id
	}
}

// ---------------- SecretPayloadBinaryDB ----------------

// SecretPayloadBinaryDB represents a secret of type "binary".
// It stores raw binary data along with optional metadata.
type SecretPayloadBinaryDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Data     []byte            `json:"data" db:"data"`           // Raw binary data
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata
}

// SecretPayloadBinaryDBOption defines a functional option for configuring SecretPayloadBinaryDB.
type SecretPayloadBinaryDBOption func(*SecretPayloadBinaryDB)

// NewSecretPayloadBinaryDB creates a new SecretPayloadBinaryDB configured with the given options.
func NewSecretPayloadBinaryDB(opts ...SecretPayloadBinaryDBOption) *SecretPayloadBinaryDB {
	pb := &SecretPayloadBinaryDB{
		Meta: make(map[string]string),
	}
	for _, opt := range opts {
		opt(pb)
	}
	return pb
}

// WithSecretBinaryData sets the Data field of SecretPayloadBinaryDB.
func WithSecretBinaryData(data []byte) SecretPayloadBinaryDBOption {
	return func(pb *SecretPayloadBinaryDB) {
		pb.Data = data
	}
}

// WithSecretBinaryMeta sets the Meta map of SecretPayloadBinaryDB.
func WithSecretBinaryMeta(meta map[string]string) SecretPayloadBinaryDBOption {
	return func(pb *SecretPayloadBinaryDB) {
		pb.Meta = meta
	}
}

// WithSecretBinarySecretID sets the SecretID field of SecretPayloadBinaryDB.
func WithSecretBinarySecretID(id string) SecretPayloadBinaryDBOption {
	return func(pb *SecretPayloadBinaryDB) {
		pb.SecretID = id
	}
}

// ---------------- SecretPayloadCardDB ----------------

// SecretPayloadCardDB represents a secret of type "card".
// It stores payment card details along with optional metadata.
type SecretPayloadCardDB struct {
	SecretID string            `json:"secret_id" db:"secret_id"` // Unique secret identifier (UUID)
	Number   string            `json:"number" db:"number"`       // Card number
	Holder   string            `json:"holder" db:"holder"`       // Cardholder's name
	ExpMonth int               `json:"exp_month" db:"exp_month"` // Expiry month (1-12)
	ExpYear  int               `json:"exp_year" db:"exp_year"`   // Expiry year (four digits)
	CVV      string            `json:"cvv" db:"cvv"`             // CVV code
	Meta     map[string]string `json:"meta,omitempty" db:"meta"` // Optional metadata (e.g., bank, card type)
}

// SecretPayloadCardDBOption defines a functional option for configuring SecretPayloadCardDB.
type SecretPayloadCardDBOption func(*SecretPayloadCardDB)

// NewSecretPayloadCardDB creates a new SecretPayloadCardDB configured with the given options.
func NewSecretPayloadCardDB(opts ...SecretPayloadCardDBOption) *SecretPayloadCardDB {
	pc := &SecretPayloadCardDB{
		Meta: make(map[string]string),
	}
	for _, opt := range opts {
		opt(pc)
	}
	return pc
}

// WithSecretCardNumber sets the Number field of SecretPayloadCardDB.
func WithSecretCardNumber(number string) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.Number = number
	}
}

// WithSecretCardHolder sets the Holder field of SecretPayloadCardDB.
func WithSecretCardHolder(holder string) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.Holder = holder
	}
}

// WithSecretCardExpMonth sets the ExpMonth field of SecretPayloadCardDB.
func WithSecretCardExpMonth(month int) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.ExpMonth = month
	}
}

// WithSecretCardExpYear sets the ExpYear field of SecretPayloadCardDB.
func WithSecretCardExpYear(year int) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.ExpYear = year
	}
}

// WithSecretCardCVV sets the CVV field of SecretPayloadCardDB.
func WithSecretCardCVV(cvv string) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.CVV = cvv
	}
}

// WithSecretCardMeta sets the Meta map of SecretPayloadCardDB.
func WithSecretCardMeta(meta map[string]string) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.Meta = meta
	}
}

// WithSecretCardSecretID sets the SecretID field of SecretPayloadCardDB.
func WithSecretCardSecretID(id string) SecretPayloadCardDBOption {
	return func(pc *SecretPayloadCardDB) {
		pc.SecretID = id
	}
}
