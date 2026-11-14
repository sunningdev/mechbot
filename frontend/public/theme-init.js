(function () {
  var t;
  try {
    t = localStorage.getItem("mechbot-theme");
  } catch (e) {}
  if (t !== "dark" && t !== "light") t = matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
  document.documentElement.dataset.theme = t;
  if (t === "light") document.querySelector('meta[name="theme-color"]').setAttribute("content", "#f5f7fa");
})();
