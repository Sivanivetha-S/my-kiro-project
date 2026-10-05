/**
 * app.js — StudentHub SPA shell and hash-based router.
 *
 * Routing table (design.md § 6.1):
 *   #dashboard           → pages/dashboard.js
 *   #students            → pages/students.js  (list)
 *   #students/new        → pages/students.js  (add form)
 *   #students/:id        → pages/student-detail.js
 *   #students/:id/edit   → pages/students.js  (edit form)
 *   #courses             → pages/courses.js
 *   #marks               → pages/marks.js
 *   #attendance          → pages/attendance.js
 */

// ----------------------------------------------------------------
// Router
// ----------------------------------------------------------------

const routes = [
  { pattern: /^dashboard$/,            page: 'dashboard',       params: () => ({}) },
  { pattern: /^students\/new$/,         page: 'students',        params: () => ({ mode: 'new' }) },
  { pattern: /^students\/(\d+)\/edit$/, page: 'students',        params: (m) => ({ mode: 'edit', id: m[1] }) },
  { pattern: /^students\/(\d+)$/,       page: 'student-detail',  params: (m) => ({ id: m[1] }) },
  { pattern: /^students$/,              page: 'students',        params: () => ({}) },
  { pattern: /^courses$/,              page: 'courses',         params: () => ({}) },
  { pattern: /^marks$/,                page: 'marks',           params: () => ({}) },
  { pattern: /^attendance$/,           page: 'attendance',      params: () => ({}) },
];

let currentController = null; // AbortController for in-flight navigations

async function navigate() {
  // Cancel any previous page load.
  currentController?.abort();
  currentController = new AbortController();

  const raw = window.location.hash.replace(/^#\/?/, '') || 'dashboard';
  setActiveNav(raw.split('/')[0]);

  let matched = null;
  let params = {};
  for (const route of routes) {
    const m = raw.match(route.pattern);
    if (m) {
      matched = route;
      params = route.params(m);
      break;
    }
  }

  const app = document.getElementById('app');
  if (!app) return;

  app.innerHTML = '<div class="page-loading" aria-live="polite" aria-label="Loading…"><div class="spinner"></div></div>';

  if (!matched) {
    app.innerHTML = '<div class="content-card"><h1 class="page-title">404 — Page Not Found</h1><p><a href="#dashboard">Go to dashboard</a></p></div>';
    return;
  }

  try {
    const mod = await import(`./pages/${matched.page}.js`);
    if (currentController.signal.aborted) return;
    await mod.render(app, params);
  } catch (err) {
    if (currentController.signal.aborted) return;
    console.error('Page load error:', err);
    app.innerHTML = `<div class="content-card error-state"><p>Failed to load page. <a href="#dashboard">Go home</a></p></div>`;
  }

  // Move focus to main for screen readers.
  app.focus();
}

// ----------------------------------------------------------------
// Active nav link highlighting
// ----------------------------------------------------------------

function setActiveNav(route) {
  document.querySelectorAll('.nav-link').forEach((link) => {
    const linkRoute = link.dataset.route;
    const isActive = linkRoute === route;
    link.classList.toggle('is-active', isActive);
    link.setAttribute('aria-current', isActive ? 'page' : 'false');
  });
}

// ----------------------------------------------------------------
// Mobile sidebar toggle
// ----------------------------------------------------------------

function initSidebar() {
  const sidebar  = document.getElementById('sidebar');
  const overlay  = document.getElementById('sidebar-overlay');
  const menuBtn  = document.getElementById('menu-btn');
  const closeBtn = document.getElementById('sidebar-close');

  if (!sidebar || !overlay || !menuBtn || !closeBtn) return;

  function openSidebar() {
    sidebar.classList.add('is-open');
    overlay.classList.add('is-visible');
    overlay.removeAttribute('aria-hidden');
    menuBtn.setAttribute('aria-expanded', 'true');
    closeBtn.setAttribute('aria-expanded', 'true');
    closeBtn.focus();
  }

  function closeSidebar() {
    sidebar.classList.remove('is-open');
    overlay.classList.remove('is-visible');
    overlay.setAttribute('aria-hidden', 'true');
    menuBtn.setAttribute('aria-expanded', 'false');
    closeBtn.setAttribute('aria-expanded', 'false');
    menuBtn.focus();
  }

  menuBtn.addEventListener('click', openSidebar);
  closeBtn.addEventListener('click', closeSidebar);
  overlay.addEventListener('click', closeSidebar);

  sidebar.querySelectorAll('.nav-link').forEach((link) => {
    link.addEventListener('click', () => {
      if (window.innerWidth < 1024) closeSidebar();
    });
  });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && sidebar.classList.contains('is-open')) closeSidebar();
  });
}

// ----------------------------------------------------------------
// Bootstrap
// ----------------------------------------------------------------

document.addEventListener('DOMContentLoaded', () => {
  initSidebar();
  navigate();
});

window.addEventListener('hashchange', navigate);
