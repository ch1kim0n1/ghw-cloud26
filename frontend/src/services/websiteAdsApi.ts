import { request } from "./apiClient";
import type { WebsiteAdJob } from "../types/WebsiteAd";

export interface CreateWebsiteAdPayload {
  product_id?: string;
  product_name?: string;
  product_description?: string;
  article_headline: string;
  article_body: string;
  brand_style?: string;
}

export function listWebsiteAds(): Promise<{ jobs: WebsiteAdJob[] }> {
  return request<{ jobs: WebsiteAdJob[] }>("/api/website-ads");
}

export function createWebsiteAd(payload: CreateWebsiteAdPayload): Promise<WebsiteAdJob> {
  return request<WebsiteAdJob>("/api/website-ads", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}
