/**
 * api.js — Centralized HTTP client for StudentHub.
 *
 * This is the ONLY file in the frontend that calls fetch().
 * All page modules import { api } and call api.get/post/put/delete.
 *
 * design.md § 6.6
 */

const BASE = '/api/v1';

/**
 * Core request function.
 * @param {string} method  HTTP method
 * @param {string} path    API path (without /api/v1 prefix)
 * @param {any}    [body]  Request body (will be JSON-encoded)
 * @returns {Promise<any>} Parsed JSON or null for 204 responses
 * @throws {{ status: number, error: string, code: string, fields?: Array }} on non-2xx
 */
async function request(method, path, body = null) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body !== null) {
    opts.body = JSON.stringify(body);
  }

  let res;
  try {
    res = await fetch(BASE + path, opts);
  } catch (networkErr) {
    throw { status: 0, error: 'Network error — server may be unreachable.', code: 'NETWORK_ERROR' };
  }

  if (res.status === 204) return null;

  let data;
  try {
    data = await res.json();
  } catch {
    data = { error: 'Invalid response from server.', code: 'PARSE_ERROR' };
  }

  if (!res.ok) {
    throw { status: res.status, error: data.error ?? 'Unknown error', code: data.code ?? 'UNKNOWN', fields: data.fields };
  }

  return data;
}

export const api = {
  get:    (path)         => request('GET',    path),
  post:   (path, body)   => request('POST',   path, body),
  put:    (path, body)   => request('PUT',    path, body),
  delete: (path)         => request('DELETE', path),
};

// ---- Convenience wrappers for all StudentHub endpoints ----

// Students
export const studentsApi = {
  list:   (params = {}) => api.get('/students' + toQuery(params)),
  get:    (id)          => api.get(`/students/${id}`),
  create: (data)        => api.post('/students', data),
  update: (id, data)    => api.put(`/students/${id}`, data),
  delete: (id)          => api.delete(`/students/${id}`),
  marks:      (id)      => api.get(`/students/${id}/marks`),
  attendance: (id)      => api.get(`/students/${id}/attendance`),
  courses:    (id)      => api.get(`/students/${id}/courses`),
};

// Courses
export const coursesApi = {
  list:   ()       => api.get('/courses'),
  get:    (id)     => api.get(`/courses/${id}`),
  create: (data)   => api.post('/courses', data),
  update: (id, d)  => api.put(`/courses/${id}`, d),
  delete: (id)     => api.delete(`/courses/${id}`),
  students: (id)   => api.get(`/courses/${id}/students`),
};

// Enrollments
export const enrollmentsApi = {
  create: (studentId, courseId) => api.post('/enrollments', { student_id: studentId, course_id: courseId }),
  delete: (id)                  => api.delete(`/enrollments/${id}`),
};

// Marks
export const marksApi = {
  create: (data)        => api.post('/marks', data),
  update: (id, marks)   => api.put(`/marks/${id}`, { marks }),
};

// Attendance
export const attendanceApi = {
  create: (data)                            => api.post('/attendance', data),
  update: (id, total_classes, attended)     => api.put(`/attendance/${id}`, { total_classes, attended }),
  low:    ()                                => api.get('/attendance/low'),
};

// Dashboard
export const dashboardApi = {
  get: () => api.get('/dashboard'),
};

/** Convert a plain object to a URL query string, skipping empty values. */
function toQuery(params) {
  const q = Object.entries(params)
    .filter(([, v]) => v !== '' && v !== null && v !== undefined)
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join('&');
  return q ? `?${q}` : '';
}
