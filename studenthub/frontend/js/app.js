/**
 * app.js — StudentHub SPA shell and hash-based router.
 *
 * Routing table:
 *   #landing             → pages/landing.js   (entry point — shown first)
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
  { pattern: /^landing$/,               page: 'landing',        params: () => ({}) },
  { pattern: /^dashboard$/,             page: 'dashboard',      params: () => ({}) },
  { pattern: /^students\/new$/,          page: 'students',       params: () => ({ mode: 'new' }) },
  { pattern: /^students\/(\d+)\/edit$/,  page: 'students',       params: (m) => ({ mode: 'edit', id: m[1] }) },
  { pattern: /^students\/(\d+)$/,        page: 'student-detail', params: (m) => ({ id: m[1] }) },
  { pattern: /^students$/,               page: 'students',       params: () => ({}) },
  { pattern: /^courses$/,               page: 'courses',        params: () => ({}) },
  { pattern: /^marks$/,                 page: 'marks',          params: () => ({}) },
  { pattern: /^attendance$/,            page: 'attendance',     params: () => ({}) },
];

/** Routes that render inside the full app shell (sidebar + topbar). */
const APP_ROUTES = new Set(['dashboard', 'students', 'student-detail', 'courses', 'marks', 'attendance']);

let currentController = null;

// ----------------------------------------------------------------
// Shell visibility
// When on the landing page the sidebar/topbar shell is hidden so the
// landing page occupies the full viewport. On any app route it is shown.
// ----------------------------------------------------------------

function setShellVisible(visible) {
  const layout = document.getElementById('layout');
  if (!layout) return;
  if (visible) {
    layout.classList.remove('shell--hidden');
  } else {
    layout.classList.add('shell--hidden');
  }
}

// ----------------------------------------------------------------
// Router
// ----------------------------------------------------------------

async function navigate() {
  currentController?.abort();
  currentController = new AbortController();

  // Default route: show the landing page when no hash is present.
  const raw = window.location.hash.replace(/^#\/?/, '') || 'landing';
  const routeKey = raw.split('/')[0];

  let matched = null;
  let params  = {};
  for (const route of routes) {
    const m = raw.match(route.pattern);
    if (m) {
      matched = route;
      params  = route.params(m);
      break;
    }
  }

  const isLanding = matched?.page === 'landing';

  // Show/hide the app shell.
  setShellVisible(!isLanding);

  // Highlight the active nav link (only meaningful on app routes).
  setActiveNav(routeKey);

  if (isLanding) {
    // Landing page renders into its own full-page container, not #app.
    // First ensure the landing root is visible.
    const landingRoot = document.getElementById('landing-root');
    if (landingRoot) {
      landingRoot.hidden = false;
      landingRoot.removeAttribute('aria-hidden');
    }
    const mod = await import('./pages/landing.js');
    if (currentController.signal.aborted) return;
    mod.render();
    return;
  }

  // Leaving the landing page — hide it.
  const landingRoot = document.getElementById('landing-root');
  if (landingRoot && !landingRoot.hidden) {
    const landingMod = await import('./pages/landing.js').catch(() => null);
    landingMod?.hide?.();
  }

  // ---- App shell routes ----
  const app = document.getElementById('app');
  if (!app) return;

  app.innerHTML = '<div class="page-loading" aria-live="polite" aria-label="Loading…"><div class="spinner"></div></div>';

  if (!matched) {
    app.innerHTML = `<div class="content-card">
      <h1 class="page-title">404 — Page Not Found</h1>
      <p><a href="#dashboard">Go to dashboard</a></p>
    </div>`;
    return;
  }

  try {
    const mod = await import(`./pages/${matched.page}.js`);
    if (currentController.signal.aborted) return;
    await mod.render(app, params);
  } catch (err) {
    if (currentController.signal.aborted) return;
    console.error('Page load error:', err);
    app.innerHTML = `<div class="content-card error-state">
      <p>Failed to load page. <a href="#dashboard">Go home</a></p>
    </div>`;
  }

  app.focus();
}

// ----------------------------------------------------------------
// Active nav link highlighting
// ----------------------------------------------------------------

function setActiveNav(route) {
  document.querySelectorAll('.nav-link').forEach((link) => {
    const linkRoute = link.dataset.route;
    const isActive  = linkRoute === route;
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
