import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { adminPagePath } from "./adminPath";
import { captureRef } from "./api";
import SeoHead from "./components/SeoHead";
import { useLang } from "./i18n";
import Landing from "./pages/Landing";
import Admin from "./pages/Admin";
import { BlogList, BlogPost } from "./pages/Blog";
import ConsoleLayout from "./pages/ConsoleLayout";
import Legal from "./pages/Legal";
import Dashboard from "./pages/Dashboard";
import LiqMap from "./pages/LiqMap";
import Membership from "./pages/Membership";
import NotFound from "./pages/NotFound";
import Referral from "./pages/Referral";
import "@fontsource-variable/inter";
import "@fontsource/jetbrains-mono/400.css";
import "@fontsource/jetbrains-mono/600.css";
import "./styles.css";
import "./liqmap.css";

captureRef();

function App() {
  useLang();
  const adminRoute = adminPagePath();
  return (
    <>
      <SeoHead />
      <Routes>
        <Route path="/" element={<Landing />} />
        {adminRoute ? <Route path={adminRoute} element={<Admin />} /> : null}
        <Route path="/blog" element={<BlogList />} />
        <Route path="/blog/:slug" element={<BlogPost />} />
        <Route path="/terms" element={<Legal kind="terms" />} />
        <Route path="/privacy" element={<Legal kind="privacy" />} />
        <Route path="/app" element={<ConsoleLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="map" element={<LiqMap />} />
          <Route path="pro" element={<Membership />} />
          <Route path="referral" element={<Referral />} />
        </Route>
        <Route path="*" element={<NotFound />} />
      </Routes>
    </>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
