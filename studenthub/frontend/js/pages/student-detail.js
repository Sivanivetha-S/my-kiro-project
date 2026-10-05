/**
 * student-detail.js — Full student profile page.
 * Unchanged API/logic — UI improvements: profile hero, progress bars.
 */

import { studentsApi, coursesApi, enrollmentsApi } from '../api.js';
import { escHtml, gradeBadge, formatDate } from '../utils.js';
import { toast } from '../components/toast.js';

export async function render(container, params) {
  const id = params.id;
  container.innerHTML = `
    <div class="page-header">
      <a href="#students" class="btn btn--outline btn--sm">← Back to Students</a>
      <h1 class="page-title" id="student-name-heading" style="font-size:1.25rem">Loading profile…</h1>
    </div>
    <div id="profile-content">
      <div class="skeleton skeleton--card" style="height:140px;margin-bottom:1.5rem"></div>
      <div class="skeleton skeleton--row"></div>
      <div class="skeleton skeleton--row"></div>
    </div>
  `;

  try {
    const [profile, allCourses] = await Promise.all([
      studentsApi.get(id),
      coursesApi.list(),
    ]);
    renderProfile(container, profile, allCourses);
  } catch (err) {
    if (err.status === 404) {
      container.querySelector('#profile-content').innerHTML =
        '<div class="content-card error-state"><p>Student not found.</p></div>';
    } else {
      toast.error('Failed to load student profile.');
    }
  }
}

function renderProfile(container, profile, allCourses) {
  const nameEl = document.getElementById('student-name-heading');
  if (nameEl) nameEl.textContent = profile.full_name;

  const enrolledCourseIds = new Set((profile.courses ?? []).map(c => c.course_id));
  const availableCourses  = (allCourses ?? []).filter(c => !enrolledCourseIds.has(c.id));

  const avgMarks   = profile.average_marks   != null ? profile.average_marks.toFixed(2)   : '—';
  const overallAtt = profile.overall_attendance != null ? profile.overall_attendance.toFixed(2) + '%' : '—';
  const attLow     = profile.overall_attendance != null && profile.overall_attendance < 75;

  const content = document.getElementById('profile-content');
  content.innerHTML = `
    <!-- Profile Hero -->
    <div class="profile-hero" role="region" aria-label="Student profile summary">
      <div class="profile-hero__avatar" aria-hidden="true">
        ${escHtml(profile.full_name.charAt(0).toUpperCase())}
      </div>
      <div class="profile-hero__info">
        <div class="profile-hero__name">${escHtml(profile.full_name)}</div>
        <div class="profile-hero__meta">
          <span>🎓 ${escHtml(profile.department)}</span>
          <span>📅 Year ${escHtml(String(profile.year))} · Section ${escHtml(profile.section)}</span>
          <span>🪪 <code style="color:rgba(255,255,255,0.75);background:rgba(255,255,255,0.12);padding:1px 6px;border-radius:4px">${escHtml(profile.student_id)}</code></span>
        </div>
      </div>
      <div class="profile-hero__stats">
        <div class="profile-hero__stat">
          <div class="profile-hero__stat-val">${avgMarks}</div>
          <div class="profile-hero__stat-lbl">Avg Marks</div>
        </div>
        <div class="profile-hero__stat">
          <div class="profile-hero__stat-val" style="${attLow ? 'color:#fca5a5' : ''}">${overallAtt}</div>
          <div class="profile-hero__stat-lbl">Attendance</div>
        </div>
        <div class="profile-hero__stat">
          <div class="profile-hero__stat-val">${(profile.courses ?? []).length}</div>
          <div class="profile-hero__stat-lbl">Courses</div>
        </div>
      </div>
    </div>

    <div class="profile-grid">
      <!-- Info card -->
      <section class="content-card" aria-labelledby="info-heading">
        <h2 class="section-title" id="info-heading">Student Information</h2>
        <dl class="info-list">
          <div class="info-row"><dt>Student ID</dt><dd><code>${escHtml(profile.student_id)}</code></dd></div>
          <div class="info-row"><dt>Full Name</dt><dd>${escHtml(profile.full_name)}</dd></div>
          <div class="info-row"><dt>Email</dt><dd>${escHtml(profile.email)}</dd></div>
          <div class="info-row"><dt>Phone</dt><dd>${escHtml(profile.phone || '—')}</dd></div>
          <div class="info-row"><dt>Department</dt><dd>${escHtml(profile.department)}</dd></div>
          <div class="info-row"><dt>Year</dt><dd>Year ${escHtml(String(profile.year))}</dd></div>
          <div class="info-row"><dt>Section</dt><dd>${escHtml(profile.section)}</dd></div>
          <div class="info-row"><dt>Date of Birth</dt><dd>${formatDate(profile.dob)}</dd></div>
        </dl>
        <div class="card-actions">
          <a href="#students/${escHtml(String(profile.id))}/edit" class="btn btn--outline btn--sm">✏️ Edit Student</a>
        </div>
      </section>

      <!-- Courses table -->
      <section class="content-card" aria-labelledby="courses-heading">
        <div class="section-header">
          <h2 class="section-title" id="courses-heading">Enrolled Courses (${(profile.courses ?? []).length})</h2>
          ${availableCourses.length ? `
          <div class="assign-form">
            <label for="assign-course-select" class="sr-only">Assign course</label>
            <select id="assign-course-select" class="input input--sm" aria-label="Select a course to assign">
              <option value="">Assign a course…</option>
              ${availableCourses.map(c =>
                `<option value="${c.id}">${escHtml(c.course_code)} — ${escHtml(c.course_name)}</option>`
              ).join('')}
            </select>
            <button class="btn btn--primary btn--sm" id="assign-btn">+ Assign</button>
          </div>` : ''}
        </div>
        <div id="courses-table-wrap">
          ${renderCoursesTable(profile.courses ?? [])}
        </div>
      </section>
    </div>
  `;

  // Assign course.
  const assignBtn = document.getElementById('assign-btn');
  if (assignBtn) {
    assignBtn.addEventListener('click', async () => {
      const sel      = document.getElementById('assign-course-select');
      const courseId = parseInt(sel.value, 10);
      if (!courseId) return;
      assignBtn.disabled    = true;
      assignBtn.textContent = 'Assigning…';
      try {
        await enrollmentsApi.create(profile.id, courseId);
        toast.success('Course assigned successfully.');
        const [updated, courses] = await Promise.all([studentsApi.get(profile.id), coursesApi.list()]);
        renderProfile(container, updated, courses);
      } catch (err) {
        toast.error(err.error ?? 'Failed to assign course.');
        assignBtn.disabled    = false;
        assignBtn.textContent = '+ Assign';
      }
    });
  }

  // Remove enrollment buttons.
  document.querySelectorAll('.remove-enrollment-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const courseName = btn.dataset.course;
      if (!confirm(`Remove enrollment from "${courseName}"?\n\nThis will also delete marks and attendance for this course.`)) return;
      try {
        await enrollmentsApi.delete(btn.dataset.enrollmentId);
        toast.success('Enrollment removed.');
        const [updated, courses] = await Promise.all([studentsApi.get(profile.id), coursesApi.list()]);
        renderProfile(container, updated, courses);
      } catch (err) {
        toast.error(err.error ?? 'Failed to remove enrollment.');
      }
    });
  });
}

function renderCoursesTable(courses) {
  if (!courses.length) {
    return `<div class="empty-state" style="padding:2rem 1rem"><p>No courses enrolled yet. Use the assign dropdown above.</p></div>`;
  }
  return `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Code</th>
            <th scope="col">Course</th>
            <th scope="col">Credits</th>
            <th scope="col">Marks</th>
            <th scope="col">Grade</th>
            <th scope="col">Attendance</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          ${courses.map(c => `
            <tr>
              <td><code>${escHtml(c.course_code)}</code></td>
              <td>${escHtml(c.course_name)}</td>
              <td>${escHtml(String(c.credits))}</td>
              <td>${c.marks != null ? `<strong>${escHtml(String(c.marks))}</strong>` : '<span class="text-muted">—</span>'}</td>
              <td>${gradeBadge(c.grade)}</td>
              <td>${attProgressCell(c.attendance_percentage)}</td>
              <td>
                <button class="btn btn--sm btn--danger remove-enrollment-btn"
                  data-enrollment-id="${c.enrollment_id}"
                  data-course="${escHtml(c.course_name)}"
                  aria-label="Remove enrollment from ${escHtml(c.course_name)}">
                  Remove
                </button>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}

function attProgressCell(pct) {
  if (pct == null) return '<span class="text-muted">—</span>';
  const isLow   = pct < 75;
  const fillCls = isLow ? 'att-progress-fill--low' : '';
  const width   = Math.min(100, Math.max(0, pct)).toFixed(1);
  const label   = pct.toFixed(1) + '%';
  const colour  = isLow ? 'var(--color-danger)' : 'var(--color-neutral-700)';
  return `
    <div class="att-progress-wrap">
      <div class="att-progress-bar">
        <div class="att-progress-fill ${fillCls}" style="width:${width}%"></div>
      </div>
      <span class="att-progress-label" style="color:${colour}">${label}</span>
    </div>`;
}
