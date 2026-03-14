// web/static/js/matching.js
// Depends on: storage.js (getStoredUsername/getAccessToken), ui helpers if you had them

let queuePollTimer = null;
let queueWaitTimer = null;
let queueStartedAt = null;
let wasInQueue = false;

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

function redirectToChat(chatId, partner, matchHint = '') {
  // Save last chat for "history button" in profile (stub)
  try {
    localStorage.setItem('lastChat', JSON.stringify({ chat_id: chatId, partner: partner, match_hint: matchHint }));
  } catch (_) {}

  const url = `/static/html/chat.html?chat_id=${encodeURIComponent(chatId)}&peer=${encodeURIComponent(partner)}&match_hint=${encodeURIComponent(matchHint)}`;
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
    if (chatId && partner) {
      redirectToChat(chatId, partner, matchHint);
      return;
    }
    console.log('[MATCH] match_found but missing fields:', msg);
  }
}

// Buttons / UI binding
function bindMatchingEvents() {
  const btnStart = document.getElementById('btn-start-search');
  const btnLeave = document.getElementById('btn-leave-search');

  if (btnStart) {
    btnStart.onclick = async () => {
      const username = getStoredUsername();
      const mode = parseInt(document.getElementById('match-mode').value, 10);

      if (!username) {
        setMatchStatus('Login first', 'error');
        setMatchResult('No username found. Please login first.');
        return;
      }

      setMatchResult('Joining queue...');

      try {
        const r = await fetch('/v1/matching/join', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + getAccessToken()
          },
          body: JSON.stringify({ username, mode })
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
          setMatchResult({ status: r.status, body: data });
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
  stopQueuePolling,
  stopQueueVisuals,
  resetState: () => { wasInQueue = false; }
};
