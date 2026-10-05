/**
 * validators.js — Client-side validation helpers.
 *
 * Mirrors the server-side rules in validators/*.go.
 * Every function returns an array of { field, message } objects.
 * An empty array means the input is valid.
 *
 * design.md § 6.7, requirements.md § 7
 */

import { DEPARTMENTS } from './utils.js';

const STUDENT_ID_RE  = /^[A-Za-z0-9\-]{3,20}$/;
const EMAIL_RE       = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PHONE_RE       = /^\+?[0-9\s\-()+]{7,20}$/;
const SECTION_RE     = /^[A-Za-z0-9]{1,10}$/;
const COURSE_CODE_RE = /^[A-Z0-9\-]{2,20}$/;

/** Validate all student form fields. */
export function validateStudentForm(data) {
  const errs = [];

  // student_id
  const sid = (data.student_id ?? '').trim();
  if (!sid) {
    errs.push({ field: 'student_id', message: 'Student ID is required.' });
  } else if (!STUDENT_ID_RE.test(sid)) {
    errs.push({ field: 'student_id', message: 'Student ID must be 3–20 alphanumeric characters or hyphens.' });
  }

  // full_name
  const name = (data.full_name ?? '').trim();
  if (!name) {
    errs.push({ field: 'full_name', message: 'Full name is required.' });
  } else if (name.length < 2 || name.length > 100) {
    errs.push({ field: 'full_name', message: 'Full name must be 2–100 characters.' });
  } else if (!/^[A-Za-z\u00C0-\u024F\s\-']+$/.test(name)) {
    errs.push({ field: 'full_name', message: 'Full name may only contain letters, spaces, hyphens, and apostrophes.' });
  }

  // email
  const email = (data.email ?? '').trim();
  if (!email) {
    errs.push({ field: 'email', message: 'Email is required.' });
  } else if (email.length > 255) {
    errs.push({ field: 'email', message: 'Email must be at most 255 characters.' });
  } else if (!EMAIL_RE.test(email)) {
    errs.push({ field: 'email', message: 'Email must be a valid email address.' });
  }

  // phone (optional)
  const phone = (data.phone ?? '').trim();
  if (phone && !PHONE_RE.test(phone)) {
    errs.push({ field: 'phone', message: 'Phone must be 7–20 characters: digits, spaces, hyphens, parentheses, or leading +.' });
  }

  // department
  const dept = (data.department ?? '').trim();
  if (!dept) {
    errs.push({ field: 'department', message: 'Department is required.' });
  } else if (!DEPARTMENTS.includes(dept)) {
    errs.push({ field: 'department', message: 'Department must be one of the defined department list.' });
  }

  // year
  const year = parseInt(data.year, 10);
  if (!data.year && data.year !== 0) {
    errs.push({ field: 'year', message: 'Year is required.' });
  } else if (isNaN(year) || year < 1 || year > 6) {
    errs.push({ field: 'year', message: 'Year must be an integer between 1 and 6.' });
  }

  // section
  const section = (data.section ?? '').trim();
  if (!section) {
    errs.push({ field: 'section', message: 'Section is required.' });
  } else if (!SECTION_RE.test(section)) {
    errs.push({ field: 'section', message: 'Section must be 1–10 alphanumeric characters.' });
  }

  // dob
  const dob = (data.dob ?? '').trim();
  if (!dob) {
    errs.push({ field: 'dob', message: 'Date of birth is required.' });
  } else if (!/^\d{4}-\d{2}-\d{2}$/.test(dob)) {
    errs.push({ field: 'dob', message: 'Date of birth must be in YYYY-MM-DD format.' });
  } else {
    const dobDate = new Date(dob);
    const minDate = new Date();
    minDate.setFullYear(minDate.getFullYear() - 15);
    if (dobDate > minDate) {
      errs.push({ field: 'dob', message: 'Student must be at least 15 years old.' });
    }
  }

  return errs;
}

/** Validate course form fields. */
export function validateCourseForm(data) {
  const errs = [];

  const code = (data.course_code ?? '').trim();
  if (!code) {
    errs.push({ field: 'course_code', message: 'Course code is required.' });
  } else if (!COURSE_CODE_RE.test(code)) {
    errs.push({ field: 'course_code', message: 'Course code must be 2–20 uppercase letters, digits, or hyphens.' });
  }

  const cname = (data.course_name ?? '').trim();
  if (!cname) {
    errs.push({ field: 'course_name', message: 'Course name is required.' });
  } else if (cname.length < 3 || cname.length > 150) {
    errs.push({ field: 'course_name', message: 'Course name must be 3–150 characters.' });
  }

  const credits = parseInt(data.credits, 10);
  if (!data.credits) {
    errs.push({ field: 'credits', message: 'Credits is required.' });
  } else if (isNaN(credits) || credits < 1 || credits > 6) {
    errs.push({ field: 'credits', message: 'Credits must be an integer between 1 and 6.' });
  }

  const dept = (data.department ?? '').trim();
  if (!dept) {
    errs.push({ field: 'department', message: 'Department is required.' });
  } else if (!DEPARTMENTS.includes(dept)) {
    errs.push({ field: 'department', message: 'Department must be one of the defined department list.' });
  }

  return errs;
}

/** Validate marks form. */
export function validateMarksForm(data) {
  const errs = [];
  const marks = parseFloat(data.marks);
  if (data.marks === '' || data.marks === null || data.marks === undefined) {
    errs.push({ field: 'marks', message: 'Marks is required.' });
  } else if (isNaN(marks) || marks < 0 || marks > 100) {
    errs.push({ field: 'marks', message: 'Marks must be between 0 and 100.' });
  }
  return errs;
}

/** Validate attendance form. */
export function validateAttendanceForm(data) {
  const errs = [];
  const total = parseInt(data.total_classes, 10);
  const attended = parseInt(data.attended, 10);

  if (!data.total_classes) {
    errs.push({ field: 'total_classes', message: 'Total classes is required.' });
  } else if (isNaN(total) || total < 1) {
    errs.push({ field: 'total_classes', message: 'Total classes must be at least 1.' });
  }

  if (data.attended === '' || data.attended === null || data.attended === undefined) {
    errs.push({ field: 'attended', message: 'Attended classes is required.' });
  } else if (isNaN(attended) || attended < 0) {
    errs.push({ field: 'attended', message: 'Attended classes cannot be negative.' });
  } else if (!isNaN(total) && total >= 1 && attended > total) {
    errs.push({ field: 'attended', message: 'Attended classes cannot exceed total classes.' });
  }

  return errs;
}
