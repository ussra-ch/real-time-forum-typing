import { ws } from "./websocket.js"
import { fetchUser } from "./users.js"
import { main } from "./main.js";
import { isAuthenticated } from "./login.js";
import { triggerUserLogout } from "./logout.js";
import { Username } from "./login.js";
export let toool = {
    offset: 0,
    limit: 10
}


function throttle(func, delay) {
    let lastCall = 0;
    return function (...args) {
        const now = Date.now();
        if (now - lastCall >= delay) {
            lastCall = now;
            func.apply(this, args);
        }
    };
}

export function mesaageDiv(user, userId, receiverId) {
    const body = document.querySelector('body')
    toool.offset = 0

    if (document.getElementById('message')) {

        document.getElementById('message').remove()
    }
    document.getElementById('content').style.filter = 'blur(30px)';

    var done = true
    const conversation = document.createElement('div')

    conversation.id = 'message'
    conversation.innerHTML = `
        <div class="head">
            <input type="hidden" id="message_id" name="message_id" value="${receiverId}">
            <input type="hidden" id="name" name="message_id" value="${user}">
            <h3>${user}</h3>
        </div>
        <div class="body" id="chat-body"></div>
        <div id="footer"></div>
        <form class="input-area">
            <input type="text" placeholder="..." class="chat-input" required>
            <button type="submit" class="send-btn"><i class="fa-solid fa-paper-plane"></i></button>
        </form>
    `
    body.append(conversation)
    const container = document.getElementById('chat-body')


    fetchMessages(userId, receiverId, toool.offset, toool.limit, user, Username)
    let lastCall = 0;
    const delay = 500;

    container.addEventListener("scroll", () => {
        if (container.scrollTop === 0) {
            const now = Date.now();
            const canCall = now - lastCall >= delay;

            if (canCall) {
                lastCall = now;
                toool.offset += toool.limit;
                throttle(fetchMessages, 500)(userId, receiverId, toool.offset, toool.limit, user, Username);
            }
        }
    });



    conversation.querySelector('.input-area').addEventListener('submit', (e) => {
        e.preventDefault()
        isAuthenticated().then(auth => {
            if (!auth) {
                triggerUserLogout()
                main()
            } else {
                const input = conversation.querySelector('.chat-input')
                const message = input.value.trim()
                let chatBody = document.getElementById('chat-body')

                if (message !== "") {
                    let newMsg = document.createElement('div')
                    newMsg.className = 'messageReceived'
                    let msgContent = document.createElement('h3')
                    let messagProfil = document.createElement('div')
                    messagProfil.className = 'messagProfil'
                    let profile = document.createElement('div')
                    profile.className = 'profile'
                    if (document.querySelector('.profile')) {
                            profile.innerHTML = `
                                    <i class="fa-solid fa-user"></i>
                                `
                    }
                    messagProfil.appendChild(profile)
                    let h7 = document.createElement('h7')
                    h7.textContent = Username
                    messagProfil.appendChild(h7)
                    msgContent.appendChild(messagProfil)
                    toool.offset++
                    newMsg.append(msgContent)

                    let msgDiv = document.createElement('h3')
                    msgDiv.textContent = `${message}`
                    newMsg.append(msgDiv)

                    let timeDiv = document.createElement('h7')
                    timeDiv.textContent = `${formatDate(Date.now())}`
                    newMsg.append(timeDiv)

                    chatBody.append(newMsg)

                    const payload = {
                        "senderId": userId,
                        "receiverId": receiverId,
                        "messageContent": input.value,
                        "type": "message",
                    };

                    ws.send(JSON.stringify(payload));
                    fetchUser()
                    input.value = ''
                    const container = document.getElementById('chat-body')
                    container.scrollTop = container.scrollHeight;

                }
            }
        })

    })
    conversation.querySelector('.input-area').addEventListener('input', (e) => {
        isAuthenticated().then(auth => {
            if (!auth) {
                triggerUserLogout()
                main()
            } else {
                const payload = {
                    "senderId": userId,
                    "receiverId": receiverId,
                    "type": "typing",
                };

                ws.send(JSON.stringify(payload));
            }
        })


    })


    const deleteButton = document.createElement('button')
    deleteButton.innerHTML = `<i class="fa-solid fa-xmark"></i>`
    deleteButton.id = 'closeConversation'
    let conversationHead = document.querySelector('#message .head')

    conversationHead.append(deleteButton)


    deleteButton.addEventListener('click', (e) => {
        document.getElementById('content').style.filter = 'none';
        let isConversationOpen = {
            senderId: userId,
            receiverId: receiverId,
            type: "CloseConversation",
            ws: ws
        }
        ws.send(JSON.stringify(isConversationOpen));
        conversation.remove();
        document.getElementById('user').style.display='block'

    });

}

function fetchMessages(userId, receiverId, offset, limit, name) {
    const body = document.getElementById('chat-body');
    var previousScrollHeight = body.scrollHeight;

    fetch(`/api/fetchMessages?offset=${offset}&limit=${limit}&sender=${receiverId} `, {
        method: 'GET'
    })
        .then(response => response.json())
        .then(messages => {

            if (!messages) {
                return
            }


            for (const message of messages) {
                console.log(message);

                if (message.content != "") {
                    if (message.userId == userId && message.sender_id == receiverId) {
                        let newMsg = document.createElement('div')
                        newMsg.className = 'messageSent'

                        newMsg.innerHTML = `
                                            <div class="messagProfil">
                                                <div class="profile">
                                                </div>
                                                   <h7>${name}</h7>
                                            </div>
                                            <h3>${message.content}</h3>
                                            <h7>${formatDate(message.time)}</h7>`
                        body.prepend(newMsg)
                        if (document.querySelector('.profile')) {
                                document.querySelector('.profile').innerHTML = `
                                    <i class="fa-solid fa-user"></i>
                                `
                        }
                    } else if (message.userId == receiverId && message.sender_id == userId) {
                        let newMsg = document.createElement('div')
                        newMsg.className = 'messageReceived'
                        newMsg.innerHTML = `
                                            <div class="messagProfil">
                                                <div class="profile">
                                                </div>
                                                   <h7>${message.receiverName}</h7>
                                            </div>
                                            <h3>${message.content}</h3>
                                            <h7>${formatDate(message.time)}</h7>`
                        body.prepend(newMsg)
                            document.querySelector('.profile').innerHTML = `
                                    <i class="fa-solid fa-user"></i>
                                `
                    }
                }
            }
            const newScrollHeight = body.scrollHeight;
            body.scrollTop += (newScrollHeight - previousScrollHeight);
        })
}

export function formatDate(timestampInSeconds) {
    const isoString = timestampInSeconds;
    const date = new Date(isoString);
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    const formattedDate = `${year}-${month}-${day}`;
    const formattedTime = `${hours}:${minutes}`;
    const formattedDateTime = `${formattedDate} ${formattedTime}`;

    return formattedDateTime;
}