/* Amana client portal — a separate self-contained app (client token, /api/portal).
   Reached at #/portal; clients log in with credentials issued by staff and chat
   with the company + see their own installment contracts. */
(function () {
  window.AM = window.AM || {};
  var React = window.React;
  var ui = window.AM.ui, html = ui.html, Icon = ui.Icon, api = window.AM.api, fmt = window.AM.fmt;
  var useState = React.useState, useEffect = React.useEffect;

  function Portal(p) {
    var au = useState(api.portal.isAuthed()), authed = au[0], setAuthed = au[1];
    function onAuthed() { setAuthed(true); }
    function logout() { api.portal.logout(); setAuthed(false); }
    return authed
      ? html`<${PortalApp} logout=${logout} toggleTheme=${p.toggleTheme} theme=${p.theme}/>`
      : html`<${PortalLogin} onAuthed=${onAuthed}/>`;
  }

  function PortalLogin(p) {
    var f = useState({ email: '', password: '' }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    var er = useState(''), err = er[0], setErr = er[1];
    function submit(e) {
      if (e) e.preventDefault();
      if (!v.email.trim() || !v.password) { setErr('Введите email и пароль'); return; }
      setBusy(true); setErr('');
      api.portal.login(v.email.trim(), v.password).then(function () { p.onAuthed(); })
        .catch(function (ex) { setBusy(false); setErr(ex.status === 401 ? 'Неверный email или пароль' : ex.message); });
    }
    var inp = function (k, ph, type) { return html`<input class="input" type=${type || 'text'} value=${v[k]} placeholder=${ph}
      onInput=${function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }}/>`; };
    return html`<div class="portal-auth">
      <form class="portal-auth-card" onSubmit=${submit}>
        <div class="portal-brand"><span class="brand-badge"><${Icon} name="logo" size=${20}/></span>
          <div><div class="brand-name">Амана</div><div class="brand-sub">Кабинет клиента</div></div></div>
        <div class="portal-auth-title">Вход в личный кабинет</div>
        <div class="portal-auth-sub">Войдите по данным, которые выдал вам менеджер.</div>
        ${err ? html`<div class="banner banner-warn" style=${{ marginBottom: 12 }}>${err}</div>` : null}
        <${ui.Field} label="Email">${inp('email', 'client@mail.ru')}<//>
        <${ui.Field} label="Пароль">${inp('password', '••••••••', 'password')}<//>
        <button class="btn btn-primary btn-block" disabled=${busy} type="submit">${busy ? html`<${ui.Spinner}/>` : 'Войти'}</button>
        <a class="portal-back" href="#/dashboard">← Вернуться на сайт</a>
      </form>
    </div>`;
  }

  function PortalApp(p) {
    var me = useState(null), profile = me[0], setProfile = me[1];
    var tb = useState('chat'), active = tb[0], setActive = tb[1];
    useEffect(function () {
      api.portal.me().then(setProfile).catch(function (e) { if (e.status === 401) p.logout(); });
    }, []);
    return html`<div class="portal-shell">
      <header class="portal-top">
        <div class="portal-brand"><span class="brand-badge"><${Icon} name="logo" size=${18}/></span>
          <div><div class="brand-name">Амана</div><div class="brand-sub">Кабинет клиента</div></div></div>
        <div style=${{ flex: 1 }}></div>
        ${profile ? html`<div class="portal-user">${profile.full_name}</div>` : null}
        <button class="icon-btn" onClick=${p.toggleTheme} title="Тема"><${Icon} name=${p.theme === 'dark' ? 'sun' : 'moon'} size=${18}/></button>
        <button class="icon-btn" onClick=${p.logout} title="Выйти"><${Icon} name="logout" size=${18}/></button>
      </header>
      <div class="portal-tabs">
        <button class=${'tab ' + (active === 'chat' ? 'active' : '')} onClick=${function () { setActive('chat'); }}>Чат с менеджером</button>
        <button class=${'tab ' + (active === 'contracts' ? 'active' : '')} onClick=${function () { setActive('contracts'); }}>Мои рассрочки</button>
      </div>
      <main class="portal-main">
        ${active === 'chat'
          ? html`<div class="card portal-chat-card">
              <${ui.ChatThread} threadKey="portal" meKind="client"
                load=${api.portal.messages} onSend=${api.portal.send}
                placeholder="Напишите менеджеру…" emptyText="Задайте вопрос — менеджер ответит здесь."/>
            </div>`
          : html`<${PortalContracts}/>`}
      </main>
    </div>`;
  }

  function PortalContracts() {
    var sel = useState(null), id = sel[0], setId = sel[1];
    if (id) return html`<${PortalContractDetail} id=${id} onBack=${function () { setId(null); }}/>`;
    return html`<${PortalContractList} onOpen=${setId}/>`;
  }

  function PortalContractList(p) {
    var st = useState({ list: null, loading: true, err: '' }), s = st[0], set = st[1];
    useEffect(function () {
      api.portal.contracts().then(function (r) { set({ list: r || [], loading: false, err: '' }); })
        .catch(function (e) { set({ list: [], loading: false, err: e.message }); });
    }, []);
    if (s.loading) return html`<${ui.Loading}/>`;
    if (s.err) return html`<div class="banner banner-warn">${s.err}</div>`;
    if (!s.list.length) return html`<div class="card"><${ui.Empty} icon="contracts" title="Рассрочек пока нет" text="Оставьте заявку — менеджер оформит для вас договор."/></div>`;
    return html`<div class="entity-list">${s.list.map(function (c) {
      return html`<button key=${c.id} class="card card-pad portal-contract-card" onClick=${function () { p.onOpen(c.id); }}>
        <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12 }}>
          <div style=${{ textAlign: 'left' }}><div style=${{ fontWeight: 700 }}>Договор ${c.id.slice(0, 8)}</div>
            <div class="page-sub">Оформлен ${fmt.date(c.created_at)} · ${c.installments} платежей</div></div>
          <${ui.StatusChip} map="contractStatus" value=${c.status}/>
        </div>
        <div class="compact-fields compact-fields-2" style=${{ marginTop: 14 }}>
          <div class="compact-field"><span>Цена</span><strong class="amana-num">${fmt.money(c.sale_price)}</strong></div>
          <div class="compact-field"><span>Остаток</span><strong class="amana-num">${fmt.money(c.outstanding)}</strong></div>
        </div>
        <div class="portal-open-link">Открыть и посмотреть график <${Icon} name="arrow" size=${15}/></div>
      </button>`;
    })}</div>`;
  }

  function PortalContractDetail(p) {
    var st = useState({ d: null, loading: true, err: '' }), s = st[0], set = st[1];
    useEffect(function () {
      api.portal.contract(p.id).then(function (r) { set({ d: r, loading: false, err: '' }); })
        .catch(function (e) { set({ d: null, loading: false, err: e.message }); });
    }, [p.id]);
    return html`<div>
      <button class="btn btn-soft btn-sm" onClick=${p.onBack}><${Icon} name="back" size=${15}/> К списку</button>
      ${s.loading ? html`<div style=${{ marginTop: 16 }}><${ui.Loading}/></div>`
        : s.err ? html`<div class="banner banner-warn" style=${{ marginTop: 16 }}>${s.err}</div>`
        : (function () {
          var d = s.d;
          var financed = parseFloat(d.financed_amount) || 0, paid = parseFloat(d.paid_amount) || 0;
          var progress = financed > 0 ? Math.min(100, Math.round(paid / financed * 100)) : 0;
          return html`<div class="grid" style=${{ gap: 14, marginTop: 14 }}>
            <div class="card card-pad">
              <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12 }}>
                <div><div style=${{ fontWeight: 700, fontSize: 18 }}>Договор ${d.id.slice(0, 8)}</div>
                  <div class="page-sub">Оформлен ${fmt.date(d.created_at)}</div></div>
                <${ui.StatusChip} map="contractStatus" value=${d.status}/>
              </div>
              ${d.has_next && d.status === 'active' ? html`<div class="portal-next">
                  <div><div class="portal-next-label">Следующий платёж</div>
                    <div class="portal-next-amt amana-num">${fmt.money(d.next_due_amount)}</div></div>
                  <div class="portal-next-date">до ${fmt.dateLong(d.next_due_date)}</div>
                </div>`
                : (d.status === 'completed' ? html`<div class="banner banner-accent" style=${{ marginTop: 14 }}>
                    <${Icon} name="check" size=${17}/> Рассрочка полностью оплачена. БаракаЛлаху фика!</div>` : null)}
              ${d.has_overdue ? html`<div class="banner banner-warn" style=${{ marginTop: 10 }}>
                <${Icon} name="info" size=${16}/> Есть просроченный платёж — свяжитесь с менеджером в чате.</div>` : null}
              <div class="portal-progress-row" style=${{ marginTop: 16 }}>
                <div class="progress" style=${{ flex: 1 }}><i style=${{ width: progress + '%' }}></i></div>
                <span class="amana-num" style=${{ fontSize: 13, fontWeight: 600 }}>${progress}%</span>
              </div>
              <div class="compact-fields compact-fields-3" style=${{ marginTop: 14 }}>
                <div class="compact-field"><span>Цена</span><strong class="amana-num">${fmt.money(d.sale_price)}</strong></div>
                <div class="compact-field"><span>Оплачено</span><strong class="amana-num">${fmt.money(d.paid_amount)}</strong></div>
                <div class="compact-field"><span>Остаток</span><strong class="amana-num">${fmt.money(d.outstanding)}</strong></div>
              </div>
              <div class="banner banner-accent" style=${{ marginTop: 14 }}>
                <${Icon} name="shield" size=${17}/> Цена зафиксирована при оформлении. Долг не растёт со временем — 0% риба.</div>
            </div>
            <div class="table-card">
              <div class="table-card-head">График платежей — сколько и когда</div>
              <table class="data-table"><thead><tr><th>№</th><th>Дата</th><th>Сумма</th><th>Статус</th></tr></thead>
                <tbody>${d.schedule.map(function (it) {
                  return html`<tr key=${it.number} class=${'data-row ' + (it.status === 'overdue' ? 'row-overdue' : '')}>
                    <td><strong>${it.number}</strong></td>
                    <td>${fmt.dateLong(it.due_date)}</td>
                    <td class="table-money amana-num">${fmt.money(it.amount)}</td>
                    <td><${ui.StatusChip} map="installmentStatus" value=${it.status}/></td>
                  </tr>`;
                })}</tbody></table>
            </div>
            ${d.payments.length ? html`<div class="table-card">
              <div class="table-card-head">История платежей</div>
              <table class="data-table"><thead><tr><th>Дата</th><th>Сумма</th></tr></thead>
                <tbody>${d.payments.map(function (pm, i) {
                  return html`<tr key=${i} class="data-row"><td>${fmt.dateTime(pm.paid_at)}</td>
                    <td class="table-money amana-num delta-pos">+${fmt.money(pm.amount)}</td></tr>`;
                })}</tbody></table>
            </div>` : null}
          </div>`;
        })()}
    </div>`;
  }

  window.AM.Portal = Portal;
})();
