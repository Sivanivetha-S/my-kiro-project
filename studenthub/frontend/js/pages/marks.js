/**
 * marks.js — Marks entry and view page.
 * Unchanged API/logic — UI improvements only (animated grade preview).
 */

import { studentsApi, marksApi } from '../api.js';
import { escHtml, showEmpty, gradeBadge, previewGrade, buildOptions, GRADE_CLASS } from '../utils.js';
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
              min="0" max="100" step="0.01" placeholder="Enter marks (0–100)"
              aria-required="true" aria-describedby="err-marks grade-preview-region" />
            <span id="err-marks" data-error="marks" class="field-error" role="alert" hidden></span>
            <!-- Animated grade preview -->
            <div class="grade-preview-wrap" id="grade-preview-region" aria-live="polite">
              <span class="grade-preview-label" id="grade-preview-label" hidden>Predicted grade:</span>
              <span id="grade-preview-pill"></span>
            </div>
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
  const submitBtn  = document.getElementById('marks-submit-btn');
  const form       = document.getElementById('marks-form');

  let existingMarksId    = null;
  let currentStudentMarks = [];

  // Load students.
  try {
    const students = await studentsApi.list();
    studentSel.innerHTML = buildOptions(
      students.map(s => ({ value: s.id, label: `${s.student_id} — ${s.full_name}` })),
      '', 'Select student…'
    );
  } catch {
    studentSel.innerHTML = '<option value="">Failed to load students</option>';
  }

  // On student change.
  studentSel.addEventListener('change', async () => {
    const sid = studentSel.value;
    courseSel.innerHTML = '<option value="">Loading…</option>';
    courseSel.disabled  = true;
    submitBtn.disabled  = true;
    existingMarksId     = null;
    marksInput.value    = '';
    setGradePreview('');
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

  // On course change: prefill.
  courseSel.addEventListener('change', () => {
    const cid       = parseInt(courseSel.value, 10);
    existingMarksId = null;
    marksInput.value = '';
    setGradePreview('');
    submitBtn.disabled = !cid;

    if (!cid) return;
    const existing = currentStudentMarks.find(m => m.course_id === cid);
    if (existing) {
      marksInput.value = existing.marks;
      existingMarksId  = existing.marks_id;
      setGradePreview(previewGrade(existing.marks));
    }
  });

  // Live animated grade preview.
  marksInput.addEventListener('input', () => setGradePreview(previewGrade(marksInput.value)));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);
    const errs = validateMarksForm({ marks: marksInput.value });
    if (errs.length) { displayErrors(form, errs); return; }

    const sid   = parseInt(studentSel.value, 10);
    const cid   = parseInt(courseSel.value, 10);
    const marks = parseFloat(marksInput.value);

    submitBtn.disabled    = true;
    submitBtn.textContent = 'Saving…';
    try {
      if (existingMarksId) {
        await marksApi.update(existingMarksId, marks);
        toast.success('Marks updated successfully.');
      } else {
        await marksApi.create({ student_id: sid, course_id: cid, marks });
        toast.success('Marks recorded successfully.');
      }
      const marksResp = await studentsApi.marks(sid);
      currentStudentMarks = marksResp.marks ?? [];
      renderMarksTable(currentStudentMarks);
      const updated = currentStudentMarks.find(m => m.course_id === cid);
      if (updated) existingMarksId = updated.marks_id;
    } catch (err) {
      if (err.fields) displayErrors(form, err.fields);
      else toast.error(err.error ?? 'Failed to save marks.');
    } finally {
      submitBtn.disabled    = false;
      submitBtn.textContent = 'Save Marks';
    }
  });
}

/** Render the animated grade pill. Clearing the DOM child forces the CSS animation to replay. */
function setGradePreview(grade) {
  const pill        = document.getElementById('grade-preview-pill');
  const label       = document.getElementById('grade-preview-label');
  if (!pill) return;

  if (!grade) {
    pill.innerHTML = '';
    if (label) label.hidden = true;
    return;
  }

  if (label) label.hidden = false;

  const cls = GRADE_CLASS[grade] ?? 'grade--f';
  const gradeLabels = {
    'A+': 'Outstanding', 'A': 'Excellent', 'B': 'Good',
    'C':  'Satisfactory', 'D': 'Pass',    'F': 'Fail',
  };
  const desc = gradeLabels[grade] ?? '';

  // Remove and re-add to replay the animation.
  const newPill = document.createElement('span');
  newPill.className = `grade-preview-pill badge ${cls}`;
  newPill.setAttribute('aria-label', `Predicted grade: ${grade} — ${desc}`);
  newPill.textContent = `${grade}  ${desc}`;
  pill.innerHTML = '';
  pill.appendChild(newPill);
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
              <td><strong>${escHtml(String(m.marks))}</strong></td>
              <td>${gradeBadge(m.grade)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}
