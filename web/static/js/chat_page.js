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
  let mediaRecorder = null;
  let recordingChunks = [];
  let recordingStartedAt = 0;

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
    const attachBtn = document.getElementById('chat-attach');
    const voiceBtn = document.getElementById('chat-voice');
    if (!input || !sendBtn || !blockBtn || !attachBtn || !voiceBtn) return;

    const disabled = fastChat || isBlocked;
    input.disabled = disabled;
    sendBtn.disabled = disabled;
    attachBtn.disabled = disabled;
    voiceBtn.disabled = disabled;

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

  function escapeHtml(value) {
    return String(value || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function formatFileSize(bytes) {
    const size = Number(bytes || 0);
    if (!size) return '';
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }

  function formatDuration(seconds) {
    const total = Math.max(0, Number(seconds || 0));
    const mins = Math.floor(total / 60);
    const secs = total % 60;
    return `${mins}:${String(secs).padStart(2, '0')}`;
  }

  function normalizeMessage(raw) {
    const msg = raw || {};
    return {
      sender: msg.sender || msg.from || '',
      content: msg.content || msg.text || '',
      created_at: msg.created_at || msg.createdAt || '',
      message_type: msg.message_type || msg.messageType || 'text',
      file_url: msg.file_url || msg.fileUrl || '',
      file_name: msg.file_name || msg.fileName || '',
      mime_type: msg.mime_type || msg.mimeType || '',
      file_size_bytes: msg.file_size_bytes || msg.fileSizeBytes || 0,
      duration_seconds: msg.duration_seconds || msg.durationSeconds || 0
    };
  }

  function buildAttachmentHtml(msg) {
    const kind = String(msg.message_type || 'text').toLowerCase();
    const fileUrl = escapeHtml(msg.file_url || '');
    const fileName = escapeHtml(msg.file_name || 'Attachment');

    if (kind === 'image' && fileUrl) {
      return `
        <div class="msg-attachment">
          <a href="${fileUrl}" target="_blank" rel="noopener noreferrer">
            <img class="msg-image" src="${fileUrl}" alt="${fileName}" />
          </a>
        </div>
      `;
    }

    if (kind === 'voice' && fileUrl) {
      const meta = msg.duration_seconds ? `Voice message • ${formatDuration(msg.duration_seconds)}` : 'Voice message';
      return `
        <div class="msg-attachment">
          <audio class="msg-audio" controls preload="metadata" src="${fileUrl}"></audio>
          <div class="msg-meta-inline">${escapeHtml(meta)}</div>
        </div>
      `;
    }

    if ((kind === 'file' || fileUrl) && fileUrl) {
      const extra = [formatFileSize(msg.file_size_bytes), msg.mime_type || '']
        .filter(Boolean)
        .join(' • ');
      return `
        <div class="msg-attachment">
          <a class="msg-file" href="${fileUrl}" target="_blank" rel="noopener noreferrer" download="${fileName}">
            <div>
              <strong>${fileName}</strong>
              <span>${escapeHtml(extra || 'File attachment')}</span>
            </div>
          </a>
        </div>
      `;
    }

    return `<div>${escapeHtml(msg.content || '')}</div>`;
  }

  function renderMessage(msg, me) {
    const normalized = normalizeMessage(msg);
    const box = document.getElementById('chat-messages');
    const div = document.createElement('div');
    div.className = 'msg' + (me ? ' me' : '');

    const kind = String(normalized.message_type || 'text').toLowerCase();
    if (kind !== 'text') {
      div.classList.add('has-attachment', `attachment-${kind}`);
    }

    if (kind === 'text') {
      div.textContent = normalized.content || '';
    } else {
      div.innerHTML = buildAttachmentHtml(normalized);
      if (normalized.content) {
        const caption = document.createElement('div');
        caption.textContent = normalized.content;
        div.appendChild(caption);
      }
    }

    if (normalized.created_at) {
      const meta = document.createElement('div');
      meta.className = 'msg-meta';
      meta.textContent = new Date(normalized.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      div.appendChild(meta);
    }

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
      for (const item of messages) {
        const msg = normalizeMessage(item);
        const sender = msg.sender || '';
        if (sender === me) {
          didSendFirstMessage = true;
        }
        renderMessage(msg, sender === me);
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
    if (didSendFirstMessage || fastChat || isBlocked) {
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

  function handleSuccessfulOwnMessage(renderedMessage) {
    renderMessage(renderedMessage, true);
    if (!didSendFirstMessage) {
      didSendFirstMessage = true;
      hideStarters();
    }
  }

  async function sendMessage() {
    if (isBlocked) return;

    const input = document.getElementById('chat-text');
    const text = input.value.trim();
    if (!text) return;

    input.value = '';
    const createdAt = new Date().toISOString();
    const sentByWs = sendWS({
      type: 'chat_message',
      chat_id: chatId,
      text: text
    });

    if (sentByWs) {
      handleSuccessfulOwnMessage({
        content: text,
        created_at: createdAt,
        message_type: 'text'
      });
      return;
    }

    const persisted = await persistMessageHttp(text);
    if (persisted) {
      handleSuccessfulOwnMessage({
        content: text,
        created_at: createdAt,
        message_type: 'text'
      });
      return;
    }

    input.value = text;
    alert('Message was not sent. Connection issue, try again.');
  }

  async function uploadAttachment(file, preferredType, durationSeconds) {
    if (!file || !chatId || fastChat || isBlocked) return;

    const me = getStoredUsername();
    if (!me) return;

    const form = new FormData();
    form.append('file', file);
    form.append('chat_id', String(chatId));
    form.append('sender', me);
    if (preferredType) form.append('message_type', preferredType);
    if (durationSeconds) form.append('duration_seconds', String(durationSeconds));

    try {
      const r = await fetch('/v1/chat/upload', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer ' + getAccessToken(),
          'X-Refresh-Token': getRefreshToken()
        },
        body: form
      });

      const newAccess = r.headers.get('X-New-Access-Token');
      if (newAccess) saveTokens(newAccess, getRefreshToken());

      const body = await r.json().catch(() => ({}));
      if (!r.ok) {
        alert(body.error || 'Attachment upload failed.');
        return;
      }

      handleSuccessfulOwnMessage({
        content: '',
        created_at: new Date().toISOString(),
        message_type: body.message_type || preferredType || 'file',
        file_url: body.file_url || '',
        file_name: body.file_name || file.name,
        mime_type: body.mime_type || file.type || '',
        file_size_bytes: body.file_size_bytes || file.size || 0,
        duration_seconds: body.duration_seconds || durationSeconds || 0
      });
    } catch (e) {
      console.error('uploadAttachment error', e);
      alert('Could not upload the attachment.');
    }
  }

  async function startVoiceRecording() {
    if (!navigator.mediaDevices || typeof MediaRecorder === 'undefined') {
      alert('Voice recording is not supported in this browser.');
      return;
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      recordingChunks = [];
      recordingStartedAt = Date.now();
      mediaRecorder = new MediaRecorder(stream);

      mediaRecorder.ondataavailable = (event) => {
        if (event.data && event.data.size > 0) {
          recordingChunks.push(event.data);
        }
      };

      mediaRecorder.onstop = async () => {
        const durationSeconds = Math.max(1, Math.round((Date.now() - recordingStartedAt) / 1000));
        const blob = new Blob(recordingChunks, { type: mediaRecorder.mimeType || 'audio/webm' });
        const voiceFile = new File([blob], `voice-${Date.now()}.webm`, { type: blob.type || 'audio/webm' });
        stream.getTracks().forEach((track) => track.stop());
        mediaRecorder = null;
        recordingChunks = [];
        setRecordingUI(false);
        await uploadAttachment(voiceFile, 'voice', durationSeconds);
      };

      mediaRecorder.start();
      setRecordingUI(true);
    } catch (e) {
      console.error('startVoiceRecording error', e);
      alert('Microphone access was denied or unavailable.');
    }
  }

  function setRecordingUI(recording) {
    const voiceBtn = document.getElementById('chat-voice');
    if (!voiceBtn) return;
    voiceBtn.classList.toggle('recording', recording);
    voiceBtn.title = recording ? 'Stop recording' : 'Record voice message';
  }

  async function toggleVoiceRecording() {
    if (fastChat || isBlocked) return;

    if (mediaRecorder && mediaRecorder.state === 'recording') {
      mediaRecorder.stop();
      return;
    }

    await startVoiceRecording();
  }

  function handleFilePicked(event) {
    const file = event.target.files && event.target.files[0];
    if (!file) return;

    const kind = (file.type || '').startsWith('image/') ? 'image' : 'file';
    uploadAttachment(file, kind, 0);
    event.target.value = '';
  }

  function handleWsEvent(msg) {
    if (!msg || !msg.type) return;

    if (msg.type === 'chat_message') {
      if (msg.chat_id && chatId && String(msg.chat_id) !== String(chatId)) return;
      renderMessage(normalizeMessage({
        sender: msg.from || '',
        content: msg.text || '',
        created_at: new Date().toISOString(),
        message_type: msg.message_type || 'text',
        file_url: msg.file_url || '',
        file_name: msg.file_name || '',
        mime_type: msg.mime_type || '',
        file_size_bytes: msg.file_size_bytes || 0,
        duration_seconds: msg.duration_seconds || 0
      }), false);
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
    document.getElementById('chat-attach').onclick = () => document.getElementById('chat-file-input').click();
    document.getElementById('chat-file-input').addEventListener('change', handleFilePicked);
    document.getElementById('chat-voice').onclick = toggleVoiceRecording;
    document.getElementById('chat-text').addEventListener('keydown', (e) => {
      if (e.key === 'Enter') sendMessage();
    });

    document.getElementById('chat-exit').onclick = () => {
      window.location.href = '/';
    };
  }

  init();
})();
