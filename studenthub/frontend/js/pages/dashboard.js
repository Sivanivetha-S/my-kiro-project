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
        <h2 class="section-title" id="low-att-heading">Low Attendance Students</h2>
        <div id="low-att-table"><div class="skeleton skeleton--row"></div></div>
      </section>
      <section class="content-card" aria-labelledby="recent-heading">
        <h2 class="section-title" id="recent-heading">Recent Students</h2>
        <div id="recent-table"><div class="skeleton skeleton--row"></div></div>
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
    container.querySelector('.stat-grid').innerHTML =
      '<p class="error-state">Could not load dashboard. Try refreshing.</p>';
  }
}

function renderStats(data) {
  const grid = document.getElementById('stat-grid');
  if (!grid) return;
  // All four cards use the unified blue design language.
  // --stat-accent drives only the left border hue; layout is identical.
  grid.innerHTML = `
    ${statCard('👥', 'Total Students',  data.total_students,                             '#2563EB')}
    ${statCard('📚', 'Total Courses',   data.total_courses,                              '#3B82F6')}
    ${statCard('📅', 'Avg Attendance',  (data.average_attendance ?? 0).toFixed(2) + '%', '#1D4ED8')}
    ${statCard('⚠️', 'Low Attendance',  (data.low_attendance_students ?? []).length,     '#60A5FA')}
  `;
}

function statCard(icon, label, value, accent) {
  return `
    <div class="stat-card" style="--stat-accent: ${accent}">
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
    '<div class="stat-card skeleton" style="height:96px"></div>'
  ).join('');
}

function renderLowAttendance(students) {
  const el = document.getElementById('low-att-table');
  if (!el) return;
  if (!students.length) {
    el.innerHTML = '<p class="empty-state">No students with low attendance. 🎉</p>';
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
            <th scope="col">Attendance %</th>
          </tr>
        </thead>
        <tbody>
          ${students.map(s => `
            <tr>
              <td><a href="#students/${escHtml(String(s.student_id))}">${escHtml(s.full_name)}</a></td>
              <td>${escHtml(s.department)}</td>
              <td>${escHtml(String(s.year))}</td>
              <td>${attendanceBadge(s.overall_attendance)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}

function renderRecentStudents(students) {
  const el = document.getElementById('recent-table');
  if (!el) return;
  if (!students.length) {
    el.innerHTML = '<p class="empty-state">No students yet.</p>';
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
            <th scope="col">Year</th>
            <th scope="col">Added</th>
          </tr>
        </thead>
        <tbody>
          ${students.map(s => `
            <tr>
              <td><a href="#students/${escHtml(String(s.id))}">${escHtml(s.full_name)}</a></td>
              <td><code>${escHtml(s.student_id)}</code></td>
              <td>${escHtml(s.department)}</td>
              <td>${escHtml(String(s.year))}</td>
              <td>${formatDateTime(s.created_at)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}
