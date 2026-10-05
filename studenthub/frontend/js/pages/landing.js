/**
 * landing.js — StudentHub landing page.
 *
 * Renders into #landing-root (full-viewport, outside the app shell).
 * The "Enter StudentHub" CTA navigates to #dashboard.
 *
 * The hero background is a pure-CSS layered gradient that evokes a
 * clean, modern educational environment — no external image URL required,
 * so the page works offline and never has broken-image issues.
 */

export function render() {
  const root = document.getElementById('landing-root');
  if (!root) return;

  root.hidden        = false;
  root.removeAttribute('aria-hidden');

  root.innerHTML = `
    <div class="lp">

      <!-- ── Hero ── -->
      <section class="lp__hero" aria-labelledby="lp-heading">
        <div class="lp__hero-overlay" aria-hidden="true"></div>

        <!-- Decorative floating shapes -->
        <div class="lp__shapes" aria-hidden="true">
          <div class="lp__shape lp__shape--1"></div>
          <div class="lp__shape lp__shape--2"></div>
          <div class="lp__shape lp__shape--3"></div>
          <div class="lp__shape lp__shape--4"></div>
        </div>

        <div class="lp__hero-content">
          <!-- Brand -->
          <div class="lp__brand" aria-hidden="true">
            <span class="lp__brand-icon">🎓</span>
            <span class="lp__brand-name">StudentHub</span>
          </div>

          <!-- Eyebrow -->
          <p class="lp__eyebrow">Student Management System</p>

          <!-- Main heading -->
          <h1 class="lp__heading" id="lp-heading">
            Welcome to<br>StudentHub
          </h1>

          <!-- Supporting text -->
          <p class="lp__sub">
            A simple and smart platform to manage students, courses,
            marks and attendance — all in one place.
          </p>

          <!-- CTA -->
          <a
            href="#dashboard"
            class="lp__cta"
            role="button"
            aria-label="Enter StudentHub and go to the dashboard"
          >
            Enter StudentHub
            <span class="lp__cta-arrow" aria-hidden="true">→</span>
          </a>
        </div>

        <!-- Scroll hint -->
        <div class="lp__scroll-hint" aria-hidden="true">&#8964;</div>
      </section>

      <!-- ── Feature strip ── -->
      <section class="lp__features" aria-label="Key features">
        <div class="lp__features-inner">

          <div class="lp__feature">
            <div class="lp__feature-icon" aria-hidden="true">👥</div>
            <h2 class="lp__feature-title">Student Records</h2>
            <p class="lp__feature-desc">Add, edit, search and filter students across departments and years.</p>
          </div>

          <div class="lp__feature">
            <div class="lp__feature-icon" aria-hidden="true">📚</div>
            <h2 class="lp__feature-title">Course Management</h2>
            <p class="lp__feature-desc">Create courses, assign credits and enroll students in a few clicks.</p>
          </div>

          <div class="lp__feature">
            <div class="lp__feature-icon" aria-hidden="true">📝</div>
            <h2 class="lp__feature-title">Marks &amp; Grades</h2>
            <p class="lp__feature-desc">Record final marks and get instant letter grades calculated automatically.</p>
          </div>

          <div class="lp__feature">
            <div class="lp__feature-icon" aria-hidden="true">📅</div>
            <h2 class="lp__feature-title">Attendance Tracking</h2>
            <p class="lp__feature-desc">Track class attendance per course and spot low-attendance students instantly.</p>
          </div>

        </div>
      </section>

      <!-- ── Footer ── -->
      <footer class="lp__footer">
        <p>StudentHub &copy; 2026 &mdash; Student Management System</p>
      </footer>

    </div>
  `;
}

/**
 * Called by app.js when navigating away from the landing page.
 * Hides the landing root so the app shell becomes visible.
 */
export function hide() {
  const root = document.getElementById('landing-root');
  if (!root) return;
  root.hidden = true;
  root.setAttribute('aria-hidden', 'true');
  root.innerHTML = '';
}
