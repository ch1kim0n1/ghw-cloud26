import { request } from "./apiClient";
import type { Product } from "../types/Product";

export function listProducts(): Promise<{ products: Product[] }> {
  return request<{ products: Product[] }>("/api/products");
}

export function createProduct(formData: FormData): Promise<Product> {
  return request<Product>("/api/products", {
    method: "POST",
    body: formData,
  });
}
