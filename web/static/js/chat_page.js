// web/static/js/chat_page.js

(function () {
  let chatId = null;
  let peer = null;

  function qs(name) {
    return new URLSearchParams(window.location.search).get(name);
  }

  function setPeerUI() {
    const nameEl = document.getElementById('chat-peer-name');
    const avaEl = document.getElementById('chat-peer-avatar');
    const subEl = document.getElementById('chat-peer-sub');

    nameEl.textContent = peer || 'No chat';
    avaEl.textContent = peer ? peer.charAt(0).toUpperCase() : '?';
    subEl.textContent = chatId ? ('Chat #' + chatId) : 'No chat_id';
  }

  function addMessage(text, me) {
    const box = document.getElementById('chat-messages');
    const div = document.createElement('div');
    div.className = 'msg' + (me ? ' me' : '');
    div.textContent = text;
    box.appendChild(div);
    box.scrollTop = box.scrollHeight;
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
        addMessage(text, sender === me);
      }
    } catch (e) {
      console.error('loadMessageHistory error', e);
    }
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
      return;
    }

    const persisted = await persistMessageHttp(text);
    if (persisted) {
      addMessage(text, true);
      return;
    }

    input.value = text;
    alert('Message was not sent. Connection issue, try again.');
  }

  function handleWsEvent(msg) {
    if (!msg || !msg.type) return;

    if (msg.type === 'chat_message') {
      // expected: {type:"chat_message", chat_id:"4", from:"user", text:"hi"}
      if (msg.chat_id && chatId && String(msg.chat_id) !== String(chatId)) return;
      addMessage(msg.text || '', false);
    }
  }

  function init() {
    chatId = qs('chat_id');
    peer = qs('peer');

    // connect ws using stored username
    const username = getStoredUsername();
    if (!username) {
      // not logged in -> back to main
      window.location.href = '/';
      return;
    }

    if (window.AppWS) window.AppWS.connect(username);

    // register handler for ws.js dispatch
    window.AppChat = { handleWsEvent };

    // Also keep lastChat for profile button
    try {
      localStorage.setItem('lastChat', JSON.stringify({ chat_id: chatId, partner: peer }));
    } catch (_) {}

    setPeerUI();
    loadMessageHistory();

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
