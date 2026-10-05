/**
 * marks.js — Marks entry and view page.
 * design.md § 6.3 (Marks Page), requirements.md US-14 through US-17
 */

import { studentsApi, marksApi } from '../api.js';
import { escHtml, showEmpty, gradeBadge, previewGrade, buildOptions } from '../utils.js';
import { toast } from '../components/toast.js';
import { displayErrors, clearErrors } from '../components/form.js';
import { validateMarksForm } from '../validators.js';

export async function render(container) {
  container.innerHTML = `
    <div class="page-header">
      <h1 class="page-title">Marks</h1>
    </div>
    <div class="content-card">
      <h2 class="section-title">Record / Update Marks</h2>
      <form id="marks-form" novalidate class="marks-form">
        <div class="form-grid">
          <div class="form-group">
            <label for="student-select" class="form-label">Student <span aria-hidden="true">*</span></label>
            <select id="student-select" class="input" aria-required="true" aria-label="Select student">
              <option value="">Loading students…</option>
            </select>
          </div>
          <div class="form-group">
            <label for="course-select" class="form-label">Course <span aria-hidden="true">*</span></label>
            <select id="course-select" class="input" aria-required="true" aria-label="Select course" disabled>
              <option value="">Select student first…</option>
            </select>
          </div>
          <div class="form-group">
            <label for="marks-input" class="form-label">Marks (0–100) <span aria-hidden="true">*</span></label>
            <input id="marks-input" name="marks" type="number" class="input"
              min="0" max="100" step="0.01" placeholder="Enter marks"
              aria-required="true" aria-describedby="err-marks marks-preview" />
            <span id="err-marks" data-error="marks" class="field-error" role="alert" hidden></span>
            <span id="marks-preview" class="marks-preview" aria-live="polite"></span>
          </div>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn--primary" id="marks-submit-btn" disabled>Save Marks</button>
        </div>
      </form>
    </div>
    <div class="content-card" id="marks-table-card" hidden>
      <h2 class="section-title">Marks for <span id="marks-student-name"></span></h2>
      <div id="marks-table-wrap"></div>
    </div>
  `;

  const studentSel = document.getElementById('student-select');
  const courseSel  = document.getElementById('course-select');
  const marksInput = document.getElementById('marks-input');
  const preview    = document.getElementById('marks-preview');
  const submitBtn  = document.getElementById('marks-submit-btn');
  const form       = document.getElementById('marks-form');

  // State for existing record (to decide POST vs PUT).
  let existingMarksId = null;
  let currentStudentMarks = [];

  // Load all students.
  try {
    const students = await studentsApi.list();
    studentSel.innerHTML = buildOptions(
      students.map(s => ({ value: s.id, label: `${s.student_id} — ${s.full_name}` })),
      '', 'Select student…'
    );
  } catch {
    studentSel.innerHTML = '<option value="">Failed to load students</option>';
  }

  // On student change: load their enrolled courses + existing marks.
  studentSel.addEventListener('change', async () => {
    const sid = studentSel.value;
    courseSel.innerHTML = '<option value="">Loading…</option>';
    courseSel.disabled = true;
    submitBtn.disabled = true;
    existingMarksId = null;
    marksInput.value = '';
    preview.innerHTML = '';
    document.getElementById('marks-table-card').hidden = true;

    if (!sid) { courseSel.innerHTML = '<option value="">Select student first…</option>'; return; }

    try {
      const [courses, marksResp] = await Promise.all([
        studentsApi.courses(sid),
        studentsApi.marks(sid),
      ]);
      currentStudentMarks = marksResp.marks ?? [];
      document.getElementById('marks-student-name').textContent = marksResp.student_name ?? '';

      courseSel.innerHTML = buildOptions(
        courses.map(c => ({ value: c.id, label: `${c.course_code} — ${c.course_name}` })),
        '', 'Select course…'
      );
      courseSel.disabled = false;

      renderMarksTable(currentStudentMarks);
    } catch {
      courseSel.innerHTML = '<option value="">No enrolled courses</option>';
    }
  });

  // On course change: prefill existing marks.
  courseSel.addEventListener('change', () => {
    const cid = parseInt(courseSel.value, 10);
    existingMarksId = null;
    marksInput.value = '';
    preview.innerHTML = '';
    submitBtn.disabled = !cid;

    if (!cid) return;
    const existing = currentStudentMarks.find(m => m.course_id === cid);
    if (existing) {
      marksInput.value = existing.marks;
      existingMarksId = existing.marks_id;
      preview.innerHTML = `Grade preview: ${gradeBadge(previewGrade(existing.marks))}`;
    }
  });

  // Live grade preview.
  marksInput.addEventListener('input', () => {
    const g = previewGrade(marksInput.value);
    preview.innerHTML = g ? `Grade preview: ${gradeBadge(g)}` : '';
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);
    const errs = validateMarksForm({ marks: marksInput.value });
    if (errs.length) { displayErrors(form, errs); return; }

    const sid = parseInt(studentSel.value, 10);
    const cid = parseInt(courseSel.value, 10);
    const marks = parseFloat(marksInput.value);

    submitBtn.disabled = true;
    submitBtn.textContent = 'Saving…';
    try {
      if (existingMarksId) {
        await marksApi.update(existingMarksId, marks);
        toast.success('Marks updated.');
      } else {
        await marksApi.create({ student_id: sid, course_id: cid, marks });
        toast.success('Marks recorded.');
      }
      // Refresh marks table.
      const marksResp = await studentsApi.marks(sid);
      currentStudentMarks = marksResp.marks ?? [];
      renderMarksTable(currentStudentMarks);
      // Update existing id in case it was a new record.
      const updated = currentStudentMarks.find(m => m.course_id === cid);
      if (updated) existingMarksId = updated.marks_id;
    } catch (err) {
      if (err.fields) displayErrors(form, err.fields);
      else toast.error(err.error ?? 'Failed to save marks.');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Save Marks';
    }
  });
}

function renderMarksTable(marks) {
  const card = document.getElementById('marks-table-card');
  const wrap = document.getElementById('marks-table-wrap');
  card.hidden = false;

  if (!marks.length) {
    showEmpty(wrap, 'No marks recorded for this student yet.');
    return;
  }
  wrap.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Code</th>
            <th scope="col">Course</th>
            <th scope="col">Marks</th>
            <th scope="col">Grade</th>
          </tr>
        </thead>
        <tbody>
          ${marks.map(m => `
            <tr>
              <td><code>${escHtml(m.course_code)}</code></td>
              <td>${escHtml(m.course_name)}</td>
              <td>${escHtml(String(m.marks))}</td>
              <td>${gradeBadge(m.grade)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}
