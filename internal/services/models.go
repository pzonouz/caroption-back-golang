package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Setting struct {
	ID    pgtype.Text `json:"id"`
	Key   pgtype.Text `json:"key"`
	Value pgtype.Text `json:"value"`
}

type Child struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name,omitempty"`
	Slug        pgtype.Text `json:"slug"`
	ParentID    uuid.UUID   `json:"parentId"`
	Description pgtype.Text `json:"description"`
	Priority    pgtype.Text `json:"priority"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type AccountChild struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name,omitempty"`
	ParentID  uuid.UUID   `json:"parentId"`
	Code      pgtype.Text `json:"code"`
	CreatedAt time.Time   `json:"createdAt"`
}

type Category struct {
	ID          uuid.UUID   `json:"id"`
	Name        pgtype.Text `json:"name"`
	ParentID    *uuid.UUID  `json:"parentId"`
	ParentName  pgtype.Text `json:"parentName"`
	Description pgtype.Text `json:"description"`
	Priority    pgtype.Text `json:"priority"`
	ImageID     uuid.UUID   `json:"imageId"`
	ImageUrl    pgtype.Text `json:"imageUrl"`
	Show        bool        `json:"show"`
	Children    []Child     `json:"children"`
	Slug        pgtype.Text `json:"slug"`
	TotalCount  int32       `json:"total_count"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Account struct {
	ID         uuid.UUID      `json:"id"`
	Code       pgtype.Text    `json:"code"`
	Name       pgtype.Text    `json:"name"`
	ParentId   *uuid.UUID     `json:"parentId"`
	ParentName pgtype.Text    `json:"parentName"`
	Locked     pgtype.Bool    `json:"locked"`
	Children   []AccountChild `json:"children"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type Entity struct {
	ID          uuid.UUID     `json:"id"`
	Name        pgtype.Text   `json:"name"`
	Description pgtype.Text   `json:"description"`
	ImageID     uuid.UUID     `json:"imageId"`
	ImageUrl    pgtype.Text   `json:"imageUrl"`
	Price       pgtype.Text   `json:"price"`
	Priority    pgtype.Text   `json:"priority"`
	ParentID    uuid.UUID     `json:"parentId"`
	ParentName  pgtype.Text   `json:"parentName"`
	Keywords    []pgtype.Text `json:"keywords"`
	Show        bool          `json:"show"`
	EntitySlug  pgtype.Text   `json:"entitySlug"`
	Children    []ChildEntity `json:"children"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type ChildEntity struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	EntitySlug string      `json:"entitySlug"`
	ParentId   uuid.UUID   `json:"parentId"`
	ParentName pgtype.Text `json:"parentName"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type Image struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ImageUrl   string    `json:"imageUrl"`
	CategoryID uuid.UUID `json:"categoryId"`
	ProductID  uuid.UUID `json:"productId"`
	EntityID   uuid.UUID `json:"EntityId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Article struct {
	ID             uuid.UUID     `json:"id"`
	Name           pgtype.Text   `json:"name"`
	Description    pgtype.Text   `json:"description"`
	ImageID        uuid.UUID     `json:"imageId"`
	ImageUrl       pgtype.Text   `json:"imageUrl"`
	Slug           pgtype.Text   `json:"slug"`
	ShowInProducts bool          `json:"showInProducts"`
	Keywords       []pgtype.Text `json:"keywords"`
	CategoryID     uuid.UUID     `json:"categoryId"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type Product struct {
	ID                     uuid.UUID               `json:"id"`
	Name                   pgtype.Text             `json:"name"`
	Description            pgtype.Text             `json:"description"`
	Info                   pgtype.Text             `json:"info"`
	BuyPrice               pgtype.Text             `json:"buyPrice"`
	SellPrice              pgtype.Text             `json:"sellPrice"`
	Count                  pgtype.Text             `json:"count"`
	CurrentCount           pgtype.Text             `json:"currentCount"`
	CategoryID             uuid.UUID               `json:"categoryId"`
	CategoryName           pgtype.Text             `json:"categoryName"`
	BrandID                uuid.UUID               `json:"brandId"`
	BrandName              pgtype.Text             `json:"brandName"`
	EntityID               uuid.UUID               `json:"entityId"`
	Slug                   pgtype.Text             `json:"slug"`
	ImageID                uuid.UUID               `json:"imageId"`
	ImageIDs               []uuid.UUID             `json:"imageIds"`
	Images                 []any                   `json:"images"`
	ImageUrl               pgtype.Text             `json:"imageUrl"`
	Parameters             []Parameter             `json:"parameters"`
	ProductParameterValues []ProductParameterValue `json:"productParameterValues"`
	Generatable            pgtype.Bool             `json:"generatable"`
	Generated              pgtype.Bool             `json:"generated"`
	Keywords               []pgtype.Text           `json:"keywords"`
	Show                   pgtype.Bool             `json:"show"`
	IsService              pgtype.Bool             `json:"isService"`
	Position               pgtype.Text             `json:"position"`
	Code                   pgtype.Text             `json:"code"`
	CreatedAt              time.Time               `json:"createdAt"`
	UpdatedAt              time.Time               `json:"updatedAt"`
}

type Brand struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description pgtype.Text `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type ParameterGroup struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	CategoryId   uuid.UUID   `json:"categoryId"`
	CategoryName pgtype.Text `json:"categoryName"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type Parameter struct {
	ID               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	Description      pgtype.Text   `json:"description"`
	Type             pgtype.Text   `json:"type"`
	ParameterGroupId uuid.UUID     `json:"parameterGroupId"`
	Selectables      []pgtype.Text `json:"selectables"`
	ParameterGroup   pgtype.Text   `json:"parameterGroup"`
	Priority         pgtype.Text   `json:"priority"`
	CreatedAt        time.Time     `json:"createdAt"`
}

type ProductParameterValue struct {
	ID              uuid.UUID   `json:"id"`
	ProductID       uuid.UUID   `json:"productId"`
	ParameterId     uuid.UUID   `json:"parameterId"`
	TextValue       pgtype.Text `json:"textValue"`
	BoolValue       pgtype.Bool `json:"boolValue"`
	SelectableValue pgtype.Text `json:"selectableValue"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type User struct {
	ID           uuid.UUID   `json:"id"`
	Password     pgtype.Text `json:"password"`
	Email        pgtype.Text `json:"email"`
	Token        pgtype.Text `json:"token"`
	TokenExpires *time.Time  `json:"tokenExpires"`
	IsAdmin      bool        `json:"isAdmin"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type InvoiceItem struct {
	ID          uuid.UUID   `json:"id"`
	InvoiceID   uuid.UUID   `json:"invoiceId"`
	ProductID   uuid.UUID   `json:"productId"`
	Price       pgtype.Text `json:"price"`
	Discount    pgtype.Text `json:"discount"`
	Count       pgtype.Text `json:"count"`
	ProductName pgtype.Text `json:"productName"`
	Description pgtype.Text `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Invoice struct {
	ID            uuid.UUID     `json:"id"`
	PersonID      uuid.UUID     `json:"personId"`
	PersonName    pgtype.Text   `json:"personName"`
	Type          string        `json:"type"`
	Total         string        `json:"total"`
	Description   string        `json:"description"`
	Discount      pgtype.Text   `json:"discount"`
	Notes         string        `json:"notes,omitempty"`
	Number        pgtype.Text   `json:"number"`
	VoucherNumber pgtype.Text   `json:"voucherNumber"`
	Items         []InvoiceItem `json:"items"`
	Date          time.Time     `json:"date"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}
type InvoiceProductsGroup struct {
	ID         uuid.UUID   `json:"id"`
	Name       pgtype.Text `json:"name"`
	ProductIds []string    `json:"productIds"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type Bank struct {
	ID           uuid.UUID   `json:"id"`
	Name         pgtype.Text `json:"name"`
	Branch       pgtype.Text `json:"branch"`
	Number       pgtype.Text `json:"number"`
	Shaba        pgtype.Text `json:"shaba"`
	FirstBalance pgtype.Text `json:"firstBalance"`
	Balance      pgtype.Text `json:"balance"`
	Type         pgtype.Text `json:"type"`
	AccountId    uuid.UUID   `json:"accountId"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type Check struct {
	ID          uuid.UUID   `json:"id"`
	Type        pgtype.Text `json:"type"`
	Name        pgtype.Text `json:"name"`
	CheckNumber pgtype.Text `json:"checkNumber"`
	BankName    pgtype.Text `json:"bankName"`
	DueDate     time.Time   `json:"dueDate"`
	Amount      pgtype.Text `json:"amount"`
	IssuerName  pgtype.Text `json:"issuerName"`
	Sayyad      pgtype.Text `json:"sayyad"`
	Description pgtype.Text `json:"description"`
	PersonId    pgtype.Text `json:"personId"`
	PersonName  pgtype.Text `json:"personName"`
	CheckBandId *uuid.UUID  `json:"checkBandId"`
	Status      pgtype.Text `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CheckBand struct {
	ID          uuid.UUID   `json:"id"`
	Serial      pgtype.Text `json:"serial"`
	Name        pgtype.Text `json:"name"`
	Count       pgtype.Text `json:"count"`
	StartNumber pgtype.Text `json:"startNumber"`
	EndNumber   pgtype.Text `json:"endNumber"`
	UsedNumbers []string    `json:"usedNumbers"`
	BankId      pgtype.Text `json:"bankId"`
	BankName    pgtype.Text `json:"bankName"`
	BankBranch  pgtype.Text `json:"bankBranch"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Person struct {
	ID          uuid.UUID   `json:"id"`
	FirstName   pgtype.Text `json:"firstName"`
	Name        pgtype.Text `json:"name"`
	Address     pgtype.Text `json:"address"`
	PhoneNumber pgtype.Text `json:"phoneNumber"`
	Debit       pgtype.Text `json:"debit"`
	Credit      pgtype.Text `json:"credit"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type VoucherItem struct {
	ID          uuid.UUID   `json:"id"`
	VoucherId   uuid.UUID   `json:"voucherId"`
	PersonId    *uuid.UUID  `json:"personId"`
	AccountId   *uuid.UUID  `json:"accountId"`
	Debit       pgtype.Text `json:"debit"`
	Credit      pgtype.Text `json:"credit"`
	Description pgtype.Text `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type VoucherItemForView struct {
	ID              uuid.UUID   `json:"id"`
	PersonName      pgtype.Text `json:"personName"`
	PersonFirstName pgtype.Text `json:"personFirstName"`
	PersonPhone     pgtype.Text `json:"personPhone"`
	AccountName     pgtype.Text `json:"accountName"`
	AccountCode     pgtype.Text `json:"accountCode"`
	VoucherNumber   pgtype.Text `json:"voucherNumber"`
	Debit           pgtype.Text `json:"debit"`
	Credit          pgtype.Text `json:"credit"`
	Description     pgtype.Text `json:"description"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

type Voucher struct {
	ID            uuid.UUID     `json:"id"`
	VoucherNumber pgtype.Text   `json:"voucherNumber"`
	Date          time.Time     `json:"date"`
	Description   pgtype.Text   `json:"description"`
	IsInvoice     bool          `json:"isInvoice"`
	Items         []VoucherItem `json:"items"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type VoucherForView struct {
	ID            uuid.UUID            `json:"id"`
	VoucherNumber pgtype.Text          `json:"voucherNumber"`
	Date          time.Time            `json:"date"`
	Description   pgtype.Text          `json:"description"`
	IsInvoice     bool                 `json:"isInvoice"`
	Items         []VoucherItemForView `json:"items"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

type SellInvoiceSettlement struct {
	InvoiceId *uuid.UUID  `json:"invoiceId"`
	Date      time.Time   `json:"date"`
	BankId    *uuid.UUID  `json:"bankId"`
	Bank      pgtype.Text `json:"bank"`
	Checks    []uuid.UUID `json:"checks"`
	Cash      pgtype.Text `json:"cash"`
	POS       bool        `json:"pos"`
}

type BuyInvoiceSettlement struct {
	InvoiceId *uuid.UUID  `json:"invoiceId"`
	Date      time.Time   `json:"date"`
	BankId    *uuid.UUID  `json:"bankId"`
	Bank      pgtype.Text `json:"bank"`
	Checks    []uuid.UUID `json:"checks"`
	Cash      pgtype.Text `json:"cash"`
	POS       bool        `json:"pos"`
}

type BankToBank struct {
	OriginBankId      *uuid.UUID  `json:"originBankId"`
	DestinationBankId *uuid.UUID  `json:"destinationBankId"`
	Date              time.Time   `json:"date"`
	Total             pgtype.Text `json:"total"`
	Description       pgtype.Text `json:"description"`
}

type CashToBank struct {
	BankId      *uuid.UUID  `json:"bankId"`
	Date        time.Time   `json:"date"`
	Total       pgtype.Text `json:"total"`
	Description pgtype.Text `json:"description"`
}

type BankToCash struct {
	BankId      *uuid.UUID  `json:"bankId"`
	Date        time.Time   `json:"date"`
	Total       pgtype.Text `json:"total"`
	Description pgtype.Text `json:"description"`
}

type BankToCost struct {
	BankId      *uuid.UUID  `json:"bankId"`
	AccountId   *uuid.UUID  `json:"accountId"`
	Date        time.Time   `json:"date"`
	Total       pgtype.Text `json:"total"`
	Description pgtype.Text `json:"description"`
}

type ReceiptFromPerson struct {
	Date     time.Time   `json:"date"`
	BankId   pgtype.Text `json:"bankId"`
	Bank     pgtype.Text `json:"bank"`
	PersonId pgtype.Text `json:"personId"`
	Checks   []uuid.UUID `json:"checks"`
	Cash     pgtype.Text `json:"cash"`
	POS      bool        `json:"pos"`
}

type PaymentToPerson struct {
	Date     time.Time   `json:"date"`
	BankId   pgtype.Text `json:"bankId"`
	Bank     pgtype.Text `json:"bank"`
	PersonId pgtype.Text `json:"personId"`
	Checks   []uuid.UUID `json:"checks"`
	Cash     pgtype.Text `json:"cash"`
	POS      bool        `json:"pos"`
}

type PersonActivitiesDetail struct {
	Debit         pgtype.Text `json:"debit"`
	Credit        pgtype.Text `json:"credit"`
	Description   pgtype.Text `json:"description"`
	VoucherNumber pgtype.Text `json:"voucherNumber"`
	VoucherDate   time.Time   `json:"voucherDate"`
	Balance       pgtype.Text `json:"balance"`
	NetBalance    pgtype.Text `json:"netBalance"`
	BalanceType   pgtype.Text `json:"balanceType"`
}

type ProductActivitiesDetail struct {
	InPrice       pgtype.Text `json:"inPrice"`
	InCount       pgtype.Text `json:"inCount"`
	OutPrice      pgtype.Text `json:"outPrice"`
	OutCount      pgtype.Text `json:"outCount"`
	Description   pgtype.Text `json:"description"`
	VoucherNumber pgtype.Text `json:"voucherNumber"`
	InvoiceNumber pgtype.Text `json:"invoiceNumber"`
	VoucherDate   time.Time   `json:"voucherDate"`
	Count         pgtype.Text `json:"count"`
}

type BankActivitiesDetail struct {
	Debit         pgtype.Text `json:"debit"`
	Credit        pgtype.Text `json:"credit"`
	Description   pgtype.Text `json:"description"`
	VoucherNumber pgtype.Text `json:"voucherNumber"`
	VoucherDate   time.Time   `json:"voucherDate"`
	Balance       pgtype.Text `json:"balance"`
	NetBalance    pgtype.Text `json:"netBalance"`
	BalanceType   pgtype.Text `json:"balanceType"`
	BankId        pgtype.Text `json:"bankId"`
}

type CashActivitiesDetail struct {
	Debit         pgtype.Text `json:"debit"`
	Credit        pgtype.Text `json:"credit"`
	AccountCode   pgtype.Text `json:"accountCode"`
	Description   pgtype.Text `json:"description"`
	VoucherNumber pgtype.Text `json:"voucherNumber"`
	VoucherDate   time.Time   `json:"voucherDate"`
	Balance       pgtype.Text `json:"balance"`
	NetBalance    pgtype.Text `json:"netBalance"`
	BalanceType   pgtype.Text `json:"balanceType"`
}
