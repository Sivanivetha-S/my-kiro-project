/**
 * toast.js — Non-blocking notification toasts.
 *
 * Usage:
 *   import { toast } from '../components/toast.js';
 *   toast.success('Student created!');
 *   toast.error('Failed to load data.');
 */

let container = null;

function getContainer() {
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    container.setAttribute('aria-live', 'polite');
    container.setAttribute('aria-atomic', 'false');
    document.body.appendChild(container);
  }
  return container;
}

/**
 * Show a toast notification.
 * @param {'success'|'error'|'info'} type
 * @param {string} message
 * @param {number} duration  ms before auto-dismiss (default 4000)
 */
function show(type, message, duration = 4000) {
  const c = getContainer();
  const el = document.createElement('div');
  el.className = `toast toast--${type}`;
  el.setAttribute('role', 'status');
  el.innerHTML = `
    <span class="toast__message">${escHtml(message)}</span>
    <button class="toast__close" aria-label="Dismiss notification">✕</button>
  `;

  const dismiss = () => {
    el.classList.add('toast--leaving');
    el.addEventListener('animationend', () => el.remove(), { once: true });
  };

  el.querySelector('.toast__close').addEventListener('click', dismiss);
  c.appendChild(el);

  setTimeout(dismiss, duration);
}

function escHtml(s) {
  return String(s ?? '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

export const toast = {
  success: (msg, dur) => show('success', msg, dur),
  error:   (msg, dur) => show('error',   msg, dur),
  info:    (msg, dur) => show('info',    msg, dur),
};
