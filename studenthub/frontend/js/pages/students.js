/**
 * students.js — Students list page + Add/Edit modal form.
 * design.md § 6.3, requirements.md US-02 through US-07
 */

import { studentsApi } from '../api.js';
import { escHtml, debounce, showSkeleton, showEmpty, DEPARTMENTS, buildOptions } from '../utils.js';
import { toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { displayErrors, clearErrors, collectForm } from '../components/form.js';
import { validateStudentForm } from '../validators.js';

let modal = null;

export async function render(container, params = {}) {
  container.innerHTML = `
    <div class="page-header">
      <h1 class="page-title">Students</h1>
      <button class="btn btn--primary" id="add-student-btn">+ Add Student</button>
    </div>
    <div class="filter-bar">
      <div class="search-wrap">
        <label for="search-input" class="sr-only">Search students</label>
        <input id="search-input" type="search" class="input" placeholder="Search by name, email, or ID…" aria-label="Search students" />
      </div>
      <div class="filter-row">
        <label for="dept-filter" class="sr-only">Filter by department</label>
        <select id="dept-filter" class="input" aria-label="Filter by department">
          <option value="">All Departments</option>
          ${DEPARTMENTS.map(d => `<option value="${escHtml(d)}">${escHtml(d)}</option>`).join('')}
        </select>
        <label for="year-filter" class="sr-only">Filter by year</label>
        <select id="year-filter" class="input" aria-label="Filter by year">
          <option value="">All Years</option>
          ${[1,2,3,4,5,6].map(y => `<option value="${y}">Year ${y}</option>`).join('')}
        </select>
      </div>
    </div>
    <div class="content-card">
      <div id="students-table-wrap"></div>
    </div>
  `;

  const wrap   = document.getElementById('students-table-wrap');
  const search = document.getElementById('search-input');
  const dept   = document.getElementById('dept-filter');
  const year   = document.getElementById('year-filter');

  const loadStudents = async () => {
    showSkeleton(wrap);
    try {
      const students = await studentsApi.list({
        search: search.value.trim(),
        department: dept.value,
        year: year.value,
      });
      renderTable(wrap, students);
    } catch {
      showEmpty(wrap, 'Failed to load students.');
    }
  };

  search.addEventListener('input', debounce(loadStudents, 300));
  dept.addEventListener('change', loadStudents);
  year.addEventListener('change', loadStudents);

  document.getElementById('add-student-btn').addEventListener('click', () => {
    openStudentModal(null, loadStudents);
  });

  await loadStudents();

  // Handle initial params (e.g. #students/new from routing).
  if (params.mode === 'new') {
    openStudentModal(null, loadStudents);
  } else if (params.mode === 'edit' && params.id) {
    try {
      const s = await studentsApi.get(params.id);
      openStudentModal(s, loadStudents);
    } catch {
      toast.error('Student not found.');
    }
  }
}

function renderTable(wrap, students) {
  if (!students.length) {
    wrap.innerHTML = `
      <div class="empty-state">
        <p>No students match your search. Try adjusting the filters.</p>
      </div>`;
    return;
  }
  wrap.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Student ID</th>
            <th scope="col">Name</th>
            <th scope="col">Email</th>
            <th scope="col">Department</th>
            <th scope="col">Year</th>
            <th scope="col">Section</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          ${students.map(s => `
            <tr data-href="#students/${escHtml(String(s.id))}" title="View ${escHtml(s.full_name)}">
              <td><code>${escHtml(s.student_id)}</code></td>
              <td><strong>${escHtml(s.full_name)}</strong></td>
              <td>${escHtml(s.email)}</td>
              <td>${escHtml(s.department)}</td>
              <td>Year ${escHtml(String(s.year))}</td>
              <td>${escHtml(s.section)}</td>
              <td class="actions" onclick="event.stopPropagation()">
                <a href="#students/${escHtml(String(s.id))}" class="btn btn--sm btn--outline" aria-label="View ${escHtml(s.full_name)}">View</a>
                <button class="btn btn--sm btn--outline edit-btn" data-id="${s.id}" aria-label="Edit ${escHtml(s.full_name)}">Edit</button>
                <button class="btn btn--sm btn--danger delete-btn" data-id="${s.id}" data-name="${escHtml(s.full_name)}" aria-label="Delete ${escHtml(s.full_name)}">Delete</button>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;

  // Wire clickable rows — but not when clicking action buttons.
  wrap.querySelectorAll('tr[data-href]').forEach(row => {
    row.style.cursor = 'pointer';
    row.setAttribute('tabindex', '0');
    row.addEventListener('click', (e) => {
      if (e.target.closest('.actions')) return;
      window.location.hash = row.dataset.href;
    });
    row.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.target.closest('.actions')) {
        e.preventDefault();
        window.location.hash = row.dataset.href;
      }
    });
  });

  // Wire edit buttons.
  wrap.querySelectorAll('.edit-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      try {
        const s = await studentsApi.get(btn.dataset.id);
        openStudentModal(s, () => {
          // Re-render table after edit — use same search params.
          const search = document.getElementById('search-input');
          const dept   = document.getElementById('dept-filter');
          const year   = document.getElementById('year-filter');
          studentsApi.list({ search: search?.value.trim() ?? '', department: dept?.value ?? '', year: year?.value ?? '' })
            .then(list => renderTable(wrap, list))
            .catch(() => {});
        });
      } catch { toast.error('Could not load student.'); }
    });
  });

  // Wire delete buttons.
  wrap.querySelectorAll('.delete-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const name = btn.dataset.name;
      if (!confirm(`Delete student "${name}"? This will also remove their marks and attendance records.`)) return;
      try {
        await studentsApi.delete(btn.dataset.id);
        toast.success('Student deleted.');
        // Refresh table.
        const search = document.getElementById('search-input');
        const dept   = document.getElementById('dept-filter');
        const year   = document.getElementById('year-filter');
        const students = await studentsApi.list({
          search: search?.value.trim() ?? '',
          department: dept?.value ?? '',
          year: year?.value ?? '',
        });
        renderTable(wrap, students);
      } catch (err) {
        toast.error(err.error ?? 'Failed to delete student.');
      }
    });
  });
}

function buildStudentForm(student = null) {
  const s = student ?? {};
  const deptOptions = buildOptions(
    DEPARTMENTS.map(d => ({ value: d, label: d })),
    s.department ?? '',
    'Select department…'
  );
  const yearOptions = buildOptions(
    [1,2,3,4,5,6].map(y => ({ value: y, label: `Year ${y}` })),
    s.year ?? '',
    'Select year…'
  );
  return `
    <form id="student-form" novalidate>
      <div class="form-grid">
        <div class="form-group">
          <label for="student_id" class="form-label">Student ID <span aria-hidden="true">*</span></label>
          <input id="student_id" name="student_id" type="text" class="input"
            value="${escHtml(s.student_id ?? '')}"
            aria-required="true" aria-describedby="err-student_id"
            ${student ? 'readonly' : ''} />
          <span id="err-student_id" data-error="student_id" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="full_name" class="form-label">Full Name <span aria-hidden="true">*</span></label>
          <input id="full_name" name="full_name" type="text" class="input"
            value="${escHtml(s.full_name ?? '')}"
            aria-required="true" aria-describedby="err-full_name" />
          <span id="err-full_name" data-error="full_name" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="email" class="form-label">Email <span aria-hidden="true">*</span></label>
          <input id="email" name="email" type="email" class="input"
            value="${escHtml(s.email ?? '')}"
            aria-required="true" aria-describedby="err-email" />
          <span id="err-email" data-error="email" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="phone" class="form-label">Phone</label>
          <input id="phone" name="phone" type="tel" class="input"
            value="${escHtml(s.phone ?? '')}"
            aria-describedby="err-phone" />
          <span id="err-phone" data-error="phone" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="department" class="form-label">Department <span aria-hidden="true">*</span></label>
          <select id="department" name="department" class="input"
            aria-required="true" aria-describedby="err-department">
            ${deptOptions}
          </select>
          <span id="err-department" data-error="department" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="year" class="form-label">Year <span aria-hidden="true">*</span></label>
          <select id="year" name="year" class="input"
            aria-required="true" aria-describedby="err-year">
            ${yearOptions}
          </select>
          <span id="err-year" data-error="year" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="section" class="form-label">Section <span aria-hidden="true">*</span></label>
          <input id="section" name="section" type="text" class="input"
            value="${escHtml(s.section ?? '')}"
            aria-required="true" aria-describedby="err-section" />
          <span id="err-section" data-error="section" class="field-error" role="alert" hidden></span>
        </div>
        <div class="form-group">
          <label for="dob" class="form-label">Date of Birth <span aria-hidden="true">*</span></label>
          <input id="dob" name="dob" type="date" class="input"
            value="${escHtml(s.dob ?? '')}"
            aria-required="true" aria-describedby="err-dob" />
          <span id="err-dob" data-error="dob" class="field-error" role="alert" hidden></span>
        </div>
      </div>
      <div class="form-actions">
        <button type="submit" class="btn btn--primary">${student ? 'Save Changes' : 'Add Student'}</button>
        <button type="button" class="btn btn--outline" id="modal-cancel-btn">Cancel</button>
      </div>
    </form>
  `;
}

function openStudentModal(student = null, onSuccess) {
  const title = student ? 'Edit Student' : 'Add Student';
  modal?.destroy();
  modal = new Modal(title, buildStudentForm(student));
  modal.open();

  const form = modal.body.querySelector('#student-form');
  modal.body.querySelector('#modal-cancel-btn').addEventListener('click', () => modal.close());

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);
    const data = collectForm(form);
    data.year = parseInt(data.year, 10);

    const errs = validateStudentForm(data);
    if (errs.length) { displayErrors(form, errs); return; }

    const submitBtn = form.querySelector('[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Saving…';

    try {
      if (student) {
        await studentsApi.update(student.id, data);
        toast.success('Student updated.');
      } else {
        await studentsApi.create(data);
        toast.success('Student added.');
      }
      modal.close();
      onSuccess?.();
    } catch (err) {
      submitBtn.disabled = false;
      submitBtn.textContent = student ? 'Save Changes' : 'Add Student';
      if (err.fields) {
        displayErrors(form, err.fields);
      } else if (err.status === 409) {
        displayErrors(form, [{ field: 'student_id', message: err.error }]);
      } else {
        toast.error(err.error ?? 'Failed to save student.');
      }
    }
  });
}
