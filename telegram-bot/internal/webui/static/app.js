'use strict';

// Веб-инбокс менеджера: вход по паролю, опрос списка чатов и активной переписки,
// отправка ответов. Все данные клиента вставляются через textContent (без innerHTML),
// чтобы исключить XSS из текста сообщений.

const TOKEN_KEY = 'amana_mgr_token';
let token = localStorage.getItem(TOKEN_KEY) || '';
let authRequired = false;
let activeChatId = null;
let activeName = '';
let lastThreadCount = -1;
let inboxTimer = null;
let threadTimer = null;

const $ = (id) => document.getElementById(id);

// ---- HTTP ------------------------------------------------------------------

async function api(path, opts = {}) {
  const headers = Object.assign({}, opts.headers || {});
  if (token) headers['Authorization'] = 'Bearer ' + token;
  if (opts.body) headers['Content-Type'] = 'application/json';
  const res = await fetch(path, Object.assign({}, opts, { headers }));
  if (res.status === 401) {
    logout();
    throw new Error('unauthorized');
  }
  return res;
}

// ---- вход ------------------------------------------------------------------

function showLogin() {
  stopPolling();
  $('app').classList.add('hidden');
  $('login').classList.remove('hidden');
  $('login-password').focus();
}

function logout() {
  token = '';
  localStorage.removeItem(TOKEN_KEY);
  showLogin();
}

$('login-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  $('login-error').textContent = '';
  try {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: $('login-password').value }),
    });
    if (!res.ok) {
      $('login-error').textContent = 'Неверный пароль';
      return;
    }
    const data = await res.json();
    token = data.token || '';
    if (token) localStorage.setItem(TOKEN_KEY, token);
    startApp();
  } catch {
    $('login-error').textContent = 'Сервер недоступен';
  }
});

// ---- инициализация ---------------------------------------------------------

async function init() {
  try {
    const cfg = await (await fetch('/api/config')).json();
    authRequired = !!cfg.auth_required;
  } catch {
    authRequired = false;
  }
  if (authRequired && !token) {
    showLogin();
    return;
  }
  startApp();
}

function startApp() {
  $('login').classList.add('hidden');
  $('app').classList.remove('hidden');
  loadInbox();
  inboxTimer = setInterval(loadInbox, 3000);
}

function stopPolling() {
  clearInterval(inboxTimer);
  clearInterval(threadTimer);
  inboxTimer = threadTimer = null;
}

// ---- список чатов ----------------------------------------------------------

async function loadInbox() {
  try {
    const res = await api('/api/chats');
    const chats = await res.json();
    setConn(true);
    renderInbox(chats || []);
  } catch (e) {
    if (e.message !== 'unauthorized') setConn(false);
  }
}

function renderInbox(chats) {
  const box = $('chats');
  box.textContent = '';
  if (!chats.length) {
    const empty = el('div', 'chats-empty', 'Пока нет ни одного заказчика. Они появятся, как только напишут боту.');
    box.appendChild(empty);
    return;
  }
  for (const c of chats) {
    const item = el('div', 'chat-item');
    if (c.chat_id === activeChatId) item.classList.add('active');
    item.appendChild(el('div', 'avatar', initials(c.full_name)));

    const body = el('div', 'chat-body');
    const row = el('div', 'chat-row');
    row.appendChild(el('div', 'chat-name', c.full_name));
    row.appendChild(el('div', 'chat-time', formatTime(c.last_at)));
    body.appendChild(row);

    const preview = el('div', 'chat-preview');
    if (c.last_sender === 'manager' && c.last_text) {
      preview.appendChild(el('span', 'you', 'Вы: '));
    }
    preview.appendChild(document.createTextNode(c.last_text || c.phone || ''));
    body.appendChild(preview);
    item.appendChild(body);

    if (c.unread > 0) item.appendChild(el('div', 'badge', String(c.unread)));

    item.addEventListener('click', () => openChat(c.chat_id, c.full_name, c.phone));
    box.appendChild(item);
  }
}

// ---- переписка -------------------------------------------------------------

async function openChat(id, name, phone) {
  const switching = id !== activeChatId;
  activeChatId = id;
  activeName = name;
  if (switching) lastThreadCount = -1;

  $('empty').classList.add('hidden');
  $('conversation').classList.remove('hidden');
  document.querySelector('.app').classList.add('has-active');

  $('head-avatar').textContent = initials(name);
  $('head-name').textContent = name;
  $('head-phone').textContent = phone || '';
  $('reply-error').textContent = '';

  await loadThread(true);
  try {
    await api('/api/chats/' + id + '/read', { method: 'POST' });
  } catch {}
  loadInbox();

  clearInterval(threadTimer);
  threadTimer = setInterval(() => loadThread(false), 2500);
  $('reply-input').focus();
}

async function loadThread(force) {
  if (activeChatId === null) return;
  try {
    const res = await api('/api/chats/' + activeChatId + '/messages');
    const data = await res.json();
    setConn(true);
    const msgs = data.messages || [];
    if (!force && msgs.length === lastThreadCount) return; // нечего перерисовывать
    lastThreadCount = msgs.length;
    renderMessages(msgs);
  } catch (e) {
    if (e.message !== 'unauthorized') setConn(false);
  }
}

function renderMessages(msgs) {
  const box = $('messages');
  box.textContent = '';
  let lastDay = '';
  for (const m of msgs) {
    const day = dayLabel(m.created_at);
    if (day !== lastDay) {
      box.appendChild(el('div', 'day-sep', day));
      lastDay = day;
    }
    const bubble = el('div', 'msg ' + (m.sender === 'manager' ? 'manager' : 'client'));
    bubble.appendChild(document.createTextNode(m.text));
    bubble.appendChild(el('span', 'meta', formatTime(m.created_at)));
    box.appendChild(bubble);
  }
  box.scrollTop = box.scrollHeight;
}

// ---- отправка ответа -------------------------------------------------------

const replyInput = $('reply-input');

$('reply-form').addEventListener('submit', sendReply);
replyInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendReply(e);
  }
});
replyInput.addEventListener('input', () => {
  replyInput.style.height = 'auto';
  replyInput.style.height = Math.min(replyInput.scrollHeight, 140) + 'px';
});

async function sendReply(e) {
  e.preventDefault();
  const text = replyInput.value.trim();
  if (!text || activeChatId === null) return;
  $('reply-error').textContent = '';
  $('reply-send').disabled = true;
  try {
    const res = await api('/api/chats/' + activeChatId + '/messages', {
      method: 'POST',
      body: JSON.stringify({ text }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      $('reply-error').textContent = err.error === 'failed to deliver to telegram'
        ? 'Не удалось доставить сообщение в Telegram'
        : (err.error || 'Не удалось отправить');
      return;
    }
    replyInput.value = '';
    replyInput.style.height = 'auto';
    await loadThread(true);
    loadInbox();
  } catch {
    $('reply-error').textContent = 'Сеть недоступна';
  } finally {
    $('reply-send').disabled = false;
    replyInput.focus();
  }
}

// ---- помощники -------------------------------------------------------------

function setConn(ok) {
  const c = $('conn');
  c.classList.toggle('ok', ok);
  c.classList.toggle('err', !ok);
}

function el(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
}

function initials(name) {
  const parts = (name || '?').trim().split(/\s+/);
  let s = (parts[0] && parts[0][0]) || '?';
  if (parts.length > 1 && parts[1][0]) s += parts[1][0];
  return s.toUpperCase();
}

function formatTime(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d)) return '';
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  if (sameDay) {
    return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  }
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
}

function dayLabel(iso) {
  const d = new Date(iso);
  if (isNaN(d)) return '';
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return 'Сегодня';
  const y = new Date(now);
  y.setDate(now.getDate() - 1);
  if (d.toDateString() === y.toDateString()) return 'Вчера';
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' });
}

init();
