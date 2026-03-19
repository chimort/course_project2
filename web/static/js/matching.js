// web/static/js/matching.js
// Depends on: storage.js (getStoredUsername/getAccessToken), ui helpers if you had them

let queuePollTimer = null;
let queueWaitTimer = null;
let queueStartedAt = null;
let wasInQueue = false;
let matchingInterestsLoaded = false;
let matchingUserInterests = [];

function setMatchStatus(text, type = 'idle') {
  const el = document.getElementById('match-status');
  if (!el) return;
  el.textContent = text;
  el.className = `match-status ${type}`;
}

function setMatchResult(obj) {
  const box = document.getElementById('match-result');
  if (!box) return;
  box.textContent = typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2);
}

function startQueueVisuals() {
  document.getElementById('queue-loader').style.display = 'block';
  document.getElementById('queue-meta').style.display = 'flex';

  queueStartedAt = Date.now();

  if (queueWaitTimer) clearInterval(queueWaitTimer);
  queueWaitTimer = setInterval(() => {
    const sec = Math.floor((Date.now() - queueStartedAt) / 1000);
    document.getElementById('queue-wait-time').textContent = `${sec}s`;
  }, 1000);
}

function stopQueueVisuals() {
  document.getElementById('queue-loader').style.display = 'none';
  document.getElementById('queue-meta').style.display = 'none';

  if (queueWaitTimer) {
    clearInterval(queueWaitTimer);
    queueWaitTimer = null;
  }
  queueStartedAt = null;
}

function stopQueuePolling() {
  if (queuePollTimer) {
    clearInterval(queuePollTimer);
    queuePollTimer = null;
  }
}

async function refreshQueueState(username) {
  try {
    const r = await fetch('/v1/matching/list');
    const data = await r.json();
    const list = data.usernames || data.Usernames || [];
    document.getElementById('queue-count').textContent = list.length;

    const isInQueue = list.includes(username);

    if (isInQueue) {
      wasInQueue = true;
      setMatchStatus('Searching for a partner...', 'searching');
    } else if (wasInQueue) {
      // user disappeared from queue -> likely matched
      stopQueuePolling();
      stopQueueVisuals();
      wasInQueue = false;

      setMatchStatus('You may have been matched. Waiting for WS event...', 'success');
      setMatchResult('Removed from queue. Expect "match_found" via WebSocket.');
    } else {
      setMatchStatus('Not searching', 'idle');
    }
  } catch (e) {
    setMatchStatus('Queue polling error', 'error');
    setMatchResult('Network error while checking queue: ' + e.message);
  }
}

function startQueuePolling(username) {
  stopQueuePolling();
  refreshQueueState(username);
  queuePollTimer = setInterval(() => refreshQueueState(username), 3000);
}

function redirectToChat(chatId, partner, matchHint = '', fastChat = false) {
  // Save last chat for "history button" in profile (stub)
  try {
    localStorage.setItem('lastChat', JSON.stringify({
      chat_id: chatId,
      partner: partner,
      match_hint: matchHint,
      fast_chat: fastChat
    }));
  } catch (_) {}

  const url = `/static/html/chat.html?chat_id=${encodeURIComponent(chatId)}&peer=${encodeURIComponent(partner)}&match_hint=${encodeURIComponent(matchHint)}&fast_chat=${fastChat ? '1' : '0'}`;
  window.location.href = url;
}

function handleWsEvent(msg) {
  // Expected payload:
  // { type: "match_found", chat_id: "4", partner: "username2" }
  if (!msg || !msg.type) return;

  if (msg.type === 'match_found') {
    const chatId = msg.chat_id || msg.chatId;
    const partner = msg.partner || msg.peer || msg.username;
    const matchHint = msg.match_hint || msg.matchHint || '';
    const fastChat = !!(msg.fast_chat || msg.fastChat);
    if (chatId && partner) {
      redirectToChat(chatId, partner, matchHint, fastChat);
      return;
    }
    console.log('[MATCH] match_found but missing fields:', msg);
  }
}

function extractInterestNames(user) {
  const interests = Array.isArray(user?.interests) ? user.interests : [];
  return Array.from(new Set(interests.map(it => String(it?.name || it?.interest || '').toLowerCase()).filter(Boolean)));
}

function formatInterestLabel(value) {
  const item = (window.AppInterests?.catalog || []).find(entry => entry.value === value);
  if (item) return item.label;
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function renderMatchingInterestChips(query = '') {
  const chipsWrap = document.getElementById('match-interest-chips');
  const selectedInput = document.getElementById('match-topic-interest');
  if (!chipsWrap || !selectedInput) return;

  const current = selectedInput.value || '';
  const q = String(query || '').trim().toLowerCase();
  const interests = matchingUserInterests.filter(value => !q || formatInterestLabel(value).toLowerCase().includes(q) || value.includes(q));

  chipsWrap.innerHTML = '';

  if (interests.length === 0) {
    chipsWrap.innerHTML = '<div class="match-interest-empty">No matching interests</div>';
    return;
  }

  for (const interest of interests) {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'match-interest-chip' + (interest === current ? ' active' : '');
    btn.textContent = formatInterestLabel(interest);
    btn.onclick = () => {
      selectedInput.value = interest;
      renderMatchingInterestChips(document.getElementById('match-interest-search')?.value || '');
    };
    chipsWrap.appendChild(btn);
  }
}

function applyMatchingModeAvailability(userInterests) {
  matchingUserInterests = Array.isArray(userInterests) ? userInterests : [];

  const interestCard = document.querySelector('.match-mode-card[data-mode-value="3"]');
  const modeInput = document.getElementById('match-mode');
  const topicInput = document.getElementById('match-topic-interest');

  if (interestCard) {
    const canUseInterestMode = matchingUserInterests.length > 0;
    interestCard.disabled = !canUseInterestMode;
    interestCard.classList.toggle('disabled', !canUseInterestMode);
    if (!canUseInterestMode && modeInput?.value === '3') {
      modeInput.value = '1';
      document.querySelector('.match-mode-card[data-mode-value="1"]')?.classList.add('active');
      interestCard.classList.remove('active');
    }
  }

  if (topicInput) {
    if (!matchingUserInterests.includes(topicInput.value)) {
      topicInput.value = matchingUserInterests[0] || '';
    }
  }

  renderMatchingInterestChips();
  updateLanguageModeVisibility();
}

async function refreshMatchingModeOptions() {
  const token = getAccessToken();
  if (!token) return;

  try {
    const r = await fetch('/v1/profile', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token,
        'X-Refresh-Token': getRefreshToken()
      }
    });

    const newAccess = r.headers.get('X-New-Access-Token');
    if (newAccess) saveTokens(newAccess, getRefreshToken());

    const body = await r.json().catch(() => ({}));
    const user = body.user || (body.body && body.body.user);
    if (!r.ok || !user) return;

    applyMatchingModeAvailability(extractInterestNames(user));
    matchingInterestsLoaded = true;
  } catch (e) {
    console.error('refreshMatchingModeOptions error', e);
  }
}

function updateLanguageModeVisibility() {
  const modeEl = document.getElementById('match-mode');
  const fieldEl = document.getElementById('language-match-mode-field');
  const interestFieldEl = document.getElementById('interest-topic-field');
  if (!modeEl || !fieldEl || !interestFieldEl) return;

  fieldEl.style.display = String(modeEl.value) === '2' ? 'block' : 'none';
  interestFieldEl.style.display = String(modeEl.value) === '3' ? 'block' : 'none';
}

// Buttons / UI binding
function bindMatchingEvents() {
  const btnStart = document.getElementById('btn-start-search');
  const btnLeave = document.getElementById('btn-leave-search');
  const modeInput = document.getElementById('match-mode');
  const modeCards = Array.from(document.querySelectorAll('.match-mode-card'));
  const topicSearch = document.getElementById('match-interest-search');

  modeCards.forEach(card => {
    card.addEventListener('click', () => {
      if (card.disabled) return;
      modeCards.forEach(item => item.classList.remove('active'));
      card.classList.add('active');
      modeInput.value = card.dataset.modeValue || '1';
      updateLanguageModeVisibility();
    });
  });

  if (topicSearch) {
    topicSearch.addEventListener('input', () => {
      renderMatchingInterestChips(topicSearch.value);
    });
  }

  bindChipFilterGroup('chat-sort', 'chat-history-sort', () => loadChatHistory());
  bindChipFilterGroup('chat-mode', 'chat-history-mode-filter', () => loadChatHistory());

  updateLanguageModeVisibility();

  refreshMatchingModeOptions();

  if (btnStart) {
    btnStart.onclick = async () => {
      const username = getStoredUsername();
      if (!matchingInterestsLoaded) {
        await refreshMatchingModeOptions();
      }
      const mode = parseInt(document.getElementById('match-mode').value, 10);
      const languageMode = parseInt(document.getElementById('language-match-mode')?.value || '1', 10);
      const topicInterest = String(document.getElementById('match-topic-interest')?.value || '').trim().toLowerCase();

      if (!username) {
        setMatchStatus('Login first', 'error');
        setMatchResult('No username found. Please login first.');
        return;
      }

      setMatchResult('Joining queue...');

      if (mode === 3 && !topicInterest) {
        setMatchStatus('Choose an interest', 'error');
        setMatchResult('Pick one of your interests for interest-focused matching.');
        return;
      }

      try {
        const r = await fetch('/v1/matching/join', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + getAccessToken()
          },
          body: JSON.stringify({ username, mode, language_mode: languageMode, topic_interest: topicInterest })
        });

        const data = await r.json().catch(() => ({}));

        if (r.ok && data.ok) {
          wasInQueue = true;
          setMatchStatus('Searching for a partner...', 'searching');
          startQueueVisuals();
          startQueuePolling(username);
          setMatchResult('In queue. Waiting for match...');
        } else {
          setMatchStatus('Failed to join queue', 'error');
          if (mode === 3 && r.ok) {
            setMatchResult(`You can search by ${topicInterest || 'this interest'} only if it is selected in your profile.`);
          } else {
            setMatchResult({ status: r.status, body: data });
          }
        }
      } catch (e) {
        setMatchStatus('Join request failed', 'error');
        setMatchResult('Network error: ' + e.message);
      }
    };
  }

  if (btnLeave) {
    btnLeave.onclick = async () => {
      const username = getStoredUsername();

      if (!username) {
        setMatchStatus('No username found', 'error');
        setMatchResult('Cannot leave queue: no username');
        return;
      }

      setMatchStatus('Leaving queue...', 'idle');
      setMatchResult('Leaving...');

      try {
        const r = await fetch('/v1/matching/leave', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + getAccessToken()
          },
          body: JSON.stringify({ username })
        });

        const data = await r.json().catch(() => ({}));

        if (r.ok && data.ok) {
          wasInQueue = false;
          stopQueuePolling();
          stopQueueVisuals();
          setMatchStatus('You left the queue', 'idle');
          document.getElementById('queue-count').textContent = '0';
          document.getElementById('queue-wait-time').textContent = '0s';
          setMatchResult('Left queue.');
        } else {
          setMatchStatus('Failed to leave queue', 'error');
          setMatchResult({ status: r.status, body: data });
        }
      } catch (e) {
        setMatchStatus('Leave request failed', 'error');
        setMatchResult('Network error: ' + e.message);
      }
    };
  }
}

// expose for ws.js dispatch
window.AppMatching = {
  bindMatchingEvents,
  handleWsEvent,
  refreshModes: refreshMatchingModeOptions,
  stopQueuePolling,
  stopQueueVisuals,
  resetState: () => { wasInQueue = false; }
};

function bindChipFilterGroup(groupName, inputId, onChange) {
  const input = document.getElementById(inputId);
  const buttons = Array.from(document.querySelectorAll(`[data-filter-group="${groupName}"]`));
  if (!input || buttons.length === 0) return;

  buttons.forEach(btn => {
    btn.addEventListener('click', () => {
      buttons.forEach(item => item.classList.remove('active'));
      btn.classList.add('active');
      input.value = btn.dataset.filterValue || '';
      onChange?.();
    });
  });
}
