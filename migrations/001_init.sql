CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    balance DECIMAL(15,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_account INT REFERENCES accounts(id),
    to_account INT REFERENCES accounts(id),
    amount DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS cards (
    id SERIAL PRIMARY KEY,
    account_id INT REFERENCES accounts(id),
    card_number_enc TEXT NOT NULL,
    card_expiry_enc TEXT NOT NULL,
    cvv_hash TEXT NOT NULL,
    hmac TEXT NOT NULL,
    owner_id INT REFERENCES users(id)
    );

CREATE TABLE IF NOT EXISTS credits (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    account_id INT REFERENCES accounts(id),
    amount DECIMAL(15,2),
    rate DECIMAL(5,2),
    term_months INT,
    monthly_payment DECIMAL(15,2),
    status VARCHAR(20) DEFAULT 'active'
    );

CREATE TABLE IF NOT EXISTS payment_schedules (
    id SERIAL PRIMARY KEY,
    credit_id INT REFERENCES credits(id),
    due_date DATE,
    amount DECIMAL(15,2),
    paid BOOLEAN DEFAULT FALSE
    );