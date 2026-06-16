/* Amana authenticated screens — all wired to the Go API. */
(function () {
  window.AM = window.AM || {};
  var React = window.React;
  var ui = window.AM.ui, html = ui.html, Icon = ui.Icon, api = AM.api, fmt = AM.fmt;
  var useState = React.useState, useEffect = React.useEffect, useRef = React.useRef;

  /* ---- data loading hook ---- */
  function useAsync(fn, deps) {
    var d = useState(null), data = d[0], setData = d[1];
    var l = useState(true), loading = l[0], setLoading = l[1];
    var e = useState(''), err = e[0], setErr = e[1];
    var t = useState(0), tick = t[0], setTick = t[1];
    useEffect(function () {
      var alive = true; setLoading(true); setErr('');
      fn().then(function (r) { if (alive) { setData(r); setLoading(false); } })
        .catch(function (ex) { if (alive) { setErr(ex.message || 'Ошибка загрузки'); setLoading(false); } });
      return function () { alive = false; };
    }, (deps || []).concat([tick]));
    return { data: data, loading: loading, err: err, reload: function () { setTick(tick + 1); } };
  }

  function Guard(p) {
    if (p.loading) return html`<${ui.Loading}/>`;
    if (p.err) return html`<div class="banner banner-warn">${p.err}</div>`;
    return p.children;
  }

  function PageHead(p) {
    return html`<div class="page-head">
      <div><div class="page-title">${p.title}</div>${p.sub ? html`<div class="page-sub">${p.sub}</div>` : null}</div>
      ${p.actions || null}
    </div>`;
  }

  /* ================= DASHBOARD =================
     Priority (visual weight, top→bottom): overdue (names/amounts/days) →
     this week's expected vs collected (+ collection rate) → today's agenda
     (namaz-aware) → portfolio balance. Money is computed on the backend
     (GET /api/app/dashboard) and only displayed here. Per-block loading/error. */
  function Dashboard(ctx) {
    var dash = useAsync(api.dashboard);     // overdue + week + portfolio (one source)
    var rem = useAsync(api.listReminders);  // today's agenda (separate source)
    var qp = useState(false), quickPay = qp[0], setQuickPay = qp[1];
    var cm = useState(false), clientOpen = cm[0], setClientOpen = cm[1];
    var ab = useState({}), agendaBusy = ab[0], setAgendaBusy = ab[1];
    var agenda = agendaForToday(rem.data || []);
    function quickReminderAction(reminder, action) {
      setAgendaBusy(function (prev) {
        var next = Object.assign({}, prev);
        next[reminder.id] = action;
        return next;
      });
      var req = action === 'complete' ? api.completeReminder(reminder.id) : api.cancelReminder(reminder.id);
      req.then(function () {
        setAgendaBusy(function (prev) {
          var next = Object.assign({}, prev);
          delete next[reminder.id];
          return next;
        });
        rem.reload();
        ctx.toast(action === 'complete' ? 'Задача выполнена' : 'Задача отменена');
      }).catch(function (e) {
        setAgendaBusy(function (prev) {
          var next = Object.assign({}, prev);
          delete next[reminder.id];
          return next;
        });
        ctx.toast(e.message, true);
      });
    }

    var head = html`<${PageHead} title="Дашборд" sub="Просрочки, платежи на неделю и задачи на сегодня"
      actions=${html`<button class="btn btn-primary" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> Новый договор</button>`}/>`;

    if (dash.loading && !dash.data) {
      return html`<div>${head}
        <div class="grid" style=${{ gap: 18 }}>
          <div class="dash-grid-main">
            <section class="dash-panel dash-loading"><${ui.Skeleton} h=${18} w="22%"/><div style=${{ height: 14 }}></div><${ui.Skeleton} h=${58} style=${{ marginBottom: 10 }}/><${ui.Skeleton} h=${58} style=${{ marginBottom: 10 }}/><${ui.Skeleton} h=${58}/></section>
            <section class="dash-panel dash-loading"><${ui.Skeleton} h=${18} w="34%"/><div style=${{ height: 14 }}></div><${ui.Skeleton} h=${88} style=${{ marginBottom: 12 }}/><${ui.Skeleton} h=${42} style=${{ marginBottom: 8 }}/><${ui.Skeleton} h=${42}/></section>
          </div>
          <div class="dash-grid-lower">
            <section class="dash-panel dash-loading"><${ui.Skeleton} h=${18} w="30%"/><div style=${{ height: 14 }}></div><${ui.Skeleton} h=${52} style=${{ marginBottom: 8 }}/><${ui.Skeleton} h=${52}/></section>
            <section class="dash-panel dash-loading"><${ui.Skeleton} h=${18} w="32%"/><div style=${{ height: 14 }}></div><${ui.Skeleton} h=${104}/></section>
          </div>
        </div></div>`;
    }

    var d = dash.data;
    var emptyOrg = d && d.portfolio.active_contracts === 0 && d.overdue.length === 0 && d.week.upcoming.length === 0;
    if (emptyOrg) {
      return html`<div>${head}
        <section class="dash-panel dash-panel-empty">
          <${ui.Empty} icon="contracts" title="Добро пожаловать в «Аману»"
            text="После создания клиентов и договоров здесь появятся просрочки, ближайшие платежи и задачи на сегодня."
            action=${html`<div style=${{ display: 'flex', gap: 10, justifyContent: 'center', marginTop: 16 }}>
              <button class="btn btn-ghost" onClick=${function () { setClientOpen(true); }}><${Icon} name="clients" size=${17}/> Новый клиент</button>
              <button class="btn btn-primary" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> Новый договор</button>
            </div>`}/>
        </section>
        ${clientOpen ? html`<${ClientModal} ctx=${ctx} onClose=${function () { setClientOpen(false); }} onSaved=${function () { setClientOpen(false); ctx.toast('Клиент добавлен'); }}/>` : null}
      </div>`;
    }

    var rate = weekRate(d);

    return html`<div>
      ${head}

      ${dash.err ? html`<${ui.BlockError} message=${'Не удалось загрузить сводку: ' + dash.err} onRetry=${dash.reload}/>`
        : html`<div class="dash-stack">
            <section class="dash-panel dash-panel-week dash-panel-week-large">
              <div class="dash-section-head">
                <div>
                  <div class="dash-section-kicker dash-section-kicker-strong">Неделя к получению</div>
                  <h3>Платежи на этой неделе</h3>
                </div>
                <div class="dash-week-inline-rate amana-num">${rate}% собрано</div>
              </div>
              <div class="dash-week-layout">
                <div class="dash-week-summary">
                  <div class="dash-meter-block">
                    <div class="dash-meter-top">
                      <div>
                        <div class="dash-meter-label">Получено</div>
                        <div class="dash-meter-value dash-meter-value-strong amana-num">${fmt.money(d.week.collected)}</div>
                      </div>
                      <div class="dash-meter-side dash-meter-side-expected">
                        <span>К получению</span>
                        <strong class="amana-num">${fmt.money(d.week.expected)}</strong>
                      </div>
                    </div>
                    <div class="progress dash-progress"><i style=${{ width: rate + '%' }}></i></div>
                  </div>
                  <div class="dash-overdue-mini">
                    <div class="dash-overdue-mini-head">
                      <div class="dash-ops-label">Просроченные платежи</div>
                      <div class="dash-mini-note">${Math.min(d.overdue.length, 2)} из ${d.overdue.length}</div>
                    </div>
                    ${d.overdue.length === 0
                      ? html`<div class="dash-empty-note">Просроченных платежей нет.</div>`
                      : html`<div class="dash-overdue-mini-list">${d.overdue.slice(0, 2).map(function (o) {
                          return html`<button key=${o.contract_id} class="dash-overdue-mini-row" onClick=${function () { ctx.go('contract', o.contract_id); }}>
                            <div class="dash-overdue-mini-main">
                              <div class="dash-overdue-mini-name">${o.client_name || 'Клиент'}</div>
                              <div class="dash-overdue-mini-meta">${o.days_overdue} ${plural(o.days_overdue, 'день', 'дня', 'дней')} просрочки</div>
                            </div>
                            <div class="dash-overdue-mini-amount amana-num">${fmt.money(o.outstanding)}</div>
                            <div class="dash-overdue-mini-arrow"><${Icon} name="arrow" size=${15}/></div>
                          </button>`;
                        })}</div>`}
                  </div>
                </div>
                <div class="dash-week-list dash-week-list-box">
                  <div class="dash-week-list-head">
                    <div class="dash-week-list-title">Ближайшие платежи недели</div>
                    <div class="dash-week-count amana-num">${d.week.upcoming.length}</div>
                  </div>
                  ${d.week.upcoming.length === 0
                    ? html`<div class="dash-empty-note">На этой неделе новых платежей по графику нет.</div>`
                    : html`<div class="dash-ledger">${d.week.upcoming.map(function (u) {
                        return html`<button key=${u.contract_id + u.due_date} class="dash-ledger-row" onClick=${function () { ctx.go('contract', u.contract_id); }}>
                          <div class="dash-ledger-date">
                            <span class="dash-ledger-day">Срок: ${fmt.date(u.due_date)}</span>
                            <span class="dash-ledger-client">${u.client_name || 'Клиент'}</span>
                          </div>
                          <div class="dash-ledger-amount amana-num">${fmt.money(u.amount)}</div>
                          <${ui.StatusChip} map="installmentStatus" value=${u.status}/>
                        </button>`;
                      })}</div>`}
                </div>
              </div>
              ${d.overdue.length > 0 ? html`<button class="dash-week-alert" onClick=${function () { ctx.go('schedule'); }}>
                <span><${Icon} name="clock" size=${16}/> Есть просроченные платежи</span>
                <span class="dash-week-alert-link">Открыть календарь</span>
              </button>` : null}
            </section>

            <div class="dash-grid-lower">
              <section class="dash-panel dash-panel-agenda">
                <div class="dash-section-head">
                  <div>
                    <div class="dash-section-kicker">Повестка на сегодня</div>
                    <h3>События на сегодня</h3>
                  </div>
                  <div class="dash-agenda-count amana-num">${rem.loading ? '…' : agenda.length}</div>
                </div>
                ${rem.loading ? html`<${ui.Skeleton} h=${54} style=${{ marginBottom: 8 }}/><${ui.Skeleton} h=${54}/>`
                  : rem.err ? html`<${ui.BlockError} message="Не удалось загрузить задачи" onRetry=${rem.reload}/>`
                  : agenda.length === 0 ? html`<div class="dash-empty-note">На сегодня задач нет.</div>`
                  : html`<div class="dash-agenda-list">${agenda.map(function (r) {
                      var busy = agendaBusy[r.id] || '';
                      return html`<button key=${r.id} class="dash-agenda-row dash-agenda-row-link" onClick=${function () { ctx.go('reminder', r.id); }}>
                        <div class="dash-agenda-time amana-num">${fmt.time(r.scheduled_at)}</div>
                        <div class="dash-agenda-track"><span></span></div>
                        <div class="dash-agenda-body">
                          <div class="dash-agenda-topline">
                            <div class="dash-agenda-title">
                              <span class="dash-agenda-icon"><${Icon} name=${r.type === 'delivery' ? 'truck' : r.type === 'call' ? 'phone' : 'coins'} size=${16}/></span>
                              <span>${ui.labels.reminderType[r.type] || r.type}</span>
                            </div>
                            <${ReminderQuickActions} reminder=${r} busy=${busy} variant="dashboard"
                              onComplete=${function (item) { quickReminderAction(item, 'complete'); }}
                              onCancel=${function (item) { quickReminderAction(item, 'cancel'); }}/>
                          </div>
                          ${r.note ? html`<div class="dash-agenda-note">${r.note}</div>` : null}
                          ${r.was_shifted ? html`<div class="dash-agenda-shift">Перенесено из-за намаза: ${r.reason}</div>` : null}
                        </div>
                      </button>`;
                    })}</div>`}
              </section>

              <section class="dash-panel dash-panel-ops">
                <div class="dash-section-head">
                  <div>
                    <div class="dash-section-kicker dash-section-kicker-strong">Активный портфель</div>
                    <h3>Всего к получению</h3>
                  </div>
                </div>
                <div class="dash-portfolio-band">
                  <div class="dash-portfolio-value amana-num">${fmt.money(d.portfolio.outstanding)}</div>
                  <div class="dash-portfolio-contracts amana-num">${d.portfolio.active_contracts}</div>
                  <div class="dash-portfolio-contracts-note">${plural(d.portfolio.active_contracts, 'активный договор', 'активных договора', 'активных договоров')}</div>
                </div>
                <div class="dash-ops-title">Быстрые действия</div>
                <div class="dash-ops-actions dash-ops-actions-grid">
                  <button class="btn btn-primary dash-action-btn" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> <span>Создать договор</span></button>
                  <button class="btn btn-ghost dash-action-btn" onClick=${function () { setClientOpen(true); }}><${Icon} name="clients" size=${17}/> <span>Добавить клиента</span></button>
                  <button class="btn btn-ghost dash-action-btn" onClick=${function () { setQuickPay(true); }}><${Icon} name="coins" size=${17}/> <span>Внести платёж</span></button>
                </div>
              </section>
            </div>
          </div>`}

      ${quickPay ? html`<${QuickPay} ctx=${ctx} onClose=${function () { setQuickPay(false); }} onDone=${function () { setQuickPay(false); dash.reload(); ctx.toast('Платёж принят'); }}/>` : null}
      ${clientOpen ? html`<${ClientModal} ctx=${ctx} onClose=${function () { setClientOpen(false); }} onSaved=${function () { setClientOpen(false); ctx.toast('Клиент добавлен'); }}/>` : null}
    </div>`;
  }

  function agendaForToday(list) {
    return (list || []).filter(function (r) { return isToday(r.scheduled_at) && r.base_status === 'scheduled'; });
  }
  function weekRate(d) {
    var r = parseFloat((d && d.week && d.week.collection_rate_percent) || '0') || 0;
    return Math.max(0, Math.min(100, Math.round(r)));
  }
  function isToday(iso) {
    var d = new Date(iso), n = new Date();
    return d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate();
  }
  function plural(n, one, few, many) {
    var m10 = n % 10, m100 = n % 100;
    if (m10 === 1 && m100 !== 11) return one;
    if (m10 >= 2 && m10 <= 4 && (m100 < 10 || m100 >= 20)) return few;
    return many;
  }
  function QuickPay(p) {
    var st = useAsync(api.listContracts);
    var s = useState(null), selected = s[0], setSel = s[1];
    if (selected) return html`<${PaymentModal} c=${selected} ctx=${p.ctx} onClose=${p.onClose} onDone=${p.onDone}/>`;
    var active = (st.data || []).filter(function (c) { return c.status === 'active'; });
    return html`<${ui.Modal} title="Принять платёж — выберите договор" onClose=${p.onClose} width=${460}>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${active.length === 0 ? html`<${ui.Empty} icon="contracts" title="Нет активных договоров"/>`
          : html`<div style=${{ display: 'flex', flexDirection: 'column', gap: 8 }}>${active.map(function (c) {
            return html`<div key=${c.id} class="select-card" onClick=${function () { setSel({ id: c.id, outstanding: c.outstanding }); }}>
              <div style=${{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
                <span class="amana-num" style=${{ fontWeight: 600 }}>${fmt.money(c.sale_price)}</span>
                <span class="amana-num" style=${{ color: 'var(--fg-muted)' }}>остаток ${fmt.money(c.outstanding)}</span></div></div>`;
          })}</div>`}
      <//>
    <//>`;
  }

  /* ================= CLIENTS ================= */
  function Clients(ctx) {
    var st = useAsync(api.listClients);
    var contracts = useAsync(api.listContracts);
    var q = useState(''), query = q[0], setQuery = q[1];
    var m = useState(false), open = m[0], setOpen = m[1];
    var list = (st.data || []).filter(function (c) { return c.full_name.toLowerCase().indexOf(query.toLowerCase()) >= 0; });
    return html`<div>
      <${PageHead} title="Клиенты" sub=${(st.data || []).length + ' клиентов'}
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новый клиент</button>`}/>
      <div class="search" style=${{ marginBottom: 16 }}><${Icon} name="search" size=${17} style=${{ color: 'var(--fg-subtle)' }}/>
        <input placeholder="Поиск по имени…" value=${query} onInput=${function (e) { setQuery(e.target.value); }}/></div>
      <${Guard} loading=${st.loading || contracts.loading} err=${st.err || contracts.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="clients" title="Клиентов нет"/></div>`
          : html`<div class="table-card clients-table-card"><table class="data-table clients-data-table"><thead><tr>
              <th>Клиент</th><th>Контакты</th><th>Договоры</th><th>Остаток</th><th></th>
            </tr></thead><tbody>${list.map(function (c) {
              var rel = (contracts.data || []).filter(function (x) { return x.client_id === c.id; });
              var activeCount = rel.filter(function (x) { return x.status === 'active'; }).length;
              var outstanding = rel.reduce(function (sum, x) { return sum + parseFloat(x.outstanding || '0'); }, 0);
              return html`<tr key=${c.id} class="data-row clients-data-row row-link" onClick=${function () { ctx.go('client', c.id); }}>
                <td>
                  <div class="table-primary">
                    <span class="table-avatar">${initials(c.full_name)}</span>
                    <div>
                      <div class="table-title">${c.full_name}</div>
                      <div class="table-subline">Добавлен ${fmt.date(c.created_at)}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="table-stack">
                    <div class="table-kv"><span>Телефон</span><strong class="amana-num">${c.phone || 'Не указан'}</strong></div>
                    <div class="table-kv"><span>Документ</span><strong>${c.document || 'Не указан'}</strong></div>
                  </div>
                </td>
                <td>
                  <div class="table-metric-pack">
                    <span class="compact-chip compact-chip-strong amana-num">${rel.length}</span>
                    <span class="compact-chip">${activeCount} активных</span>
                  </div>
                </td>
                <td><strong class="table-money amana-num">${fmt.money(outstanding)}</strong></td>
                <td class="table-arrow"><${Icon} name="arrow" size=${16}/></td>
              </tr>`;
            })}</tbody></table></div>`}
      <//>
      ${open ? html`<${ClientModal} onClose=${function () { setOpen(false); }} onSaved=${function () { setOpen(false); st.reload(); ctx.toast('Клиент добавлен'); }} ctx=${ctx}/>` : null}
    </div>`;
  }
  function ClientModal(p) {
    var f = useState({ full_name: '', phone: '', document: '' }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function save() {
      if (!v.full_name.trim()) { p.ctx.toast('Укажите ФИО', true); return; }
      setBusy(true);
      api.createClient({ full_name: v.full_name.trim(), phone: v.phone.trim(), document: v.document.trim() })
        .then(p.onSaved).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var inp = function (k, ph) { return html`<input class="input" value=${v[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }}/>`; };
    return html`<${ui.Modal} title="Новый клиент" onClose=${p.onClose}>
      <${ui.Field} label="ФИО">${inp('full_name', 'Магомед Алиев')}<//>
      <${ui.Field} label="Телефон">${inp('phone', '+7 928 000-00-00')}<//>
      <${ui.Field} label="Документ" hint="Нужен для договора">${inp('document', 'Паспорт 96 00 123456')}<//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : 'Сохранить'}</button>
    <//>`;
  }

  /* ================= CATALOG ================= */
  function Catalog(ctx) {
    var st = useAsync(api.listProducts);
    var contracts = useAsync(api.listContracts);
    var m = useState(false), open = m[0], setOpen = m[1];
    var list = st.data || [];
    return html`<div>
      <${PageHead} title="Каталог" sub=${list.length + ' товаров'}
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новый товар</button>`}/>
      <${Guard} loading=${st.loading || contracts.loading} err=${st.err || contracts.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="catalog" title="Каталог пуст"/></div>`
          : html`<div class="table-card"><table class="data-table"><thead><tr>
              <th>Товар</th><th>Категория и статус</th><th>Закупка</th><th>На складе</th><th>Договоры</th><th>Выдано</th><th></th>
            </tr></thead><tbody>${list.map(function (pr) {
              var rel = (contracts.data || []).filter(function (x) { return x.product_id === pr.id; });
              var activeCount = rel.filter(function (x) { return x.status === 'active'; }).length;
              var financed = rel.reduce(function (sum, x) { return sum + parseFloat(x.financed_amount || '0'); }, 0);
              return html`<tr key=${pr.id} class="data-row row-link" onClick=${function () { ctx.go('product', pr.id); }}>
                <td>
                  <div class="table-primary">
                    <span class="table-avatar table-avatar-icon"><${Icon} name="package" size=${16}/></span>
                    <div>
                      <div class="table-title">${pr.name}</div>
                      <div class="table-subline">${pr.can_be_financed ? 'Доступен для рассрочки' : (pr.halal_status === 'haram' ? 'Недоступен (харам)' : 'Нет в наличии')}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="table-stack">
                    <div class="table-kv"><span>Категория</span><strong>${pr.category || 'Не указана'}</strong></div>
                    <div class="compact-chip-group">
                      <${ui.StatusChip} map="halal" value=${pr.halal_status}/>
                      <span class=${'compact-chip ' + (pr.can_be_financed ? 'compact-chip-ok' : 'compact-chip-muted')}>${pr.can_be_financed ? 'Можно' : 'Нельзя'}</span>
                    </div>
                  </div>
                </td>
                <td><strong class="table-money amana-num">${fmt.money(pr.cost_price)}</strong></td>
                <td>
                  <span class=${'compact-chip amana-num ' + (pr.in_stock ? 'compact-chip-strong' : 'compact-chip-muted')}>${pr.stock} шт</span>
                </td>
                <td>
                  <div class="table-metric-pack">
                    <span class="compact-chip compact-chip-strong amana-num">${rel.length}</span>
                    <span class="compact-chip">${activeCount} активных</span>
                  </div>
                </td>
                <td><strong class="table-money amana-num">${fmt.money(financed)}</strong></td>
                <td class="table-arrow"><${Icon} name="arrow" size=${16}/></td>
              </tr>`;
            })}</tbody></table></div>`}
      <//>
      ${open ? html`<${ProductModal} onClose=${function () { setOpen(false); }} onSaved=${function () { setOpen(false); st.reload(); ctx.toast('Товар добавлен'); }} ctx=${ctx}/>` : null}
    </div>`;
  }
  function ProductModal(p) {
    var f = useState({ name: '', category: '', cost_price: '', halal_status: 'halal', stock: '0' }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function save() {
      if (!v.name.trim() || !v.cost_price.trim()) { p.ctx.toast('Заполните название и цену', true); return; }
      var stock = parseInt(String(v.stock).replace(/\D/g, ''), 10) || 0;
      setBusy(true);
      api.createProduct({ name: v.name.trim(), category: v.category.trim(), cost_price: v.cost_price.replace(',', '.').trim(), halal_status: v.halal_status, stock: stock })
        .then(p.onSaved).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var inp = function (k, ph) { return html`<input class="input" value=${v[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }}/>`; };
    return html`<${ui.Modal} title="Новый товар" onClose=${p.onClose}>
      <${ui.Field} label="Название">${inp('name', 'Диван угловой')}<//>
      <${ui.Field} label="Категория">${inp('category', 'Мебель')}<//>
      <div class="grid" style=${{ gridTemplateColumns: 'repeat(2,1fr)' }}>
        <${ui.Field} label="Закупочная цена, ₽">${inp('cost_price', '85000')}<//>
        <${ui.Field} label="Остаток на складе, шт">${inp('stock', '0')}<//>
      </div>
      <${ui.Field} label="Халяль-статус">
        <select class="select" value=${v.halal_status} onChange=${function (e) { set(Object.assign({}, v, { halal_status: e.target.value })); }}>
          <option value="halal">Халяль</option><option value="doubtful">Сомнительно</option><option value="haram">Харам</option></select>
      <//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : 'Сохранить'}</button>
    <//>`;
  }

  /* StockModal: receipt (+), writeoff (−), adjustment (±) — logs a movement. */
  function StockModal(p) {
    var f = useState({ reason: 'receipt', qty: '', dir: '+', note: '' }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function upd(o) { set(Object.assign({}, v, o)); }
    function save() {
      var qty = parseInt(String(v.qty).replace(/\D/g, ''), 10);
      if (!qty || qty <= 0) { p.ctx.toast('Укажите количество', true); return; }
      var delta = v.reason === 'receipt' ? qty : v.reason === 'writeoff' ? -qty : (v.dir === '-' ? -qty : qty);
      setBusy(true);
      api.adjustStock(p.product.id, { delta: delta, reason: v.reason, note: v.note.trim() })
        .then(p.onSaved).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var inp = function (k, ph) { return html`<input class="input" value=${v[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; upd(o); }}/>`; };
    return html`<${ui.Modal} title="Движение по складу" onClose=${p.onClose}>
      <div style=${{ fontSize: 13, color: 'var(--fg-muted)', marginBottom: 14 }}>${p.product.name} — сейчас на складе <b class="amana-num">${p.product.stock} шт</b></div>
      <${ui.Field} label="Операция">
        <select class="select" value=${v.reason} onChange=${function (e) { upd({ reason: e.target.value }); }}>
          <option value="receipt">Поступление (+)</option>
          <option value="writeoff">Списание (−)</option>
          <option value="adjustment">Корректировка (±)</option>
        </select>
      <//>
      <div class="grid" style=${{ gridTemplateColumns: v.reason === 'adjustment' ? '120px 1fr' : '1fr' }}>
        ${v.reason === 'adjustment' ? html`<${ui.Field} label="Направление">
          <select class="select" value=${v.dir} onChange=${function (e) { upd({ dir: e.target.value }); }}>
            <option value="+">Добавить</option><option value="-">Убавить</option></select>
        <//>` : null}
        <${ui.Field} label="Количество, шт">${inp('qty', '5')}<//>
      </div>
      <${ui.Field} label="Комментарий (необязательно)">${inp('note', 'Поступление от поставщика')}<//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : 'Применить'}</button>
    <//>`;
  }

  /* ================= CONTRACTS LIST ================= */
  function Contracts(ctx) {
    var st = useAsync(api.listContracts);
    var clients = useAsync(api.listClients);
    var products = useAsync(api.listProducts);
    var fl = useState('all'), filter = fl[0], setFilter = fl[1];
    var list = (st.data || []).filter(function (c) { return filter === 'all' || c.status === filter; });
    var tabs = [['all', 'Все'], ['active', 'Активные'], ['completed', 'Завершённые'], ['cancelled', 'Отменённые']];
    return html`<div>
      <${PageHead} title="Договоры" sub="Рассрочка по модели мурабаха"
        actions=${html`<button class="btn btn-primary" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> Новый договор</button>`}/>
      <div class="tabs" style=${{ marginBottom: 16, maxWidth: 460 }}>
        ${tabs.map(function (t) { return html`<button key=${t[0]} class=${'tab ' + (filter === t[0] ? 'active' : '')} onClick=${function () { setFilter(t[0]); }}>${t[1]}</button>`; })}
      </div>
      <${Guard} loading=${st.loading || clients.loading || products.loading} err=${st.err || clients.err || products.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="contracts" title="Договоров нет" text="Оформите первый договор рассрочки"
            action=${html`<button class="btn btn-primary" style=${{ marginTop: 14 }} onClick=${function () { ctx.go('contract-new'); }}>Новый договор</button>`}/></div>`
          : html`<div class="entity-list">${list.map(function (c) {
              var client = findByID(clients.data, c.client_id);
              var product = findByID(products.data, c.product_id);
              var progress = Math.round(parseFloat(c.progress_percent || '0'));
              return html`<button key=${c.id} class="entity-card entity-card-contract row-link" onClick=${function () { ctx.go('contract', c.id); }}>
                <div class="entity-card-main">
                  <div class="entity-contract-side">
                    <div class="entity-contract-id">${c.id.slice(0, 8)}</div>
                    <div class="entity-subline">Создан ${fmt.date(c.created_at)}</div>
                  </div>
                  <div class="entity-copy">
                    <div class="entity-headline">
                      <div>
                        <div class="entity-title">${client ? client.full_name : 'Клиент не найден'}</div>
                        <div class="entity-subline">${product ? product.name : 'Товар не найден'}</div>
                      </div>
                      <div class="entity-badges">
                        <${ui.StatusChip} map="contractStatus" value=${c.status}/>
                      </div>
                    </div>
                    <div class="entity-contract-grid">
                      <div class="entity-stat">
                        <span class="entity-stat-label">Цена продажи</span>
                        <strong class="amana-num">${fmt.money(c.sale_price)}</strong>
                      </div>
                      <div class="entity-stat">
                        <span class="entity-stat-label">К выдаче</span>
                        <strong class="amana-num">${fmt.money(c.financed_amount)}</strong>
                      </div>
                      <div class="entity-stat entity-stat-strong">
                        <span class="entity-stat-label">Остаток</span>
                        <strong class="amana-num">${fmt.money(c.outstanding)}</strong>
                      </div>
                    </div>
                    <div class="entity-progress-row">
                      <div class="progress entity-progress"><i style=${{ width: progress + '%' }}></i></div>
                      <span class="entity-progress-label amana-num">${progress}% оплачено</span>
                      <span class="entity-open-link">Открыть договор <${Icon} name="arrow" size=${15}/></span>
                    </div>
                  </div>
                </div>
              </button>`;
            })}</div>`}
      <//>
    </div>`;
  }

  function ClientCard(ctx) {
    var st = useAsync(function () { return api.getClient(ctx.route.id); }, [ctx.route.id]);
    var contracts = useAsync(api.listContracts);
    return html`<div>
      <div style=${{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
        <button class="icon-btn" onClick=${function () { ctx.go('clients'); }}><${Icon} name="back" size=${20}/></button>
        <div class="page-title" style=${{ fontSize: 22 }}>Клиент</div>
      </div>
      <${Guard} loading=${st.loading || contracts.loading} err=${st.err || contracts.err}>
        ${(function () {
          var c = st.data;
          if (!c) return null;
          var rel = (contracts.data || []).filter(function (x) { return x.client_id === c.id; });
          var active = rel.filter(function (x) { return x.status === 'active'; }).length;
          var outstanding = rel.reduce(function (sum, x) { return sum + parseFloat(x.outstanding || '0'); }, 0);
          return html`<div class="grid" style=${{ gap: 16 }}>
            <div class="card card-pad">
              <div class="table-primary">
                <span class="table-avatar">${initials(c.full_name)}</span>
                <div>
                  <div class="page-title" style=${{ fontSize: 24 }}>${c.full_name}</div>
                  <div class="page-sub">Клиент добавлен ${fmt.date(c.created_at)}</div>
                </div>
              </div>
              <div class="compact-fields compact-fields-2" style=${{ marginTop: 18 }}>
                <div class="compact-field"><span>Телефон</span><strong class="amana-num">${c.phone || 'Не указан'}</strong></div>
                <div class="compact-field"><span>Документ</span><strong>${c.document || 'Не указан'}</strong></div>
                <div class="compact-field"><span>Договоров</span><strong class="amana-num">${rel.length}</strong></div>
                <div class="compact-field"><span>Активных</span><strong class="amana-num">${active}</strong></div>
              </div>
              <div class="banner banner-accent" style=${{ marginTop: 14 }}>
                <${Icon} name="coins" size=${17}/> Остаток по всем договорам: <b class="amana-num">${fmt.money(outstanding)}</b>
              </div>
            </div>
            <div class="table-card">
              <div class="table-card-head">Договоры клиента</div>
              ${rel.length === 0 ? html`<div class="card"><${ui.Empty} icon="contracts" title="Договоров пока нет"/></div>`
                : html`<table class="data-table"><thead><tr><th>Договор</th><th>Статус</th><th>Цена</th><th>Остаток</th></tr></thead>
                    <tbody>${rel.map(function (x) {
                      return html`<tr key=${x.id} class="data-row row-link" onClick=${function () { ctx.go('contract', x.id); }}>
                        <td><strong class="table-code">${x.id.slice(0, 8)}</strong></td>
                        <td><${ui.StatusChip} map="contractStatus" value=${x.status}/></td>
                        <td class="table-money amana-num">${fmt.money(x.sale_price)}</td>
                        <td class="table-money amana-num">${fmt.money(x.outstanding)}</td>
                      </tr>`;
                    })}</tbody></table>`}
            </div>
          </div>`;
        })()}
      <//>
    </div>`;
  }

  function ProductCard(ctx) {
    var st = useAsync(function () { return api.getProduct(ctx.route.id); }, [ctx.route.id]);
    var contracts = useAsync(api.listContracts);
    var moves = useAsync(api.listStockMovements);
    var m = useState(false), open = m[0], setOpen = m[1];
    return html`<div>
      <div style=${{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
        <button class="icon-btn" onClick=${function () { ctx.go('catalog'); }}><${Icon} name="back" size=${20}/></button>
        <div class="page-title" style=${{ fontSize: 22 }}>Товар</div>
      </div>
      <${Guard} loading=${st.loading || contracts.loading} err=${st.err || contracts.err}>
        ${(function () {
          var pr = st.data;
          if (!pr) return null;
          var rel = (contracts.data || []).filter(function (x) { return x.product_id === pr.id; });
          var active = rel.filter(function (x) { return x.status === 'active'; }).length;
          var financed = rel.reduce(function (sum, x) { return sum + parseFloat(x.financed_amount || '0'); }, 0);
          var prMoves = (moves.data || []).filter(function (x) { return x.product_id === pr.id; });
          return html`<div class="grid" style=${{ gap: 16 }}>
            <div class="card card-pad">
              <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}>
                <div class="table-primary">
                  <span class="table-avatar table-avatar-icon"><${Icon} name="package" size=${18}/></span>
                  <div>
                    <div class="page-title" style=${{ fontSize: 24 }}>${pr.name}</div>
                    <div class="page-sub">${pr.category || 'Категория не указана'}</div>
                  </div>
                </div>
                <button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="truck" size=${16}/> Движение по складу</button>
              </div>
              <div class="compact-chip-group" style=${{ marginTop: 14 }}>
                <${ui.StatusChip} map="halal" value=${pr.halal_status}/>
                <span class=${'compact-chip ' + (pr.in_stock ? 'compact-chip-ok' : 'compact-chip-muted')}>${pr.in_stock ? 'В наличии: ' + pr.stock + ' шт' : 'Нет в наличии'}</span>
                <span class=${'compact-chip ' + (pr.can_be_financed ? 'compact-chip-ok' : 'compact-chip-muted')}>${pr.can_be_financed ? 'Можно оформить в рассрочку' : 'Нельзя оформить в рассрочку'}</span>
              </div>
              ${!pr.in_stock && pr.halal_status !== 'haram' ? html`<div class="banner banner-warn" style=${{ marginTop: 14 }}><${Icon} name="info" size=${17}/> Товара нет на складе — оформить рассрочку нельзя, пока не пополните остаток.</div>` : null}
              <div class="compact-fields compact-fields-2" style=${{ marginTop: 18 }}>
                <div class="compact-field"><span>Закупочная цена</span><strong class="amana-num">${fmt.money(pr.cost_price)}</strong></div>
                <div class="compact-field"><span>На складе</span><strong class="amana-num">${pr.stock} шт</strong></div>
                <div class="compact-field"><span>Всего договоров</span><strong class="amana-num">${rel.length}</strong></div>
                <div class="compact-field"><span>Активных договоров</span><strong class="amana-num">${active}</strong></div>
                <div class="compact-field"><span>Выдано в рассрочку</span><strong class="amana-num">${fmt.money(financed)}</strong></div>
              </div>
            </div>
            <div class="table-card">
              <div class="table-card-head">Договоры по товару</div>
              ${rel.length === 0 ? html`<div class="card"><${ui.Empty} icon="contracts" title="Договоров пока нет"/></div>`
                : html`<table class="data-table"><thead><tr><th>Договор</th><th>Статус</th><th>Цена</th><th>Остаток</th></tr></thead>
                    <tbody>${rel.map(function (x) {
                      return html`<tr key=${x.id} class="data-row row-link" onClick=${function () { ctx.go('contract', x.id); }}>
                        <td><strong class="table-code">${x.id.slice(0, 8)}</strong></td>
                        <td><${ui.StatusChip} map="contractStatus" value=${x.status}/></td>
                        <td class="table-money amana-num">${fmt.money(x.sale_price)}</td>
                        <td class="table-money amana-num">${fmt.money(x.outstanding)}</td>
                      </tr>`;
                    })}</tbody></table>`}
            </div>
            <div class="table-card">
              <div class="table-card-head">Движение по складу (товарооборот)</div>
              ${prMoves.length === 0 ? html`<div class="card"><${ui.Empty} icon="truck" title="Движений пока нет" text="Примите товар на склад, чтобы появилась история"/></div>`
                : html`<table class="data-table"><thead><tr><th>Дата</th><th>Операция</th><th>Изменение</th><th>Остаток</th><th>Комментарий</th></tr></thead>
                    <tbody>${prMoves.map(function (x) {
                      return html`<tr key=${x.id} class="data-row">
                        <td>${fmt.dateTime(x.created_at)}</td>
                        <td><${ui.StatusChip} map="stockReason" value=${x.reason}/></td>
                        <td><strong class=${'amana-num ' + (x.delta >= 0 ? 'delta-pos' : 'delta-neg')}>${x.delta > 0 ? '+' : ''}${x.delta} шт</strong></td>
                        <td class="amana-num">${x.balance_after} шт</td>
                        <td style=${{ color: 'var(--fg-muted)' }}>${x.note || '—'}</td>
                      </tr>`;
                    })}</tbody></table>`}
            </div>
          </div>`;
        })()}
      <//>
      ${open && st.data ? html`<${StockModal} product=${st.data} ctx=${ctx} onClose=${function () { setOpen(false); }}
        onSaved=${function () { setOpen(false); st.reload(); moves.reload(); ctx.toast('Склад обновлён'); }}/>` : null}
    </div>`;
  }

  function initials(name) {
    return String(name || '').split(/\s+/).filter(Boolean).slice(0, 2).map(function (x) { return x[0]; }).join('').toUpperCase() || 'К';
  }
  function findByID(list, id) {
    return (list || []).filter(function (x) { return x.id === id; })[0] || null;
  }
  function normalizeReminder(r) {
    return Object.assign({}, r, { date: new Date(r.scheduled_at) });
  }
  function reminderStatusText(r) {
    return ui.labels.reminderStatus[(r && r.status) || 'scheduled'] || (r && r.status) || 'Запланирована';
  }
  function canQuickActReminder(r) {
    return !!r && r.base_status === 'scheduled';
  }
  function stopEvent(e) {
    e.preventDefault();
    e.stopPropagation();
  }
  function ReminderQuickActions(p) {
    var reminder = p.reminder;
    if (!canQuickActReminder(reminder)) return null;
    var busy = p.busy || '';
    return html`<div class=${'reminder-quick-actions ' + (p.variant || '')}>
      <button class="reminder-quick-btn done" disabled=${busy === 'complete'} title="Выполнить задачу"
        onClick=${function (e) { stopEvent(e); p.onComplete && p.onComplete(reminder); }}>
        <${Icon} name="check" size=${p.iconSize || 12}/>
      </button>
      <button class="reminder-quick-btn cancel" disabled=${busy === 'cancel'} title="Отменить задачу"
        onClick=${function (e) { stopEvent(e); p.onCancel && p.onCancel(reminder); }}>
        <${Icon} name="x" size=${p.iconSize || 12}/>
      </button>
    </div>`;
  }
  function startOfDay(d) {
    var x = new Date(d); x.setHours(0, 0, 0, 0); return x;
  }
  function sameDay(a, b) {
    return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
  }
  function addDays(d, n) {
    var x = new Date(d); x.setDate(x.getDate() + n); return startOfDay(x);
  }
  function addMonths(d, n) {
    var x = new Date(d); x.setMonth(x.getMonth() + n, 1); return startOfDay(x);
  }
  function startOfWeek(d) {
    var x = startOfDay(d), day = x.getDay(), diff = day === 0 ? -6 : 1 - day;
    return addDays(x, diff);
  }
  function isWithinWeek(d, ref) {
    var start = startOfWeek(ref), end = addDays(start, 7);
    return d >= start && d < end;
  }
  function buildWeekDays(ref) {
    var start = startOfWeek(ref), days = [];
    for (var i = 0; i < 7; i++) days.push(addDays(start, i));
    return days;
  }
  function buildMonthGrid(ref) {
    var monthStart = new Date(ref.getFullYear(), ref.getMonth(), 1);
    var gridStart = startOfWeek(monthStart);
    var days = [];
    for (var i = 0; i < 42; i++) {
      var day = addDays(gridStart, i);
      days.push({ key: day.toISOString(), date: day, isCurrentMonth: day.getMonth() === ref.getMonth() });
    }
    return days;
  }
  function eventsForDay(list, day) {
    return list.filter(function (r) { return sameDay(r.date, day); }).slice(0, 6);
  }
  function shortReminderLabel(type) {
    if (type === 'delivery') return 'Доставка';
    if (type === 'payment_followup') return 'Платеж';
    return 'Звонок';
  }
  function reminderTone(type) {
    return type === 'delivery' ? 'delivery' : type === 'payment_followup' ? 'payment_followup' : 'call';
  }

  /* ================= CONTRACT WIZARD ================= */
  function ContractWizard(ctx) {
    var clients = useAsync(api.listClients);
    var products = useAsync(api.listProducts);
    var s = useState({ step: 1, clientId: '', productId: '', markupMode: 'sum', markupSum: '15000', markupPct: '15',
      down: '0', installments: 6, cadence: 'monthly', start: defaultStart() }), w = s[0], setW = s[1];
    var pv = useState(null), preview = pv[0], setPreview = pv[1];
    var pvErr = useState(''), previewErr = pvErr[0], setPreviewErr = pvErr[1];
    var bz = useState(false), busy = bz[0], setBusy = bz[1];
    var set = function (o) { setW(Object.assign({}, w, o)); };
    var product = (products.data || []).filter(function (p) { return p.id === w.productId; })[0];

    function termsBody() {
      var b = { cost_price: product ? product.cost_price : '0', down_payment: w.down || '0',
        installments: Number(w.installments), cadence: w.cadence, start_date: w.start };
      if (w.markupMode === 'pct') b.markup_percent = w.markupPct; else b.markup_amount = w.markupSum || '0';
      return b;
    }
    useEffect(function () {
      if (w.step !== 4 || !product) return;
      var alive = true; setPreview(null); setPreviewErr('');
      var t = setTimeout(function () {
        api.previewContract(termsBody()).then(function (r) { if (alive) setPreview(r); })
          .catch(function (e) { if (alive) setPreviewErr(e.message); });
      }, 250);
      return function () { alive = false; clearTimeout(t); };
    }, [w.step, w.productId, w.markupMode, w.markupSum, w.markupPct, w.down, w.installments, w.cadence, w.start]);

    function create() {
      setBusy(true);
      var body = Object.assign({ client_id: w.clientId, product_id: w.productId }, termsBody());
      api.createContract(body).then(function (c) { ctx.toast('Договор оформлен'); ctx.go('contract', c.id); })
        .catch(function (e) { setBusy(false); ctx.toast(e.message, true); });
    }
    var canNext = w.step === 1 ? !!w.clientId : w.step === 2 ? !!(product && product.can_be_financed) : true;
    var stepNames = ['Клиент', 'Товар', 'Условия', 'Предпросмотр', 'Подтверждение'];

    return html`<div>
      <div style=${{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
        <button class="icon-btn" onClick=${function () { ctx.go('contracts'); }}><${Icon} name="back" size=${20}/></button>
        <div class="page-title" style=${{ fontSize: 22 }}>Новый договор</div>
      </div>
      <div class="stepper">
        ${stepNames.map(function (nm, i) {
          var n = i + 1, active = n === w.step, done = n < w.step;
          return html`<${React.Fragment} key=${n}>
            ${i > 0 ? html`<div class="step-line"></div>` : null}
            <div class=${'step-dot ' + (active ? 'active' : done ? 'done' : '')}>
              <span class="b">${done ? html`<${Icon} name="check" size=${15}/>` : n}</span><span class="lbl">${nm}</span></div>
          <//>`;
        })}
      </div>
      <div class="card card-pad">
        ${w.step === 1 ? html`<${WizClient} clients=${clients} w=${w} set=${set}/>`
          : w.step === 2 ? html`<${WizProduct} products=${products} w=${w} set=${set}/>`
          : w.step === 3 ? html`<${WizTerms} w=${w} set=${set} product=${product}/>`
          : w.step === 4 ? html`<${WizPreview} preview=${preview} err=${previewErr}/>`
          : html`<${WizConfirm} w=${w} product=${product} preview=${preview} clients=${clients.data || []}/>`}
      </div>
      <div style=${{ display: 'flex', justifyContent: 'space-between', marginTop: 16 }}>
        <button class="btn btn-ghost" disabled=${w.step === 1} onClick=${function () { set({ step: w.step - 1 }); }}>Назад</button>
        ${w.step < 5
          ? html`<button class="btn btn-primary" disabled=${!canNext} onClick=${function () { set({ step: w.step + 1 }); }}>Далее <${Icon} name="arrow" size=${17}/></button>`
          : html`<button class="btn btn-primary" disabled=${busy || !preview} onClick=${create}>${busy ? html`<${ui.Spinner}/>` : 'Создать договор'}</button>`}
      </div>
    </div>`;
  }
  function WizClient(p) {
    return html`<${Guard} loading=${p.clients.loading} err=${p.clients.err}>
      <div style=${{ fontWeight: 700, marginBottom: 12 }}>Выберите клиента</div>
      <div class="grid" style=${{ gridTemplateColumns: 'repeat(2,1fr)' }}>
        ${(p.clients.data || []).map(function (c) {
          return html`<div key=${c.id} class=${'select-card ' + (p.w.clientId === c.id ? 'sel' : '')} onClick=${function () { p.set({ clientId: c.id }); }}>
            <div style=${{ fontWeight: 600 }}>${c.full_name}</div>
            <div style=${{ fontSize: 12.5, color: 'var(--fg-subtle)' }}>${c.phone || c.document || '—'}</div></div>`;
        })}
      </div>
      ${(p.clients.data || []).length === 0 ? html`<div class="banner banner-info">Сначала добавьте клиента в разделе «Клиенты».</div>` : null}
    <//>`;
  }
  function WizProduct(p) {
    return html`<${Guard} loading=${p.products.loading} err=${p.products.err}>
      <div style=${{ fontWeight: 700, marginBottom: 12 }}>Выберите товар</div>
      <div class="grid" style=${{ gridTemplateColumns: 'repeat(2,1fr)' }}>
        ${(p.products.data || []).map(function (pr) {
          var haram = pr.halal_status === 'haram';
          var blocked = !pr.can_be_financed; // харам или нет на складе
          var reason = haram ? 'Договор на «харам» оформить нельзя' : (!pr.in_stock ? 'Нет на складе — пополните остаток' : '');
          return html`<div key=${pr.id} class=${'select-card ' + (p.w.productId === pr.id ? 'sel ' : '') + (blocked ? 'disabled' : '')}
            onClick=${function () { if (!blocked) p.set({ productId: pr.id }); }}>
            <div style=${{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
              <div style=${{ fontWeight: 600 }}>${pr.name}</div><${ui.StatusChip} map="halal" value=${pr.halal_status}/></div>
            <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 4 }}>
              <div class="amana-num" style=${{ fontSize: 13, color: 'var(--fg-muted)' }}>${fmt.money(pr.cost_price)}</div>
              <span class=${'compact-chip amana-num ' + (pr.in_stock ? 'compact-chip-ok' : 'compact-chip-muted')}>${pr.stock} шт</span>
            </div>
            ${reason ? html`<div style=${{ fontSize: 12, color: 'var(--haram-fg)', marginTop: 4 }}>${reason}</div>` : null}</div>`;
        })}
      </div>
      ${(p.products.data || []).length > 0 && (p.products.data || []).every(function (pr) { return !pr.can_be_financed; })
        ? html`<div class="banner banner-warn" style=${{ marginTop: 12 }}>Нет товаров, доступных для рассрочки. Пополните склад в разделе «Каталог».</div>` : null}
    <//>`;
  }
  function WizTerms(p) {
    var w = p.w, set = p.set;
    var fld = function (k, ph) { return html`<input class="input" value=${w[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; set(o); }}/>`; };
    return html`<div>
      <div style=${{ fontWeight: 700, marginBottom: 4 }}>Условия рассрочки</div>
      <div style=${{ fontSize: 13, color: 'var(--fg-muted)', marginBottom: 16 }}>Закупочная цена: <b class="amana-num">${p.product ? fmt.money(p.product.cost_price) : '—'}</b></div>
      <div class="grid" style=${{ gridTemplateColumns: 'repeat(2,1fr)' }}>
        <${ui.Field} label="Наценка">
          <div style=${{ display: 'flex', gap: 8 }}>
            <select class="select" style=${{ width: 110 }} value=${w.markupMode} onChange=${function (e) { set({ markupMode: e.target.value }); }}>
              <option value="sum">Сумма ₽</option><option value="pct">Процент %</option></select>
            ${w.markupMode === 'sum' ? fld('markupSum', '15000') : fld('markupPct', '15')}
          </div>
        <//>
        <${ui.Field} label="Первоначальный взнос, ₽">${fld('down', '0')}<//>
        <${ui.Field} label="Число платежей">${fld('installments', '6')}<//>
        <${ui.Field} label="Периодичность">
          <select class="select" value=${w.cadence} onChange=${function (e) { set({ cadence: e.target.value }); }}>
            <option value="monthly">Ежемесячно</option><option value="weekly">Еженедельно</option></select>
        <//>
        <${ui.Field} label="Дата первого платежа">
          <input class="input" type="date" value=${w.start} onInput=${function (e) { set({ start: e.target.value }); }}/>
        <//>
      </div>
    </div>`;
  }
  function WizPreview(p) {
    if (p.err) return html`<div class="banner banner-warn">${p.err}</div>`;
    if (!p.preview) return html`<${ui.Loading} label="Считаем график…"/>`;
    var pv = p.preview, c = pv.comparison;
    return html`<div>
      <div class="grid" style=${{ gridTemplateColumns: 'repeat(3,1fr)', marginBottom: 18 }}>
        ${[['Цена продажи', pv.sale_price], ['К рассрочке', pv.financed_amount], ['Платежей', String(pv.schedule.length)]].map(function (r, i) {
          return html`<div key=${i} class="card card-pad"><div class="kpi"><div class="v amana-num" style=${{ fontSize: 22 }}>${i < 2 ? fmt.money(r[1]) : r[1]}</div><div class="l">${r[0]}</div></div></div>`;
        })}
      </div>
      <div class="banner banner-accent" style=${{ marginBottom: 16 }}><${Icon} name="shield" size=${18}/>
        <div><b>Цена зафиксирована.</b> Долг не растёт со временем — 0% риба. ${c ? html`<span> С обычным кредитом (${c.annual_rate_percent}%) переплата составила бы <b class="amana-num">${fmt.money(c.overpayment)}</b>.</span>` : null}</div></div>
      <div style=${{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        ${pv.schedule.map(function (it) {
          return html`<div key=${it.number} class="timeline-row">
            <span class="tl-num">${it.number}</span>
            <div style=${{ flex: 1 }}><div style=${{ fontWeight: 600 }}>${fmt.dateLong(it.due_date)}</div></div>
            <div class="amana-num" style=${{ fontWeight: 600 }}>${fmt.money(it.amount)}</div></div>`;
        })}
      </div>
    </div>`;
  }
  function WizConfirm(p) {
    var client = (p.clients || []).filter(function (c) { return c.id === p.w.clientId; })[0];
    var rows = [['Клиент', client ? client.full_name : '—'], ['Товар', p.product ? p.product.name : '—'],
      ['Цена продажи', p.preview ? fmt.money(p.preview.sale_price) : '—'], ['К рассрочке', p.preview ? fmt.money(p.preview.financed_amount) : '—'],
      ['Первый взнос', fmt.money((p.w.down || '0'))], ['Платежей', String(p.w.installments) + ' × ' + (p.w.cadence === 'weekly' ? 'нед.' : 'мес.')]];
    return html`<div>
      <div style=${{ fontWeight: 700, marginBottom: 14 }}>Подтверждение</div>
      <div class="card" style=${{ overflow: 'hidden' }}><table class="table"><tbody>
        ${rows.map(function (r, i) { return html`<tr key=${i}><td style=${{ color: 'var(--fg-muted)' }}>${r[0]}</td><td style=${{ textAlign: 'right', fontWeight: 600 }} class="amana-num">${r[1]}</td></tr>`; })}
      </tbody></table></div>
      <div class="banner banner-info" style=${{ marginTop: 14 }}><${Icon} name="check" size=${17}/> Договор будет создан сразу в статусе «Активен».</div>
    </div>`;
  }

  /* ================= CONTRACT CARD ================= */
  function ContractCard(ctx) {
    var st = useAsync(function () { return api.getContract(ctx.route.id); }, [ctx.route.id]);
    var tabS = useState('schedule'), tab = tabS[0], setTab = tabS[1];
    var payS = useState(null), payOpen = payS[0], setPayOpen = payS[1];
    var c = st.data;
    function reloadToast(msg) { return function () { st.reload(); ctx.toast(msg); }; }
    function doSettle() {
      ctx.confirm({ title: 'Досрочное погашение', text: 'Остаток будет погашен полностью, без штрафа. Договор завершится.', okLabel: 'Погасить',
        onOk: function () { api.settleContract(c.id).then(reloadToast('Договор завершён')).catch(function (e) { ctx.toast(e.message, true); }); } });
    }
    function doCancel() {
      ctx.confirm({ title: 'Отменить договор', text: 'Договор будет переведён в статус «Отменён». Действие необратимо.', okLabel: 'Отменить договор', danger: true,
        onOk: function () { api.cancelContract(c.id).then(reloadToast('Договор отменён')).catch(function (e) { ctx.toast(e.message, true); }); } });
    }
    return html`<div>
      <div style=${{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
        <button class="icon-btn" onClick=${function () { ctx.go('contracts'); }}><${Icon} name="back" size=${20}/></button>
        <div class="page-title" style=${{ fontSize: 22 }}>Договор</div>
      </div>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${c ? html`<div>
          ${c.has_overdue ? html`<div class="banner banner-warn" style=${{ marginBottom: 16 }}>
            <${Icon} name="clock" size=${18}/><div><b>Есть просрочка.</b> Долг не растёт со временем — сумма обязательства зафиксирована при создании.</div>
          </div>` : null}
          <div class="card card-pad" style=${{ marginBottom: 16 }}>
            <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 14 }}>
              <div>
                <div style=${{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
                  <${ui.StatusChip} map="contractStatus" value=${c.status}/>
                  <span style=${{ fontSize: 12.5, color: 'var(--fg-subtle)', fontFamily: 'monospace' }}>${c.id.slice(0, 8)}</span></div>
                <div class="amana-num" style=${{ fontSize: 30, fontWeight: 700, letterSpacing: '-.02em' }}>${fmt.money(c.sale_price)}</div>
                <div style=${{ color: 'var(--fg-muted)', fontSize: 13.5 }}>остаток <b class="amana-num" style=${{ color: 'var(--fg)' }}>${fmt.money(c.outstanding)}</b> · оплачено ${fmt.money(c.paid_amount)}</div>
              </div>
              <div style=${{ display: 'flex', gap: 9, flexWrap: 'wrap' }}>
                <button class="btn btn-soft btn-sm" onClick=${function () { api.downloadContractPdf(c.id).catch(function (e) { ctx.toast(e.message, true); }); }}><${Icon} name="doc" size=${15}/> Скачать PDF</button>
                ${c.status === 'active' ? html`<button class="btn btn-primary btn-sm" onClick=${function () { setPayOpen({ amount: '' }); }}>Принять платёж</button>` : null}
                ${c.status === 'active' ? html`<button class="btn btn-ghost btn-sm" onClick=${doSettle}>Досрочно погасить</button>` : null}
                ${(c.status === 'active' && ctx.isOwner) ? html`<button class="btn btn-danger btn-sm" onClick=${doCancel}>Отменить</button>` : null}
              </div>
            </div>
            <div style=${{ marginTop: 16 }}><div class="progress" style=${{ height: 10 }}><i style=${{ width: (c.progress_percent || 0) + '%' }}></i></div>
              <div style=${{ fontSize: 12.5, color: 'var(--fg-subtle)', marginTop: 6 }}>Прогресс ${Math.round(c.progress_percent || 0)}%</div></div>
          </div>

          <div class="tabs" style=${{ maxWidth: 320, marginBottom: 14 }}>
            <button class=${'tab ' + (tab === 'schedule' ? 'active' : '')} onClick=${function () { setTab('schedule'); }}>График</button>
            <button class=${'tab ' + (tab === 'payments' ? 'active' : '')} onClick=${function () { setTab('payments'); }}>Платежи (${c.payments.length})</button>
          </div>
          <div class="card card-pad">
            ${tab === 'schedule'
              ? c.schedule.map(function (it) {
                return html`<div key=${it.number} class="timeline-row">
                  <span class="tl-num">${it.number}</span>
                  <div style=${{ flex: 1 }}><div style=${{ fontWeight: 600 }}>${fmt.dateLong(it.due_date)}</div></div>
                  <div class="amana-num" style=${{ fontWeight: 600, marginRight: 12 }}>${fmt.money(it.amount)}</div>
                  <${ui.StatusChip} map="installmentStatus" value=${it.status}/></div>`;
              })
              : c.payments.length === 0 ? html`<${ui.Empty} icon="coins" title="Платежей пока нет"/>`
                : c.payments.map(function (pm, i) {
                  return html`<div key=${i} class="timeline-row">
                    <span class="tl-num" style=${{ background: 'var(--halal-bg)', color: 'var(--halal-fg)' }}><${Icon} name="check" size=${15}/></span>
                    <div style=${{ flex: 1 }}><div style=${{ fontWeight: 600 }}>${fmt.dateTime(pm.paid_at)}</div></div>
                    <div class="amana-num" style=${{ fontWeight: 600, color: 'var(--halal-fg)' }}>+${fmt.money(pm.amount)}</div></div>`;
                })}
          </div>
        </div>` : null}
      <//>
      ${payOpen ? html`<${PaymentModal} c=${c} onClose=${function () { setPayOpen(null); }} onDone=${function () { setPayOpen(null); st.reload(); ctx.toast('Платёж принят'); }} ctx=${ctx}/>` : null}
    </div>`;
  }
  function PaymentModal(p) {
    var a = useState(''), amount = a[0], setAmount = a[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    var max = parseFloat(p.c.outstanding);
    function pay() {
      var n = parseFloat(String(amount).replace(',', '.'));
      if (!(n > 0)) { p.ctx.toast('Введите сумму', true); return; }
      if (n > max) { p.ctx.toast('Больше остатка нельзя', true); return; }
      setBusy(true);
      api.registerPayment(p.c.id, String(amount).replace(',', '.')).then(p.onDone).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    return html`<${ui.Modal} title="Принять платёж" onClose=${p.onClose}>
      <div style=${{ fontSize: 13.5, color: 'var(--fg-muted)' }}>Остаток к оплате: <b class="amana-num" style=${{ color: 'var(--fg)' }}>${fmt.money(p.c.outstanding)}</b></div>
      <${ui.Field} label="Сумма, ₽" hint="Любая сумма в пределах остатка">
        <input class="input" value=${amount} placeholder=${fmt.num(p.c.outstanding)} onInput=${function (e) { setAmount(e.target.value); }} onKeyDown=${function (e) { if (e.key === 'Enter') pay(); }}/>
      <//>
      <div style=${{ display: 'flex', gap: 8 }}>
        <button class="btn btn-soft btn-sm" onClick=${function () { setAmount(p.c.outstanding); }}>Весь остаток</button>
      </div>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${pay}>${busy ? html`<${ui.Spinner}/>` : 'Принять платёж'}</button>
    <//>`;
  }

  /* ================= SCHEDULE ================= */
  function Schedule(ctx) {
    var st = useAsync(api.listReminders);
    var vw = useState('week'), view = vw[0], setView = vw[1];
    var pv = useState('week'), previousView = pv[0], setPreviousView = pv[1];
    var ds = useState(startOfDay(new Date())), selectedDate = ds[0], setSelectedDate = ds[1];
    var m = useState(false), open = m[0], setOpen = m[1];
    var rb = useState({}), reminderBusy = rb[0], setReminderBusy = rb[1];
    var list = st.data || [];
    var normalized = list.map(normalizeReminder).sort(function (a, b) { return a.date - b.date; });
    var monthDays = buildMonthGrid(selectedDate);
    var weekDays = buildWeekDays(selectedDate);
    var dayEvents = normalized.filter(function (r) { return sameDay(r.date, selectedDate); });
    var weekEvents = normalized.filter(function (r) { return isWithinWeek(r.date, selectedDate); });
    var monthEvents = normalized.filter(function (r) { return r.date.getMonth() === selectedDate.getMonth() && r.date.getFullYear() === selectedDate.getFullYear(); });

    function shiftRange(dir) {
      if (view === 'day') setSelectedDate(addDays(selectedDate, dir));
      else if (view === 'week') setSelectedDate(addDays(selectedDate, dir * 7));
      else setSelectedDate(addMonths(selectedDate, dir));
    }
    function openDayFrom(fromView, date) {
      setPreviousView(fromView);
      setSelectedDate(date);
      setView('day');
    }
    function switchView(nextView) {
      if (nextView !== 'day') setPreviousView(nextView);
      setView(nextView);
    }
    function quickReminderAction(reminder, action) {
      setReminderBusy(function (prev) {
        var next = Object.assign({}, prev);
        next[reminder.id] = action;
        return next;
      });
      var req = action === 'complete' ? api.completeReminder(reminder.id) : api.cancelReminder(reminder.id);
      req.then(function () {
        setReminderBusy(function (prev) {
          var next = Object.assign({}, prev);
          delete next[reminder.id];
          return next;
        });
        st.reload();
        ctx.toast(action === 'complete' ? 'Задача выполнена' : 'Задача отменена');
      }).catch(function (e) {
        setReminderBusy(function (prev) {
          var next = Object.assign({}, prev);
          delete next[reminder.id];
          return next;
        });
        ctx.toast(e.message, true);
      });
    }

    function titleByView() {
      if (view === 'day') return selectedDate.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
      if (view === 'week') {
        var start = startOfWeek(selectedDate), end = addDays(start, 6);
        return start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' }) + ' - ' +
          end.toLocaleDateString('ru-RU', { day: 'numeric', month: end.getMonth() === start.getMonth() ? undefined : 'long', year: 'numeric' });
      }
      return selectedDate.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
    }

    return html`<div>
      <${PageHead} title="Календарь" sub="Задачи мимо времён намаза"
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новая задача</button>`}/>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="calendar" title="Задач пока нет" text="Создайте звонок или доставку — система обойдёт окна намаза"/></div>`
          : html`<div class="calendar-shell">
              <section class="calendar-hero">
                <div class="calendar-toolbar">
                  <div>
                    <div class="calendar-kicker">Планирование</div>
                    <div class="calendar-range-title">${titleByView()}</div>
                  </div>
                  <div class="calendar-toolbar-actions">
                    <div class="calendar-nav">
                      <button class="icon-btn calendar-nav-btn" onClick=${function () { shiftRange(-1); }}><${Icon} name="back" size=${17}/></button>
                      <button class="btn btn-ghost btn-sm" onClick=${function () { setSelectedDate(startOfDay(new Date())); }}>Сегодня</button>
                      <button class="icon-btn calendar-nav-btn" onClick=${function () { shiftRange(1); }}><${Icon} name="arrow" size=${17}/></button>
                    </div>
                    <div class="calendar-view-switch">
                      ${[['day', 'День'], ['week', 'Неделя'], ['month', 'Месяц']].map(function (v) {
                        return html`<button key=${v[0]} class=${'calendar-view-btn ' + (view === v[0] ? 'active' : '')} onClick=${function () { switchView(v[0]); }}>${v[1]}</button>`;
                      })}
                    </div>
                  </div>
                </div>
                <div class="calendar-summary-grid">
                  <div class="calendar-summary-card">
                    <span>На выбранный день</span>
                    <strong class="amana-num">${dayEvents.length}</strong>
                    <small>${plural(dayEvents.length, 'задача', 'задачи', 'задач')}</small>
                  </div>
                  <div class="calendar-summary-card">
                    <span>На неделю</span>
                    <strong class="amana-num">${weekEvents.length}</strong>
                    <small>${plural(weekEvents.length, 'событие', 'события', 'событий')}</small>
                  </div>
                  <div class="calendar-summary-card">
                    <span>На месяц</span>
                    <strong class="amana-num">${monthEvents.length}</strong>
                    <small>${plural(monthEvents.length, 'событие', 'события', 'событий')}</small>
                  </div>
                </div>
              </section>
              <div class="calendar-layout calendar-layout-full">
                <section class="calendar-main">
                  ${view === 'month' ? html`<div class="calendar-month">
                      <div class="calendar-weekdays">${['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'].map(function (label) {
                        return html`<div key=${label} class="calendar-weekday">${label}</div>`;
                      })}</div>
                      <div class="calendar-month-grid">${monthDays.map(function (day) {
                        var events = eventsForDay(normalized, day.date);
                        var selected = sameDay(day.date, selectedDate);
                        return html`<div key=${day.key} class=${'calendar-day-cell ' + (day.isCurrentMonth ? '' : 'muted ') + (selected ? 'selected' : '')} onClick=${function () { openDayFrom('month', day.date); }}>
                          <div class="calendar-day-top">
                            <span class="calendar-day-number">${day.date.getDate()}</span>
                            ${events.length ? html`<span class="calendar-day-count amana-num">${events.length}</span>` : null}
                          </div>
                          <div class="calendar-day-events">
                            ${events.slice(0, 3).map(function (r) { return html`<button key=${r.id} class=${'calendar-dot-note calendar-inline-link calendar-month-task ' + reminderTone(r.type)} onClick=${function (e) { e.stopPropagation(); ctx.go('reminder', r.id); }}>
                              <span class="calendar-month-task-label">${fmt.time(r.scheduled_at)} ${shortReminderLabel(r.type)}</span>
                              <${ReminderQuickActions} reminder=${r} busy=${reminderBusy[r.id] || ''} variant="month"
                                onComplete=${function (item) { quickReminderAction(item, 'complete'); }}
                                onCancel=${function (item) { quickReminderAction(item, 'cancel'); }}/>
                            </button>`; })}
                          </div>
                        </div>`;
                      })}</div>
                    </div>`
                    : view === 'week' ? html`<div class="calendar-week-board">
                        ${weekDays.map(function (day) {
                          var events = eventsForDay(normalized, day);
                          return html`<div key=${day.toISOString()} class=${'calendar-week-column ' + (sameDay(day, selectedDate) ? 'selected' : '')} onClick=${function () { openDayFrom('week', day); }}>
                            <div class="calendar-week-head">
                              <div class="calendar-week-name">${day.toLocaleDateString('ru-RU', { weekday: 'short' })}</div>
                              <div class="calendar-week-date amana-num">${day.getDate()}</div>
                            </div>
                            <div class="calendar-week-events">
                              ${events.length === 0 ? html`<div class="calendar-empty-mini">Нет задач</div>` : events.map(function (r) {
                                return html`<button key=${r.id} class=${'calendar-event-chip calendar-inline-link ' + reminderTone(r.type)} onClick=${function (e) { e.stopPropagation(); ctx.go('reminder', r.id); }}>
                                  <div class="calendar-week-task-top">
                                    <strong class="amana-num">${fmt.time(r.scheduled_at)}</strong>
                                    <${ReminderQuickActions} reminder=${r} busy=${reminderBusy[r.id] || ''} variant="week"
                                      onComplete=${function (item) { quickReminderAction(item, 'complete'); }}
                                      onCancel=${function (item) { quickReminderAction(item, 'cancel'); }}/>
                                  </div>
                                  <span>${ui.labels.reminderType[r.type] || r.type}</span>
                                </button>`;
                              })}
                            </div>
                          </div>`;
                        })}
                      </div>`
                    : html`<div class="calendar-day-board">
                        <div class="calendar-day-board-head">
                          <div>
                            <div class="calendar-kicker">Выбранный день</div>
                            <div class="calendar-day-board-title">${selectedDate.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })}</div>
                          </div>
                          <div class="calendar-day-board-actions">
                            <button class="btn btn-sm calendar-back-btn" onClick=${function () { setView(previousView || 'week'); }}><${Icon} name="back" size=${15}/> Назад к ${previousView === 'month' ? 'месяцу' : 'неделе'}</button>
                            <div class="calendar-day-badge amana-num">${dayEvents.length}</div>
                          </div>
                        </div>
                        <div class="calendar-day-events-list">
                          ${dayEvents.length === 0 ? html`<div class="calendar-empty-state">На этот день задач нет.</div>` : dayEvents.map(function (r) {
                            return html`<button key=${r.id} class="calendar-event-row calendar-inline-link" onClick=${function () { ctx.go('reminder', r.id); }}>
                              <div class="calendar-event-time">
                                <span class="amana-num">${fmt.time(r.scheduled_at)}</span>
                                <small>${r.duration_minutes} мин</small>
                              </div>
                              <div class="calendar-event-body">
                                <div class="calendar-event-headline">
                                  <div class="calendar-event-title">
                                    <span class=${'calendar-event-icon ' + reminderTone(r.type)}><${Icon} name=${r.type === 'delivery' ? 'truck' : r.type === 'call' ? 'phone' : 'coins'} size=${16}/></span>
                                    <span>${ui.labels.reminderType[r.type] || r.type}</span>
                                    ${r.was_shifted ? html`<span class="chip chip-partially_paid">Перенесено</span>` : html`<span class="chip chip-paid">По плану</span>`}
                                  </div>
                                  <${ReminderQuickActions} reminder=${r} busy=${reminderBusy[r.id] || ''} variant="day"
                                    onComplete=${function (item) { quickReminderAction(item, 'complete'); }}
                                    onCancel=${function (item) { quickReminderAction(item, 'cancel'); }}/>
                                </div>
                                ${r.note ? html`<div class="calendar-event-note">${r.note}</div>` : null}
                                <div class="calendar-event-meta">
                                  <span>Назначено на ${fmt.dateTime(r.scheduled_at)}</span>
                                  ${r.was_shifted ? html`<span>${r.reason}</span>` : null}
                                </div>
                              </div>
                            </button>`;
                          })}
                        </div>
                      </div>`}
                </section>
              </div>
            </div>`}
      <//>
      ${open ? html`<${ReminderModal} ctx=${ctx} onClose=${function () { setOpen(false); }} onDone=${function () { setOpen(false); st.reload(); ctx.toast('Задача создана'); }}/>` : null}
    </div>`;
  }
  function ReminderModal(p) {
    var clients = useAsync(api.listClients);
    var contracts = useAsync(api.listContracts);
    var initial = p.initial || {};
    var f = useState({
      type: initial.type || 'call',
      client_id: initial.client_id || '',
      contract_id: initial.contract_id || '',
      note: initial.note || '',
      date: initial.scheduled_at ? initial.scheduled_at.slice(0, 10) : defaultStart(),
      time: initial.scheduled_at ? new Date(initial.scheduled_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '13:00',
      duration_minutes: initial.duration_minutes || 20,
    }), v = f[0], set = f[1];
    var pv = useState(null), slot = pv[0], setSlot = pv[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function iso() { return v.date + 'T' + v.time + ':00+03:00'; }
    useEffect(function () {
      var alive = true;
      var t = setTimeout(function () {
        api.previewSlot({ desired_at: iso(), duration_minutes: Number(v.duration_minutes) }).then(function (r) { if (alive) setSlot(r); }).catch(function () {});
      }, 200);
      return function () { alive = false; clearTimeout(t); };
    }, [v.date, v.time, v.duration_minutes]);
    function save() {
      setBusy(true);
      var body = {
        type: v.type,
        client_id: v.client_id || undefined,
        contract_id: v.contract_id || undefined,
        note: v.note.trim(),
        desired_at: iso(),
        duration_minutes: Number(v.duration_minutes)
      };
      var req = initial.id ? api.updateReminder(initial.id, body) : api.createReminder(body);
      req.then(p.onDone).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var upd = function (k) { return function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }; };
    return html`<${ui.Modal} title=${initial.id ? 'Редактировать задачу' : 'Новая задача'} onClose=${p.onClose} width=${460}>
      <${ui.Field} label="Тип">
        <select class="select" value=${v.type} onChange=${upd('type')}>
          <option value="call">Звонок</option><option value="delivery">Доставка</option><option value="payment_followup">Контакт по платежу</option></select>
      <//>
      <${ui.Field} label="Клиент (необязательно)">
        <select class="select" value=${v.client_id} onChange=${upd('client_id')}>
          <option value="">— не выбран —</option>
          ${(clients.data || []).map(function (c) { return html`<option key=${c.id} value=${c.id}>${c.full_name}</option>`; })}</select>
      <//>
      <${ui.Field} label="Договор (необязательно)">
        <select class="select" value=${v.contract_id} onChange=${upd('contract_id')}>
          <option value="">— не выбран —</option>
          ${(contracts.data || []).map(function (c) { return html`<option key=${c.id} value=${c.id}>${(c.client_name || c.id.slice(0, 8)) + ' · остаток ' + fmt.money(c.outstanding)}</option>`; })}</select>
      <//>
      <div style=${{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        <${ui.Field} label="Дата"><input class="input" type="date" value=${v.date} onInput=${upd('date')}/><//>
        <${ui.Field} label="Время"><input class="input" type="time" value=${v.time} onInput=${upd('time')}/><//>
      </div>
      <${ui.Field} label="Длительность (мин)"><input class="input" type="number" value=${v.duration_minutes} onInput=${upd('duration_minutes')}/><//>
      <${ui.Field} label="Заметка"><input class="input" value=${v.note} placeholder="Доставка дивана" onInput=${upd('note')}/><//>
      ${slot ? (slot.was_shifted
        ? html`<div class="banner banner-warn" style=${{ fontSize: 13 }}><${Icon} name="moon2" size=${17}/><div>Будет перенесено на <b>${fmt.time(slot.scheduled_at)}</b> — ${slot.reason}</div></div>`
        : html`<div class="banner banner-accent" style=${{ fontSize: 13 }}><${Icon} name="check" size=${17}/> Время свободно — переноса не требуется.</div>`) : null}
      <button class="btn btn-primary btn-block" disabled=${busy || clients.loading || contracts.loading} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : initial.id ? 'Сохранить изменения' : 'Создать задачу'}</button>
    <//>`;
  }

  function ReminderCard(ctx) {
    var st = useAsync(function () { return api.getReminder(ctx.route.id); }, [ctx.route.id]);
    var clients = useAsync(api.listClients);
    var contracts = useAsync(api.listContracts);
    var ed = useState(false), editOpen = ed[0], setEditOpen = ed[1];
    var bz = useState(''), busy = bz[0], setBusy = bz[1];
    function runAction(action, okText) {
      if (!st.data) return;
      setBusy(action);
      (action === 'complete' ? api.completeReminder(st.data.id) : api.cancelReminder(st.data.id))
        .then(function () { setBusy(''); st.reload(); ctx.toast(okText); })
        .catch(function (e) { setBusy(''); ctx.toast(e.message, true); });
    }
    return html`<div>
      <div style=${{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
        <button class="icon-btn" onClick=${function () { ctx.go('schedule'); }}><${Icon} name="back" size=${20}/></button>
        <div class="page-title" style=${{ fontSize: 22 }}>Задача</div>
      </div>
      <${Guard} loading=${st.loading || clients.loading || contracts.loading} err=${st.err || clients.err || contracts.err}>
        ${(function () {
          var reminder = st.data;
          if (!reminder) return html`<div class="card"><${ui.Empty} icon="calendar" title="Задача не найдена"/></div>`;
          var client = findByID(clients.data, reminder.client_id);
          var contract = findByID(contracts.data, reminder.contract_id);
          var canMutate = reminder.base_status === 'scheduled';
          return html`<div class="grid reminder-card-grid" style=${{ gap: 14 }}>
            <div class="card card-pad reminder-card-compact">
              <div style=${{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>
                <div>
                  <div style=${{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
                    <span class=${'calendar-event-icon ' + reminderTone(reminder.type)}><${Icon} name=${reminder.type === 'delivery' ? 'truck' : reminder.type === 'call' ? 'phone' : 'coins'} size=${16}/></span>
                    <div class="reminder-card-title">${ui.labels.reminderType[reminder.type] || reminder.type}</div>
                  </div>
                  <div class="page-sub">Назначено на ${fmt.dateTime(reminder.scheduled_at)}</div>
                </div>
                <div style=${{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <${ui.StatusChip} map="reminderStatus" value=${reminder.status}/>
                  ${reminder.was_shifted ? html`<span class="chip chip-partially_paid">Перенесено</span>` : html`<span class="chip chip-paid">По плану</span>`}
                </div>
              </div>
              ${reminder.note ? html`<div class="banner banner-accent reminder-note-banner" style=${{ marginTop: 14 }}><${Icon} name="doc" size=${16}/> ${reminder.note}</div>` : null}
              <div class="compact-fields compact-fields-2 reminder-compact-fields" style=${{ marginTop: 14 }}>
                <div class="compact-field"><span>Дата и время</span><strong class="amana-num">${fmt.dateTime(reminder.scheduled_at)}</strong></div>
                <div class="compact-field"><span>Статус</span><strong>${reminderStatusText(reminder)}</strong></div>
                <div class="compact-field"><span>Длительность</span><strong class="amana-num">${reminder.duration_minutes} мин</strong></div>
                <div class="compact-field"><span>Клиент</span><strong>${client ? client.full_name : reminder.client_id ? 'Привязан клиент' : 'Не указан'}</strong></div>
                <div class="compact-field"><span>Договор</span><strong>${contract ? contract.id.slice(0, 8) : reminder.contract_id ? reminder.contract_id.slice(0, 8) : 'Не указан'}</strong></div>
                <div class="compact-field"><span>Желаемое время</span><strong class="amana-num">${fmt.dateTime(reminder.desired_at)}</strong></div>
              </div>
              ${reminder.was_shifted ? html`<div class="banner banner-warn reminder-note-banner" style=${{ marginTop: 14 }}><${Icon} name="moon2" size=${16}/> ${reminder.reason}</div>` : null}
              ${reminder.completed_at ? html`<div class="banner banner-accent reminder-note-banner" style=${{ marginTop: 14 }}><${Icon} name="check" size=${16}/> Выполнена ${fmt.dateTime(reminder.completed_at)}</div>` : null}
              ${reminder.cancelled_at ? html`<div class="banner banner-warn reminder-note-banner" style=${{ marginTop: 14 }}><${Icon} name="x" size=${16}/> Отменена ${fmt.dateTime(reminder.cancelled_at)}</div>` : null}
            </div>
            <div class="reminder-actions-stack">
              <section class="reminder-decision-panel">
                <div class="reminder-decision-head">
                  <div>
                    <div class="reminder-decision-kicker">Главное действие</div>
                  </div>
                  ${canMutate ? html`<span class="chip chip-pending">Ожидает решения</span>` : html`<${ui.StatusChip} map="reminderStatus" value=${reminder.status}/>`}
                </div>
                ${canMutate
                  ? html`<div class="reminder-decision-grid">
                      <button class="reminder-decision-btn complete" disabled=${busy === 'complete'} onClick=${function () { runAction('complete', 'Задача отмечена выполненной'); }}>
                        <span class="reminder-decision-icon"><${Icon} name="check" size=${18}/></span>
                        <span class="reminder-decision-copy">
                          <strong>${busy === 'complete' ? 'Сохраняем выполнение…' : 'Выполнить задачу'}</strong>
                          <small>Закрыть задачу как сделанную и убрать её из активных.</small>
                        </span>
                      </button>
                      <button class="reminder-decision-btn cancel" disabled=${busy === 'cancel'} onClick=${function () { runAction('cancel', 'Задача отменена'); }}>
                        <span class="reminder-decision-icon"><${Icon} name="x" size=${18}/></span>
                        <span class="reminder-decision-copy">
                          <strong>${busy === 'cancel' ? 'Отменяем задачу…' : 'Отменить задачу'}</strong>
                          <small>Закрыть задачу без выполнения, если она больше не нужна.</small>
                        </span>
                      </button>
                    </div>`
                  : html`<div class="reminder-decision-locked">
                      <${Icon} name=${reminder.status === 'completed' ? 'check' : 'x'} size=${18}/>
                      <div>
                        <strong>${reminder.status === 'completed' ? 'Задача уже выполнена' : 'Задача уже отменена'}</strong>
                        <span>${reminder.status === 'completed' ? 'Основное действие уже завершено.' : 'Повторное решение по задаче больше не требуется.'}</span>
                      </div>
                    </div>`}
              </section>

              <section class="table-card reminder-support-panel">
                <div class="table-card-head">Дополнительные действия</div>
                <div class="reminder-actions-wrap reminder-actions-wrap-secondary">
                  ${canMutate ? html`<button class="btn btn-primary btn-sm" onClick=${function () { setEditOpen(true); }}><${Icon} name="edit" size=${15}/> Редактировать</button>` : null}
                  ${client ? html`<button class="btn btn-ghost btn-sm" onClick=${function () { ctx.go('client', client.id); }}><${Icon} name="clients" size=${15}/> Открыть клиента</button>` : null}
                  ${contract ? html`<button class="btn btn-ghost btn-sm" onClick=${function () { ctx.go('contract', contract.id); }}><${Icon} name="contracts" size=${15}/> Открыть договор</button>` : null}
                  <button class="btn btn-ghost btn-sm" onClick=${function () { ctx.go('schedule'); }}><${Icon} name="calendar" size=${15}/> К календарю</button>
                </div>
              </section>
            </div>
          </div>`;
        })()}
      <//>
      ${editOpen && st.data ? html`<${ReminderModal} ctx=${ctx} initial=${st.data} onDone=${function () { setEditOpen(false); st.reload(); ctx.toast('Задача обновлена'); }} onClose=${function () { setEditOpen(false); }}/>` : null}
    </div>`;
  }

  /* ================= DEVELOPERS (API KEYS) ================= */
  function Developers(ctx) {
    if (!ctx.isOwner) return html`<div class="card"><${ui.Empty} icon="shield" title="Доступ только для владельца"/></div>`;
    var st = useAsync(api.listApiKeys);
    var rev = useState(null), revealed = rev[0], setRevealed = rev[1];
    var nm = useState(''), name = nm[0], setName = nm[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function create() {
      if (!name.trim()) { ctx.toast('Укажите название', true); return; }
      setBusy(true);
      api.createApiKey(name.trim()).then(function (r) { setBusy(false); setName(''); setRevealed(r); st.reload(); }).catch(function (e) { setBusy(false); ctx.toast(e.message, true); });
    }
    function revoke(id) {
      ctx.confirm({ title: 'Отозвать ключ', text: 'Ключ перестанет работать немедленно.', okLabel: 'Отозвать', danger: true,
        onOk: function () { api.revokeApiKey(id).then(function () { st.reload(); ctx.toast('Ключ отозван'); }).catch(function (e) { ctx.toast(e.message, true); }); } });
    }
    var list = st.data || [];
    return html`<div>
      <${PageHead} title="Разработчикам" sub="Ключи для публичного API (/api/v1)"
        actions=${html`<a class="btn btn-ghost" href="/swagger/" target="_blank"><${Icon} name="doc" size=${17}/> Swagger</a>`}/>
      <div class="card card-pad" style=${{ marginBottom: 16, display: 'flex', gap: 10, alignItems: 'flex-end' }}>
        <${ui.Field} label="Название ключа" style=${{ flex: 1 }}>
          <input class="input" value=${name} placeholder="Маркетплейс «Беркат»" onInput=${function (e) { setName(e.target.value); }} onKeyDown=${function (e) { if (e.key === 'Enter') create(); }}/>
        <//>
        <button class="btn btn-primary" disabled=${busy} onClick=${create}>${busy ? html`<${ui.Spinner}/>` : 'Создать ключ'}</button>
      </div>
      ${revealed ? html`<div class="banner banner-accent" style=${{ marginBottom: 16, flexDirection: 'column', alignItems: 'stretch', gap: 8 }}>
        <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}><b>Ключ создан — скопируйте сейчас, он показывается один раз</b>
          <button class="icon-btn" onClick=${function () { setRevealed(null); }}><${Icon} name="x" size=${16}/></button></div>
        <code style=${{ fontFamily: 'monospace', background: 'var(--surface)', padding: '10px 12px', borderRadius: 10, border: '1px solid var(--border)', wordBreak: 'break-all' }}>${revealed.key}</code>
        <button class="btn btn-soft btn-sm" style=${{ alignSelf: 'flex-start' }} onClick=${function () { try { navigator.clipboard.writeText(revealed.key); ctx.toast('Скопировано'); } catch (e) {} }}><${Icon} name="copy" size=${15}/> Скопировать</button>
      </div>` : null}
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="key" title="Ключей нет"/></div>`
          : html`<div class="card" style=${{ overflow: 'hidden' }}><table class="table"><thead><tr><th>Название</th><th>Префикс</th><th>Создан</th><th>Статус</th><th></th></tr></thead>
            <tbody>${list.map(function (k) {
              return html`<tr key=${k.id}>
                <td style=${{ fontWeight: 600 }}>${k.name}</td>
                <td style=${{ fontFamily: 'monospace', fontSize: 12.5, color: 'var(--fg-muted)' }}>${k.prefix}…</td>
                <td>${fmt.date(k.created_at)}</td>
                <td><span class=${'chip ' + (k.active ? 'chip-paid' : 'chip-cancelled')}>${k.active ? 'Активен' : 'Отозван'}</span></td>
                <td style=${{ textAlign: 'right' }}>${k.active ? html`<button class="icon-btn" onClick=${function () { revoke(k.id); }} title="Отозвать"><${Icon} name="trash" size=${17}/></button>` : null}</td></tr>`;
            })}</tbody></table></div>`}
      <//>
    </div>`;
  }

  /* ================= SETTINGS ================= */
  function Settings(ctx) {
    var me = useAsync(api.me);
    var tgLink = useAsync(api.telegramLink);
    var u = me.data || {};
    var tg = tgLink.data;

    function copyLink() {
      var url = tg && tg.url;
      if (!url) return;
      var ok = function () { ctx.toast('Ссылка скопирована'); };
      var fail = function () { ctx.toast('Не удалось скопировать — выделите вручную', true); };
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) navigator.clipboard.writeText(url).then(ok, fail);
        else fail();
      } catch (e) { fail(); }
    }

    return html`<div>
      <${PageHead} title="Настройки"/>
      <${Guard} loading=${me.loading} err=${me.err}>
        <div class="grid settings-grid">
          ${tg && tg.available ? html`<div class="card card-pad" style=${{ gridColumn: '1 / -1' }}>
            <div style=${{ fontWeight: 700, marginBottom: 4 }}>Telegram-бот · ваша персональная ссылка</div>
            <div style=${{ fontSize: 13, color: 'var(--fg-muted)', marginBottom: 14 }}>Покажите QR клиенту или отправьте ссылку. Он откроет бота, представится (ФИО и телефон) — и его сообщения придут именно вам в «Чат», а ваши ответы вернутся ему в Telegram.</div>
            <div class="tg-qr-wrap">
              ${tg.qr ? html`<img class="tg-qr" src=${tg.qr} alt="QR-код Telegram-бота" width="180" height="180"/>` : null}
              <div class="tg-qr-side">
                <input class="input tg-link-input" readonly value=${tg.url} onFocus=${function (e) { e.target.select(); }}/>
                <div class="tg-qr-actions">
                  <button class="btn btn-primary btn-sm" onClick=${copyLink}>Скопировать ссылку</button>
                  <a class="btn btn-ghost btn-sm" href=${tg.url} target="_blank" rel="noopener">Открыть бота</a>
                </div>
              </div>
            </div>
          </div>` : null}
          <div class="card card-pad">
            <div style=${{ fontWeight: 700, marginBottom: 12 }}>Профиль</div>
            ${[['Имя', u.full_name], ['Email', u.email], ['Роль', u.role === 'owner' ? 'Владелец' : 'Менеджер']].map(function (r, i) {
              return html`<div key=${i} style=${{ display: 'flex', justifyContent: 'space-between', padding: '9px 0', borderBottom: '1px solid var(--border)' }}>
                <span style=${{ color: 'var(--fg-muted)' }}>${r[0]}</span><span style=${{ fontWeight: 600 }}>${r[1] || '—'}</span></div>`;
            })}
          </div>
          <div class="card card-pad">
            <div style=${{ fontWeight: 700, marginBottom: 12 }}>Намаз и рассрочка</div>
            <div style=${{ fontSize: 13.5, color: 'var(--fg-muted)', lineHeight: 1.7 }}>
              Координаты: <b style=${{ color: 'var(--fg)' }}>Грозный (43.32, 45.69)</b><br/>
              Мазхаб: <b style=${{ color: 'var(--fg)' }}>шафиитский</b><br/>
              Окно джума: <b style=${{ color: 'var(--fg)' }}>12:30–14:00</b><br/>
              Буфер после молитвы: <b style=${{ color: 'var(--fg)' }}>20 мин</b><br/>
              Параметры задаются через переменные окружения <code>PRAYER_*</code>.
            </div>
          </div>
          <div class="card card-pad">
            <div style=${{ fontWeight: 700, marginBottom: 12 }}>Тема оформления</div>
            <button class="btn btn-ghost" onClick=${ctx.toggleTheme}><${Icon} name=${ctx.theme === 'dark' ? 'sun' : 'moon'} size=${17}/> ${ctx.theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'}</button>
          </div>
        </div>
      <//>
    </div>`;
  }

  function defaultStart() {
    var d = new Date(); d.setMonth(d.getMonth() + 1);
    return d.toISOString().slice(0, 10);
  }

  function Chat(ctx) {
    var chats = useAsync(api.listChats);
    var clients = useAsync(api.listClients);
    var tgLink = useAsync(api.telegramLink);
    var q = useState(''), query = q[0], setQuery = q[1];
    var sel = useState(ctx.route.id || null), selId = sel[0], setSelId = sel[1];
    var convMap = {};
    (chats.data || []).forEach(function (c) { convMap[c.client_id] = c; });
    var allClients = (clients.data || []).slice().sort(function (a, b) {
      var ca = convMap[a.id], cb = convMap[b.id];
      var ta = ca ? new Date(ca.last_message_at || 0).getTime() : 0;
      var tb = cb ? new Date(cb.last_message_at || 0).getTime() : 0;
      if (ta !== tb) return tb - ta;
      return String(a.full_name || '').localeCompare(String(b.full_name || ''), 'ru');
    });
    var filtered = allClients.filter(function (c) {
      var hay = [c.full_name, c.phone, c.document].join(' ').toLowerCase();
      return hay.indexOf(query.trim().toLowerCase()) >= 0;
    });

    useEffect(function () {
      if (selId) return;
      if (filtered.length > 0) setSelId(filtered[0].id);
    }, [filtered.length, selId]);

    useEffect(function () {
      if (!selId || filtered.some(function (c) { return c.id === selId; })) return;
      setSelId(filtered.length ? filtered[0].id : null);
    }, [query, filtered.length, selId]);

    var current = findByID(clients.data, selId);
    var currentConv = selId ? convMap[selId] : null;
    var activeCount = Object.keys(convMap).length;

    function previewText(c) {
      if (!c) return 'Диалог ещё не начат';
      return (c.last_sender === 'staff' ? 'Вы: ' : '') + (c.last_message || 'Без текста');
    }

    function copyLink() {
      var url = tgLink.data && tgLink.data.url;
      if (!url) return;
      var ok = function () { ctx.toast('Ссылка скопирована'); };
      var fail = function () { ctx.toast('Не удалось скопировать — выделите ссылку вручную', true); };
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) navigator.clipboard.writeText(url).then(ok, fail);
        else fail();
      } catch (e) { fail(); }
    }

    var tg = tgLink.data;

    return html`<div>
      <${PageHead} title="Чат" sub=${activeCount ? ('Активных диалогов: ' + activeCount) : 'Внутренняя переписка с клиентами'}/>
      ${tg && tg.available ? html`<div class="tg-link">
          <div class="tg-link-info">
            <div class="tg-link-title"><${Icon} name="chat" size=${15}/> Ваша ссылка на Telegram-бот</div>
            <div class="tg-link-sub">Отправьте её клиенту: он откроет бота, представится (ФИО и телефон) — и его сообщения придут сюда, а ваши ответы вернутся ему в Telegram.</div>
          </div>
          <div class="tg-link-row">
            <input class="input tg-link-input" readonly value=${tg.url} onFocus=${function (e) { e.target.select(); }}/>
            <button class="btn btn-primary btn-sm" onClick=${copyLink}>Скопировать</button>
            <a class="btn btn-ghost btn-sm" href=${tg.url} target="_blank" rel="noopener">Открыть</a>
          </div>
        </div>` : null}
      <${Guard} loading=${chats.loading || clients.loading} err=${chats.err || clients.err}>
        ${(clients.data || []).length === 0
          ? html`<div class="card"><${ui.Empty} icon="clients" title="Клиентов пока нет" text="Добавьте клиента, чтобы начать переписку."/></div>`
          : html`<div class="chat-workspace">
              <div class="chat-sidebar">
                <div class="chat-sidebar-head">
                  <div>
                    <div class="chat-sidebar-title">Клиенты</div>
                    <div class="chat-sidebar-sub">${allClients.length} в базе · ${activeCount} с перепиской</div>
                  </div>
                  <span class="chat-counter">${filtered.length}</span>
                </div>
                <div class="chat-search">
                  <${Icon} name="search" size=${16}/>
                  <input class="input" value=${query} placeholder="Поиск по имени, телефону, документу"
                    onInput=${function (e) { setQuery(e.target.value); }}/>
                </div>
                <div class="chat-client-list">
                  ${filtered.length === 0 ? html`<div class="chat-list-empty">Ничего не найдено</div>` : filtered.map(function (client) {
                    var conv = convMap[client.id];
                    return html`<button key=${client.id} class=${'chat-client-item ' + (selId === client.id ? 'active' : '')}
                      onClick=${function () { setSelId(client.id); }}>
                      <span class="chat-avatar chat-avatar-logo"><${Icon} name="logo" size=${19} sw=${1.9}/></span>
                      <div class="chat-client-copy">
                        <div class="chat-client-top">
                          <span class="chat-client-name">${client.full_name}</span>
                          <span class="chat-client-time">${conv ? fmt.time(conv.last_message_at) : 'Новый'}</span>
                        </div>
                        <div class="chat-client-preview">${previewText(conv)}</div>
                        <div class="chat-client-meta">${client.phone || 'Телефон не указан'}</div>
                      </div>
                    </button>`;
                  })}
                </div>
              </div>
              <div class="chat-main-card">
                ${current ? html`<div class="chat-main-head">
                    <div class="chat-main-person">
                      <span class="chat-avatar chat-avatar-logo chat-avatar-lg"><${Icon} name="logo" size=${24} sw=${1.9}/></span>
                      <div>
                        <div class="chat-main-name">${current.full_name}</div>
                        <div class="chat-main-sub">${current.phone || 'Телефон не указан'}${current.document ? ' · ' + current.document : ''}</div>
                      </div>
                    </div>
                    <div class="chat-main-actions">
                      <span class=${'compact-chip ' + (currentConv ? 'compact-chip-ok' : 'compact-chip-muted')}>${currentConv ? 'Диалог активен' : 'Новый диалог'}</span>
                      <button class="btn btn-ghost btn-sm" onClick=${function () { ctx.go('client', current.id); }}>
                        <${Icon} name="arrow" size=${14}/> Карточка клиента
                      </button>
                    </div>
                  </div>
                  <div class="chat-main-summary">
                    ${currentConv ? html`<span>Последнее сообщение ${fmt.date(currentConv.last_message_at)} в ${fmt.time(currentConv.last_message_at)}</span>` : html`<span>Переписка ещё не начата. Можно отправить первое сообщение.</span>`}
                  </div>
                  <${ui.ChatThread} threadKey=${selId} meKind="staff"
                    load=${function () { return api.chatThread(selId); }}
                    onSend=${function (body) { return api.sendChatMessage(selId, body).then(function (r) { chats.reload(); return r; }); }}
                    onError=${function (e) { ctx.toast(e.message, true); }}
                    placeholder="Напишите клиенту коротко и по делу…"
                    emptyText="Сообщений пока нет. Отправьте первое сообщение клиенту."/>`
                  : html`<div class="card"><${ui.Empty} icon="chat" title="Выберите клиента" text="Слева можно открыть существующий диалог или начать новый."/></div>`}
              </div>
            </div>`}
      <//>
    </div>`;
  }

  /* ================= FINANCE (доходы и расходы) =================
     Income is derived on the backend from contracts (продажа − закупка);
     expenses = cost of goods (auto) + manual entries. The P&L summary and the
     per-sale breakdown come from GET /finance/report; the manual expense list
     from GET /finance/expenses. */
  function Finance(ctx) {
    var rep = useAsync(api.financeReport);
    var exp = useAsync(api.listExpenses);
    var products = useAsync(api.listProducts);
    var m = useState(false), open = m[0], setOpen = m[1];
    var pg = useState(1), salesPage = pg[0], setSalesPage = pg[1];
    var r = rep.data;
    var productName = function (id) { var p = findByID(products.data, id); return p ? p.name : '—'; };

    function removeExpense(e) {
      ctx.confirm({
        title: 'Удалить расход?', text: e.category + ' — ' + fmt.money(e.amount), okLabel: 'Удалить', danger: true,
        onOk: function () {
          api.deleteExpense(e.id).then(function () { exp.reload(); rep.reload(); ctx.toast('Расход удалён'); })
            .catch(function (ex) { ctx.toast(ex.message, true); });
        },
      });
    }

    return html`<div>
      <${PageHead} title="Финансы" sub="Доходы и расходы — доход = продажа − закупка"
        actions=${html`<div style=${{ display: 'flex', gap: 9 }}>
          <button class="btn btn-soft" onClick=${function () { api.downloadFinanceReportPdf().catch(function (e) { ctx.toast(e.message, true); }); }}><${Icon} name="doc" size=${16}/> Отчёт PDF</button>
          <button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Добавить расход</button>
        </div>`}/>
      <${Guard} loading=${rep.loading} err=${rep.err}>
        ${r ? html`<div>
          <div class="grid" style=${{ gridTemplateColumns: 'repeat(5, minmax(0,1fr))', gap: 12, marginBottom: 16 }}>
            ${(function () {
              var netNeg = parseFloat(r.net_profit) < 0;
              var cards = [
                ['Выручка', r.revenue, 'Сумма продаж', '', false],
                ['Себестоимость', r.cost_of_goods, 'Закупка проданного', 'delta-neg', false],
                ['Валовая прибыль', r.gross_profit, 'Продажа − закупка', 'delta-pos', false],
                ['Прочие расходы', r.other_expenses, 'Аренда, ремонт…', 'delta-neg', false],
                ['Чистая прибыль', r.net_profit, 'Итог после расходов', netNeg ? 'delta-neg' : 'delta-pos', true],
              ];
              return cards.map(function (k, i) {
                return html`<div key=${i} class="card card-pad" style=${k[4] ? { borderColor: 'var(--accent-bd)', background: 'var(--accent-soft)' } : null}>
                  <div class="kpi">
                    <div class=${'v amana-num ' + k[3]} style=${{ fontSize: 20 }}>${fmt.money(k[1])}</div>
                    <div class="l">${k[0]}</div>
                    <div style=${{ fontSize: 11.5, color: 'var(--fg-subtle)', marginTop: 3 }}>${k[2]}</div>
                  </div>
                </div>`;
              });
            })()}
          </div>
          <div class="grid" style=${{ gridTemplateColumns: '1fr 1fr', gap: 16, alignItems: 'start' }}>
            <div class="table-card">
              <div class="table-card-head finance-panel-head finance-panel-head-expenses">
                <span class="finance-panel-kicker">Расходы</span>
                <strong>Прочие расходы</strong>
              </div>
              <${Guard} loading=${exp.loading} err=${exp.err}>
                ${(exp.data || []).length === 0 ? html`<div class="card"><${ui.Empty} icon="coins" title="Расходов нет" text="Добавьте аренду, ремонт, логистику и т.д."/></div>`
                  : html`<table class="data-table"><thead><tr><th>Категория</th><th>Дата</th><th>Сумма</th><th></th></tr></thead>
                      <tbody>${(exp.data || []).map(function (e) {
                        return html`<tr key=${e.id} class="data-row">
                          <td><div class="table-title">${e.category}</div>${e.note ? html`<div class="table-subline">${e.note}</div>` : null}</td>
                          <td>${fmt.date(e.spent_at)}</td>
                          <td><strong class="table-money amana-num delta-neg">−${fmt.money(e.amount)}</strong></td>
                          <td class="table-arrow"><button class="icon-btn" title="Удалить" onClick=${function () { removeExpense(e); }}><${Icon} name="trash" size=${15}/></button></td>
                        </tr>`;
                      })}</tbody></table>`}
              <//>
            </div>
            <div class="table-card">
              <div class="table-card-head finance-panel-head finance-panel-head-sales">
                <span class="finance-panel-kicker">Продажи</span>
                <strong>Доходы от продаж</strong>
              </div>
              ${(function () {
                var sales = (r.sales || []).filter(function (s) { return s.status !== 'cancelled'; });
                var perPage = 4;
                var totalPages = Math.max(1, Math.ceil(sales.length / perPage));
                var currentPage = Math.min(salesPage, totalPages);
                var visibleSales = sales.slice((currentPage - 1) * perPage, currentPage * perPage);
                return sales.length === 0 ? html`<div class="card"><${ui.Empty} icon="contracts" title="Продаж пока нет"/></div>`
                  : html`<div class="finance-sales-list">${visibleSales.map(function (s) {
                      return html`<button key=${s.contract_id} class="finance-sale-card row-link" onClick=${function () { ctx.go('contract', s.contract_id); }}>
                        <div class="finance-sale-main">
                          <div class="finance-sale-top">
                            <div>
                              <div class="finance-sale-title">${productName(s.product_id)}</div>
                              <div class="finance-sale-subline">Договор ${s.contract_id.slice(0, 8)} · ${fmt.date(s.created_at)}</div>
                            </div>
                            <strong class="finance-sale-profit amana-num delta-pos">+${fmt.money(s.profit)}</strong>
                          </div>
                          <div class="finance-sale-meta">
                            <span>Продажа <b class="amana-num">${fmt.money(s.sale_price)}</b></span>
                            <span>Закупка <b class="amana-num">${fmt.money(s.cost_price)}</b></span>
                            <${ui.StatusChip} map="contractStatus" value=${s.status}/>
                          </div>
                        </div>
                        <div class="finance-sale-arrow"><${Icon} name="arrow" size=${16}/></div>
                      </button>`;
                    })}
                    ${totalPages > 1 ? html`<div class="finance-sales-pagination">
                      <button class="btn btn-ghost btn-sm" disabled=${currentPage === 1} onClick=${function () { setSalesPage(currentPage - 1); }}>
                        <${Icon} name="back" size=${14}/> Назад
                      </button>
                      <div class="finance-sales-page-indicator">Страница <b class="amana-num">${currentPage}</b> из <b class="amana-num">${totalPages}</b></div>
                      <button class="btn btn-ghost btn-sm" disabled=${currentPage === totalPages} onClick=${function () { setSalesPage(currentPage + 1); }}>
                        Вперёд <${Icon} name="arrow" size=${14}/>
                      </button>
                    </div>` : null}
                  </div>`;
              })()}
            </div>
          </div>
        </div>` : null}
      <//>
      ${open ? html`<${ExpenseModal} ctx=${ctx} onClose=${function () { setOpen(false); }}
        onSaved=${function () { setOpen(false); exp.reload(); rep.reload(); ctx.toast('Расход добавлен'); }}/>` : null}
    </div>`;
  }
  function ExpenseModal(p) {
    var f = useState({ category: 'Аренда', amount: '', note: '', spent_at: new Date().toISOString().slice(0, 10) }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function upd(o) { set(Object.assign({}, v, o)); }
    function save() {
      if (!v.category.trim() || !v.amount.trim()) { p.ctx.toast('Заполните категорию и сумму', true); return; }
      setBusy(true);
      api.createExpense({ category: v.category.trim(), amount: v.amount.replace(',', '.').trim(), note: v.note.trim(), spent_at: v.spent_at })
        .then(p.onSaved).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var inp = function (k, ph) { return html`<input class="input" value=${v[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; upd(o); }}/>`; };
    var cats = ['Аренда', 'Ремонт', 'Логистика', 'Реклама', 'Зарплата', 'Налоги', 'Прочее'];
    return html`<${ui.Modal} title="Новый расход" onClose=${p.onClose}>
      <${ui.Field} label="Категория">
        <input class="input" list="amana-expense-cats" value=${v.category} placeholder="Аренда, Ремонт…"
          onInput=${function (e) { upd({ category: e.target.value }); }}/>
        <datalist id="amana-expense-cats">${cats.map(function (c) { return html`<option key=${c} value=${c}></option>`; })}</datalist>
      <//>
      <div class="grid" style=${{ gridTemplateColumns: '1fr 1fr' }}>
        <${ui.Field} label="Сумма, ₽">${inp('amount', '15000')}<//>
        <${ui.Field} label="Дата">
          <input class="input" type="date" value=${v.spent_at} onInput=${function (e) { upd({ spent_at: e.target.value }); }}/>
        <//>
      </div>
      <${ui.Field} label="Комментарий (необязательно)">${inp('note', 'Аренда зала за месяц')}<//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : 'Сохранить'}</button>
    <//>`;
  }

  AM.screens = { dashboard: Dashboard, clients: Clients, client: ClientCard, catalog: Catalog, product: ProductCard, contracts: Contracts, reminder: ReminderCard,
    'contract-new': ContractWizard, contract: ContractCard, schedule: Schedule, chat: Chat, finance: Finance,
    developers: Developers, settings: Settings };
})();
