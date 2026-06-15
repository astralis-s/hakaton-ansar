/* Amana landing page — original implementation. Sells the product and lets the
   visitor register (first-run org setup) or log in. */
(function () {
  window.AM = window.AM || {};
  var React = window.React;
  var ui = window.AM.ui, html = ui.html, Icon = ui.Icon;
  var useState = React.useState, useEffect = React.useEffect, useRef = React.useRef;

  /* scroll-reveal + cursor/parallax effects */
  function useEffects() {
    useEffect(function () {
      var io = new IntersectionObserver(function (entries) {
        entries.forEach(function (e) { if (e.isIntersecting) { e.target.classList.add('in'); io.unobserve(e.target); } });
      }, { threshold: 0.15 });
      document.querySelectorAll('.reveal').forEach(function (el) { io.observe(el); });

      var glow = document.getElementById('cursorGlow');
      var raf = 0, mx = 0, my = 0;
      function onMove(e) {
        mx = e.clientX; my = e.clientY;
        if (glow) { glow.style.left = mx + 'px'; glow.style.top = my + 'px'; glow.style.opacity = '1'; }
        var cx = window.innerWidth / 2, cy = window.innerHeight / 2;
        var dx = (mx - cx) / cx, dy = (my - cy) / cy;
        if (!raf) raf = requestAnimationFrame(function () {
          raf = 0;
          document.querySelectorAll('[data-px]').forEach(function (el) {
            var d = parseFloat(el.getAttribute('data-px')) || 0;
            el.style.transform = 'translate3d(' + (dx * d * 26).toFixed(1) + 'px,' + (dy * d * 26).toFixed(1) + 'px,0)';
          });
        });
      }
      window.addEventListener('mousemove', onMove);
      return function () { io.disconnect(); window.removeEventListener('mousemove', onMove); if (raf) cancelAnimationFrame(raf); };
    }, []);
  }

  function CompareChart() {
    var months = 12, W = 460, H = 220, pad = 14;
    var sale = 100; // illustrative: financed=100 paid off linearly
    var pts = [];
    for (var i = 0; i <= months; i++) pts.push(i);
    var x = function (i) { return pad + (i / months) * (W - pad * 2); };
    var yMax = 170;
    var y = function (v) { return H - pad - (v / yMax) * (H - pad * 2); };
    // murabaha: outstanding falls linearly to 0 (debt never grows)
    var mur = pts.map(function (i) { return sale * (1 - i / months); });
    // conventional credit: balance grows with interest then ends higher
    var credit = pts.map(function (i) { var t = i / months; return sale * (1 - t) + sale * 0.42 * Math.sin(t * Math.PI) * (1 - t * 0.2); });
    var line = function (arr) { return arr.map(function (v, i) { return (i ? 'L' : 'M') + x(i).toFixed(1) + ' ' + y(v).toFixed(1); }).join(' '); };
    var area = function (arr) { return line(arr) + ' L' + x(months) + ' ' + y(0) + ' L' + x(0) + ' ' + y(0) + ' Z'; };
    return html`<svg viewBox=${'0 0 ' + W + ' ' + H} style=${{ width: '100%', height: 'auto', display: 'block' }}>
      <path d=${area(credit)} fill="var(--haram-bg)" opacity="0.7"></path>
      <path d=${line(credit)} fill="none" stroke="var(--haram-fg)" stroke-width="2.5" stroke-dasharray="6 5"></path>
      <path d=${line(mur)} fill="none" stroke="var(--accent)" stroke-width="3"
        style=${{ strokeDasharray: 900, '--len': 900, animation: 'amDraw 1.6s ease forwards' }}></path>
      <circle cx=${x(0)} cy=${y(sale)} r="4" fill="var(--accent)"></circle>
      <circle cx=${x(months)} cy=${y(0)} r="4" fill="var(--accent)"></circle>
    </svg>`;
  }

  function FloatCard() {
    return html`<div class="hero-visual" data-px="0.6">
      <div class="float-card" style=${{ animation: 'amRise .9s cubic-bezier(.2,.8,.3,1) both .2s' }}>
        <div style=${{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <div>
            <div style=${{ fontSize: 12.5, color: 'var(--fg-subtle)' }}>Договор A-1042</div>
            <div style=${{ fontSize: 22, fontWeight: 700, letterSpacing: '-.02em' }} class="amana-num">120 000,00 ₽</div>
          </div>
          <span class="chip chip-active">Активен</span>
        </div>
        ${[['Оплачено', 75, 'var(--grad)'], ['Остаток', 25, 'var(--surface-3)']].map(function (r, i) {
          return html`<div key=${i} style=${{ marginBottom: 12 }}>
            <div style=${{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5, color: 'var(--fg-muted)', marginBottom: 5 }}>
              <span>${r[0]}</span><span class="amana-num">${r[1]}%</span></div>
            <div class="mini-bar"><i style=${{ width: r[1] + '%', background: r[2], animation: 'amBar 1s ease both ' + (0.3 + i * 0.1) + 's' }}></i></div>
          </div>`;
        })}
        <div style=${{ marginTop: 14, padding: '11px 13px', borderRadius: 12, background: 'var(--accent-soft)', border: '1px solid var(--accent-bd)', display: 'flex', alignItems: 'center', gap: 9 }}>
          <${Icon} name="shield" size=${17} style=${{ color: 'var(--accent)' }}/>
          <span style=${{ fontSize: 13, fontWeight: 600, color: 'var(--accent)' }}>0% риба · долг не растёт</span>
        </div>
      </div>
      <div class="float-card" data-px="1.1" style=${{ position: 'absolute', right: -18, bottom: -26, width: 220, padding: 16, animation: 'amRise 1s cubic-bezier(.2,.8,.3,1) both .5s' }}>
        <div style=${{ display: 'flex', alignItems: 'center', gap: 9, marginBottom: 8 }}>
          <span style=${{ width: 32, height: 32, borderRadius: 10, background: 'var(--sd-bg)', color: 'var(--sd-fg)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><${Icon} name="clock" size=${16}/></span>
          <span style=${{ fontSize: 13, fontWeight: 700 }}>Доставка</span>
        </div>
        <div style=${{ fontSize: 12.5, color: 'var(--fg-muted)', lineHeight: 1.5 }}>Перенесено мимо <b style=${{ color: 'var(--fg)' }}>Магриба</b> на 20:25</div>
      </div>
    </div>`;
  }

  var FEATURES = [
    { ico: 'shield', t: 'Мурабаха без рибы', d: 'Цена фиксируется один раз как «себестоимость + наценка». Долг не растёт со временем — никакого процента и капитализации.' },
    { ico: 'moon2', t: 'Планировщик с намазом', d: 'Звонки и доставки автоматически сдвигаются мимо пяти молитв и пятничного джума. Система объясняет, почему перенесла.' },
    { ico: 'charity', t: 'Садака вместо пени', d: 'Штраф за просрочку — фиксированная садака в отдельный реестр благотворительности. Не в выручку и не в долг клиента.' },
    { ico: 'star', t: 'Халяль-статус товара', d: 'Каждому товару присвоен статус: халяль, харам или сомнительно. На «харам» договор оформить нельзя.' },
    { ico: 'code', t: 'Публичный API', d: 'Внешние маркетплейсы создают договоры и читают статусы платежей через API-ключ. Документация — в Swagger.' },
    { ico: 'trend', t: 'Прозрачный график', d: 'Равные доли с детерминированным округлением до копейки. Сравнение «без рибы vs кредит» прямо в мастере договора.' },
  ];
  var STEPS = [
    { n: '01', t: 'Заводите товар и клиента', d: 'Каталог с халяль-статусом и карточки клиентов с документом для договора.' },
    { n: '02', t: 'Оформляете договор', d: 'Мастер считает график мурабахи на сервере и показывает предпросмотр до сохранения.' },
    { n: '03', t: 'Принимаете платежи', d: 'Любая сумма в пределах остатка. Статусы долей пересчитываются автоматически.' },
    { n: '04', t: 'Ведёте задачи мимо намаза', d: 'Календарь подсвечивает окна молитв; задачи сдвигаются на свободный слот.' },
  ];

  function AuthModal(p) {
    var s = useState('login'), mode = s[0], setMode = s[1];
    var l = useState({ email: '', password: '' }), login = l[0], setLogin = l[1];
    var r = useState({ org_name: '', owner_name: '', owner_email: '', owner_password: '' }), reg = r[0], setReg = r[1];
    var b = useState(false), busy = b[0], setBusy = b[1];
    var e = useState(''), err = e[0], setErr = e[1];

    function submit() {
      setErr(''); setBusy(true);
      var done = function () { setBusy(false); p.onAuthed(); };
      var fail = function (ex) {
        setBusy(false);
        if (ex.code === 'already_initialized') { setErr('Организация уже создана — войдите в систему.'); setMode('login'); return; }
        setErr(ex.message || 'Не удалось выполнить запрос.');
      };
      if (mode === 'login') {
        AM.api.login(login.email.trim(), login.password).then(done).catch(fail);
      } else {
        AM.api.setup({
          org_name: reg.org_name.trim(), currency: 'RUB',
          owner_name: reg.owner_name.trim(), owner_email: reg.owner_email.trim(), owner_password: reg.owner_password,
        }).then(function () { return AM.api.login(reg.owner_email.trim(), reg.owner_password); }).then(done).catch(fail);
      }
    }
    var input = function (val, on, ph, type) {
      return html`<input class="input" type=${type || 'text'} value=${val} placeholder=${ph}
        onInput=${function (ev) { on(ev.target.value); }} onKeyDown=${function (ev) { if (ev.key === 'Enter') submit(); }}/>`;
    };
    return html`<${ui.Modal} title=${mode === 'login' ? 'Вход в систему' : 'Создать организацию'} onClose=${p.onClose} width=${440}>
      <div class="tabs" style=${{ marginBottom: 4 }}>
        <button class=${'tab ' + (mode === 'login' ? 'active' : '')} onClick=${function () { setErr(''); setMode('login'); }}>Войти</button>
        <button class=${'tab ' + (mode === 'register' ? 'active' : '')} onClick=${function () { setErr(''); setMode('register'); }}>Регистрация</button>
      </div>
      ${mode === 'login'
        ? html`<${ui.Field} label="Email">${input(login.email, function (v) { setLogin(Object.assign({}, login, { email: v })); }, 'owner@amana.ru')}<//>
            <${ui.Field} label="Пароль">${input(login.password, function (v) { setLogin(Object.assign({}, login, { password: v })); }, '••••••••', 'password')}<//>`
        : html`<${ui.Field} label="Название организации">${input(reg.org_name, function (v) { setReg(Object.assign({}, reg, { org_name: v })); }, 'Грозный Мебель')}<//>
            <${ui.Field} label="Ваше имя (владелец)">${input(reg.owner_name, function (v) { setReg(Object.assign({}, reg, { owner_name: v })); }, 'Адам Магомадов')}<//>
            <${ui.Field} label="Email">${input(reg.owner_email, function (v) { setReg(Object.assign({}, reg, { owner_email: v })); }, 'owner@amana.ru')}<//>
            <${ui.Field} label="Пароль" hint="Минимум 8 символов">${input(reg.owner_password, function (v) { setReg(Object.assign({}, reg, { owner_password: v })); }, '••••••••', 'password')}<//>`}
      ${err ? html`<div class="banner banner-warn" style=${{ fontSize: 13.5 }}>${err}</div>` : null}
      <button class="btn btn-primary btn-block btn-lg" disabled=${busy} onClick=${submit}>
        ${busy ? html`<${ui.Spinner}/>` : (mode === 'login' ? 'Войти' : 'Создать и войти')}
      </button>
      <div style=${{ fontSize: 12.5, color: 'var(--fg-subtle)', textAlign: 'center' }}>
        Демо-доступ: <b>owner@amana.ru</b> / <b>owner12345</b>
      </div>
    <//>`;
  }

  function Landing(p) {
    useEffects();
    var a = useState(false), authOpen = a[0], setAuthOpen = a[1];
    var open = function () { setAuthOpen(true); };
    var scrollTo = function (sel) { var el = document.querySelector(sel); if (el) window.scrollTo({ top: el.getBoundingClientRect().top + window.scrollY - 20, behavior: 'smooth' }); };

    return html`<div class="lp">
      <div class="lp-bg"><div class="lp-blob b1"></div><div class="lp-blob b2"></div><div class="lp-blob b3"></div></div>
      <div class="lp-grain"></div>
      <div class="cursor-glow" id="cursorGlow"></div>

      <header class="lp-header">
        <div style=${{ display: 'flex', alignItems: 'center', gap: 11 }}>
          <span class="brand-badge"><${Icon} name="logo" size=${20} sw=${1.9}/></span>
          <span style=${{ fontSize: 20, fontWeight: 700, letterSpacing: '-.03em' }}>Амана</span>
        </div>
        <div style=${{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <button class="icon-btn" onClick=${p.toggleTheme} title="Тема" style=${{ border: '1px solid var(--glass-bd)', background: 'var(--glass-2)' }}>
            <${Icon} name=${p.theme === 'dark' ? 'sun' : 'moon'} size=${18}/></button>
          <button class="btn btn-primary" onClick=${open}>Войти</button>
        </div>
      </header>

      <section class="lp-section hero">
        <div data-px="0.25">
          <div class="eyebrow"><span class="dot"></span>БЕЗ РИБЫ · МОДЕЛЬ МУРАБАХА</div>
          <h1>Рассрочка без процента — <span class="grad-text">честно</span> и прозрачно</h1>
          <p class="lead">CRM для продаж в рассрочку по исламской модели: фиксированная наценка, понятный график платежей,
            планирование задач мимо времён намаза и прозрачный реестр садаки. Долг <b style=${{ color: 'var(--fg)' }}>не растёт</b> со временем.</p>
          <div style=${{ display: 'flex', gap: 14, flexWrap: 'wrap' }}>
            <button class="btn btn-primary btn-lg" onClick=${open}>Начать бесплатно <${Icon} name="arrow" size=${19}/></button>
            <button class="btn btn-ghost btn-lg" onClick=${function () { scrollTo('#how'); }}>Как это работает</button>
          </div>
          <div class="hero-stats">
            <div class="hero-stat"><div class="v grad-text">0%</div><div class="l">риба / процент</div></div>
            <div class="vrule"></div>
            <div class="hero-stat"><div class="v">5×</div><div class="l">молитв учтены</div></div>
            <div class="vrule"></div>
            <div class="hero-stat"><div class="v">100%</div><div class="l">копейки сходятся</div></div>
          </div>
        </div>
        <${FloatCard}/>
      </section>

      <section class="lp-section" id="how" style=${{ padding: '70px 6vw 10px' }}>
        <div class="reveal" style=${{ textAlign: 'center', maxWidth: 640, margin: '0 auto' }}>
          <div class="section-kicker">Возможности</div>
          <h2 class="section-title" style=${{ marginTop: 10 }}>Финансовая модель, которой нет у обычных CRM</h2>
          <p style=${{ color: 'var(--fg-muted)', marginTop: 14, fontSize: 17 }}>Мы переписали саму логику рассрочки под нормы ислама — от фиксированной цены до планирования задач.</p>
        </div>
        <div class="feature-grid">
          ${FEATURES.map(function (f, i) {
            return html`<div key=${i} class=${'feature reveal d' + (i % 3 + 1)}>
              <div class="ico"><${Icon} name=${f.ico} size=${24}/></div>
              <h3>${f.t}</h3><p>${f.d}</p></div>`;
          })}
        </div>
      </section>

      <section class="lp-section" style=${{ padding: '70px 6vw 10px' }}>
        <div class="compare-wrap">
          <div class="reveal">
            <div class="section-kicker">Анти-риба</div>
            <h2 class="section-title" style=${{ marginTop: 10 }}>Долг не растёт. Точка.</h2>
            <p style=${{ color: 'var(--fg-muted)', margin: '14px 0 22px', fontSize: 16.5, lineHeight: 1.6 }}>
              В обычном кредите остаток капитализируется процентом — чем дольше платите, тем больше переплата.
              В мурабахе сумма обязательства фиксируется при создании и <b style=${{ color: 'var(--fg)' }}>не зависит от времени</b>.</p>
            <div style=${{ display: 'flex', gap: 22, flexWrap: 'wrap' }}>
              <div class="legend"><span class="sw" style=${{ background: 'var(--accent)' }}></span>Мурабаха (фиксировано)</div>
              <div class="legend"><span class="sw" style=${{ background: 'var(--haram-fg)' }}></span>Кредит с процентом</div>
            </div>
          </div>
          <div class="reveal d2 card card-pad"><${CompareChart}/></div>
        </div>
      </section>

      <section class="lp-section" style=${{ padding: '70px 6vw 10px' }}>
        <div class="reveal" style=${{ textAlign: 'center', maxWidth: 600, margin: '0 auto' }}>
          <div class="section-kicker">Как это работает</div>
          <h2 class="section-title" style=${{ marginTop: 10 }}>Четыре шага до первого договора</h2>
        </div>
        <div class="steps-row">
          ${STEPS.map(function (st, i) {
            return html`<div key=${i} class=${'step-card reveal d' + (i % 4 + 1)}><div class="n">${st.n}</div><h4>${st.t}</h4><p>${st.d}</p></div>`;
          })}
        </div>
      </section>

      <section class="lp-section">
        <div class="cta-band reveal">
          <div class="cta-sheen"></div>
          <h2>Честная рассрочка начинается здесь</h2>
          <p>Создайте организацию за минуту и оформите первый договор мурабахи уже сегодня.</p>
          <button class="btn btn-lg" style=${{ background: '#fff', color: 'var(--accent)', fontWeight: 700 }} onClick=${open}>Создать организацию <${Icon} name="arrow" size=${19}/></button>
        </div>
      </section>

      <footer class="lp-footer">
        <div style=${{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span class="brand-badge" style=${{ width: 30, height: 30 }}><${Icon} name="logo" size=${17}/></span>
          <span style=${{ fontWeight: 700 }}>Амана</span>
          <span style=${{ color: 'var(--fg-subtle)' }}>· CRM честной рассрочки</span>
        </div>
        <div style=${{ display: 'flex', gap: 18, fontSize: 14 }}>
          <a href="/swagger/" target="_blank">API · Swagger</a>
          <span style=${{ color: 'var(--fg-subtle)' }}>Грозный, ЧР</span>
        </div>
      </footer>

      ${authOpen ? html`<${AuthModal} onClose=${function () { setAuthOpen(false); }} onAuthed=${p.onAuthed}/>` : null}
    </div>`;
  }

  AM.Landing = Landing;
})();
