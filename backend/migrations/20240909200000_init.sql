-- to migrate with using goose run in CMD:
-- $HOME/.goose/bin/goose -dir migrations postgres "user=postgres dbname=tenders sslmode=disable password=12345678 host=localhost" up

-- +goose Up

-- +goose StatementBegin
CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    type organization_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES employee(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    status tender_status NOT NULL,
    service_type service_type ARRAY NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Cancelled'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE author_type AS ENUM (
    'User',
    'Organization'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE bid (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    status bid_status NOT NULL,
    author_type author_type NOT NULL,
    author_id UUID NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE decision AS ENUM (-- +goose StatementBegin
    'Approved',
    'Rejected'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE bid_decision (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    decision decision NOT NULL
)
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE bid_decision;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE decision;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE bid;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE author_type;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE bid_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE tender;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE service_type;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE tender_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE organization_responsible;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE organization;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE organization_type;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE employee;
-- +goose StatementEnd