CREATE TABLE users (
    login VARCHAR(50) PRIMARY KEY,
    password TEXT NOT NULL  -- bcrypt
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL CHECK (price > 0),
    image_url VARCHAR(500) NOT NULL,
    owner VARCHAR(50) NOT NULL REFERENCES users(login) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_owner ON posts(owner);
CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_posts_price ON posts(price);