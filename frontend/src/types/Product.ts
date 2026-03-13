export interface Product {
  id: string;
  name: string;
  description?: string;
  category?: string;
  context_keywords?: string[];
  source_url?: string;
  image_path?: string;
  created_at: string;
}
