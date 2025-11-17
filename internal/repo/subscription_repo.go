package repo

import (
	"errors"
	"strconv"
	"time"

	"subscriptions/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(s *model.Subscription) error {
	if s == nil {
		return errors.New("nil subscription")
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	_, err := r.db.NamedExec(`
		INSERT INTO subscriptions
		  (id, service_name, price, user_id, start_month, start_year, end_month, end_year, created_at, updated_at)
		VALUES
		  (:id, :service_name, :price, :user_id, :start_month, :start_year, :end_month, :end_year, :created_at, :updated_at)
	`, s)
	return err
}

func (r *SubscriptionRepo) List() ([]model.Subscription, error) {
	var out []model.Subscription
	err := r.db.Select(&out, `
		SELECT id, service_name, price, user_id, start_month, start_year, end_month, end_year, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
	`)
	return out, err
}

func (r *SubscriptionRepo) GetByID(id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.Get(&s, `
		SELECT id, service_name, price, user_id, start_month, start_year, end_month, end_year, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriptionRepo) Update(s *model.Subscription) error {
	if s == nil {
		return errors.New("nil subscription")
	}
	s.UpdatedAt = time.Now().UTC()
	_, err := r.db.NamedExec(`
		UPDATE subscriptions
		SET service_name = :service_name, price = :price, user_id = :user_id,
		    start_month = :start_month, start_year = :start_year,
		    end_month = :end_month, end_year = :end_year,
		    updated_at = :updated_at
		WHERE id = :id
	`, s)
	return err
}

func (r *SubscriptionRepo) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM subscriptions WHERE id = $1", id)
	return err
}

func (r *SubscriptionRepo) Summary(fromMonth, fromYear, toMonth, toYear int, userID *uuid.UUID, serviceName *string) (int, error) {
	if toYear < fromYear || (toYear == fromYear && toMonth < fromMonth) {
		return 0, errors.New("invalid period")
	}

	
	q := `SELECT id, service_name, price, user_id, start_month, start_year, end_month, end_year FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if userID != nil {
		q += " AND user_id = $" + itoa(argIdx)
		args = append(args, *userID)
		argIdx++
	}
	if serviceName != nil && *serviceName != "" {
		q += " AND service_name ILIKE $" + itoa(argIdx)
		args = append(args, "%"+*serviceName+"%")
		argIdx++
	}

	var subs []model.Subscription
	if err := r.db.Select(&subs, q, args...); err != nil {
		return 0, err
	}

	
	toIndex := func(month, year int) int {
		return year*12 + month
	}
	fromIdx := toIndex(fromMonth, fromYear)
	toIdx := toIndex(toMonth, toYear)

	total := 0
	for _, s := range subs {
		subStartIdx := toIndex(s.StartMonth, s.StartYear)
		subEndIdx := toIndex(12, 9999) 
		if s.EndMonth != nil && s.EndYear != nil {
			subEndIdx = toIndex(*s.EndMonth, *s.EndYear)
		}
		
		start := max(subStartIdx, fromIdx)
		end := min(subEndIdx, toIdx)
		if end >= start {
			months := end - start + 1
			total += months * s.Price
		}
	}
	return total, nil
}

func itoa(i int) string {
	
	return strconv.Itoa(i)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
