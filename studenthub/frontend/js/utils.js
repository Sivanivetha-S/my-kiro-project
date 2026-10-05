/**
 * utils.js — Shared pure helper functions.
 * No fetch() calls, no DOM side-effects beyond returning HTML strings.
 */

/** Departments list — must match backend validators/common.go */
export const DEPARTMENTS = [
  'Computer Science',
  'Information Technology',
  'Electronics and Communication',
  'Mechanical Engineering',
  'Civil Engineering',
  'Business Administration',
  'Mathematics',
  'Physics',
];

/** Grade → CSS class suffix mapping for badge colors. */
export const GRADE_CLASS = {
  'A+': 'grade--aplus',
  'A':  'grade--a',
  'B':  'grade--b',
  'C':  'grade--c',
  'D':  'grade--d',
  'F':  'grade--f',
};

/**
 * Format an ISO date string (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SSZ) to a
 * human-readable short date like "Jun 15, 2004".
 */
export function formatDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso.substring(0, 10));
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/**
 * Format a UTC timestamp string to a human-readable date.
 */
export function formatDateTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/**
 * Return an HTML string for a grade badge.
 * @param {string|null} grade
 */
export function gradeBadge(grade) {
  if (!grade) return '<span class="badge badge--neutral">—</span>';
  const cls = GRADE_CLASS[grade] ?? 'grade--f';
  return `<span class="badge ${cls}">${escHtml(grade)}</span>`;
}

/**
 * Return an HTML string for an attendance percentage display.
 * Red if below 75%, green otherwise.
 * @param {number|null} pct
 */
export function attendanceBadge(pct) {
  if (pct === null || pct === undefined) return '<span class="badge badge--neutral">—</span>';
  const cls = pct < 75 ? 'badge--danger' : 'badge--success';
  return `<span class="badge ${cls}">${pct.toFixed(2)}%</span>`;
}

/**
 * Escape HTML special characters to prevent XSS when inserting
 * user-supplied strings via innerHTML.
 */
export function escHtml(str) {
  if (str === null || str === undefined) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/**
 * Debounce a function call: only invoke fn after `delay` ms of quiet time.
 */
export function debounce(fn, delay = 300) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

/**
 * Show a loading skeleton inside a container.
 * @param {HTMLElement} container
 * @param {number} rows
 */
export function showSkeleton(container, rows = 5) {
  container.innerHTML = Array.from({ length: rows }, () =>
    '<div class="skeleton skeleton--row"></div>'
  ).join('');
}

/**
 * Show an empty-state message inside a container.
 * @param {HTMLElement} container
 * @param {string} message
 */
export function showEmpty(container, message = 'No records found.') {
  container.innerHTML = `<div class="empty-state"><p>${escHtml(message)}</p></div>`;
}

/**
 * Show an error state inside a container.
 * @param {HTMLElement} container
 * @param {string} message
 */
export function showError(container, message = 'Failed to load data.') {
  container.innerHTML = `<div class="error-state"><p>${escHtml(message)}</p></div>`;
}

/**
 * Build a <select> element's <option> list from an array.
 * @param {Array<{value, label}>} items
 * @param {string} selected  currently selected value
 * @param {string} placeholder  first blank option label
 */
export function buildOptions(items, selected = '', placeholder = 'Select…') {
  const opts = [`<option value="">${escHtml(placeholder)}</option>`];
  for (const { value, label } of items) {
    const sel = String(value) === String(selected) ? ' selected' : '';
    opts.push(`<option value="${escHtml(String(value))}"${sel}>${escHtml(label)}</option>`);
  }
  return opts.join('');
}

/**
 * Compute live attendance percentage from two input values.
 * Returns null if inputs are invalid.
 */
export function liveAttendancePct(total, attended) {
  const t = parseInt(total, 10);
  const a = parseInt(attended, 10);
  if (isNaN(t) || isNaN(a) || t <= 0) return null;
  if (a < 0 || a > t) return null;
  return Math.round((a / t) * 10000) / 100;
}

/**
 * Compute the preview grade from a marks value.
 * Returns '' for invalid input.
 */
export function previewGrade(marks) {
  const m = parseFloat(marks);
  if (isNaN(m) || m < 0 || m > 100) return '';
  if (m >= 90) return 'A+';
  if (m >= 80) return 'A';
  if (m >= 70) return 'B';
  if (m >= 60) return 'C';
  if (m >= 50) return 'D';
  return 'F';
}
