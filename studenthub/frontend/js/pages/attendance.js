/**
 * attendance.js — Attendance entry and view page.
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
            <div class="att-live-preview" id="att-preview" aria-live="polite" hidden>
              <div class="att-live-preview__label">Live Attendance</div>
              <div class="att-live-preview__pct" id="att-pct-val">—</div>
              <div class="att-live-preview__bar-wrap">
                <div class="att-live-preview__bar-fill" id="att-pct-bar" style="width:0%"></div>
              </div>
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

  const studentSel = document.getElementById('att-student-select');
  const courseSel  = document.getElementById('att-course-select');
  const totalInput = document.getElementById('total-classes');
  const attInput   = document.getElementById('attended');
  const preview    = document.getElementById('att-preview');
  const pctVal     = document.getElementById('att-pct-val');
  const pctBar     = document.getElementById('att-pct-bar');
  const submitBtn  = document.getElementById('att-submit-btn');
  const form       = document.getElementById('att-form');

  // existingAttId tracks whether this is a POST (null) or PUT (id).
  let existingAttId     = null;
  let currentAttRecords = [];

  // ── Helper: enable/disable the Save button based on current select values ──
  function refreshSubmitState() {
    submitBtn.disabled = !studentSel.value || !courseSel.value;
  }

  // ── Load all students ────────────────────────────────────────────────────
  try {
    const students = await studentsApi.list();
    studentSel.innerHTML = buildOptions(
      students.map(s => ({ value: s.id, label: `${s.student_id} — ${s.full_name}` })),
      '', 'Select student…'
    );
  } catch {
    studentSel.innerHTML = '<option value="">Failed to load students</option>';
  }

  // ── On student change ────────────────────────────────────────────────────
  studentSel.addEventListener('change', async () => {
    const sid = studentSel.value;

    // Reset course dropdown and related state.
    courseSel.innerHTML    = '<option value="">Loading…</option>';
    courseSel.disabled     = true;
    existingAttId          = null;
    totalInput.value       = '';
    attInput.value         = '';
    preview.hidden         = true;
    refreshSubmitState();
    document.getElementById('att-table-card').hidden = true;

    if (!sid) {
      courseSel.innerHTML = '<option value="">Select student first…</option>';
      return;
    }

    try {
      const [courses, attResp] = await Promise.all([
        studentsApi.courses(sid),
        studentsApi.attendance(sid),
      ]);

      currentAttRecords = attResp.attendance ?? [];
      document.getElementById('att-student-name').textContent = attResp.student_name ?? '';

      if (!courses.length) {
        courseSel.innerHTML = '<option value="">No enrolled courses</option>';
        courseSel.disabled  = true;
        refreshSubmitState();
        renderAttTable(currentAttRecords);
        return;
      }

      courseSel.innerHTML = buildOptions(
        courses.map(c => ({ value: c.id, label: `${c.course_code} — ${c.course_name}` })),
        '', 'Select course…'
      );
      courseSel.disabled = false;

      // *** KEY FIX: enable Save only after courses are loaded ***
      refreshSubmitState();
      renderAttTable(currentAttRecords);
    } catch (err) {
      courseSel.innerHTML = '<option value="">Failed to load courses</option>';
      refreshSubmitState();
    }
  });

  // ── On course change: prefill if existing record found ──────────────────
  courseSel.addEventListener('change', () => {
    existingAttId    = null;
    totalInput.value = '';
    attInput.value   = '';
    preview.hidden   = true;

    refreshSubmitState();   // re-check button state

    const cid = parseInt(courseSel.value, 10);
    if (!cid) return;

    const existing = currentAttRecords.find(a => a.course_id === cid);
    if (existing) {
      totalInput.value = existing.total_classes;
      attInput.value   = existing.attended;
      existingAttId    = existing.attendance_id;
      updatePreview();
    }
  });

  // ── Live animated percentage preview ────────────────────────────────────
  function updatePreview() {
    const pct = liveAttendancePct(totalInput.value, attInput.value);
    if (pct === null) { preview.hidden = true; return; }

    preview.hidden = false;
    const isLow    = pct < 75;
    const width    = Math.min(100, Math.max(0, pct)).toFixed(1);

    pctVal.textContent = pct.toFixed(2) + '%';
    pctVal.className   = 'att-live-preview__pct' + (isLow ? ' att-live-preview__pct--low' : '');
    pctBar.style.width = width + '%';
    pctBar.className   = 'att-live-preview__bar-fill' + (isLow ? ' att-live-preview__bar-fill--low' : '');
  }
  totalInput.addEventListener('input', updatePreview);
  attInput.addEventListener('input', updatePreview);

  // ── Form submit ──────────────────────────────────────────────────────────
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(form);

    // Read values fresh at submit time.
    const sid   = parseInt(studentSel.value, 10);
    const cid   = parseInt(courseSel.value, 10);
    const total = parseInt(totalInput.value, 10);
    const att   = parseInt(attInput.value, 10);

    // Guard: both selects must have valid values.
    if (!sid || !cid) {
      toast.error('Please select both a student and a course.');
      return;
    }

    // Client-side validation.
    const errs = validateAttendanceForm({ total_classes: totalInput.value, attended: attInput.value });
    if (errs.length) { displayErrors(form, errs); return; }

    submitBtn.disabled    = true;
    submitBtn.textContent = 'Saving…';

    try {
      if (existingAttId) {
        // PUT — update existing record.
        await attendanceApi.update(existingAttId, total, att);
        toast.success('Attendance updated successfully.');
      } else {
        // POST — create new record.
        await attendanceApi.create({ student_id: sid, course_id: cid, total_classes: total, attended: att });
        toast.success('Attendance recorded successfully.');
      }

      // Refresh the attendance table.
      const attResp = await studentsApi.attendance(sid);
      currentAttRecords = attResp.attendance ?? [];
      renderAttTable(currentAttRecords);

      // Update existingAttId so the next save is a PUT.
      const updated = currentAttRecords.find(a => a.course_id === cid);
      if (updated) existingAttId = updated.attendance_id;

    } catch (err) {
      if (err.fields) {
        displayErrors(form, err.fields);
      } else if (err.status === 409) {
        // Duplicate — switch to update mode automatically.
        toast.info('Record already exists. Fetching existing record…');
        const attResp = await studentsApi.attendance(sid).catch(() => null);
        if (attResp) {
          currentAttRecords = attResp.attendance ?? [];
          const found = currentAttRecords.find(a => a.course_id === cid);
          if (found) {
            existingAttId    = found.attendance_id;
            totalInput.value = found.total_classes;
            attInput.value   = found.attended;
            updatePreview();
            renderAttTable(currentAttRecords);
            toast.info('Existing record loaded. Edit the values and save again.');
          }
        }
      } else {
        toast.error(err.error ?? 'Failed to save attendance. Please try again.');
      }
    } finally {
      submitBtn.disabled    = false;
      submitBtn.textContent = 'Save Attendance';
    }
  });
}

// ── Render attendance table ──────────────────────────────────────────────
function renderAttTable(records) {
  const card = document.getElementById('att-table-card');
  const wrap = document.getElementById('att-table-wrap');
  if (!card || !wrap) return;
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
            <th scope="col">Attendance</th>
          </tr>
        </thead>
        <tbody>
          ${records.map(a => `
            <tr class="${a.percentage < 75 ? 'row--warn' : ''}">
              <td><code>${escHtml(a.course_code)}</code></td>
              <td>${escHtml(a.course_name)}</td>
              <td>${escHtml(String(a.total_classes))}</td>
              <td>${escHtml(String(a.attended))}</td>
              <td>${attendanceProgressCell(a.percentage)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    </div>
  `;
}

function attendanceProgressCell(pct) {
  const isLow   = pct < 75;
  const fillCls = isLow ? 'att-progress-fill--low' : '';
  const width   = Math.min(100, Math.max(0, pct)).toFixed(1);
  return `
    <div class="att-progress-wrap">
      <div class="att-progress-bar">
        <div class="att-progress-fill ${fillCls}" style="width:${width}%"></div>
      </div>
      ${attendanceBadge(pct)}
    </div>`;
}
