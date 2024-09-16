package storage

import (
	"backend/config"
	"backend/entities"
	"backend/entities/decision"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"time"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(cfg *config.Config) *Storage {
	return &Storage{db: cfg.GetDB()}
}

func (s Storage) CreateTender(
	name string,
	description string,
	serviceType []string,
	organizationId string,
) (entities.Tender, error) {
	query := "INSERT INTO tender " +
		"(name, description, service_type, organization_id, status, version, created_at, updated_at)" +
		" VALUES ($1, $2, $3, $4, 'Created', 1, $5, $5) RETURNING id"
	var insertedId string
	creationTime := time.Now().UTC()
	err := s.db.QueryRow(
		query,
		name,
		description,
		serviceType,
		organizationId,
		creationTime,
	).Scan(&insertedId)
	if err != nil {
		return entities.Tender{}, err
	}
	return entities.Tender{
		Id:             insertedId,
		Name:           name,
		Description:    description,
		ServiceType:    serviceType,
		OrganizationId: organizationId,
		CreatedAt:      creationTime,
		UpdatedAt:      creationTime,
		Status:         "Created",
		Version:        1,
	}, nil
}

func (s Storage) FilterTenders(
	limit int,
	offset int,
	serviceType []string,
) ([]entities.Tender, error) {
	filters := "status='Published'"
	for _, item := range serviceType {
		filters += fmt.Sprintf("AND '%s'=ANY(service_type)", item)
	}
	query := "SELECT id, name, description, status, service_type, version, created_at, updated_at, organization_id " +
		"FROM tender WHERE " + filters + " ORDER BY name OFFSET $1 LIMIT $2"
	rows, err := s.db.Query(query, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tenders := make([]entities.Tender, 0)
	for rows.Next() {
		var tender entities.Tender
		err := rows.Scan(
			&tender.Id,
			&tender.Name,
			&tender.Description,
			&tender.Status,
			pq.Array(&tender.ServiceType),
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt,
			&tender.OrganizationId,
		)
		if err != nil {
			return nil, err
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (s Storage) GetUserId(username string) (string, error) {
	query := "SELECT id FROM employee WHERE username=$1"
	var id string
	err := s.db.QueryRow(query, username).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s Storage) FilterUsersTenders(
	limit int,
	offset int,
	userId string,
) ([]entities.Tender, error) {
	query := `
SELECT
	t.id,
	t.name,
	t.description,
	t.status,
	t.service_type,
	t.version,
	t.created_at,
	t.updated_at,
	t.organization_id
FROM tender AS t
JOIN organization_responsible AS o
ON t.organization_id=o.organization_id
WHERE o.user_id=$1
ORDER BY t.id
OFFSET $2
LIMIT $3
	`
	rows, err := s.db.Query(query, userId, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tenders := make([]entities.Tender, 0)
	for rows.Next() {
		var tender entities.Tender
		err := rows.Scan(
			&tender.Id,
			&tender.Name,
			&tender.Description,
			&tender.Status,
			pq.Array(&tender.ServiceType),
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt,
			&tender.OrganizationId,
		)
		if err != nil {
			return nil, err
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (s Storage) CheckOrganizationResponsible(
	userId string,
	organizationId string,
) (bool, error) {
	query := "SELECT COUNT(*) FROM organization_responsible WHERE user_id=$1 AND organization_id=$2"
	var count int
	err := s.db.QueryRow(query, userId, organizationId).Scan(&count)
	return count > 0, err
}

func (s Storage) GetTender(id string) (entities.Tender, error) {
	query := "SELECT name, description, status, service_type, version, created_at, updated_at, organization_id " +
		"FROM tender WHERE id=$1"
	var tender entities.Tender
	tender.Id = id
	err := s.db.QueryRow(query, id).Scan(
		&tender.Name,
		&tender.Description,
		&tender.Status,
		pq.Array(&tender.ServiceType),
		&tender.Version,
		&tender.CreatedAt,
		&tender.UpdatedAt,
		&tender.OrganizationId,
	)
	return tender, err
}

func (s Storage) PatchTender(
	id string,
	name *string,
	description *string,
	status *string,
	serviceType []string,
) (entities.Tender, error) {
	if name == nil && description == nil && status == nil && serviceType == nil {
		return s.GetTender(id)
	}
	tender, err := s.GetTender(id)
	if err != nil {
		return tender, err
	}
	query := sq.Update("tender")
	query = query.Set("version", tender.Version+1)
	query = query.Set("updated_at", time.Now().UTC())
	if name != nil {
		query = query.Set("name", name)
	}
	if description != nil {
		query = query.Set("description", description)
	}
	if status != nil {
		query = query.Set("status", status)
	}
	if serviceType != nil {
		query = query.Set("service_type", serviceType)
	}
	query = query.Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return entities.Tender{}, err
	}
	_, err = s.db.Exec(sqlQuery, args...)
	if err != nil {
		return entities.Tender{}, err
	}
	return s.GetTender(id)
}

func pointerToSQLNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{"", false}
	}
	return sql.NullString{*s, true}
}

func (s Storage) GetOrganization(id string) (entities.Organization, error) {
	query := "SELECT name, description, type, created_at, updated_at FROM organization WHERE id=$1"
	var org entities.Organization
	err := s.db.QueryRow(query, id).Scan(
		&org.Name,
		&org.Description,
		&org.Type,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return entities.Organization{}, err
	}
	org.Id = id
	return org, nil
}

func (s Storage) GetUser(id string) (entities.Employee, error) {
	query := "SELECT username, first_name, last_name, created_at, updated_at FROM employee WHERE id=$1"
	var user entities.Employee
	err := s.db.QueryRow(query, id).Scan(
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return entities.Employee{}, err
	}
	user.Id = id
	return user, nil
}

func (s Storage) CreateBid(
	name string,
	description string,
	authorType string,
	authorId string,
	tenderId string,
) (entities.Bid, error) {
	query := "INSERT INTO bid (name, description, status, author_type, author_id, version, created_at, updated_at, tender_id) " +
		"VALUES ($1, $2, 'Created', $3, $4, 1, $5, $6, $7) RETURNING id"
	var insertedId string
	creationTime := time.Now().UTC()
	err := s.db.QueryRow(
		query,
		name,
		description,
		authorType,
		authorId,
		creationTime,
		creationTime,
		tenderId,
	).Scan(&insertedId)
	if err != nil {
		return entities.Bid{}, err
	}
	return entities.Bid{
		Id:          insertedId,
		TenderId:    tenderId,
		Name:        name,
		Description: description,
		Status:      "Created",
		AuthorType:  authorType,
		AuthorId:    authorId,
		Version:     1,
		CreatedAt:   creationTime,
		UpdatedAt:   creationTime,
	}, nil
}

func (s Storage) GetMyBids(
	userId string,
	limit int,
	offset int,
) ([]entities.Bid, error) {
	query := "SELECT id, tender_id, name, description, status, version, created_at, updated_at " +
		"FROM bid WHERE author_type='User' AND author_id=$1 ORDER BY name LIMIT $2 OFFSET $3"
	rows, err := s.db.Query(query, userId, limit, offset)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	bids := make([]entities.Bid, 0)
	for rows.Next() {
		var bid entities.Bid
		err := rows.Scan(
			&bid.Id,
			&bid.TenderId,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.Version,
			&bid.CreatedAt,
			&bid.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bid.AuthorId = userId
		bid.AuthorType = "User"
		bids = append(bids, bid)
	}
	return bids, nil
}

func (s Storage) GetBidsByTender(
	tenderId string,
	limit int,
	offset int,
) ([]entities.Bid, error) {
	query := "SELECT id, name, description, status, version, author_type, author_id, created_at, updated_at " +
		"FROM bid WHERE tender_id=$1 ORDER BY name LIMIT $2 OFFSET $3"
	rows, err := s.db.Query(query, tenderId, limit, offset)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	bids := make([]entities.Bid, 0)
	for rows.Next() {
		var bid entities.Bid
		err := rows.Scan(
			&bid.Id,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.Version,
			&bid.AuthorType,
			&bid.AuthorId,
			&bid.CreatedAt,
			&bid.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bid.TenderId = tenderId
		bids = append(bids, bid)
	}
	return bids, nil
}

func (s Storage) GetBid(id string) (entities.Bid, error) {
	query := "SELECT tender_id, name, description, status, author_type, author_id, version, created_at, updated_at " +
		"FROM bid WHERE id=$1"
	var bid entities.Bid
	err := s.db.QueryRow(query, id).Scan(
		&bid.TenderId,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.AuthorType,
		&bid.AuthorId,
		&bid.Version,
		&bid.CreatedAt,
		&bid.UpdatedAt,
	)
	if err != nil {
		return entities.Bid{}, err
	}
	bid.Id = id
	return bid, nil
}

func (s Storage) PatchBid(
	id string,
	name *string,
	description *string,
	status *string,
) (entities.Bid, error) {
	if name == nil && description == nil && status == nil {
		return s.GetBid(id)
	}
	bid, err := s.GetBid(id)
	if err != nil {
		return bid, err
	}
	query := sq.Update("bid")
	query = query.Set("version", bid.Version+1)
	query = query.Set("updated_at", time.Now().UTC())
	if name != nil {
		query = query.Set("name", name)
	}
	if description != nil {
		query = query.Set("description", description)
	}
	if status != nil {
		query = query.Set("status", status)
	}
	query = query.Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return entities.Bid{}, err
	}
	_, err = s.db.Exec(sqlQuery, args...)
	if err != nil {
		return entities.Bid{}, err
	}
	return s.GetBid(id)
}

func (s Storage) GetDecision(bidId string) (string, error) {
	query := "SELECT decision FROM bid_decision WHERE bid_id=$1"
	rows, err := s.db.Query(query, bidId)
	defer rows.Close()
	if err != nil {
		return "", err
	}
	for rows.Next() {
		var decision string
		err := rows.Scan(&decision)
		if err != nil {
			return "", err
		}
		return decision, nil
	}
	return decision.UNKNOWN, nil
}

func (s Storage) SetDecision(bidId string, decision string) error {
	query := "INSERT INTO bid_decision (bid_id, decision) VALUES ($1, $2)"
	_, err := s.db.Exec(query, bidId, decision)
	return err
}
