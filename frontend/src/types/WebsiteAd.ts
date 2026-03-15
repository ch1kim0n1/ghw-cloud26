export interface WebsiteAdJob {
  id: string;
  product_id?: string;
  product_name: string;
  product_description?: string;
  article_headline: string;
  article_body: string;
  brand_style?: string;
  prompt: string;
  status: string;
  banner_image_url?: string;
  vertical_image_url?: string;
  created_at: string;
  updated_at: string;
}
