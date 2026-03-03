let queuePollTimer = null;
let queueWaitTimer = null;
let queueStartedAt = null;
let wasInQueue = false;

function setMatchStatus(text, type = 'idle') {
  const el = document.getElementById('match-status');
  el.textContent = text;
  el.className = `match-status ${type}`;
}

function setMatchResult(obj) {
  const box = document.getElementById('match-result');
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
      stopQueuePolling();
      stopQueueVisuals();
      wasInQueue = false;
      setMatchStatus('You may have been matched. Check your chat.', 'success');
      setMatchResult('User disappeared from queue. Possible match found.');
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

async function handleStartSearch() {
  const username = getStoredUsername();
  const mode = parseInt(document.getElementById('match-mode').value, 10);

  if (!username) {
    setMatchStatus('Login first', 'error');
    setMatchResult('No username found. Please login first.');
    return;
  }

  setMatchResult('...joining');

  try {
    const r = await fetch('/v1/matching/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({ username, mode })
    });

    const data = await r.json();
    setMatchResult({ status: r.status, body: data });

    if (r.ok && data.ok) {
      wasInQueue = true;
      setMatchStatus('Searching for a partner...', 'searching');
      startQueueVisuals();
      startQueuePolling(username);
    } else {
      setMatchStatus('Failed to join queue', 'error');
    }
  } catch (e) {
    setMatchStatus('Join request failed', 'error');
    setMatchResult('Network error: ' + e.message);
  }
}

async function handleLeaveSearch() {
  const username = getStoredUsername();

  if (!username) {
    setMatchStatus('No username found', 'error');
    setMatchResult('Cannot leave queue: no username');
    return;
  }

  setMatchStatus('Leaving queue...', 'idle');
  setMatchResult('...leaving');

  try {
    const r = await fetch('/v1/matching/leave', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({ username })
    });

    const data = await r.json();
    setMatchResult({ status: r.status, body: data });

    if (r.ok && data.ok) {
      wasInQueue = false;
      stopQueuePolling();
      stopQueueVisuals();
      setMatchStatus('You left the queue', 'idle');
      document.getElementById('queue-count').textContent = '0';
      document.getElementById('queue-wait-time').textContent = '0s';
    } else {
      setMatchStatus('Failed to leave queue', 'error');
    }
  } catch (e) {
    setMatchStatus('Leave request failed', 'error');
    setMatchResult('Network error: ' + e.message);
  }
}

function bindMatchingEvents() {
  document.getElementById('btn-matching').onclick = () => showPanel('matching');
  document.getElementById('btn-start-search').onclick = handleStartSearch;
  document.getElementById('btn-leave-search').onclick = handleLeaveSearch;
}