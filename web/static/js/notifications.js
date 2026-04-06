// web/static/js/notifications.js

(function () {
  function isSupported() {
    return typeof window !== 'undefined' && 'Notification' in window;
  }

  async function requestPermissionIfNeeded() {
    if (!isSupported()) return 'unsupported';
    if (Notification.permission === 'granted') return 'granted';
    if (Notification.permission === 'denied') return 'denied';

    try {
      return await Notification.requestPermission();
    } catch (e) {
      console.error('Notification permission error', e);
      return 'default';
    }
  }

  function currentChatId() {
    try {
      return new URLSearchParams(window.location.search).get('chat_id') || '';
    } catch (_) {
      return '';
    }
  }

  function isSameVisibleChat(msg) {
    const onChatPage = window.location.pathname.includes('/static/html/chat.html');
    if (!onChatPage) return false;
    if (document.visibilityState !== 'visible') return false;
    return String(currentChatId()) === String(msg.chat_id || msg.chatId || '');
  }

  function buildNotificationBody(msg) {
    const kind = String(msg.message_type || msg.messageType || 'text').toLowerCase();
    if (kind === 'image') return 'Sent you an image';
    if (kind === 'voice') return 'Sent you a voice message';
    if (kind === 'file') return 'Sent you a file';
    return msg.text || msg.content || 'Sent you a message';
  }

  function showMessageNotification(msg) {
    if (!isSupported() || Notification.permission !== 'granted') return;
    if (!msg || msg.type !== 'chat_message') return;
    if (isSameVisibleChat(msg)) return;

    const sender = msg.from || msg.sender || 'New message';
    if (typeof window.isUserMuted === 'function' && window.isUserMuted(sender)) return;

    const chatId = msg.chat_id || msg.chatId || '';
    const fastChat = !!(msg.fast_chat || msg.fastChat);
    const notification = new Notification(`Message from ${sender}`, {
      body: buildNotificationBody(msg),
      tag: chatId ? `chat-${chatId}` : undefined
    });

    notification.onclick = () => {
      window.focus();
      if (chatId) {
        const url = `/static/html/chat.html?chat_id=${encodeURIComponent(chatId)}&peer=${encodeURIComponent(sender)}&fast_chat=${fastChat ? '1' : '0'}`;
        window.location.href = url;
      }
      notification.close();
    };
  }

  window.AppNotifications = {
    requestPermissionIfNeeded,
    handleWsEvent(msg) {
      showMessageNotification(msg);
    }
  };
})();
