package infra

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// Store implements usecase.TaxpayerStore and usecase.RegimeStore with GORM
// on the unit of work in ctx (pgdb.Conn), as cc_app: RLS shows only the
// context user's rows. It never begins, commits or rolls back, and never
// sets session state (P0-30 D-3): the use case owns the transaction.
type Store struct{}

var (
	_ usecase.TaxpayerStore = Store{}
	_ usecase.RegimeStore   = Store{}
)

// NewStore returns a Store. It holds no pool.
func NewStore() Store { return Store{} }

// newTaxpayerRow is exactly the columns cc_app may INSERT.
type newTaxpayerRow struct {
	ID                 string `gorm:"column:id"`
	UserID             string `gorm:"column:user_id"`
	PersonType         string `gorm:"column:person_type"`
	DisplayName        string `gorm:"column:display_name"`
	RFCEncrypted       []byte `gorm:"column:rfc_encrypted"`
	RFCKeyID           int16  `gorm:"column:rfc_key_id"`
	RFCHash            []byte `gorm:"column:rfc_hash"`
	RFCMasked          string `gorm:"column:rfc_masked"`
	RFCCheckDigitValid bool   `gorm:"column:rfc_check_digit_valid"`
}

func (newTaxpayerRow) TableName() string { return "taxpayers" }

// taxpayerRow is what the use cases read: never the RFC ciphertext or hash.
type taxpayerRow struct {
	ID          string     `gorm:"column:id"`
	PersonType  string     `gorm:"column:person_type"`
	DisplayName string     `gorm:"column:display_name"`
	RFCMasked   string     `gorm:"column:rfc_masked"`
	ArchivedAt  *time.Time `gorm:"column:archived_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (taxpayerRow) TableName() string { return "taxpayers" }

var taxpayerColumns = []string{"id", "person_type", "display_name", "rfc_masked", "archived_at", "created_at"}

func (r taxpayerRow) toStored() usecase.StoredTaxpayer {
	t := usecase.StoredTaxpayer{
		ID: r.ID, PersonType: domain.PersonType(r.PersonType), DisplayName: r.DisplayName,
		RFCMasked: r.RFCMasked, CreatedAt: r.CreatedAt,
	}
	if r.ArchivedAt != nil {
		t.ArchivedAt = *r.ArchivedAt
	}
	return t
}

// lockClause renders a Lock for the named table ("" = the query's only one).
func lockClause(lock usecase.Lock, table string) (clause.Locking, bool) {
	var of clause.Table
	if table != "" {
		of = clause.Table{Name: table}
	}
	switch lock {
	case usecase.LockShare:
		return clause.Locking{Strength: clause.LockingStrengthShare, Table: of}, true
	case usecase.LockUpdate:
		return clause.Locking{Strength: "NO KEY UPDATE", Table: of}, true
	}
	return clause.Locking{}, false
}

// CountForCreate implements usecase.TaxpayerStore through the migration's
// taxpayer_count_for_create(): it locks the user row, then counts.
func (Store) CountForCreate(ctx context.Context) (int, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	var count int
	if err := db.Raw("SELECT taxpayer_count_for_create()").Scan(&count).Error; err != nil {
		return 0, translate("count taxpayers", err)
	}
	return count, nil
}

// Insert implements usecase.TaxpayerStore.
func (Store) Insert(ctx context.Context, t usecase.NewTaxpayer) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	row := newTaxpayerRow{
		ID: t.ID, UserID: t.UserID, PersonType: string(t.PersonType), DisplayName: t.DisplayName.String(),
		RFCEncrypted: t.RFC.Ciphertext, RFCKeyID: t.RFC.KeyID, RFCHash: t.RFC.Hash, RFCMasked: t.RFC.Masked,
		RFCCheckDigitValid: t.RFC.CheckDigitValid,
	}
	return translate("insert taxpayer", db.Create(&row).Error)
}

// Find implements usecase.TaxpayerStore.
func (Store) Find(ctx context.Context, id string, lock usecase.Lock) (usecase.StoredTaxpayer, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredTaxpayer{}, err
	}
	query := db.Select(taxpayerColumns).Where("id = ?", id).Limit(1)
	if locking, ok := lockClause(lock, ""); ok {
		query = query.Clauses(locking)
	}
	var rows []taxpayerRow
	if err := query.Find(&rows).Error; err != nil {
		return usecase.StoredTaxpayer{}, translate("read taxpayer", err)
	}
	if len(rows) == 0 {
		return usecase.StoredTaxpayer{}, usecase.ErrTaxpayerNotFound
	}
	return rows[0].toStored(), nil
}

// List implements usecase.TaxpayerStore.
func (Store) List(ctx context.Context, includeArchived bool) ([]usecase.StoredTaxpayer, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	query := db.Select(taxpayerColumns).Order("created_at, id")
	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}
	var rows []taxpayerRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, translate("list taxpayers", err)
	}
	out := make([]usecase.StoredTaxpayer, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toStored())
	}
	return out, nil
}

// Rename implements usecase.TaxpayerStore.
func (Store) Rename(ctx context.Context, id string, name domain.DisplayName) error {
	return updateTaxpayer(ctx, "rename taxpayer", id, map[string]any{"display_name": name.String()})
}

// ReplaceRFC implements usecase.TaxpayerStore: the five rfc_* columns move
// together (OQ-1).
func (Store) ReplaceRFC(ctx context.Context, id string, rfc usecase.SealedRFC) error {
	return updateTaxpayer(ctx, "replace rfc", id, map[string]any{
		"rfc_encrypted":         bytea(rfc.Ciphertext),
		"rfc_key_id":            rfc.KeyID,
		"rfc_hash":              bytea(rfc.Hash),
		"rfc_masked":            rfc.Masked,
		"rfc_check_digit_valid": rfc.CheckDigitValid,
	})
}

// SetArchivedAt implements usecase.TaxpayerStore.
func (Store) SetArchivedAt(ctx context.Context, id string, at time.Time) error {
	var archivedAt *time.Time
	if !at.IsZero() {
		archivedAt = &at
	}
	return updateTaxpayer(ctx, "set archived_at", id, map[string]any{"archived_at": archivedAt})
}

func updateTaxpayer(ctx context.Context, op, id string, columns map[string]any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Model(&taxpayerRow{}).Where("id = ?", id).Updates(columns)
	if result.Error != nil {
		return translate(op, result.Error)
	}
	if result.RowsAffected == 0 {
		return usecase.ErrTaxpayerNotFound
	}
	return nil
}

// Delete implements usecase.TaxpayerStore.
func (Store) Delete(ctx context.Context, id string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	result := db.Where("id = ?", id).Delete(&taxpayerRow{})
	if result.Error != nil {
		return false, translate("delete taxpayer", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// regimeRow is a taxpayer_regimes row joined with its catalog name and
// treatment.
type regimeRow struct {
	TaxpayerID  string
	RegimeCode  string
	Name        string
	Treatment   string
	Options     string
	StartPeriod string
	EndPeriod   *string
}

// regimeSelect joins the catalog on (code, person_type): 626 PF and 626 PM
// are different regimes.
const regimeSelect = `SELECT tr.taxpayer_id, tr.regime_code, r.name, r.treatment, tr.options::text AS options,
       tr.start_period, tr.end_period
FROM taxpayer_regimes tr
JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type`

func (r regimeRow) toHeld() (usecase.HeldRegime, error) {
	options, err := decodeOptions([]byte(r.Options))
	if err != nil {
		return usecase.HeldRegime{}, err
	}
	window := domain.Window{}
	if window.Start, err = domain.ParsePeriod(r.StartPeriod); err != nil {
		return usecase.HeldRegime{}, fmt.Errorf("taxpayers infra: stored start_period: %w", err)
	}
	if r.EndPeriod != nil {
		if window.End, err = domain.ParsePeriod(*r.EndPeriod); err != nil {
			return usecase.HeldRegime{}, fmt.Errorf("taxpayers infra: stored end_period: %w", err)
		}
	}
	return usecase.HeldRegime{
		TaxpayerID: r.TaxpayerID, Code: r.RegimeCode, Name: r.Name, Treatment: r.Treatment,
		Options: options, Window: window,
	}, nil
}

// Regimes implements usecase.RegimeStore: one query for every taxpayer.
func (Store) Regimes(ctx context.Context, taxpayerIDs ...string) ([]usecase.HeldRegime, error) {
	if len(taxpayerIDs) == 0 {
		return nil, nil
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []regimeRow
	query := regimeSelect + "\nWHERE tr.taxpayer_id IN ?\nORDER BY tr.taxpayer_id, tr.start_period, tr.regime_code"
	if err := db.Raw(query, taxpayerIDs).Scan(&rows).Error; err != nil {
		return nil, translate("list regimes", err)
	}
	return toHeldRegimes(rows)
}

func toHeldRegimes(rows []regimeRow) ([]usecase.HeldRegime, error) {
	out := make([]usecase.HeldRegime, 0, len(rows))
	for _, r := range rows {
		held, err := r.toHeld()
		if err != nil {
			return nil, err
		}
		out = append(out, held)
	}
	return out, nil
}

// regimeLocks lock only the taxpayer_regimes row, never the catalog row.
var regimeLocks = map[usecase.Lock]string{
	usecase.LockShare:  "\nFOR SHARE OF tr",
	usecase.LockUpdate: "\nFOR NO KEY UPDATE OF tr",
}

// FindRegime implements usecase.RegimeStore.
func (Store) FindRegime(ctx context.Context, taxpayerID, code string, lock usecase.Lock) (usecase.HeldRegime, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.HeldRegime{}, err
	}
	var rows []regimeRow
	query := regimeSelect + "\nWHERE tr.taxpayer_id = ? AND tr.regime_code = ?" + regimeLocks[lock]
	if err := db.Raw(query, taxpayerID, code).Scan(&rows).Error; err != nil {
		return usecase.HeldRegime{}, translate("read regime", err)
	}
	if len(rows) == 0 {
		return usecase.HeldRegime{}, usecase.ErrRegimeNotHeld
	}
	return rows[0].toHeld()
}

// InsertRegime implements usecase.RegimeStore.
func (Store) InsertRegime(ctx context.Context, r usecase.NewRegime) error {
	options, err := encodeOptions(r.Options)
	if err != nil {
		return err
	}
	return execRegime(ctx, "insert regime", false, `INSERT INTO taxpayer_regimes
  (taxpayer_id, regime_code, user_id, person_type, options, start_period, needs_cpa_review_ack)
VALUES (?, ?, ?, ?, ?::jsonb, ?, ?)`,
		r.TaxpayerID, r.Code, r.UserID, string(r.PersonType), string(options), r.Start.String(), r.NeedsCPAReviewAck)
}

// ReopenRegime implements usecase.RegimeStore: a plain UPDATE of an ended
// row that never names start_period, so the overlap trigger checks the
// stored window (an upsert's INSERT pass would check the body's start).
func (Store) ReopenRegime(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue, ack bool) error {
	encoded, err := encodeOptions(options)
	if err != nil {
		return err
	}
	return execRegime(ctx, "reopen regime", true, `UPDATE taxpayer_regimes
SET end_period = NULL, options = ?::jsonb, needs_cpa_review_ack = ?
WHERE taxpayer_id = ? AND regime_code = ? AND end_period IS NOT NULL`,
		string(encoded), ack, taxpayerID, code)
}

// UpdateOptions implements usecase.RegimeStore.
func (Store) UpdateOptions(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue) error {
	encoded, err := encodeOptions(options)
	if err != nil {
		return err
	}
	return execRegime(ctx, "update regime options", true,
		"UPDATE taxpayer_regimes SET options = ?::jsonb WHERE taxpayer_id = ? AND regime_code = ?",
		string(encoded), taxpayerID, code)
}

// EndRegime implements usecase.RegimeStore.
func (Store) EndRegime(ctx context.Context, taxpayerID, code string, end domain.Period) error {
	return execRegime(ctx, "end regime", true,
		"UPDATE taxpayer_regimes SET end_period = ? WHERE taxpayer_id = ? AND regime_code = ?",
		end.String(), taxpayerID, code)
}

// execRegime runs one taxpayer_regimes write; for an update (mustMatch) no
// matching row is ErrRegimeNotHeld.
func execRegime(ctx context.Context, op string, mustMatch bool, statement string, args ...any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Exec(statement, args...)
	if result.Error != nil {
		return translate(op, result.Error)
	}
	if mustMatch && result.RowsAffected == 0 {
		return usecase.ErrRegimeNotHeld
	}
	return nil
}

// NoRecords implements usecase.RegimeRecords until M3 and M4 add the
// incomes and expenses tables: no regime holds records yet. Those modules
// replace it with an adapter that reads their tables.
type NoRecords struct{}

var _ usecase.RegimeRecords = NoRecords{}

// LatestRecordPeriod implements usecase.RegimeRecords.
func (NoRecords) LatestRecordPeriod(context.Context, string, string) (domain.Period, error) {
	return domain.Period{}, nil
}
