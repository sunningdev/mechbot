import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { captureRef } from "./api";
import { useLang } from "./i18n";
import Landing from "./pages/Landing";
import { BlogList, BlogPost } from "./pages/Blog";
import ConsoleLayout from "./pages/ConsoleLayout";
import Dashboard from "./pages/Dashboard";
import LiqMap from "./pages/LiqMap";
import Membership from "./pages/Membership";
import Referral from "./pages/Referral";
import "@fontsource-variable/inter";
import "@fontsource/jetbrains-mono/400.css";
import "@fontsource/jetbrains-mono/600.css";
import "@fontsource/barlow-condensed/latin-800.css";
import "./styles.css";
import "./liqmap.css";

captureRef();

function App() {
  useLang();
  return (
    <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/blog" element={<BlogList />} />
        <Route path="/blog/:slug" element={<BlogPost />} />
        <Route path="/app" element={<ConsoleLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="map" element={<LiqMap />} />
          <Route path="pro" element={<Membership />} />
          <Route path="referral" element={<Referral />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
