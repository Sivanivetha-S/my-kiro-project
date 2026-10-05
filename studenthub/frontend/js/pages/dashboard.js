/**
 * dashboard.js — Dashboard page.
 * design.md § 6.3 (Dashboard Page), requirements.md § 5.8
 */

import { dashboardApi } from '../api.js';
import { escHtml, formatDateTime, attendanceBadge } from '../utils.js';
import { toast } from '../components/toast.js';

export async function render(container) {
  container.innerHTML = `
    <div class="page-header">
      <h1 class="page-title">Dashboard</h1>
    </div>
    <div class="stat-grid" id="stat-grid">
      ${statCardSkeleton(4)}
    </div>
    <div class="dashboard-tables">
      <section class="content-card" aria-labelledby="low-att-heading">
        <h2 class="section-title" id="low-att-heading">⚠️ Low Attendance Students</h2>
        <div id="low-att-table">
          <div class="skeleton skeleton--row"></div>
          <div class="skeleton skeleton--row"></div>
        </div>
      </section>
      <section class="content-card" aria-labelledby="recent-heading">
        <h2 class="section-title" id="recent-heading">🕐 Recent Students</h2>
        <div id="recent-table">
          <div class="skeleton skeleton--row"></div>
          <div class="skeleton skeleton--row"></div>
        </div>
      </section>
    </div>
  `;

  try {
    const data = await dashboardApi.get();
    renderStats(data);
    renderLowAttendance(data.low_attendance_students ?? []);
    renderRecentStudents(data.recent_students ?? []);
  } catch (err) {
    toast.error('Failed to load dashboard data.');
    document.getElementById('stat-grid').innerHTML =
      '<p class="error-state">Could not load dashboard. Try refreshing.</p>';
  }
}

function renderStats(data) {
  const grid = document.getElementById('stat-grid');
  if (!grid) return;
  grid.innerHTML = `
    ${statCard('👥', 'Total Students',  data.total_students,                             '#2563EB', '#students')}
    ${statCard('📚', 'Total Courses',   data.total_courses,                              '#3B82F6', '#courses')}
    ${statCard('📅', 'Avg Attendance',  (data.average_attendance ?? 0).toFixed(2) + '%', '#1D4ED8', '#attendance')}
    ${statCard('⚠️', 'Low Attendance',  (data.low_attendance_students ?? []).length,     '#60A5FA', null)}
  `;
  // Wire clickable cards
  grid.querySelectorAll('.stat-card[data-href]').forEach(card => {
    card.style.cursor = 'pointer';
    card.setAttribute('role', 'link');
    card.setAttribute('tabindex', '0');
    card.addEventListener('click', () => { window.location.hash = card.dataset.href; });
    card.addEventListener('keydown', e => {
      if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); window.location.hash = card.dataset.href; }
    });
  });
}

function statCard(icon, label, value, accent, href) {
  const hrefAttr = href ? `data-href="${href}"` : '';
  const title = href ? `aria-label="${escHtml(label)}: ${escHtml(String(value))} — click to view"` : '';
  return `
    <div class="stat-card" style="--stat-accent: ${accent}" ${hrefAttr} ${title}>
      <div class="stat-card__icon" aria-hidden="true">${icon}</div>
      <div class="stat-card__body">
        <div class="stat-card__value">${escHtml(String(value))}</div>
        <div class="stat-card__label">${escHtml(label)}</div>
      </div>
    </div>
  `;
}

function statCardSkeleton(n) {
  return Array.from({ length: n }, () =>
    '<div class="stat-card skeleton" style="height:90px;border-left:4px solid #e2e8f0"></div>'
  ).join('');
}

function renderLowAttendance(students) {
  const el = document.getElementById('low-att-table');
  if (!el) return;
  if (!students.length) {
    el.innerHTML = `
      <div class="empty-state" style="padding:2rem">
        <p>No students below the attendance threshold. 🎉</p>
      </div>`;
    return;
  }
  el.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Department</th>
            <th scope="col">Year</th>
            <th scope="col">Attendance</th>
          </tr>
        </thead>
        <tbody>
          ${students.map(s => `
            <tr data-href="#students" style="cursor:pointer" title="View student profile">
              <td><strong>${escHtml(s.full_name)}</strong></td>
              <td>${escHtml(s.department)}</td>
              <td>Year ${escHtml(String(s.year))}</td>
              <td>${attendanceProgressCell(s.overall_attendance)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
  wireClickableRows(el);
}

function renderRecentStudents(students) {
  const el = document.getElementById('recent-table');
  if (!el) return;
  if (!students.length) {
    el.innerHTML = `<div class="empty-state" style="padding:2rem"><p>No students added yet.</p></div>`;
    return;
  }
  el.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Student ID</th>
            <th scope="col">Department</th>
            <th scope="col">Added</th>
          </tr>
        </thead>
        <tbody>
          ${students.map(s => `
            <tr data-href="#students/${escHtml(String(s.id))}" style="cursor:pointer" title="View ${escHtml(s.full_name)}">
              <td><strong>${escHtml(s.full_name)}</strong></td>
              <td><code>${escHtml(s.student_id)}</code></td>
              <td>${escHtml(s.department)}</td>
              <td>${formatDateTime(s.created_at)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
  wireClickableRows(el);
}

/** Render an inline progress bar + badge for an attendance percentage. */
function attendanceProgressCell(pct) {
  const isLow   = pct < 75;
  const fillCls = isLow ? 'att-progress-fill--low' : '';
  const width   = Math.min(100, Math.max(0, pct)).toFixed(1);
  const badge   = attendanceBadge(pct);
  return `
    <div class="att-progress-wrap">
      <div class="att-progress-bar">
        <div class="att-progress-fill ${fillCls}" style="width:${width}%"></div>
      </div>
      ${badge}
    </div>`;
}

/** Make tr[data-href] rows navigate on click/keyboard. */
function wireClickableRows(container) {
  container.querySelectorAll('tr[data-href]').forEach(row => {
    row.setAttribute('tabindex', '0');
    row.addEventListener('click', () => { window.location.hash = row.dataset.href; });
    row.addEventListener('keydown', e => {
      if (e.key === 'Enter') { e.preventDefault(); window.location.hash = row.dataset.href; }
    });
  });
}
