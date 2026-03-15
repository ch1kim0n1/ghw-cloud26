import { assetUrl } from "../utils/assetUrl";

export function BrandMark() {
  return <img className="brand-lockup__mark" src={assetUrl("logo.png")} alt="CAFAI logo" />;
}
