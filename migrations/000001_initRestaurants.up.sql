CREATE TABLE IF NOT EXISTS restaurants(
	id SERIAL PRIMARY KEY, 
	title VARCHAR(100) NOT NULL,
	description VARCHAR(300),
	address VARCHAR(300) NOT NULL,
	rating DECIMAL(3,2) NOT NULL,
	UNIQUE(title, address)
);