-- +goose Up
CREATE TABLE
    books (
        id VARCHAR(36) PRIMARY KEY,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL,
        title VARCHAR(255) NOT NULL,
        author VARCHAR(255) NOT NULL,
        price DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
        is_available BOOLEAN NOT NULL DEFAULT true
    );

-- +goose Down
DROP TABLE books;