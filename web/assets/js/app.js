const API = "/api/v1";

function getToken() {
  return localStorage.getItem("token");
}

function setToken(t) {
  localStorage.setItem("token", t);
}

function clearAuth() {
  localStorage.removeItem("token");
  localStorage.removeItem("user");
}

function getUser() {
  try { return JSON.parse(localStorage.getItem("user")); } catch { return null; }
}

function setUser(u) {
  localStorage.setItem("user", JSON.stringify(u));
}

function headers() {
  const h = { "Content-Type": "application/json" };
  const t = getToken();
  if (t) h["Authorization"] = "Bearer " + t;
  return h;
}

async function req(method, path, body) {
  const opt = { method, headers: headers() };
  if (body) opt.body = JSON.stringify(body);
  const r = await fetch(API + path, opt);
  const d = await r.json();
  if (r.status === 401) { clearAuth(); location.href = "/login"; }
  return d;
}

async function reqRaw(method, path, body) {
  const opt = { method, headers: headers() };
  if (body) opt.body = JSON.stringify(body);
  return fetch(API + path, opt);
}

function post(path, body) { return req("POST", path, body); }
function get(path) { return req("GET", path); }
function put(path, body) { return req("PUT", path, body); }
function del(path) { return req("DELETE", path); }

function qs(obj) {
  const p = new URLSearchParams();
  for (const k in obj) if (obj[k] !== undefined && obj[k] !== "" && obj[k] !== null) p.append(k, obj[k]);
  const s = p.toString();
  return s ? "?" + s : "";
}

function $(sel, root = document) { return root.querySelector(sel); }
function $$(sel, root = document) { return [...root.querySelectorAll(sel)]; }

function escapeHtml(s) {
  return String(s ?? "").replace(/[&<>"']/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
}

function toast(msg, type = "success") {
  const t = document.createElement("div");
  t.className = "alert alert-" + type;
  t.style.cssText = "position:fixed;top:70px;right:24px;z-index:2000;min-width:250px;box-shadow:0 4px 12px rgba(0,0,0,.15);";
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 3000);
}

function confirmDialog(msg) { return confirm(msg); }

function formatDate(d) {
  if (!d) return "-";
  const t = new Date(d);
  const pad = n => String(n).padStart(2, "0");
  return `${t.getFullYear()}-${pad(t.getMonth()+1)}-${pad(t.getDate())} ${pad(t.getHours())}:${pad(t.getMinutes())}`;
}

function formatDuration(sec) {
  if (!sec) return "00:00";
  sec = Math.floor(sec);
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  const pad = n => String(n).padStart(2, "0");
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}

function formatSize(bytes) {
  if (!bytes) return "0 B";
  const u = ["B","KB","MB","GB"];
  let i = 0; let v = bytes;
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
  return v.toFixed(v < 10 && i > 0 ? 2 : 1) + " " + u[i];
}

function renderNav(active) {
  const user = getUser();
  const nav = $$(".nav-links")[0];
  if (!nav || !user) return;
  const items = [
    { k: "dashboard", href: "/dashboard", label: "仪表盘" },
    { k: "channels", href: "/channels", label: "频道" },
    { k: "episodes", href: "/episodes", label: "节目" },
    { k: "stats", href: "/stats", label: "统计" },
    { k: "rss-view", href: "/rss-view", label: "RSS" },
  ];
  nav.innerHTML = items.map(x =>
    `<a href="${x.href}" class="${active === x.k ? "active":""}">${x.label}</a>`
  ).join("") + `
    <span style="color:var(--muted);font-size:14px">${escapeHtml(user.nickname || user.username)}</span>
    <button class="btn btn-sm btn-secondary" onclick="logout()">退出</button>`;
}

async function logout() {
  clearAuth();
  await post("/auth/logout");
  location.href = "/login";
}

function requireAuth() {
  if (!getToken()) { location.href = "/login"; return; }
}

function pageQuery(defaults = {}) {
  const p = new URLSearchParams(location.search);
  const q = {};
  for (const [k, v] of p.entries()) q[k] = v;
  return { ...defaults, ...q };
}

function updateQuery(params) {
  const p = new URLSearchParams(location.search);
  for (const k in params) {
    if (params[k] === undefined || params[k] === null || params[k] === "") p.delete(k);
    else p.set(k, params[k]);
  }
  const s = p.toString();
  history.replaceState({}, "", s ? "?" + s : location.pathname);
}
