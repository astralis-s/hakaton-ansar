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

  /* ================= DASHBOARD ================= */
  function Dashboard(ctx) {
    var contracts = useAsync(api.listContracts);
    var charity = useAsync(api.listCharity);
    var reminders = useAsync(api.listReminders);
    var list = contracts.data || [];
    var active = list.filter(function (c) { return c.status === 'active'; });
    var outstanding = active.reduce(function (a, c) { return a + parseFloat(c.outstanding || 0); }, 0);
    var totalCharity = (charity.data && charity.data.total_amount) || '0';
    var shifted = (reminders.data || []).filter(function (r) { return r.was_shifted; });

    var kpis = [
      { v: active.length, l: 'Активные договоры', ico: 'contracts' },
      { v: fmt.money(String(outstanding)), l: 'Остаток к получению', ico: 'coins' },
      { v: fmt.money(totalCharity), l: 'Собрано садаки', ico: 'charity' },
      { v: (reminders.data || []).length, l: 'Напоминаний', ico: 'calendar' },
    ];
    return html`<div>
      <${PageHead} title="Дашборд" sub="Общая картина бизнеса"
        actions=${html`<button class="btn btn-primary" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> Новый договор</button>`}/>
      <${Guard} loading=${contracts.loading} err=${contracts.err}>
        <div class="grid" style=${{ gridTemplateColumns: 'repeat(4,1fr)' }}>
          ${kpis.map(function (k, i) {
            return html`<div key=${i} class="card card-pad" style=${{ animation: 'amRise .5s ease both ' + (i * 0.05) + 's' }}>
              <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div class="kpi"><div class="v amana-num">${k.v}</div><div class="l">${k.l}</div></div>
                <span style=${{ width: 40, height: 40, borderRadius: 12, background: 'var(--grad-soft)', color: 'var(--accent)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><${Icon} name=${k.ico} size=${20}/></span>
              </div></div>`;
          })}
        </div>
        <div class="grid" style=${{ gridTemplateColumns: '1.4fr 1fr', marginTop: 16 }}>
          <div class="card card-pad">
            <div style=${{ fontWeight: 700, marginBottom: 12 }}>Последние договоры</div>
            ${list.length === 0 ? html`<${ui.Empty} icon="contracts" title="Договоров пока нет" text="Оформите первый договор рассрочки"/>`
              : html`<table class="table"><tbody>${list.slice(0, 6).map(function (c) {
                return html`<tr key=${c.id} class="row-link" onClick=${function () { ctx.go('contract', c.id); }}>
                  <td><div style=${{ fontWeight: 600 }} class="amana-num">${fmt.money(c.sale_price)}</div>
                    <div style=${{ fontSize: 12.5, color: 'var(--fg-subtle)' }}>остаток ${fmt.money(c.outstanding)}</div></td>
                  <td style=${{ width: 130 }}><div class="progress"><i style=${{ width: (c.progress_percent || 0) + '%' }}></i></div></td>
                  <td style=${{ width: 110, textAlign: 'right' }}><${ui.StatusChip} map="contractStatus" value=${c.status}/></td></tr>`;
              })}</tbody></table>`}
          </div>
          <div class="card card-pad">
            <div style=${{ fontWeight: 700, marginBottom: 12 }}>Перенесено мимо намаза</div>
            ${shifted.length === 0 ? html`<${ui.Empty} icon="calendar" title="Нет переносов" text="Задачи вне окон молитв"/>`
              : shifted.slice(0, 4).map(function (r) {
                return html`<div key=${r.id} style=${{ padding: '10px 0', borderBottom: '1px solid var(--border)' }}>
                  <div style=${{ fontWeight: 600, fontSize: 14 }}>${ui.labels.reminderType[r.type] || r.type}</div>
                  <div style=${{ fontSize: 12.5, color: 'var(--st-part-fg)' }}>${r.reason || 'перенесено'}</div></div>`;
              })}
          </div>
        </div>
      <//>
    </div>`;
  }

  /* ================= CLIENTS ================= */
  function Clients(ctx) {
    var st = useAsync(api.listClients);
    var q = useState(''), query = q[0], setQuery = q[1];
    var m = useState(false), open = m[0], setOpen = m[1];
    var list = (st.data || []).filter(function (c) { return c.full_name.toLowerCase().indexOf(query.toLowerCase()) >= 0; });
    return html`<div>
      <${PageHead} title="Клиенты" sub=${(st.data || []).length + ' клиентов'}
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новый клиент</button>`}/>
      <div class="search" style=${{ marginBottom: 16 }}><${Icon} name="search" size=${17} style=${{ color: 'var(--fg-subtle)' }}/>
        <input placeholder="Поиск по имени…" value=${query} onInput=${function (e) { setQuery(e.target.value); }}/></div>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="clients" title="Клиентов нет"/></div>`
          : html`<div class="card" style=${{ overflow: 'hidden' }}><table class="table"><thead><tr><th>Имя</th><th>Телефон</th><th>Документ</th></tr></thead>
            <tbody>${list.map(function (c) {
              return html`<tr key=${c.id}>
                <td style=${{ fontWeight: 600 }}>${c.full_name}</td>
                <td class="amana-num" style=${{ color: 'var(--fg-muted)' }}>${c.phone || '—'}</td>
                <td style=${{ color: 'var(--fg-muted)' }}>${c.document || '—'}</td></tr>`;
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
    var m = useState(false), open = m[0], setOpen = m[1];
    var list = st.data || [];
    return html`<div>
      <${PageHead} title="Каталог" sub=${list.length + ' товаров'}
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новый товар</button>`}/>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="catalog" title="Каталог пуст"/></div>`
          : html`<div class="card" style=${{ overflow: 'hidden' }}><table class="table"><thead><tr><th>Название</th><th>Категория</th><th>Закупка</th><th>Статус</th></tr></thead>
            <tbody>${list.map(function (pr) {
              return html`<tr key=${pr.id}>
                <td style=${{ fontWeight: 600 }}>${pr.name}</td>
                <td style=${{ color: 'var(--fg-muted)' }}>${pr.category || '—'}</td>
                <td class="amana-num" style=${{ fontWeight: 600 }}>${fmt.money(pr.cost_price)}</td>
                <td><${ui.StatusChip} map="halal" value=${pr.halal_status}/></td></tr>`;
            })}</tbody></table></div>`}
      <//>
      ${open ? html`<${ProductModal} onClose=${function () { setOpen(false); }} onSaved=${function () { setOpen(false); st.reload(); ctx.toast('Товар добавлен'); }} ctx=${ctx}/>` : null}
    </div>`;
  }
  function ProductModal(p) {
    var f = useState({ name: '', category: '', cost_price: '', halal_status: 'halal' }), v = f[0], set = f[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function save() {
      if (!v.name.trim() || !v.cost_price.trim()) { p.ctx.toast('Заполните название и цену', true); return; }
      setBusy(true);
      api.createProduct({ name: v.name.trim(), category: v.category.trim(), cost_price: v.cost_price.replace(',', '.').trim(), halal_status: v.halal_status })
        .then(p.onSaved).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var inp = function (k, ph) { return html`<input class="input" value=${v[k]} placeholder=${ph} onInput=${function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }}/>`; };
    return html`<${ui.Modal} title="Новый товар" onClose=${p.onClose}>
      <${ui.Field} label="Название">${inp('name', 'Диван угловой')}<//>
      <${ui.Field} label="Категория">${inp('category', 'Мебель')}<//>
      <${ui.Field} label="Закупочная цена, ₽">${inp('cost_price', '85000')}<//>
      <${ui.Field} label="Халяль-статус">
        <select class="select" value=${v.halal_status} onChange=${function (e) { set(Object.assign({}, v, { halal_status: e.target.value })); }}>
          <option value="halal">Халяль</option><option value="doubtful">Сомнительно</option><option value="haram">Харам</option></select>
      <//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${save}>${busy ? html`<${ui.Spinner}/>` : 'Сохранить'}</button>
    <//>`;
  }

  /* ================= CONTRACTS LIST ================= */
  function Contracts(ctx) {
    var st = useAsync(api.listContracts);
    var fl = useState('all'), filter = fl[0], setFilter = fl[1];
    var list = (st.data || []).filter(function (c) { return filter === 'all' || c.status === filter; });
    var tabs = [['all', 'Все'], ['active', 'Активные'], ['completed', 'Завершённые'], ['cancelled', 'Отменённые']];
    return html`<div>
      <${PageHead} title="Договоры" sub="Рассрочка по модели мурабаха"
        actions=${html`<button class="btn btn-primary" onClick=${function () { ctx.go('contract-new'); }}><${Icon} name="plus" size=${17}/> Новый договор</button>`}/>
      <div class="tabs" style=${{ marginBottom: 16, maxWidth: 460 }}>
        ${tabs.map(function (t) { return html`<button key=${t[0]} class=${'tab ' + (filter === t[0] ? 'active' : '')} onClick=${function () { setFilter(t[0]); }}>${t[1]}</button>`; })}
      </div>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="contracts" title="Договоров нет" text="Оформите первый договор рассрочки"
            action=${html`<button class="btn btn-primary" style=${{ marginTop: 14 }} onClick=${function () { ctx.go('contract-new'); }}>Новый договор</button>`}/></div>`
          : html`<div class="card" style=${{ overflow: 'hidden' }}><table class="table"><thead><tr><th>Цена продажи</th><th>Остаток</th><th>Прогресс</th><th>Статус</th></tr></thead>
            <tbody>${list.map(function (c) {
              return html`<tr key=${c.id} class="row-link" onClick=${function () { ctx.go('contract', c.id); }}>
                <td class="amana-num" style=${{ fontWeight: 600 }}>${fmt.money(c.sale_price)}</td>
                <td class="amana-num" style=${{ color: 'var(--fg-muted)' }}>${fmt.money(c.outstanding)}</td>
                <td style=${{ width: 170 }}><div style=${{ display: 'flex', alignItems: 'center', gap: 9 }}>
                  <div class="progress"><i style=${{ width: (c.progress_percent || 0) + '%' }}></i></div>
                  <span class="amana-num" style=${{ fontSize: 12.5, color: 'var(--fg-subtle)' }}>${Math.round(c.progress_percent || 0)}%</span></div></td>
                <td style=${{ width: 120 }}><${ui.StatusChip} map="contractStatus" value=${c.status}/></td></tr>`;
            })}</tbody></table></div>`}
      <//>
    </div>`;
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
          return html`<div key=${pr.id} class=${'select-card ' + (p.w.productId === pr.id ? 'sel ' : '') + (haram ? 'disabled' : '')}
            onClick=${function () { if (!haram) p.set({ productId: pr.id }); }}>
            <div style=${{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
              <div style=${{ fontWeight: 600 }}>${pr.name}</div><${ui.StatusChip} map="halal" value=${pr.halal_status}/></div>
            <div class="amana-num" style=${{ fontSize: 13, color: 'var(--fg-muted)', marginTop: 4 }}>${fmt.money(pr.cost_price)}</div>
            ${haram ? html`<div style=${{ fontSize: 12, color: 'var(--haram-fg)', marginTop: 4 }}>Договор на «харам» оформить нельзя</div>` : null}</div>`;
        })}
      </div>
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
    var sdS = useState(false), sdOpen = sdS[0], setSdOpen = sdS[1];
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
          ${c.has_overdue ? html`<div class="banner banner-warn" style=${{ marginBottom: 16, justifyContent: 'space-between', alignItems: 'center' }}>
            <div style=${{ display: 'flex', gap: 10 }}><${Icon} name="clock" size=${18}/><div><b>Есть просрочка.</b> Долг не растёт — можно начислить фиксированную садаку.</div></div>
            ${ctx.isOwner ? html`<button class="btn btn-sm" style=${{ background: 'var(--st-over-fg)', color: '#fff' }} onClick=${function () { setSdOpen(true); }}>Начислить садаку</button>` : null}
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
      ${sdOpen ? html`<${SadaqaModal} c=${c} onClose=${function () { setSdOpen(false); }} onDone=${function () { setSdOpen(false); st.reload(); ctx.toast('Садака начислена'); }} ctx=${ctx}/>` : null}
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
  function SadaqaModal(p) {
    var a = useState('500'), amount = a[0], setAmount = a[1];
    var nt = useState('Просрочка платежа'), note = nt[0], setNote = nt[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    function go() {
      var n = parseFloat(String(amount).replace(',', '.'));
      if (!(n > 0)) { p.ctx.toast('Введите сумму', true); return; }
      setBusy(true);
      api.accrueCharity(p.c.id, String(amount).replace(',', '.'), note.trim()).then(p.onDone).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    return html`<${ui.Modal} title="Начислить садаку" onClose=${p.onClose}>
      <div class="banner banner-accent" style=${{ fontSize: 13 }}><${Icon} name="charity" size=${17}/> Фиксированный сбор уходит в реестр благотворительности и <b>не меняет долг</b>.</div>
      <${ui.Field} label="Сумма, ₽"><input class="input" value=${amount} onInput=${function (e) { setAmount(e.target.value); }}/><//>
      <${ui.Field} label="Комментарий"><input class="input" value=${note} onInput=${function (e) { setNote(e.target.value); }}/><//>
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${go}>${busy ? html`<${ui.Spinner}/>` : 'Начислить'}</button>
    <//>`;
  }

  /* ================= SCHEDULE ================= */
  function Schedule(ctx) {
    var st = useAsync(api.listReminders);
    var m = useState(false), open = m[0], setOpen = m[1];
    var list = st.data || [];
    return html`<div>
      <${PageHead} title="Календарь" sub="Задачи мимо времён намаза"
        actions=${html`<button class="btn btn-primary" onClick=${function () { setOpen(true); }}><${Icon} name="plus" size=${17}/> Новая задача</button>`}/>
      <${Guard} loading=${st.loading} err=${st.err}>
        ${list.length === 0 ? html`<div class="card"><${ui.Empty} icon="calendar" title="Задач пока нет" text="Создайте звонок или доставку — система обойдёт окна намаза"/></div>`
          : html`<div class="grid" style=${{ gap: 12 }}>${list.map(function (r) {
            return html`<div key=${r.id} class="card card-pad" style=${{ display: 'flex', alignItems: 'center', gap: 14 }}>
              <span style=${{ width: 42, height: 42, borderRadius: 12, background: 'var(--grad-soft)', color: 'var(--accent)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <${Icon} name=${r.type === 'delivery' ? 'truck' : r.type === 'call' ? 'phone' : 'coins'} size=${20}/></span>
              <div style=${{ flex: 1 }}>
                <div style=${{ fontWeight: 600 }}>${ui.labels.reminderType[r.type] || r.type}${r.note ? html`<span style=${{ color: 'var(--fg-subtle)', fontWeight: 400 }}> · ${r.note}</span>` : null}</div>
                <div style=${{ fontSize: 13, color: 'var(--fg-muted)' }}>${fmt.dateTime(r.scheduled_at)}</div>
              </div>
              ${r.was_shifted ? html`<div style=${{ textAlign: 'right' }}><span class="chip chip-partially_paid">Перенесено</span>
                <div style=${{ fontSize: 11.5, color: 'var(--st-part-fg)', marginTop: 4, maxWidth: 240 }}>${r.reason}</div></div>`
                : html`<span class="chip chip-paid">Вовремя</span>`}
            </div>`;
          })}</div>`}
      <//>
      ${open ? html`<${ReminderModal} ctx=${ctx} onClose=${function () { setOpen(false); }} onDone=${function () { setOpen(false); st.reload(); ctx.toast('Задача создана'); }}/>` : null}
    </div>`;
  }
  function ReminderModal(p) {
    var clients = useAsync(api.listClients);
    var f = useState({ type: 'call', client_id: '', note: '', date: defaultStart(), time: '13:00', duration_minutes: 20 }), v = f[0], set = f[1];
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
    function create() {
      setBusy(true);
      api.createReminder({ type: v.type, client_id: v.client_id || undefined, note: v.note.trim(),
        desired_at: iso(), duration_minutes: Number(v.duration_minutes) }).then(p.onDone).catch(function (e) { setBusy(false); p.ctx.toast(e.message, true); });
    }
    var upd = function (k) { return function (e) { var o = {}; o[k] = e.target.value; set(Object.assign({}, v, o)); }; };
    return html`<${ui.Modal} title="Новая задача" onClose=${p.onClose} width=${460}>
      <${ui.Field} label="Тип">
        <select class="select" value=${v.type} onChange=${upd('type')}>
          <option value="call">Звонок</option><option value="delivery">Доставка</option><option value="payment_followup">Контакт по платежу</option></select>
      <//>
      <${ui.Field} label="Клиент (необязательно)">
        <select class="select" value=${v.client_id} onChange=${upd('client_id')}>
          <option value="">— не выбран —</option>
          ${(clients.data || []).map(function (c) { return html`<option key=${c.id} value=${c.id}>${c.full_name}</option>`; })}</select>
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
      <button class="btn btn-primary btn-block" disabled=${busy} onClick=${create}>${busy ? html`<${ui.Spinner}/>` : 'Создать задачу'}</button>
    <//>`;
  }

  /* ================= CHARITY ================= */
  function Charity(ctx) {
    var st = useAsync(api.listCharity);
    var d = st.data || { entries: [], total_amount: '0' };
    return html`<div>
      <${PageHead} title="Реестр садаки" sub="Прозрачность благотворительных сборов"/>
      <${Guard} loading=${st.loading} err=${st.err}>
        <div class="card card-pad" style=${{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div class="kpi"><div class="v amana-num grad-text">${fmt.money(d.total_amount)}</div><div class="l">Всего собрано садаки</div></div>
          <span style=${{ width: 48, height: 48, borderRadius: 14, background: 'var(--sd-bg)', color: 'var(--sd-fg)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><${Icon} name="charity" size=${24}/></span>
        </div>
        ${d.entries.length === 0 ? html`<div class="card"><${ui.Empty} icon="charity" title="Записей нет" text="Сборы появятся при начислении садаки по просрочке"/></div>`
          : html`<div class="card" style=${{ overflow: 'hidden' }}><table class="table"><thead><tr><th>Дата</th><th>Договор</th><th>Сумма</th><th>Статус</th></tr></thead>
            <tbody>${d.entries.map(function (e) {
              return html`<tr key=${e.id}>
                <td>${fmt.date(e.created_at)}</td>
                <td style=${{ fontFamily: 'monospace', fontSize: 12.5, color: 'var(--fg-muted)' }}>${e.contract_id.slice(0, 8)}</td>
                <td class="amana-num" style=${{ fontWeight: 600 }}>${fmt.money(e.amount)}</td>
                <td><span class="chip chip-sd">${ui.labels.charityStatus[e.status] || e.status}</span></td></tr>`;
            })}</tbody></table></div>`}
      <//>
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
    var u = me.data || {};
    return html`<div>
      <${PageHead} title="Настройки"/>
      <${Guard} loading=${me.loading} err=${me.err}>
        <div class="grid" style=${{ gridTemplateColumns: '1fr 1fr' }}>
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

  AM.screens = { dashboard: Dashboard, clients: Clients, catalog: Catalog, contracts: Contracts,
    'contract-new': ContractWizard, contract: ContractCard, schedule: Schedule, charity: Charity,
    developers: Developers, settings: Settings };
})();
