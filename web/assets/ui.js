/* Amana UI primitives — built on React (UMD global) + htm (no build step). */
(function () {
  window.AM = window.AM || {};
  var React = window.React;
  var html = window.htm.bind(React.createElement);

  /* ---- icon set (stroke paths) ---- */
  var ICONS = {
    logo: 'M5 5h14v14H5z|M5 5h14v14H5z@45',
    dashboard: 'M4 13h7V4H4zM13 20h7v-9h-7zM13 4v5h7V4zM4 16v4h7v-4z',
    clients: 'M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75',
    catalog: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
    contracts: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6M9 13h6M9 17h6',
    calendar: 'M3 4h18v18H3zM3 9h18M8 2v4M16 2v4',
    code: 'M16 18l6-6-6-6M8 6l-6 6 6 6',
    settings: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-2.82 1.17V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 14H4.5a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 11 4.6V4.5a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 2.82 1.17l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 11h.1a2 2 0 1 1 0 4h-.1z',
    search: 'M21 21l-4.3-4.3M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14z',
    plus: 'M12 5v14M5 12h14',
    arrow: 'M5 12h14M13 6l6 6-6 6',
    back: 'M19 12H5M11 18l-6-6 6-6',
    check: 'M20 6L9 17l-5-5',
    x: 'M18 6L6 18M6 6l12 12',
    sun: 'M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4',
    moon: 'M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z',
    logout: 'M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9',
    bell: 'M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9M13.7 21a2 2 0 0 1-3.4 0',
    moon2: 'M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9z',
    coins: 'M8 14a6 6 0 1 0 0-12 6 6 0 0 0 0 12zM16 22a6 6 0 1 0 0-12 6 6 0 0 0 0 12z',
    shield: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
    clock: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM12 6v6l4 2',
    bolt: 'M13 2L3 14h7l-1 8 10-12h-7z',
    trend: 'M22 7l-9 9-4-4-6 6M16 7h6v6',
    key: 'M21 2l-2 2m-7.6 7.6a5 5 0 1 0-1.4 1.4l3.5-3.5m4-4l-4 4m4-4l2.5 2.5M14 8l2.5 2.5',
    trash: 'M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m3 0v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6',
    copy: 'M9 9h10v10H9zM5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1',
    phone: 'M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3 19.5 19.5 0 0 1-6-6 19.8 19.8 0 0 1-3-8.6A2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1 1 .4 1.9.7 2.8a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.8.7a2 2 0 0 1 1.7 2z',
    doc: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6',
    edit: 'M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z',
    info: 'M12 8h.01M11 12h2v6h-2zM12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20z',
    star: 'M12 2l3 6.5 7 .9-5 4.8 1.3 7-6.3-3.4L5.7 21 7 14.2 2 9.4l7-.9z',
    package: 'M16.5 9.4 7.5 4.2M21 16V8a2 2 0 0 0-1-1.7l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.7l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.3 7L12 12l8.7-5M12 22V12',
    truck: 'M1 3h15v13H1zM16 8h4l3 3v5h-7zM5.5 21a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5zM18.5 21a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5z',
  };

  function Icon(props) {
    var name = props.name, size = props.size || 18;
    var spec = ICONS[name] || ICONS.dashboard;
    var children = [];
    spec.split('|').forEach(function (part, i) {
      var rot = null, d = part;
      if (part.indexOf('@') >= 0) { var s = part.split('@'); d = s[0]; rot = s[1]; }
      children.push(html`<path key=${i} d=${d} transform=${rot ? 'rotate(' + rot + ' 12 12)' : undefined}></path>`);
    });
    return html`<svg viewBox="0 0 24 24" width=${size} height=${size} fill="none" stroke="currentColor"
      stroke-width=${props.sw || 1.8} stroke-linecap="round" stroke-linejoin="round" style=${props.style}>${children}</svg>`;
  }

  function Spinner() { return html`<span class="spinner"></span>`; }
  function Loading(p) { return html`<div class="center-load"><${Spinner}/> ${p.label || 'Загрузка…'}</div>`; }
  function Skeleton(p) {
    return html`<div class="skeleton" style=${Object.assign({ height: (p.h || 16) + 'px', width: p.w || '100%', borderRadius: (p.r == null ? 8 : p.r) + 'px' }, p.style)}></div>`;
  }
  function BlockError(p) {
    return html`<div class="banner banner-warn" style=${{ justifyContent: 'space-between', alignItems: 'center' }}>
      <div style=${{ display: 'flex', gap: 10 }}><${Icon} name="x" size=${17}/><div>${p.message || 'Не удалось загрузить'}</div></div>
      ${p.onRetry ? html`<button class="btn btn-sm btn-ghost" onClick=${p.onRetry}>Повторить</button>` : null}
    </div>`;
  }

  var labels = {
    contractStatus: { active: 'Активен', completed: 'Завершён', cancelled: 'Отменён', draft: 'Черновик' },
    installmentStatus: { paid: 'Оплачен', partially_paid: 'Частично', pending: 'Предстоит', overdue: 'Просрочен' },
    reminderStatus: { scheduled: 'Запланирована', overdue: 'Просрочена', completed: 'Выполнена', cancelled: 'Отменена' },
    halal: { halal: 'Халяль', haram: 'Харам', doubtful: 'Сомнительно' },
    halalChip: { halal: 'chip-halal', haram: 'chip-haram', doubtful: 'chip-doubt' },
    reminderType: { call: 'Звонок', delivery: 'Доставка', payment_followup: 'Контакт по платежу' },
  };

  function Chip(p) {
    return html`<span class=${'chip ' + (p.cls || '')}>${p.dot ? html`<span class="dot"></span>` : null}${p.label}</span>`;
  }
  function StatusChip(p) {
    var label = (labels[p.map] && labels[p.map][p.value]) || p.value;
    var cls = 'chip-' + p.value;
    if (p.map === 'halal') cls = labels.halalChip[p.value] || 'chip-info';
    if (p.map === 'reminderStatus' && p.value === 'scheduled') cls = 'chip-pending';
    return html`<${Chip} cls=${cls} label=${label}/>`;
  }

  function Field(p) {
    return html`<div class="field" style=${p.style}>
      ${p.label ? html`<label>${p.label}</label>` : null}
      ${p.children}
      ${p.hint ? html`<span class="hint">${p.hint}</span>` : null}
    </div>`;
  }

  function Modal(p) {
    return html`<div class="modal-overlay" onClick=${function (e) { if (e.target === e.currentTarget && p.onClose) p.onClose(); }}>
      <div class="modal" style=${p.width ? { maxWidth: p.width } : null}>
        <div class="modal-head">
          <div class="modal-title">${p.title}</div>
          ${p.onClose ? html`<button class="icon-btn" onClick=${p.onClose} aria-label="Закрыть"><${Icon} name="x" size=${18}/></button>` : null}
        </div>
        <div class="modal-body">${p.children}</div>
      </div>
    </div>`;
  }

  function Empty(p) {
    return html`<div class="empty">
      <${Icon} name=${p.icon || 'package'} size=${40}/>
      <div style=${{ fontWeight: 600, color: 'var(--fg-muted)', marginBottom: 4 }}>${p.title}</div>
      ${p.text ? html`<div style=${{ fontSize: 13.5 }}>${p.text}</div>` : null}
      ${p.action || null}
    </div>`;
  }

  AM.ui = { React: React, html: html, Icon: Icon, Spinner: Spinner, Loading: Loading, Skeleton: Skeleton,
    BlockError: BlockError, Chip: Chip, StatusChip: StatusChip, Field: Field, Modal: Modal, Empty: Empty, labels: labels };
})();
