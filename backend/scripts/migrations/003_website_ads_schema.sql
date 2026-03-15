CREATE TABLE website_ad_jobs (
	id TEXT PRIMARY KEY,
	product_id TEXT REFERENCES products(id),
	product_name TEXT NOT NULL,
	product_description TEXT,
	article_headline TEXT NOT NULL,
	article_body TEXT NOT NULL,
	brand_style TEXT,
	prompt TEXT NOT NULL,
	status TEXT NOT NULL,
	banner_image_path TEXT,
	vertical_image_path TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX idx_website_ad_jobs_created_at ON website_ad_jobs(created_at DESC, id DESC);
