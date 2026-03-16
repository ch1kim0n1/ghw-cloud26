export interface VideoUploadFormState {
  campaignName: string;
  brandName: string;
  videoFile: File | null;
}

export interface WebsiteUploadFormState {
  productMode: "existing" | "custom";
  productId: string;
  productName: string;
  productDescription: string;
  articleHeadline: string;
  articleBody: string;
  brandStyle: string;
}
