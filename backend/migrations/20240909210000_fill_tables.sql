-- +goose Up

-- +goose StatementBegin
INSERT INTO employee (id, username, first_name, last_name) VALUES ('11111111-1111-1111-1111-111111111111', 'test_user_1', 'ivan', 'ivanov');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO employee (id, username, first_name, last_name) VALUES ('22222222-2222-2222-2222-222222222222', 'test_user_2', 'petr', 'petrov');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO employee (id, username, first_name, last_name) VALUES ('33333333-3333-3333-3333-333333333333', 'test_user_3', 'vlad', 'vladimirov');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO organization (id, name, description, type) VALUES ('11111111-1111-1111-1111-111111111111', 'test_ie', 'description', 'IE');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO organization (id, name, description, type) VALUES ('22222222-2222-2222-2222-222222222222', 'test_llc', 'description', 'LLC');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO organization_responsible(id, organization_id, user_id) VALUES ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO organization_responsible(id, organization_id, user_id) VALUES ('22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO organization_responsible(id, organization_id, user_id) VALUES ('33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333');
-- +goose StatementEnd