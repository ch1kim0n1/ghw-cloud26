import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import { PublicMotionProvider } from "./components/PublicMotionProvider";
import "./styles/app.css";
import "./styles/components.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <PublicMotionProvider />
      <App />
    </BrowserRouter>
  </React.StrictMode>,
);
