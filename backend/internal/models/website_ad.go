package models

type WebsiteAdJob struct {
	ID                 string `json:"id"`
	ProductID          string `json:"product_id,omitempty"`
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description,omitempty"`
	ArticleHeadline    string `json:"article_headline"`
	ArticleBody        string `json:"article_body"`
	BrandStyle         string `json:"brand_style,omitempty"`
	Prompt             string `json:"prompt"`
	Status             string `json:"status"`
	BannerImagePath    string `json:"-"`
	VerticalImagePath  string `json:"-"`
	BannerImageURL     string `json:"banner_image_url,omitempty"`
	VerticalImageURL   string `json:"vertical_image_url,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
