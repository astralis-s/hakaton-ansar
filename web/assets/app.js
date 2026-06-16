/* Amana SPA root — theme, auth, hash routing, app shell, toast & confirm. */
(function () {
  var React = window.React, ReactDOM = window.ReactDOM;
  var ui = window.AM.ui, html = ui.html, Icon = ui.Icon, api = AM.api;
  var useState = React.useState, useEffect = React.useEffect;

  var NAV = [
    ['dashboard', 'Дашборд', 'dashboard'],
    ['clients', 'Клиенты', 'clients'],
    ['catalog', 'Каталог', 'catalog'],
    ['contracts', 'Договоры', 'contracts'],
    ['chat', 'Чат', 'chat'],
    ['finance', 'Финансы', 'coins'],
    ['schedule', 'Календарь', 'calendar'],
    ['developers', 'Разработчикам', 'code', true],
    ['settings', 'Настройки', 'settings'],
  ];

  function parseHash() {
    var h = (location.hash || '#/dashboard').replace(/^#\/?/, '');
    var parts = h.split('/');
    return { page: parts[0] || 'dashboard', id: parts[1] || null };
  }

  function Root() {
    var th = useState(localStorage.getItem('amana.theme') || 'light'), theme = th[0], setTheme = th[1];
    var au = useState(api.isAuthed()), authed = au[0], setAuthed = au[1];
    var rt = useState(parseHash()), route = rt[0], setRoute = rt[1];
    var ts = useState(null), toast = ts[0], setToast = ts[1];
    var cf = useState(null), confirm = cf[0], setConfirm = cf[1];
    var meS = useState(null), me = meS[0], setMe = meS[1];

    useEffect(function () { document.documentElement.setAttribute('data-theme', theme); localStorage.setItem('amana.theme', theme); }, [theme]);
    useEffect(function () { function onHash() { setRoute(parseHash()); window.scrollTo(0, 0); } window.addEventListener('hashchange', onHash); return function () { window.removeEventListener('hashchange', onHash); }; }, []);
    useEffect(function () {
      if (!authed) { setMe(null); return; }
      api.me().then(setMe).catch(function (e) { if (e.status === 401) { api.logout(); setAuthed(false); } });
    }, [authed]);

    function toggleTheme() { setTheme(theme === 'dark' ? 'light' : 'dark'); }
    function go(page, id) { location.hash = '#/' + page + (id ? '/' + id : ''); }
    function flash(msg, err) { setToast({ msg: msg, err: err }); clearTimeout(flash._t); flash._t = setTimeout(function () { setToast(null); }, 2600); }
    function onAuthed() { setAuthed(true); go('dashboard'); flash('Вход выполнен'); }
    function logout() { api.logout(); setAuthed(false); setMe(null); location.hash = ''; }

    if (!authed) return html`<${window.AM.Landing} theme=${theme} toggleTheme=${toggleTheme} onAuthed=${onAuthed}/>`;

    var isOwner = me ? me.role === 'owner' : false;
    var ctx = { route: route, go: go, toast: flash, confirm: setConfirm, isOwner: isOwner, me: me, theme: theme, toggleTheme: toggleTheme };
    var Screen = AM.screens[route.page] || AM.screens.dashboard;
    var activePage = route.page === 'contract' || route.page === 'contract-new' ? 'contracts'
      : route.page === 'client' ? 'clients'
      : route.page === 'product' ? 'catalog'
      : route.page === 'reminder' ? 'schedule'
      : route.page;

    return html`<div class="shell">
      <aside class="sidebar">
        <div class="sidebar-panel">
          <div class="brand">
            <span class="brand-badge"><${Icon} name="logo" size=${19} sw=${1.9}/></span>
            <div class="brand-copy">
              <span class="brand-name">Амана</span>
              <span class="brand-sub">CRM для рассрочек</span>
            </div>
          </div>
          <div class="sidebar-nav">
            ${NAV.map(function (n) {
              if (n[3] && !isOwner) return null;
              return html`<a key=${n[0]} class=${'nav-item ' + (activePage === n[0] ? 'active' : '')} href=${'#/' + n[0]}><${Icon} name=${n[2]} size=${18}/> <span>${n[1]}</span></a>`;
            })}
          </div>
          <div class="sidebar-footer">
            <div class="nav-sep"></div>
            <div class="sidebar-meta">
              <div class="sidebar-meta-title">${me ? me.full_name : 'Пользователь'}</div>
              <div class="sidebar-meta-sub">${isOwner ? 'Владелец системы' : 'Менеджер системы'}</div>
            </div>
            <a class="nav-item nav-item-exit" onClick=${logout} style=${{ cursor: 'pointer' }}><${Icon} name="logout" size=${18}/> <span>Выйти</span></a>
          </div>
        </div>
      </aside>
      <div>
        <div class="topbar">
          <div style=${{ fontWeight: 700, letterSpacing: '-.02em' }}>${me ? me.full_name : ''}
            <span style=${{ marginLeft: 8 }} class=${'chip ' + (isOwner ? 'chip-paid' : 'chip-info')}>${isOwner ? 'Владелец' : 'Менеджер'}</span></div>
          <div style=${{ flex: 1 }}></div>
          <button class="icon-btn" onClick=${toggleTheme} title="Тема"><${Icon} name=${theme === 'dark' ? 'sun' : 'moon'} size=${18}/></button>
          <a class="icon-btn" href="/swagger/" target="_blank" title="API"><${Icon} name="code" size=${18}/></a>
        </div>
        <main class="main"><${Screen} ...${ctx}/></main>
      </div>

      ${toast ? html`<div class=${'toast ' + (toast.err ? 'err' : '')}><${Icon} name=${toast.err ? 'x' : 'check'} size=${17}/> ${toast.msg}</div>` : null}
      ${confirm ? html`<${ConfirmModal} c=${confirm} onClose=${function () { setConfirm(null); }}/>` : null}
    </div>`;
  }

  function ConfirmModal(p) {
    var c = p.c;
    return html`<${ui.Modal} title=${c.title} onClose=${p.onClose}>
      <div style=${{ color: 'var(--fg-muted)', lineHeight: 1.6 }}>${c.text}</div>
      <div style=${{ display: 'flex', gap: 10, justifyContent: 'flex-end', marginTop: 6 }}>
        <button class="btn btn-ghost" onClick=${p.onClose}>Отмена</button>
        <button class=${'btn ' + (c.danger ? 'btn-danger' : 'btn-primary')} onClick=${function () { p.onClose(); c.onOk(); }}>${c.okLabel || 'OK'}</button>
      </div>
    <//>`;
  }

  ReactDOM.createRoot(document.getElementById('root')).render(html`<${Root}/>`);
})();
