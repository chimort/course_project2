// web/static/js/chat_page.js

(function () {
  let chatId = null;
  let peer = null;
  let matchHint = '';
  let didSendFirstMessage = false;
  let startersConfig = null;

  function qs(name) {
    return new URLSearchParams(window.location.search).get(name);
  }

  function toHintLabel(tag) {
    const normalized = String(tag || '').trim().toLowerCase();
    const reasonMap = {
      language: 'language',
      shared_interests: 'shared interests',
      similar_age: 'similar age',
      movies: 'movies'
    };
    if (reasonMap[normalized]) return reasonMap[normalized];
    return normalized || 'common interests';
  }

  function setPeerUI() {
    const nameEl = document.getElementById('chat-peer-name');
    const avaEl = document.getElementById('chat-peer-avatar');
    const subEl = document.getElementById('chat-peer-sub');
    const hintEl = document.getElementById('chat-match-hint');

    nameEl.textContent = peer || 'No chat';
    avaEl.textContent = peer ? peer.charAt(0).toUpperCase() : '?';
    subEl.textContent = chatId ? ('Chat #' + chatId) : 'No chat_id';

    if (matchHint) {
      hintEl.textContent = 'You are most similar by: ' + toHintLabel(matchHint);
    } else {
      hintEl.textContent = '';
    }
  }

  function addMessage(text, me) {
    const box = document.getElementById('chat-messages');
    const div = document.createElement('div');
    div.className = 'msg' + (me ? ' me' : '');
    div.textContent = text;
    box.appendChild(div);
    box.scrollTop = box.scrollHeight;
  }

  function hideStarters() {
    const startersEl = document.getElementById('chat-starters');
    if (!startersEl) return;
    startersEl.style.display = 'none';
    startersEl.innerHTML = '';
  }

  async function persistMessageHttp(text) {
    const username = getStoredUsername();
    if (!chatId || !username || !text) return false;

    try {
      const r = await fetch('/v1/chat/send', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        },
        body: JSON.stringify({
          chat_id: String(chatId),
          sender: username,
          content: text
        })
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      return r.ok;
    } catch (e) {
      console.error('persistMessageHttp error', e);
      return false;
    }
  }

  async function loadMessageHistory() {
    if (!chatId) return;

    try {
      const r = await fetch(`/v1/chat/${encodeURIComponent(chatId)}/messages?limit=100`, {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        }
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      const body = await r.json().catch(() => ({}));
      const messages = body.messages || (body.body && body.body.messages) || [];

      if (!r.ok || !Array.isArray(messages)) return;

      const me = getStoredUsername();
      for (const msg of messages) {
        const text = msg.content || '';
        const sender = msg.sender || '';
        if (sender === me) {
          didSendFirstMessage = true;
        }
        addMessage(text, sender === me);
      }
    } catch (e) {
      console.error('loadMessageHistory error', e);
    }
  }

  async function markChatAsRead() {
    const username = getStoredUsername();
    if (!chatId || !username) return;

    try {
      const r = await fetch(`/v1/chat/${encodeURIComponent(chatId)}/read/${encodeURIComponent(username)}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        },
        body: JSON.stringify({})
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());
    } catch (e) {
      console.error('markChatAsRead error', e);
    }
  }

  function sampleThree(items) {
    const copy = Array.isArray(items) ? items.slice() : [];
    for (let i = copy.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      const tmp = copy[i];
      copy[i] = copy[j];
      copy[j] = tmp;
    }
    return copy.slice(0, 3);
  }

  async function loadStarterConfig() {
    if (startersConfig) return startersConfig;
    try {
      const r = await fetch('/static/data/chat_starters.json');
      if (!r.ok) return null;
      startersConfig = await r.json();
      return startersConfig;
    } catch (e) {
      console.error('loadStarterConfig error', e);
      return null;
    }
  }

  function pickStarterPool(config, hint) {
    if (!config) return [];

    const key = String(hint || '').trim().toLowerCase();

    if (config.reasons && config.reasons[key]) {
      return config.reasons[key];
    }

    if (config.interests && config.interests[key]) {
      return config.interests[key];
    }

    return config.generic || [];
  }

  async function showStartersIfNeeded() {
    if (didSendFirstMessage) {
      hideStarters();
      return;
    }

    const startersEl = document.getElementById('chat-starters');
    if (!startersEl) return;

    const config = await loadStarterConfig();
    if (!config) return;

    const pool = pickStarterPool(config, matchHint);
    const chosen = sampleThree(pool);

    if (chosen.length === 0) {
      hideStarters();
      return;
    }

    startersEl.innerHTML = '';

    for (const phrase of chosen) {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'chat-starter-chip';
      btn.textContent = phrase;
      btn.onclick = () => {
        const input = document.getElementById('chat-text');
        input.value = phrase;
        sendMessage();
      };
      startersEl.appendChild(btn);
    }

    startersEl.style.display = 'flex';
  }

  async function sendMessage() {
    const input = document.getElementById('chat-text');
    const text = input.value.trim();
    if (!text) return;

    input.value = '';
    const sentByWs = sendWS({
      type: 'chat_message',
      chat_id: chatId,
      text: text
    });

    if (sentByWs) {
      addMessage(text, true);
      if (!didSendFirstMessage) {
        didSendFirstMessage = true;
        hideStarters();
      }
      return;
    }

    const persisted = await persistMessageHttp(text);
    if (persisted) {
      addMessage(text, true);
      if (!didSendFirstMessage) {
        didSendFirstMessage = true;
        hideStarters();
      }
      return;
    }

    input.value = text;
    alert('Message was not sent. Connection issue, try again.');
  }

  function handleWsEvent(msg) {
    if (!msg || !msg.type) return;

    if (msg.type === 'chat_message') {
      if (msg.chat_id && chatId && String(msg.chat_id) !== String(chatId)) return;
      addMessage(msg.text || '', false);
      markChatAsRead();
    }
  }

  async function init() {
    chatId = qs('chat_id');
    peer = qs('peer');
    matchHint = qs('match_hint') || '';

    const username = getStoredUsername();
    if (!username) {
      window.location.href = '/';
      return;
    }

    if (window.AppWS) window.AppWS.connect(username);

    window.AppChat = { handleWsEvent };

    try {
      localStorage.setItem('lastChat', JSON.stringify({ chat_id: chatId, partner: peer, match_hint: matchHint }));
    } catch (_) {}

    setPeerUI();
    await loadMessageHistory();
    await markChatAsRead();
    await showStartersIfNeeded();

    document.getElementById('chat-send').onclick = sendMessage;
    document.getElementById('chat-text').addEventListener('keydown', (e) => {
      if (e.key === 'Enter') sendMessage();
    });

    document.getElementById('chat-exit').onclick = () => {
      window.location.href = '/';
    };
  }

  init();
})();
