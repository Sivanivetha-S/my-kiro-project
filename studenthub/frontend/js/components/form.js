/**
 * form.js — Form validation display helpers.
 *
 * design.md § 6.7
 */

/**
 * Display field-level validation errors.
 * Each error in `errors` must have { field, message }.
 * Looks for <span data-error="field-name" role="alert"> elements.
 * @param {HTMLElement} container  Form element or parent
 * @param {Array<{field:string, message:string}>} errors
 */
export function displayErrors(container, errors) {
  clearErrors(container);
  for (const { field, message } of errors) {
    const el = container.querySelector(`[data-error="${field}"]`);
    if (el) {
      el.textContent = message;
      el.hidden = false;
    }
    // Also mark the input as invalid.
    const input = container.querySelector(`[name="${field}"], #${field}`);
    if (input) {
      input.setAttribute('aria-invalid', 'true');
    }
  }
}

/**
 * Clear all displayed validation errors in a container.
 * @param {HTMLElement} container
 */
export function clearErrors(container) {
  container.querySelectorAll('[data-error]').forEach(el => {
    el.textContent = '';
    el.hidden = true;
  });
  container.querySelectorAll('[aria-invalid]').forEach(el => {
    el.removeAttribute('aria-invalid');
  });
}

/**
 * Collect form values as a plain object keyed by input name.
 * Handles text, number, select, and checkbox.
 * @param {HTMLFormElement} form
 * @returns {object}
 */
export function collectForm(form) {
  const data = {};
  const fd = new FormData(form);
  for (const [key, value] of fd.entries()) {
    data[key] = value;
  }
  return data;
}
