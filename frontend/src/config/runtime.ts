export const runtimeConfig = {
  showcaseMode: import.meta.env.VITE_SHOWCASE_MODE === "true",
  routerMode: import.meta.env.VITE_ROUTER_MODE === "hash" ? "hash" : "browser",
};
