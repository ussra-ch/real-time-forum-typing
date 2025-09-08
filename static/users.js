import { mesaageDiv } from "./message.js";
import { ws } from "./websocket.js"
import { main } from "./main.js";
import { isAuthenticated } from "./login.js";
import { triggerUserLogout } from "./logout.js";

export function fetchUser() {
  const usern = document.getElementById('users');
  fetch('/user').then(r => r.json()).then(users => {
    if (!users || !users.UserId) {
      triggerUserLogout()
      main()
      return
    }
    usern.textContent = '';

    let on = new Set()
    if (users.onlineUsers) {
      // Only show users with status 'online'
      users.onlineUsers.sort((a, b) => {
        return a.nickname.localeCompare(b.nickname);
      });

      users.onlineUsers.sort((a, b) => {
        return new Date(b.time) - new Date(a.time);
      });

      on = [...new Set(users.onlineUsers)];
    }

    for (const user of on) {
      if (users.UserId == user.userId) {
        continue
      }
      let profile = `<i class="fa-solid fa-user"></i>`
      const conversationButton = document.createElement('button')
      conversationButton.id = "conversationButton"
     

      conversationButton.style.marginRight = '0'
      const div = document.createElement('div');
      div.className = 'user-info';
      div.innerHTML = `${profile} ${user.nickname}`;
      if (user.status == 'online') {
        div.style.color = 'rgb(89, 230, 187)';
      }
      conversationButton.style.display = 'flex';
      conversationButton.style.justifyContent = 'space-between';
      conversationButton.style.alignItems = 'center'
      conversationButton.style.border = '1px solid #ccc';
      conversationButton.style.padding = '8px';
      conversationButton.style.borderRadius = '70px';
      conversationButton.style.width = '60%';
      conversationButton.style.margin = '10px'
      conversationButton.style.background = 'rgba(26, 35, 50, 0.95)';
      conversationButton.className = "user-container"
      conversationButton.append(div)
      usern.appendChild(conversationButton);

      conversationButton.addEventListener('click', () => {
        isAuthenticated().then(auth => {
          if (!auth) {
            triggerUserLogout()
            main()
          } else {
            if (document.getElementById('message')) document.getElementById('message').remove()
            let isConversationOpen = { "senderId": users.UserId, "receiverId": user.userId, "isOpen": true, "type": "OpenConversation", "ws" : ws }
            const jsonIsConversationOpen = JSON.stringify(isConversationOpen);
            ws.send(jsonIsConversationOpen);
            document.getElementById('user').style.display = 'none';
            mesaageDiv(user.nickname, users.UserId, user.userId)
          }
        })
      })
    }
  }).catch(err => {
    console.error("Error fetching users:", err);
  });
}