/**
 * student-detail.js — Full student profile page.
 * design.md § 6.3 (Student Detail Page), requirements.md US-03
 */

import { studentsApi, coursesApi, enrollmentsApi } from '../api.js';
import { escHtml, gradeBadge, attendanceBadge, formatDate } from '../utils.js';
import { toast } from '../components/toast.js';

export async function render(container, params) {
  const id = params.id;
  container.innerHTML = `
    <div class="page-header">
      <a href="#students" class="btn btn--outline btn--sm">← Back to Students</a>
      <h1 class="page-title" id="student-name-heading">Loading…</h1>
    </div>
    <div id="profile-content">
      <div class="skeleton skeleton--card"></div>
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
  document.getElementById('student-name-heading').textContent = profile.full_name;

  const enrolledCourseIds = new Set((profile.courses ?? []).map(c => c.course_id));
  const availableCourses = (allCourses ?? []).filter(c => !enrolledCourseIds.has(c.id));

  const content = document.getElementById('profile-content');
  content.innerHTML = `
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
          <div class="info-row"><dt>Year</dt><dd>${escHtml(String(profile.year))}</dd></div>
          <div class="info-row"><dt>Section</dt><dd>${escHtml(profile.section)}</dd></div>
          <div class="info-row"><dt>Date of Birth</dt><dd>${formatDate(profile.dob)}</dd></div>
        </dl>
        <div class="summary-badges">
          <div class="summary-item">
            <span class="summary-label">Avg Marks</span>
            <span class="badge ${profile.average_marks != null ? 'badge--primary' : 'badge--neutral'}">
              ${profile.average_marks != null ? profile.average_marks.toFixed(2) : '—'}
            </span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Overall Attendance</span>
            ${attendanceBadge(profile.overall_attendance)}
          </div>
        </div>
        <div class="card-actions">
          <a href="#students/${escHtml(String(profile.id))}/edit" class="btn btn--outline btn--sm">Edit Student</a>
        </div>
      </section>

      <!-- Courses table -->
      <section class="content-card" aria-labelledby="courses-heading">
        <div class="section-header">
          <h2 class="section-title" id="courses-heading">Enrolled Courses</h2>
          ${availableCourses.length ? `
          <div class="assign-form">
            <label for="assign-course-select" class="sr-only">Assign course</label>
            <select id="assign-course-select" class="input input--sm" aria-label="Select a course to assign">
              <option value="">Assign a course…</option>
              ${availableCourses.map(c =>
                `<option value="${c.id}">${escHtml(c.course_code)} — ${escHtml(c.course_name)}</option>`
              ).join('')}
            </select>
            <button class="btn btn--primary btn--sm" id="assign-btn">Assign</button>
          </div>` : ''}
        </div>
        <div id="courses-table-wrap">
          ${renderCoursesTable(profile.courses ?? [], profile.id)}
        </div>
      </section>
    </div>
  `;

  // Assign course.
  const assignBtn = document.getElementById('assign-btn');
  if (assignBtn) {
    assignBtn.addEventListener('click', async () => {
      const sel = document.getElementById('assign-course-select');
      const courseId = parseInt(sel.value, 10);
      if (!courseId) return;
      try {
        await enrollmentsApi.create(profile.id, courseId);
        toast.success('Course assigned.');
        // Refresh the profile.
        const [updated, courses] = await Promise.all([studentsApi.get(profile.id), coursesApi.list()]);
        renderProfile(container, updated, courses);
      } catch (err) {
        toast.error(err.error ?? 'Failed to assign course.');
      }
    });
  }

  // Wire remove enrollment buttons.
  document.querySelectorAll('.remove-enrollment-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const courseName = btn.dataset.course;
      if (!confirm(`Remove enrollment from "${courseName}"? This will also delete marks and attendance for this course.`)) return;
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

function renderCoursesTable(courses, studentId) {
  if (!courses.length) {
    return '<p class="empty-state">No courses enrolled yet.</p>';
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
            <th scope="col">Attendance %</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          ${courses.map(c => `
            <tr>
              <td><code>${escHtml(c.course_code)}</code></td>
              <td>${escHtml(c.course_name)}</td>
              <td>${escHtml(String(c.credits))}</td>
              <td>${c.marks != null ? escHtml(String(c.marks)) : '<span class="text-muted">—</span>'}</td>
              <td>${gradeBadge(c.grade)}</td>
              <td>${attendanceBadge(c.attendance_percentage)}</td>
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
