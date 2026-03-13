package models

type Product struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Category        string   `json:"category,omitempty"`
	ContextKeywords []string `json:"context_keywords,omitempty"`
	SourceURL       string   `json:"source_url,omitempty"`
	ImagePath       string   `json:"image_path,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}
