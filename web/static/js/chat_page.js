// web/static/js/chat_page.js

(function () {
  let chatId = null;
  let peer = null;
  let matchHint = '';
  let fastChat = false;
  let didSendFirstMessage = false;
  let startersConfig = null;
  let isBlocked = false;
  let blockedByMe = false;
  let blockedMe = false;

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
    subEl.textContent = fastChat ? 'Fast one-time chat' : (chatId ? ('Chat #' + chatId) : 'No chat_id');

    if (fastChat) {
      hintEl.textContent = 'Fast one-time chat. It is not saved.';
    } else if (isBlocked && blockedByMe) {
      hintEl.textContent = 'You blocked this user. Messaging is disabled.';
    } else if (isBlocked && blockedMe) {
      hintEl.textContent = 'This user blocked you. Messaging is disabled.';
    } else if (matchHint) {
      hintEl.textContent = 'You are most similar by: ' + toHintLabel(matchHint);
    } else {
      hintEl.textContent = '';
    }
  }

  function applyBlockedState() {
    const input = document.getElementById('chat-text');
    const sendBtn = document.getElementById('chat-send');
    const blockBtn = document.getElementById('chat-block-user');
    if (!input || !sendBtn || !blockBtn) return;

    const disabled = fastChat || isBlocked;
    input.disabled = disabled;
    sendBtn.disabled = disabled;
    if (fastChat) {
      blockBtn.style.display = 'none';
    } else if (blockedByMe) {
      blockBtn.disabled = false;
      blockBtn.textContent = 'Unblock user';
    } else if (blockedMe) {
      blockBtn.disabled = true;
      blockBtn.textContent = 'Blocked by user';
    } else {
      blockBtn.disabled = false;
      blockBtn.textContent = 'Block user';
    }

    if (disabled) {
      hideStarters();
      input.placeholder = isBlocked ? 'Messaging is disabled in this chat' : 'Message...';
    } else {
      input.placeholder = 'Message...';
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
    if (fastChat) return false;

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
    if (!chatId || fastChat) return;

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

  async function loadBlockStatus() {
    if (fastChat || !peer) return;

    const me = getStoredUsername();
    if (!me) return;

    try {
      const r = await fetch(`/v1/chat/block-status/${encodeURIComponent(me)}/${encodeURIComponent(peer)}`, {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        }
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      const body = await r.json().catch(() => ({}));
      if (!r.ok) return;

      isBlocked = !!(body.isBlocked || body.is_blocked);
      blockedByMe = !!(body.blockedByUser1 || body.blocked_by_user1);
      blockedMe = !!(body.blockedByUser2 || body.blocked_by_user2);
      setPeerUI();
      applyBlockedState();
    } catch (e) {
      console.error('loadBlockStatus error', e);
    }
  }

  async function blockCurrentUser() {
    if (fastChat || !peer) return;
    const me = getStoredUsername();
    if (!me) return;

    if (blockedByMe) {
      await unblockCurrentUser();
      return;
    }

    const confirmed = window.confirm(`Block @${peer}? They will not be able to write to you, and you will not match again.`);
    if (!confirmed) return;

    try {
      const r = await fetch('/v1/chat/block', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        },
        body: JSON.stringify({
          blocker_username: me,
          blocked_username: peer
        })
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      if (!r.ok) {
        alert('Could not block this user right now.');
        return;
      }

      isBlocked = true;
      blockedByMe = true;
      blockedMe = false;
      setPeerUI();
      applyBlockedState();
    } catch (e) {
      console.error('blockCurrentUser error', e);
      alert('Network error while blocking user.');
    }
  }

  async function unblockCurrentUser() {
    if (fastChat || !peer) return;
    const me = getStoredUsername();
    if (!me) return;

    const confirmed = window.confirm(`Unblock @${peer}? You may be able to chat and match again.`);
    if (!confirmed) return;

    try {
      const r = await fetch('/v1/chat/unblock', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        },
        body: JSON.stringify({
          blocker_username: me,
          blocked_username: peer
        })
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      if (!r.ok) {
        alert('Could not unblock this user right now.');
        return;
      }

      isBlocked = false;
      blockedByMe = false;
      blockedMe = false;
      setPeerUI();
      applyBlockedState();
      await showStartersIfNeeded();
    } catch (e) {
      console.error('unblockCurrentUser error', e);
      alert('Network error while unblocking user.');
    }
  }

  async function markChatAsRead() {
    if (fastChat) return;

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
    if (isBlocked) return;

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
    fastChat = qs('fast_chat') === '1';

    const username = getStoredUsername();
    if (!username) {
      window.location.href = '/';
      return;
    }

    if (window.AppWS) window.AppWS.connect(username);

    window.AppChat = { handleWsEvent };

    try {
      localStorage.setItem('lastChat', JSON.stringify({
        chat_id: chatId,
        partner: peer,
        match_hint: matchHint,
        fast_chat: fastChat
      }));
    } catch (_) {}

    setPeerUI();
    await loadBlockStatus();
    if (!fastChat) {
      await loadMessageHistory();
      await markChatAsRead();
    }
    await showStartersIfNeeded();
    applyBlockedState();

    document.getElementById('chat-send').onclick = sendMessage;
    document.getElementById('chat-block-user').onclick = blockCurrentUser;
    document.getElementById('chat-text').addEventListener('keydown', (e) => {
      if (e.key === 'Enter') sendMessage();
    });

    document.getElementById('chat-exit').onclick = () => {
      window.location.href = '/';
    };
  }

  init();
})();
