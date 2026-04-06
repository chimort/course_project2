// web/static/js/ws.js
// No modules, no export/import. Everything is global via window.AppWS.

(function () {
  const AppWS = {
    socket: null,
    isOpen: false,
    username: null,

    connect(username) {
      if (!username) return;

      // already connected with same username
      if (this.socket && (this.isOpen || this.socket.readyState === WebSocket.CONNECTING) && this.username === username) {
        return;
      }

      this.close();
      this.username = username;

      const proto = location.protocol === 'https:' ? 'wss' : 'ws';
      const url = `${proto}://${location.host}/ws?username=${encodeURIComponent(username)}`;

      console.log('[WS] connecting:', url);
      const ws = new WebSocket(url);
      this.socket = ws;

      ws.onopen = () => {
        this.isOpen = true;
        console.log('[WS] open');
      };

      ws.onclose = () => {
        this.isOpen = false;
        console.log('[WS] closed');
      };

      ws.onerror = (e) => {
        console.log('[WS] error', e);
      };

      ws.onmessage = (ev) => {
        let msg = null;
        try {
          msg = JSON.parse(ev.data);
        } catch (e) {
          console.log('[WS] non-json message:', ev.data);
          return;
        }

        if (window.AppChat && typeof window.AppChat.handleWsEvent === 'function') {
          window.AppChat.handleWsEvent(msg);
        }

        if (window.AppMatching && typeof window.AppMatching.handleWsEvent === 'function') {
          window.AppMatching.handleWsEvent(msg);
        }

        if (window.AppNotifications && typeof window.AppNotifications.handleWsEvent === 'function') {
          window.AppNotifications.handleWsEvent(msg);
        }

        console.log('[WS] message:', msg);
      };
    },

    send(obj) {
      if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        console.log('[WS] send skipped (not open)', obj);
        return false;
      }
      this.socket.send(JSON.stringify(obj));
      return true;
    },

    close() {
      if (this.socket) {
        try { this.socket.close(); } catch (_) {}
      }
      this.socket = null;
      this.isOpen = false;
      this.username = null;
    }
  };

  window.AppWS = AppWS;
  window.sendWS = (obj) => AppWS.send(obj);
})();
