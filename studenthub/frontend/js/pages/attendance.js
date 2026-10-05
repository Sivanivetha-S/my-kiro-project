/**
 * attendance.js — Attendance entry and view page.
 * design.md § 6.3 (Attendance Page), requirements.md US-18 through US-21
 */

import { studentsApi, attendanceApi } from '../api.js';
import { escHtml, showEmpty, attendanceBadge, liveAttendancePct, buildOptions } from '../utils.js';
import { toast } from '../components/toast.js';
import { displayErrors, clearErrors } from '../components/form.js';
import { validateAttendanceForm } from '../validators.js';

export async function render(container) {
  container.innerHTML = `
    <div class="page-header">
      <h1 class="page-title">Attendance</h1>
    </div>
    <div class="content-card">
      <h2 class="section-title">Record / Update Attendance</h2>
      <form id="att-form" novalidate class="attendance-form">
        <div class="form-grid">
          <div class="form-group">
            <label for="att-student-select" class="form-label">Student <span aria-hidden="true">*</span></label>
            <select id="att-student-select" class="input" aria-required="true">
              <option value="">Loading students…</option>
            </select>
          </div>
          <div class="form-group">
            <label for="att-course-select" class="form-label">Course <span aria-hidden="true">*</span></label>
            <select id="att-course-select" class="input" aria-required="true" disabled>
              <option value="">Select student first…</option>
            </select>
          </div>
          <div class="form-group">
            <label for="total-classes" class="form-label">Total Classes <span aria-hidden="true">*</span></label>
            <input id="total-classes" name="total_classes" type="number" class="input"
              min="1" placeholder="e.g. 48"
              aria-required="true" aria-describedby="err-total_classes" />
            <span id="err-total_classes" data-error="total_classes" class="field-error" role="alert" hidden></span>
          </div>
          <div class="form-group">
            <label for="attended" class="form-label">Attended Classes <span aria-hidden="true">*</span></label>
            <input id="attended" name="attended" type="number" class="input"
              min="0" placeholder="e.g. 36"
              aria-required="true" aria-describedby="err-attended" />
            <span id="err-attended" data-error="attended" class="field-error" role="alert" hidden></span>
          </div>
          <div class="form-group form-group--full">
            <div class="attendance-preview" id="att-preview" aria-live="polite" hidden>
              Attendance: <strong id="att-pct-val"></strong>
            </div>
          </div>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn btn--primary" id="att-submit-btn" disabled>Save Attendance</button>
        </div>
      </form>
    </div>
    <div class="content-card" id="att-table-card" hidden>
      <h2 class="section-title">Attendance for <span id="att-student-name"></span></h2>
      <div id="att-table-wrap"></div>
    </div>
  `;

  const studentSel  = document.getElementById('att-student-select');
  const courseSel   = document.getElementById('att-course-select');
  const totalInput  = document.getElementById('total-classes');
  const attInput    = document.getElementById('attended');
  const preview     = document.getElementById('att-preview');
  const pctVal      = document.getElementById('att-pct-val');
  const submitBtn   = document.getElementById('att-submit-btn');
  const form        = document.getElementById('att-form');

  let existingAttId = null;
  let currentAttRecords = [];

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

  // On student change.
  studentSel.addEventListener('change', async () => {
    const sid = studentSel.value;
    courseSel.innerHTML = '<option value="">Loading…</option>';
    courseSel.disabled = true;
    submitBtn.disabled = true;
    existingAttId = null;
    totalInput.value = '';
    attInput.value = '';
    preview.hidden = true;
    document.getElementById('att-table-card').hidden = true;

    if (!sid) { courseSel.innerHTML = '<option value="">Select student first…</option>'; return; }

    try {
      const [courses, attResp] = await Promise.all([
        studentsApi.courses(sid),
        studentsApi.attendance(sid),
      ]);
      currentAttRecords = attResp.attendance ?? [];
      document.getElementById('att-student-name').textContent = attResp.student_name ?? '';

      courseSel.innerHTML = buildOptions(
        courses.map(c => ({ value: c.id, label: `${c.course_code} — ${c.course_name}` })),
        '', 'Select course…'
      );
      courseSel.disabled = false;

      renderAttTable(currentAttRecords);
    } catch {
      courseSel.innerHTML = '<option value="">No enrolled courses</option>';
    }
  });

  // On course change: prefill existing.
  courseSel.addEventListener('change', () => {
    const cid = parseInt(courseSel.value, 10);
    existingAttId = null;
    totalInput.value = '';
    attInput.value = '';
    preview.hidden = true;
    submitBtn.disabled = !cid;

    if (!cid) return;
    const existing = currentAttRecords.find(a => a.course_id === cid);
    if (existing) {
      totalInput.value  = existing.total_classes;
      attInput.value    = existing.attended;
      existingAttId     = existing.attendance_id;
      updatePreview();
    }
  });

  // Live % preview.
  function updatePreview() {
    const pct = liveAttendancePct(totalInput.value, attInput.value);
    if (pct === null) { preview.hidden = true; return; }
    preview.hidden = false;
    pctVal.textContent = pct.toFixed(2) + '%';
    pctVal.className = pct < 75 ? 'text-danger' : 'text-success';
  }
  totalInput.addEventListener('input', updatePreview);
  attInput.addEventListener('input', updatePreview);

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);
    const errs = validateAttendanceForm({ total_classes: totalInput.value, attended: attInput.value });
    if (errs.length) { displayErrors(form, errs); return; }

    const sid   = parseInt(studentSel.value, 10);
    const cid   = parseInt(courseSel.value, 10);
    const total = parseInt(totalInput.value, 10);
    const att   = parseInt(attInput.value, 10);

    submitBtn.disabled = true;
    submitBtn.textContent = 'Saving…';
    try {
      if (existingAttId) {
        await attendanceApi.update(existingAttId, total, att);
        toast.success('Attendance updated.');
      } else {
        await attendanceApi.create({ student_id: sid, course_id: cid, total_classes: total, attended: att });
        toast.success('Attendance recorded.');
      }
      const attResp = await studentsApi.attendance(sid);
      currentAttRecords = attResp.attendance ?? [];
      renderAttTable(currentAttRecords);
      const updated = currentAttRecords.find(a => a.course_id === cid);
      if (updated) existingAttId = updated.attendance_id;
    } catch (err) {
      if (err.fields) displayErrors(form, err.fields);
      else toast.error(err.error ?? 'Failed to save attendance.');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Save Attendance';
    }
  });
}

function renderAttTable(records) {
  const card = document.getElementById('att-table-card');
  const wrap = document.getElementById('att-table-wrap');
  card.hidden = false;

  if (!records.length) {
    showEmpty(wrap, 'No attendance records for this student yet.');
    return;
  }
  wrap.innerHTML = `
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Code</th>
            <th scope="col">Course</th>
            <th scope="col">Total</th>
            <th scope="col">Attended</th>
            <th scope="col">Percentage</th>
          </tr>
        </thead>
        <tbody>
          ${records.map(a => `
            <tr class="${a.percentage < 75 ? 'row--warn' : ''}">
              <td><code>${escHtml(a.course_code)}</code></td>
              <td>${escHtml(a.course_name)}</td>
              <td>${escHtml(String(a.total_classes))}</td>
              <td>${escHtml(String(a.attended))}</td>
              <td>${attendanceBadge(a.percentage)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}
