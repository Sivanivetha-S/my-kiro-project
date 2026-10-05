/**
 * courses.js — Courses list page + Add/Edit modal.
 * design.md § 6.3 (Courses Page), requirements.md US-08 through US-11
 */

import { coursesApi } from '../api.js';
import { escHtml, showSkeleton, showEmpty, DEPARTMENTS, buildOptions } from '../utils.js';
import { toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { displayErrors, clearErrors, collectForm } from '../components/form.js';
import { validateCourseForm } from '../validators.js';

let modal = null;

export async function render(container) {
  container.innerHTML = `
    <div class="page-header">
      <h1 class="page-title">Courses</h1>
      <button class="btn btn--primary" id="add-course-btn">+ Add Course</button>
    </div>
    <div class="content-card">
      <div id="courses-table-wrap"></div>
    </div>
  `;

  document.getElementById('add-course-btn').addEventListener('click', () => {
    openCourseModal(null, loadCourses);
  });

  const wrap = document.getElementById('courses-table-wrap');
  async function loadCourses() {
    showSkeleton(wrap);
    try {
      const courses = await coursesApi.list();
      renderTable(wrap, courses, loadCourses);
    } catch {
      showEmpty(wrap, 'Failed to load courses.');
    }
  }
  await loadCourses();
}

function renderTable(wrap, courses, refresh) {
  if (!courses.length) {
    showEmpty(wrap, 'No courses yet. Add one to get started.');
    return;
  }
  wrap.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Code</th>
            <th scope="col">Course Name</th>
            <th scope="col">Department</th>
            <th scope="col">Credits</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          ${courses.map(c => `
            <tr>
              <td><code>${escHtml(c.course_code)}</code></td>
              <td>${escHtml(c.course_name)}</td>
              <td>${escHtml(c.department)}</td>
              <td>${escHtml(String(c.credits))}</td>
              <td class="actions">
                <button class="btn btn--sm btn--outline edit-btn" data-id="${c.id}" aria-label="Edit ${escHtml(c.course_code)}">Edit</button>
                <button class="btn btn--sm btn--danger delete-btn" data-id="${c.id}" data-name="${escHtml(c.course_name)}" aria-label="Delete ${escHtml(c.course_code)}">Delete</button>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;

  wrap.querySelectorAll('.edit-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      try {
        const course = await coursesApi.get(btn.dataset.id);
        openCourseModal(course, refresh);
      } catch { toast.error('Could not load course.'); }
    });
  });

  wrap.querySelectorAll('.delete-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      if (!confirm(`Delete course "${btn.dataset.name}"? This will also remove related enrollments, marks, and attendance.`)) return;
      try {
        await coursesApi.delete(btn.dataset.id);
        toast.success('Course deleted.');
        refresh();
      } catch (err) {
        toast.error(err.error ?? 'Failed to delete course.');
      }
    });
  });
}

function buildCourseForm(course = null) {
  const c = course ?? {};
  const deptOptions = buildOptions(
    DEPARTMENTS.map(d => ({ value: d, label: d })),
    c.department ?? '',
    'Select department…'
  );
  const creditsOptions = buildOptions(
    [1,2,3,4,5,6].map(n => ({ value: n, label: String(n) })),
    c.credits ?? '',
    'Select credits…'
  );
  return `
    <form id="course-form" novalidate>
      <div class="form-grid">
        <div class="form-group">
          <label for="course_code" class="form-label">Course Code <span aria-hidden="true">*</span></label>
          <input id="course_code" name="course_code" type="text" class="input"
            value="${escHtml(c.course_code ?? '')}"
            aria-required="true" aria-describedby="err-course_code"
            placeholder="e.g. CS101" />
          <span id="err-course_code" data-error="course_code" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group form-group--full">
          <label for="course_name" class="form-label">Course Name <span aria-hidden="true">*</span></label>
          <input id="course_name" name="course_name" type="text" class="input"
            value="${escHtml(c.course_name ?? '')}"
            aria-required="true" aria-describedby="err-course_name" />
          <span id="err-course_name" data-error="course_name" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="credits" class="form-label">Credits <span aria-hidden="true">*</span></label>
          <select id="credits" name="credits" class="input"
            aria-required="true" aria-describedby="err-credits">
            ${creditsOptions}
          </select>
          <span id="err-credits" data-error="credits" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="department" class="form-label">Department <span aria-hidden="true">*</span></label>
          <select id="department" name="department" class="input"
            aria-required="true" aria-describedby="err-department">
            ${deptOptions}
          </select>
          <span id="err-department" data-error="department" class="field-error" role="alert" hidden></span>
        </div>
      </div>
      <div class="form-actions">
        <button type="submit" class="btn btn--primary">${course ? 'Save Changes' : 'Add Course'}</button>
        <button type="button" class="btn btn--outline" id="modal-cancel-btn">Cancel</button>
      </div>
    </form>
  `;
}

function openCourseModal(course = null, onSuccess) {
  modal?.destroy();
  modal = new Modal(course ? 'Edit Course' : 'Add Course', buildCourseForm(course));
  modal.open();

  const form = modal.body.querySelector('#course-form');
  modal.body.querySelector('#modal-cancel-btn').addEventListener('click', () => modal.close());

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);
    const data = collectForm(form);
    data.credits = parseInt(data.credits, 10);

    const errs = validateCourseForm(data);
    if (errs.length) { displayErrors(form, errs); return; }

    const submitBtn = form.querySelector('[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Saving…';

    try {
      if (course) {
        await coursesApi.update(course.id, data);
        toast.success('Course updated.');
      } else {
        await coursesApi.create(data);
        toast.success('Course added.');
      }
      modal.close();
      onSuccess?.();
    } catch (err) {
      submitBtn.disabled = false;
      submitBtn.textContent = course ? 'Save Changes' : 'Add Course';
      if (err.fields) {
        displayErrors(form, err.fields);
      } else if (err.status === 409) {
        displayErrors(form, [{ field: 'course_code', message: err.error }]);
      } else {
        toast.error(err.error ?? 'Failed to save course.');
      }
    }
  });
}
