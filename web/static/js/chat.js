import { sendWS } from "./ws.js";

let chatId = null;
let partner = null;

export function openChat(data) {

  chatId = data.chat_id;
  partner = data.partner;

  if (window.showPanel) showPanel('chat');
  else document.getElementById("panel-chat").style.display = "block";

  document.getElementById("chat-peer-name").textContent = partner;

  document.getElementById("chat-peer-avatar").textContent =
    partner.charAt(0).toUpperCase();

  document.getElementById("chat-peer-sub").textContent =
    "Chat #" + chatId;
}

export function handleWsEvent(msg) {

  if (msg.type === "match_found") {

    openChat(msg);

  }

  if (msg.type === "chat_message") {

    addMessage(msg.from, msg.text, false);

  }

}

export function sendMessage() {

  const input =
    document.getElementById("chat-text");

  const text = input.value.trim();

  if (!text) return;

  input.value = "";

  addMessage("me", text, true);

  sendWS({
    type: "chat_message",
    chat_id: chatId,
    text: text
  });
}

function addMessage(user, text, me) {

  const box =
    document.getElementById("chat-messages");

  const div = document.createElement("div");

  div.className = "msg" + (me ? " me" : "");

  div.textContent = text;

  box.appendChild(div);

  box.scrollTop = box.scrollHeight;
}

export function initChat() {

  document
    .getElementById("chat-send")
    .onclick = sendMessage;

  document
    .getElementById("chat-text")
    .addEventListener("keydown", e => {

      if (e.key === "Enter")
        sendMessage();

    });

  document
    .getElementById("btn-chat-back")
    .onclick = () => {

      document.getElementById("panel-chat").style.display = "none";

    };

}