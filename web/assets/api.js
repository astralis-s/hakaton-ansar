/* Amana API client — talks to the Go backend on the same origin.
   Internal app uses JWT (Bearer); the public API uses X-API-Key (not needed here). */
(function () {
  window.AM = window.AM || {};
  var TOKEN_KEY = 'amana.token';

  function getToken() { try { return localStorage.getItem(TOKEN_KEY) || ''; } catch (e) { return ''; } }
  function setToken(t) { try { t ? localStorage.setItem(TOKEN_KEY, t) : localStorage.removeItem(TOKEN_KEY); } catch (e) {} }

  async function request(method, path, body) {
    var headers = { 'Content-Type': 'application/json' };
    var tok = getToken();
    if (tok) headers['Authorization'] = 'Bearer ' + tok;
    var res;
    try {
      res = await fetch('/api/app' + path, {
        method: method,
        headers: headers,
        body: body !== undefined ? JSON.stringify(body) : undefined,
      });
    } catch (e) {
      var net = new Error('Сеть недоступна. Проверьте подключение.');
      net.code = 'network'; net.status = 0; throw net;
    }
    if (res.status === 204) return null;
    var data = null;
    try { data = await res.json(); } catch (e) {}
    if (!res.ok) {
      var msg = (data && data.error && data.error.message) || ('Ошибка ' + res.status);
      var err = new Error(msg);
      err.code = (data && data.error && data.error.code) || 'error';
      err.status = res.status;
      throw err;
    }
    return data;
  }

  // download fetches a binary (PDF) with the right auth header and triggers a
  // browser download (a plain <a href> can't send the Authorization header).
  async function download(base, path, tokenGetter, filename) {
    var headers = {};
    var tok = tokenGetter();
    if (tok) headers['Authorization'] = 'Bearer ' + tok;
    var res;
    try { res = await fetch(base + path, { headers: headers }); }
    catch (e) { var n = new Error('Сеть недоступна'); n.status = 0; throw n; }
    if (!res.ok) {
      var msg = 'Ошибка ' + res.status;
      try { var d = await res.json(); msg = (d.error && d.error.message) || msg; } catch (e) {}
      var err = new Error(msg); err.status = res.status; throw err;
    }
    var blob = await res.blob();
    var url = URL.createObjectURL(blob);
    var a = document.createElement('a');
    a.href = url; a.download = filename || 'document.pdf';
    document.body.appendChild(a); a.click(); a.remove();
    setTimeout(function () { URL.revokeObjectURL(url); }, 1500);
  }

  var api = {
    getToken: getToken,
    setToken: setToken,
    isAuthed: function () { return !!getToken(); },
    logout: function () { setToken(''); },

    setup: function (p) { return request('POST', '/setup', p); },
    register: function (p) { return request('POST', '/auth/register', p); },
    login: async function (email, password) {
      var r = await request('POST', '/auth/login', { email: email, password: password });
      if (r && r.token) setToken(r.token);
      return r;
    },
    me: function () { return request('GET', '/auth/me'); },

    listClients: function () { return request('GET', '/clients'); },
    getClient: function (id) { return request('GET', '/clients/' + id); },
    createClient: function (p) { return request('POST', '/clients', p); },
    updateClient: function (id, p) { return request('PUT', '/clients/' + id, p); },

    listProducts: function () { return request('GET', '/catalog'); },
    getProduct: function (id) { return request('GET', '/catalog/' + id); },
    createProduct: function (p) { return request('POST', '/catalog', p); },
    updateProduct: function (id, p) { return request('PUT', '/catalog/' + id, p); },
    adjustStock: function (id, p) { return request('POST', '/catalog/' + id + '/stock', p); },
    listStockMovements: function () { return request('GET', '/catalog/movements'); },

    dashboard: function () { return request('GET', '/dashboard'); },
    listContracts: function () { return request('GET', '/contracts'); },
    getContract: function (id) { return request('GET', '/contracts/' + id); },
    previewContract: function (p) { return request('POST', '/contracts/preview', p); },
    createContract: function (p) { return request('POST', '/contracts', p); },
    registerPayment: function (id, amount) { return request('POST', '/contracts/' + id + '/payments', { amount: amount }); },
    settleContract: function (id) { return request('POST', '/contracts/' + id + '/settle'); },
    cancelContract: function (id) { return request('POST', '/contracts/' + id + '/cancel'); },
    downloadContractPdf: function (id) { return download('/api/app', '/contracts/' + id + '/pdf', getToken, 'dogovor-' + id.slice(0, 8) + '.pdf'); },

    listChats: function () { return request('GET', '/chats'); },
    chatThread: function (clientId) { return request('GET', '/chats/' + clientId + '/messages'); },
    sendChatMessage: function (clientId, body) { return request('POST', '/chats/' + clientId + '/messages', { body: body }); },
    telegramLink: function () { return request('GET', '/chats/telegram-link'); },

    financeReport: function () { return request('GET', '/finance/report'); },
    downloadFinanceReportPdf: function () { return download('/api/app', '/finance/report.pdf', getToken, 'finansy-otchet.pdf'); },
    listExpenses: function () { return request('GET', '/finance/expenses'); },
    createExpense: function (p) { return request('POST', '/finance/expenses', p); },
    deleteExpense: function (id) { return request('DELETE', '/finance/expenses/' + id); },

    listReminders: function () { return request('GET', '/schedule/reminders'); },
    getReminder: function (id) { return request('GET', '/schedule/reminders/' + id); },
    createReminder: function (p) { return request('POST', '/schedule/reminders', p); },
    updateReminder: function (id, p) { return request('PUT', '/schedule/reminders/' + id, p); },
    completeReminder: function (id) { return request('POST', '/schedule/reminders/' + id + '/complete'); },
    cancelReminder: function (id) { return request('POST', '/schedule/reminders/' + id + '/cancel'); },
    previewSlot: function (p) { return request('POST', '/schedule/preview', p); },

    listApiKeys: function () { return request('GET', '/api-keys'); },
    createApiKey: function (name) { return request('POST', '/api-keys', { name: name }); },
    revokeApiKey: function (id) { return request('DELETE', '/api-keys/' + id); },

    listUsers: function () { return request('GET', '/users'); },
    createUser: function (p) { return request('POST', '/users', p); },
  };

  /* ---------- formatting ---------- */
  var money = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', minimumFractionDigits: 2, maximumFractionDigits: 2 });
  var num = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  var fmt = {
    money: function (s) { var n = parseFloat(s); return isNaN(n) ? '—' : money.format(n); },
    num: function (s) { var n = parseFloat(s); return isNaN(n) ? '—' : num.format(n); },
    date: function (iso) {
      if (!iso) return '—';
      var d = new Date(iso.length === 10 ? iso + 'T00:00:00' : iso);
      return isNaN(d) ? '—' : d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
    },
    dateLong: function (iso) {
      if (!iso) return '—';
      var d = new Date(iso.length === 10 ? iso + 'T00:00:00' : iso);
      return isNaN(d) ? '—' : d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
    },
    time: function (iso) { var d = new Date(iso); return isNaN(d) ? '—' : d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }); },
    dateTime: function (iso) { return fmt.date(iso) + ', ' + fmt.time(iso); },
  };

  AM.api = api;
  AM.fmt = fmt;
})();
