/**
 * app.js — StudentHub SPA shell
 *
 * Responsibilities for T-05 (placeholder):
 *   - Renders "Welcome to StudentHub" into #app on load.
 *   - Wires up the mobile sidebar toggle (hamburger / close / overlay).
 *   - Highlights the active nav link based on the current URL hash.
 *
 * T-18 will replace the render() function with a full hash-based router
 * that loads page modules dynamically.
 */

// ----------------------------------------------------------------
// Active nav link highlighting
// ----------------------------------------------------------------

function setActiveNav() {
  const hash = window.location.hash.replace('#', '') || 'dashboard';
  // Match the route prefix (e.g. "students/123" → "students")
  const route = hash.split('/')[0];

  document.querySelectorAll('.nav-link').forEach((link) => {
    const linkRoute = link.dataset.route;
    link.classList.toggle('is-active', linkRoute === route);
    link.setAttribute('aria-current', linkRoute === route ? 'page' : 'false');
  });
}

// ----------------------------------------------------------------
// Page renderer (T-05 placeholder — replaced by router in T-18)
// ----------------------------------------------------------------

function render() {
  const app = document.getElementById('app');
  if (!app) return;

  // T-05 acceptance criterion: show "Welcome to StudentHub" in #app.
  app.innerHTML = `
    <div style="text-align: center; padding: 4rem 2rem;">
      <h1 style="font-size: 2rem; font-weight: 700; color: var(--color-neutral-900); margin-bottom: 1rem;">
        Welcome to StudentHub
      </h1>
      <p style="color: var(--color-neutral-700); max-width: 480px; margin: 0 auto;">
        A student management system for tracking students, courses, marks,
        and attendance. Use the navigation on the left to get started.
      </p>
    </div>
  `;
}

// ----------------------------------------------------------------
// Mobile sidebar toggle
// ----------------------------------------------------------------

function initSidebar() {
  const sidebar    = document.getElementById('sidebar');
  const overlay    = document.getElementById('sidebar-overlay');
  const menuBtn    = document.getElementById('menu-btn');
  const closeBtn   = document.getElementById('sidebar-close');

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

  // Close sidebar when a nav link is clicked on mobile
  sidebar.querySelectorAll('.nav-link').forEach((link) => {
    link.addEventListener('click', () => {
      if (window.innerWidth < 1024) closeSidebar();
    });
  });

  // Close sidebar on Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && sidebar.classList.contains('is-open')) {
      closeSidebar();
    }
  });
}

// ----------------------------------------------------------------
// Bootstrap
// ----------------------------------------------------------------

document.addEventListener('DOMContentLoaded', () => {
  initSidebar();
  render();
  setActiveNav();
});

// Re-highlight nav on hash change (T-18 router will also call setActiveNav)
window.addEventListener('hashchange', () => {
  setActiveNav();
  // T-18 will replace render() with a router call here.
  render();
});
