import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, HashRouter } from "react-router-dom";
import App from "./App";
import { PublicMotionProvider } from "./components/PublicMotionProvider";
import { runtimeConfig } from "./config/runtime";
import "./styles/app.css";
import "./styles/components.css";

const Router = runtimeConfig.routerMode === "hash" ? HashRouter : BrowserRouter;

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <Router>
      <PublicMotionProvider />
      <App />
    </Router>
  </React.StrictMode>,
);
