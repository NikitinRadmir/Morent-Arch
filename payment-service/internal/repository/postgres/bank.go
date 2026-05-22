package postgres

import (
	"time"

	"morent-arch/payment-service/internal/bank"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// BankRepository хранит клиентов банка и сессии в PostgreSQL.
type BankRepository struct {
	db *gorm.DB
}

func NewBankRepository(db *gorm.DB) *BankRepository {
	return &BankRepository{db: db}
}

var _ bank.ClientRegistry = (*BankRepository)(nil)

func (r *BankRepository) HasPhone(phone string) bool {
	var n int64
	r.db.Model(&BankClientRow{}).Where("phone = ?", phone).Count(&n)
	return n > 0
}

func (r *BankRepository) HasCardNumber(number string) bool {
	if number == "" {
		return false
	}
	var n int64
	r.db.Model(&BankClientRow{}).Where("card_number = ?", number).Count(&n)
	return n > 0
}

func (r *BankRepository) RegisterClient(phone, displayName, passwordHash, accountID, cardNumber, expDate, cvv, cardHolder string) {
	now := time.Now()
	row := BankClientRow{
		AccountID:    accountID,
		Phone:        phone,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CardNumber:   cardNumber,
		ExpDate:      expDate,
		CVV:          cvv,
		CardHolder:   cardHolder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_ = r.db.Create(&row).Error
}

func (r *BankRepository) VerifyLogin(phone, password string) (*bank.ClientRecord, bool) {
	c, ok := r.ClientByPhone(phone)
	if !ok {
		return nil, false
	}
	if bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password)) != nil {
		return nil, false
	}
	return c, true
}

func (r *BankRepository) CreateSession(token, accountID string) {
	_ = r.db.Create(&BankSessionRow{
		Token:     token,
		AccountID: accountID,
		CreatedAt: time.Now(),
	}).Error
}

func (r *BankRepository) DeleteSession(token string) {
	_ = r.db.Delete(&BankSessionRow{}, "token = ?", token).Error
}

func (r *BankRepository) AccountByToken(token string) (string, bool) {
	var row BankSessionRow
	if err := r.db.First(&row, "token = ?", token).Error; err != nil {
		return "", false
	}
	return row.AccountID, true
}

func (r *BankRepository) ClientByPhone(phone string) (*bank.ClientRecord, bool) {
	var row BankClientRow
	if err := r.db.First(&row, "phone = ?", phone).Error; err != nil {
		return nil, false
	}
	return rowToClient(&row), true
}

func (r *BankRepository) ClientByCardNumber(number string) (*bank.ClientRecord, bool) {
	var row BankClientRow
	if err := r.db.First(&row, "card_number = ?", number).Error; err != nil {
		return nil, false
	}
	return rowToClient(&row), true
}

func (r *BankRepository) ClientByAccountID(accountID string) (*bank.ClientRecord, bool) {
	var row BankClientRow
	if err := r.db.First(&row, "account_id = ?", accountID).Error; err != nil {
		return nil, false
	}
	return rowToClient(&row), true
}

func (r *BankRepository) EnsureClientCard(phone string, registeredAt time.Time) bool {
	var row BankClientRow
	if err := r.db.First(&row, "phone = ?", phone).Error; err != nil {
		return false
	}
	if row.CardNumber != "" {
		return true
	}
	displayName := row.DisplayName
	number, expDate, cvv, cardHolder, err := bank.GenerateVirtualCard(displayName, registeredAt, r.HasCardNumber)
	if err != nil {
		return false
	}
	res := r.db.Model(&row).Updates(map[string]any{
		"card_number": number,
		"exp_date":    expDate,
		"cvv":         cvv,
		"card_holder": cardHolder,
		"updated_at":  time.Now(),
	})
	return res.Error == nil && res.RowsAffected > 0
}

func rowToClient(row *BankClientRow) *bank.ClientRecord {
	return &bank.ClientRecord{
		AccountID:    row.AccountID,
		Phone:        row.Phone,
		DisplayName:  row.DisplayName,
		PasswordHash: []byte(row.PasswordHash),
		CardNumber:   row.CardNumber,
		ExpDate:      row.ExpDate,
		CVV:          row.CVV,
		CardHolder:   row.CardHolder,
	}
}
