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

  function sendMessage() {
    const input = document.getElementById('chat-text');
    const text = input.value.trim();
    if (!text) return;

    input.value = '';
    addMessage(text, true);

    // server side you will handle later
    sendWS({
      type: 'chat_message',
      chat_id: chatId,
      text: text
    });
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