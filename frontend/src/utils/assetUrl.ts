export function assetUrl(path: string): string {
  const normalizedPath = path.startsWith("/") ? path.slice(1) : path;
  return new URL(normalizedPath, import.meta.env.BASE_URL).toString();
}
