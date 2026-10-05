/**
 * landing.js — StudentHub landing page (redesigned).
 *
 * Sections:
 *   1. Top navbar
 *   2. Hero — heading, dual CTA, animated stat counters
 *   3. Features — 4 feature cards with hover effects
 *   4. Stats highlight strip
 *   5. How it works — 3-step process
 *   6. Final CTA band
 *   7. Footer
 *
 * All animations are CSS-driven; JS only wires the scroll-reveal
 * IntersectionObserver and the animated number counters.
 * No external libraries — pure vanilla JS.
 */

export function render() {
  const root = document.getElementById('landing-root');
  if (!root) return;

  root.hidden = false;
  root.removeAttribute('aria-hidden');

  root.innerHTML = `
<div class="lp" id="lp-root">

  <!-- ══════════════════════════════════════════
       NAV BAR
  ══════════════════════════════════════════ -->
  <header class="lp-nav" id="lp-nav" role="banner">
    <div class="lp-nav__inner">
      <a class="lp-nav__brand" href="#landing" aria-label="StudentHub home">
        <span class="lp-nav__logo" aria-hidden="true">🎓</span>
        <span class="lp-nav__name">StudentHub</span>
      </a>
      <nav class="lp-nav__links" aria-label="Landing page navigation">
        <a class="lp-nav__link" href="#lp-features">Features</a>
        <a class="lp-nav__link" href="#lp-how">How It Works</a>
      </nav>
      <a class="lp-nav__cta" href="#dashboard" aria-label="Enter the StudentHub application">
        Enter App →
      </a>
    </div>
  </header>

  <!-- ══════════════════════════════════════════
       HERO
  ══════════════════════════════════════════ -->
  <section class="lp-hero" aria-labelledby="lp-heading">
    <div class="lp-hero__bg" aria-hidden="true"></div>
    <div class="lp-hero__overlay" aria-hidden="true"></div>

    <!-- Floating blobs -->
    <div class="lp-hero__blobs" aria-hidden="true">
      <div class="lp-blob lp-blob--1"></div>
      <div class="lp-blob lp-blob--2"></div>
      <div class="lp-blob lp-blob--3"></div>
    </div>

    <!-- Floating card decorations -->
    <div class="lp-hero__deco" aria-hidden="true">
      <div class="lp-deco-card lp-deco-card--1">
        <span class="lp-deco-card__icon">📊</span>
        <span class="lp-deco-card__label">Dashboard</span>
      </div>
      <div class="lp-deco-card lp-deco-card--2">
        <span class="lp-deco-card__icon">✅</span>
        <span class="lp-deco-card__label">Grade: A+</span>
      </div>
      <div class="lp-deco-card lp-deco-card--3">
        <span class="lp-deco-card__icon">📅</span>
        <span class="lp-deco-card__label">96% Attendance</span>
      </div>
    </div>

    <div class="lp-hero__content">
      <div class="lp-hero__eyebrow">Student Management System</div>

      <h1 class="lp-hero__heading" id="lp-heading">
        Manage Students,<br>
        <span class="lp-hero__heading-accent">Courses &amp; Grades</span><br>
        All in One Place.
      </h1>

      <p class="lp-hero__sub">
        StudentHub gives administrators a single, clean interface to manage
        student records, course enrollments, marks, and attendance —
        with automatic grade calculation built in.
      </p>

      <div class="lp-hero__actions">
        <a class="lp-hero__btn lp-hero__btn--primary" href="#dashboard"
           aria-label="Enter StudentHub and go to the dashboard">
          <span>Enter StudentHub</span>
          <span class="lp-hero__btn-arrow" aria-hidden="true">→</span>
        </a>
        <a class="lp-hero__btn lp-hero__btn--ghost" href="#lp-features">
          See Features
        </a>
      </div>

      <!-- Quick stat counters -->
      <div class="lp-hero__stats">
        <div class="lp-hero__stat">
          <span class="lp-hero__stat-num" data-target="500">0</span><span class="lp-hero__stat-suffix">+</span>
          <span class="lp-hero__stat-label">Students</span>
        </div>
        <div class="lp-hero__stat-divider" aria-hidden="true"></div>
        <div class="lp-hero__stat">
          <span class="lp-hero__stat-num" data-target="50">0</span><span class="lp-hero__stat-suffix">+</span>
          <span class="lp-hero__stat-label">Courses</span>
        </div>
        <div class="lp-hero__stat-divider" aria-hidden="true"></div>
        <div class="lp-hero__stat">
          <span class="lp-hero__stat-num" data-target="100">0</span><span class="lp-hero__stat-suffix">%</span>
          <span class="lp-hero__stat-label">Automated</span>
        </div>
      </div>
    </div>

    <div class="lp-hero__scroll" aria-hidden="true">
      <div class="lp-hero__scroll-line"></div>
    </div>
  </section>

  <!-- ══════════════════════════════════════════
       FEATURES
  ══════════════════════════════════════════ -->
  <section class="lp-features" id="lp-features" aria-labelledby="lp-feat-heading">
    <div class="lp-section__inner">
      <div class="lp-section__header lp-reveal">
        <div class="lp-section__eyebrow">What You Get</div>
        <h2 class="lp-section__title" id="lp-feat-heading">Everything you need to manage academics</h2>
        <p class="lp-section__sub">Four core modules designed to work together seamlessly.</p>
      </div>

      <div class="lp-features__grid">

        <div class="lp-feat-card lp-reveal">
          <div class="lp-feat-card__icon-wrap">
            <span class="lp-feat-card__icon" aria-hidden="true">👥</span>
          </div>
          <h3 class="lp-feat-card__title">Student Records</h3>
          <p class="lp-feat-card__desc">
            Add, edit, search and filter students across departments and years.
            Full profile with all academic data in one view.
          </p>
          <div class="lp-feat-card__tag">CRUD · Search · Filter</div>
        </div>

        <div class="lp-feat-card lp-reveal">
          <div class="lp-feat-card__icon-wrap">
            <span class="lp-feat-card__icon" aria-hidden="true">📚</span>
          </div>
          <h3 class="lp-feat-card__title">Course Management</h3>
          <p class="lp-feat-card__desc">
            Create courses, assign credits, set departments, and enroll
            students instantly. Track enrolled student count per course.
          </p>
          <div class="lp-feat-card__tag">Enrollment · Credits</div>
        </div>

        <div class="lp-feat-card lp-reveal">
          <div class="lp-feat-card__icon-wrap">
            <span class="lp-feat-card__icon" aria-hidden="true">📝</span>
          </div>
          <h3 class="lp-feat-card__title">Marks &amp; Grades</h3>
          <p class="lp-feat-card__desc">
            Record final marks out of 100. Letter grades — A+, A, B, C, D, F —
            are calculated automatically. No manual calculation needed.
          </p>
          <div class="lp-feat-card__tag">Auto Grade · A+ → F</div>
        </div>

        <div class="lp-feat-card lp-reveal">
          <div class="lp-feat-card__icon-wrap">
            <span class="lp-feat-card__icon" aria-hidden="true">📅</span>
          </div>
          <h3 class="lp-feat-card__title">Attendance Tracking</h3>
          <p class="lp-feat-card__desc">
            Track total and attended classes per course. Percentage is computed
            live. Low-attendance students surface automatically in the dashboard.
          </p>
          <div class="lp-feat-card__tag">Auto % · Low-att Alerts</div>
        </div>

      </div>
    </div>
  </section>

  <!-- ══════════════════════════════════════════
       STATS BAND
  ══════════════════════════════════════════ -->
  <section class="lp-statsband" aria-label="Platform highlights">
    <div class="lp-section__inner">
      <div class="lp-statsband__grid">
        <div class="lp-statsband__item lp-reveal">
          <div class="lp-statsband__num">6</div>
          <div class="lp-statsband__label">Grade Levels</div>
        </div>
        <div class="lp-statsband__item lp-reveal">
          <div class="lp-statsband__num">8</div>
          <div class="lp-statsband__label">Departments</div>
        </div>
        <div class="lp-statsband__item lp-reveal">
          <div class="lp-statsband__num">75%</div>
          <div class="lp-statsband__label">Attendance Threshold</div>
        </div>
        <div class="lp-statsband__item lp-reveal">
          <div class="lp-statsband__num">1</div>
          <div class="lp-statsband__label">Unified Dashboard</div>
        </div>
      </div>
    </div>
  </section>

  <!-- ══════════════════════════════════════════
       HOW IT WORKS
  ══════════════════════════════════════════ -->
  <section class="lp-how" id="lp-how" aria-labelledby="lp-how-heading">
    <div class="lp-section__inner">
      <div class="lp-section__header lp-reveal">
        <div class="lp-section__eyebrow">Simple Workflow</div>
        <h2 class="lp-section__title" id="lp-how-heading">Up and running in three steps</h2>
      </div>

      <div class="lp-how__steps">
        <div class="lp-step lp-reveal">
          <div class="lp-step__num" aria-hidden="true">01</div>
          <div class="lp-step__body">
            <h3 class="lp-step__title">Add Students &amp; Courses</h3>
            <p class="lp-step__desc">Create student profiles and course records. Assign departments, credits, and years using the clean form interface.</p>
          </div>
        </div>
        <div class="lp-step__connector" aria-hidden="true"></div>
        <div class="lp-step lp-reveal">
          <div class="lp-step__num" aria-hidden="true">02</div>
          <div class="lp-step__body">
            <h3 class="lp-step__title">Enroll &amp; Record</h3>
            <p class="lp-step__desc">Assign courses to students, then record their marks and attendance per course. Grades and percentages update instantly.</p>
          </div>
        </div>
        <div class="lp-step__connector" aria-hidden="true"></div>
        <div class="lp-step lp-reveal">
          <div class="lp-step__num" aria-hidden="true">03</div>
          <div class="lp-step__body">
            <h3 class="lp-step__title">Monitor via Dashboard</h3>
            <p class="lp-step__desc">The dashboard shows totals, average attendance, and flags students with low attendance automatically. All in one place.</p>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ══════════════════════════════════════════
       FINAL CTA
  ══════════════════════════════════════════ -->
  <section class="lp-cta-band" aria-labelledby="lp-cta-heading">
    <div class="lp-cta-band__inner lp-reveal">
      <h2 class="lp-cta-band__heading" id="lp-cta-heading">Ready to get started?</h2>
      <p class="lp-cta-band__sub">Open the dashboard and start managing your institution today.</p>
      <a class="lp-cta-band__btn" href="#dashboard"
         aria-label="Enter StudentHub dashboard">
        Enter StudentHub
        <span aria-hidden="true"> →</span>
      </a>
    </div>
  </section>

  <!-- ══════════════════════════════════════════
       FOOTER
  ══════════════════════════════════════════ -->
  <footer class="lp-footer" role="contentinfo">
    <div class="lp-footer__inner">
      <div class="lp-footer__brand">
        <span aria-hidden="true">🎓</span> StudentHub
      </div>
      <p class="lp-footer__copy">© 2026 StudentHub — Student Management System</p>
    </div>
  </footer>

</div>
  `;

  // Wire up behaviours after DOM is ready
  requestAnimationFrame(() => {
    initNavScroll();
    initScrollReveal();
    initCounters();
    initSmoothScroll();
  });
}

// ─── Navbar: add scrolled class when page scrolls ────────────────
function initNavScroll() {
  const nav = document.getElementById('lp-nav');
  if (!nav) return;
  const onScroll = () => {
    nav.classList.toggle('lp-nav--scrolled', window.scrollY > 40);
  };
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();
}

// ─── Scroll-reveal via IntersectionObserver ──────────────────────
function initScrollReveal() {
  const els = document.querySelectorAll('.lp-reveal');
  if (!els.length) return;

  const obs = new IntersectionObserver((entries) => {
    entries.forEach((entry, i) => {
      if (entry.isIntersecting) {
        // Stagger siblings in the same parent
        const siblings = Array.from(entry.target.parentElement.querySelectorAll('.lp-reveal'));
        const idx = siblings.indexOf(entry.target);
        entry.target.style.transitionDelay = `${idx * 80}ms`;
        entry.target.classList.add('lp-reveal--visible');
        obs.unobserve(entry.target);
      }
    });
  }, { threshold: 0.12 });

  els.forEach(el => obs.observe(el));
}

// ─── Animated number counters in the hero ────────────────────────
function initCounters() {
  const nums = document.querySelectorAll('.lp-hero__stat-num[data-target]');
  if (!nums.length) return;

  const obs = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (!entry.isIntersecting) return;
      const el     = entry.target;
      const target = parseInt(el.dataset.target, 10);
      const dur    = 1400;
      const step   = 16;
      const steps  = Math.ceil(dur / step);
      let  current = 0;

      const tick = setInterval(() => {
        current = Math.min(current + Math.ceil(target / steps), target);
        el.textContent = current;
        if (current >= target) clearInterval(tick);
      }, step);

      obs.unobserve(el);
    });
  }, { threshold: 0.5 });

  nums.forEach(el => obs.observe(el));
}

// ─── Smooth scroll for anchor links ──────────────────────────────
function initSmoothScroll() {
  document.querySelectorAll('.lp-nav__link[href^="#lp-"]').forEach(link => {
    link.addEventListener('click', e => {
      const target = document.getElementById(link.getAttribute('href').slice(1));
      if (!target) return;
      e.preventDefault();
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    });
  });
}

/**
 * Called by app.js when navigating away from the landing page.
 */
export function hide() {
  const root = document.getElementById('landing-root');
  if (!root) return;
  // Clean up scroll listener to avoid memory leaks
  window.removeEventListener('scroll', window.__lpNavScroll);
  root.hidden = true;
  root.setAttribute('aria-hidden', 'true');
  root.innerHTML = '';
}
