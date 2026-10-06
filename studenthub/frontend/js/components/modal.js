/**
 * modal.js — Reusable accessible modal dialog.
 *
 * Usage:
 *   import { Modal } from '../components/modal.js';
 *   const m = new Modal('Edit Student', formHTML, { onConfirm, onCancel });
 *   m.open();
 *   m.close();
 *   m.destroy();
 */

export class Modal {
  /**
   * @param {string}   title      Dialog heading
   * @param {string}   bodyHTML   Inner HTML for the modal body
   * @param {object}   [options]
   * @param {Function} [options.onOpen]    Called after modal opens
   * @param {Function} [options.onClose]   Called when modal closes (any reason)
   */
  constructor(title, bodyHTML, options = {}) {
    this._opts = options;
    this._triggerEl = document.activeElement; // remember focus origin
    this._build(title, bodyHTML);
  }

  _build(title, bodyHTML) {
    // Overlay
    this._overlay = document.createElement('div');
    this._overlay.className = 'modal-overlay';
    this._overlay.setAttribute('aria-hidden', 'true');

    // Dialog
    this._dialog = document.createElement('div');
    this._dialog.className = 'modal';
    this._dialog.setAttribute('role', 'dialog');
    this._dialog.setAttribute('aria-modal', 'true');
    this._dialog.setAttribute('aria-labelledby', 'modal-title');
    this._dialog.tabIndex = -1;

    this._dialog.innerHTML = `
      <div class="modal__header">
        <h2 class="modal__title" id="modal-title">${escHtml(title)}</h2>
        <button class="modal__close" aria-label="Close dialog">✕</button>
      </div>
      <div class="modal__body">${bodyHTML}</div>
    `;

    this._overlay.appendChild(this._dialog);
    document.body.appendChild(this._overlay);

    // Event listeners
    this._dialog.querySelector('.modal__close').addEventListener('click', () => this.close());
    this._overlay.addEventListener('click', (e) => {
      if (e.target === this._overlay) this.close();
    });
    this._keyHandler = (e) => {
      if (e.key === 'Escape') { e.preventDefault(); this.close(); }
      if (e.key === 'Tab') this._trapFocus(e);
    };
  }

  /** Open the modal and move focus inside it. */
  open() {
    this._overlay.removeAttribute('aria-hidden');
    this._overlay.classList.add('modal-overlay--visible');
    document.body.classList.add('modal-open');
    document.addEventListener('keydown', this._keyHandler);

    // Move focus to the dialog.
    requestAnimationFrame(() => {
      const first = this._firstFocusable();
      (first ?? this._dialog).focus();
    });

    this._opts.onOpen?.();
    return this;
  }

  /** Close the modal and return focus to the triggering element. */
  close() {
    // Play close animation, then remove the visible class.
    this._overlay.classList.add('modal-overlay--closing');
    const done = () => {
      this._overlay.removeEventListener('animationend', done);
      this._overlay.setAttribute('aria-hidden', 'true');
      this._overlay.classList.remove('modal-overlay--visible');
      this._overlay.classList.remove('modal-overlay--closing');
      document.body.classList.remove('modal-open');
    };
    this._overlay.addEventListener('animationend', done, { once: true });
    // Fallback: if animationend never fires (e.g. reduced-motion), clean up after 220ms.
    setTimeout(() => {
      if (this._overlay.classList.contains('modal-overlay--closing')) done();
    }, 220);

    document.removeEventListener('keydown', this._keyHandler);
    this._triggerEl?.focus();
    this._opts.onClose?.();
  }

  /** Remove the modal from the DOM entirely. */
  destroy() {
    this.close();
    this._overlay.remove();
  }

  /** Expose the modal body element for external form manipulation. */
  get body() {
    return this._dialog.querySelector('.modal__body');
  }

  /** Trap keyboard focus within the modal. */
  _trapFocus(e) {
    const focusable = this._focusableEls();
    if (!focusable.length) return;
    const first = focusable[0];
    const last  = focusable[focusable.length - 1];
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus(); }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  }

  _focusableEls() {
    return Array.from(this._dialog.querySelectorAll(
      'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )).filter(el => !el.closest('[hidden]'));
  }

  _firstFocusable() {
    return this._focusableEls()[0] ?? null;
  }
}

function escHtml(s) {
  return String(s ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
