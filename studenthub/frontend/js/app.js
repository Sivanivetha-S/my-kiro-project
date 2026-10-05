/**
 * app.js — StudentHub SPA shell and hash-based router.
 *
 * Layout change: vertical sidebar replaced with horizontal top navigation.
 * All routing logic is unchanged.
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
// Router table
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

let currentController = null;

// ----------------------------------------------------------------
// Shell visibility — hides the entire app layout on the landing page
// ----------------------------------------------------------------

function setShellVisible(visible) {
  const layout = document.getElementById('layout');
  if (!layout) return;
  layout.classList.toggle('shell--hidden', !visible);
}

// ----------------------------------------------------------------
// Router — unchanged from previous implementation
// ----------------------------------------------------------------

async function navigate() {
  currentController?.abort();
  currentController = new AbortController();

  const raw      = window.location.hash.replace(/^#\/?/, '') || 'landing';
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

  setShellVisible(!isLanding);
  setActiveNav(routeKey);

  // ── Landing page ──────────────────────────────────────────────
  if (isLanding) {
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

  // ── Leaving landing — hide it ─────────────────────────────────
  const landingRoot = document.getElementById('landing-root');
  if (landingRoot && !landingRoot.hidden) {
    const landingMod = await import('./pages/landing.js').catch(() => null);
    landingMod?.hide?.();
  }

  // ── App shell routes ──────────────────────────────────────────
  const app = document.getElementById('app');
  if (!app) return;

  app.innerHTML = `<div class="page-loading" aria-live="polite" aria-label="Loading…">
    <div class="spinner"></div>
  </div>`;

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
// Marks the correct link in BOTH the desktop list and mobile menu.
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
// Top navigation — mobile hamburger menu toggle
// Replaces the old initSidebar() function.
// ----------------------------------------------------------------

function initTopNav() {
  const menuBtn    = document.getElementById('menu-btn');
  const mobileMenu = document.getElementById('mobile-menu');

  if (!menuBtn || !mobileMenu) return;

  function openMenu() {
    mobileMenu.hidden = false;
    menuBtn.setAttribute('aria-expanded', 'true');
    // Focus first link for keyboard accessibility
    const firstLink = mobileMenu.querySelector('.nav-link');
    firstLink?.focus();
  }

  function closeMenu() {
    mobileMenu.hidden = true;
    menuBtn.setAttribute('aria-expanded', 'false');
    menuBtn.focus();
  }

  function toggleMenu() {
    if (mobileMenu.hidden) {
      openMenu();
    } else {
      closeMenu();
    }
  }

  // Toggle on hamburger click
  menuBtn.addEventListener('click', toggleMenu);

  // Close when any nav link is clicked
  mobileMenu.querySelectorAll('.nav-link').forEach((link) => {
    link.addEventListener('click', closeMenu);
  });

  // Close on Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !mobileMenu.hidden) closeMenu();
  });

  // Close when clicking outside the topnav
  document.addEventListener('click', (e) => {
    const topnav = document.getElementById('topnav');
    if (topnav && !topnav.contains(e.target) && !mobileMenu.hidden) {
      closeMenu();
    }
  });

  // Close if viewport expands past mobile breakpoint
  const mq = window.matchMedia('(min-width: 768px)');
  mq.addEventListener('change', (e) => {
    if (e.matches && !mobileMenu.hidden) closeMenu();
  });
}

// ----------------------------------------------------------------
// Bootstrap
// ----------------------------------------------------------------

document.addEventListener('DOMContentLoaded', () => {
  initTopNav();
  navigate();
});

window.addEventListener('hashchange', navigate);
