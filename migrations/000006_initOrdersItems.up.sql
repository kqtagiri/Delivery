CREATE TABLE IF NOT EXISTS orders_items(
	id SERIAL PRIMARY KEY,
	number INT REFERENCES orders(number) NOT NULL,
	menu_id INT REFERENCES menu(id),
	title VARCHAR(100) NOT NULL,
	rest_title VARCHAR(100) NOT NULL,
	composition VARCHAR(300) NOT NULL,
	cost DECIMAL(9,2) NOT NULL
);