import { request } from "./apiClient";
import type { Campaign } from "../types/Campaign";

export function createCampaign(formData: FormData): Promise<Campaign> {
  return request<Campaign>("/api/campaigns", {
    method: "POST",
    body: formData,
  });
}

export function getCampaign(campaignId: string): Promise<Campaign> {
  return request<Campaign>(`/api/campaigns/${campaignId}`);
}
