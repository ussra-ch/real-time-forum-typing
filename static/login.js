import { fetchUser } from "./users.js";
import { loginDiv, content, notifications } from "./var.js";
import { formatDate } from "./message.js"
import { logout } from "./logout.js"
import { Create } from "./post.js"
import { fetchPosts } from "./post.js";
import { categories } from "./sort.js";
import { comment } from "./comment.js";
import { initWebSocket } from "./websocket.js";
import { main } from "./main.js";
import { triggerUserLogout } from "./logout.js";
import { toool } from "./message.js";
import { ws } from "./websocket.js";
export let profilPhoto
export function logindiv() {
    loginDiv.className = 'container';
    loginDiv.id = 'container';
    loginDiv.innerHTML = ` 
  <!-- Sign Up Form -->
  <div class="form-container sign-up-container">
  <form id="registerForm" action="/register" method="POST">
  <h1>Create Account</h1>
  <div id ="register-error"></div>
  <input type="text" placeholder="Nickname" name="Nickname" required />
  <input type="number" placeholder="age" name="Age" min="13" required />
  <select required name="gender">
  <option value="" disabled selected>Gender</option>
  <option value="male">Male</option>
  <option value="female">Female</option>
  </select>
  <input type="text" placeholder="First Name" name="first_name" required />
  <input type="text" placeholder="Last Name" name="last_name" required />
  <input type="email" placeholder="E-mail" name="email" required />
  <input type="password" placeholder="Password"  name="password" required />
  <button id="register">Sign Up</button>
  </form>
  </div>
  <div class="form-container sign-in-container">
  <form id="logForm" action="/login" method="POST">
  <h1>Sign in</h1>
  
  <input type="Nickname" name="Nickname" placeholder="Nickname" required />
  <input type="password" name="password" placeholder="Password" required />
  <div id = "error-message"></div>
  <button id="login">Sign In</button>
  </form>
  </div>
  <div class="overlay-container">
  <div class="overlay">
  <div class="overlay-panel overlay-left">
  <h1>Welcome Back!</h1>
  <p>To keep connected with us please login with your personal info</p>
  <button class="ghost" id="signIn">Sign In</button>
  </div>
  <div class="overlay-panel overlay-right">
  <h1>Hello, Friend!</h1>
  <p>Enter your personal details and start journey with us</p>
  <button class="ghost" id="signUp">Sign Up</button>
  </div>
  </div>
  </div>
  
  `;
    document.body.appendChild(loginDiv);
    // Use querySelector on loginDiv to get the buttons
    const signUpButton = loginDiv.querySelector('#signUp');
    const signInButton = loginDiv.querySelector('#signIn');
    const container = loginDiv;
    const form = document.getElementById('logForm')

    document.getElementById('login').addEventListener("click", (e) => {
        e.preventDefault()
        const formData = new FormData(form)
        const data = Object.fromEntries(formData.entries())

        fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(res => {
            if (!res.ok) {
                return res.json().then(errorData => {
                    throw new Error(errorData.Text || `HTTP error! Status: ${res.status}`);
                });
            }
            login();
        })
            .catch(err => {
                const ErrorDiv = document.getElementById("error-message");
                ErrorDiv.style.display = 'block'
                ErrorDiv.innerHTML = `${err.message}`;
                setTimeout(() => {
                    ErrorDiv.style.display = 'none'
                }, 2000)
            });
    })

    const registerForm = document.getElementById('registerForm')
    document.getElementById('register').addEventListener('click', (e) => {
        e.preventDefault()
        const formData = new FormData(registerForm)
        const data = Object.fromEntries(formData.entries())
        fetch('/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(res => {
            if (!res.ok) {
                return res.json().then(errorData => {
                    throw new Error(errorData.Text || `HTTP error! Status: ${res.status}`);
                });
            }
            login();
        })
            .catch(err => {
                console.error('Login error:', err)
                const ErrorDiv = document.getElementById("register-error");
                ErrorDiv.style.display = 'block'
                ErrorDiv.innerHTML = `${err.message}`;
                // document.querySelector('body').append(ErrorDiv);
                setTimeout(() => {
                    ErrorDiv.style.display = 'none'
                }, 2000)
            });
    })
    signUpButton.addEventListener('click', () => {
        container.classList.add("right-panel-active");
    });

    signInButton.addEventListener('click', () => {
        container.classList.remove("right-panel-active");
    });
}

function handleUserLogin(userId) {


    initWebSocket((msg) => {
        fetchUser()
        console.log(msg);

        let chatBody = document.getElementById('chat-body');
        if (!chatBody || msg == "") {
            return
        }
        console.log(document.getElementById('message_id').value, msg);

        let newMsg = document.createElement('div');

        if (msg.senderId == userId && document.getElementById('message_id').value == msg.receiverId) {
            newMsg.className = 'messageReceived'
            let msgContent = document.createElement('h3')

            let h7 = document.createElement('h7')
            h7.textContent = Username
            messagProfil.appendChild(h7)
            msgContent.appendChild(messagProfil)


            newMsg.append(msgContent)

            let msgDiv = document.createElement('h3')
            msgDiv.textContent = `${msg.messageContent}`
            newMsg.append(msgDiv)

            let timeDiv = document.createElement('h7')
            timeDiv.textContent = `${formatDate(Date.now())}`
            newMsg.append(timeDiv)

            chatBody.append(newMsg)


        } else {

            if (document.getElementById('message_id').value == msg.senderId) {
                newMsg.className = 'messageSent'
                let messagProfil = document.createElement('div')
                messagProfil.className = 'messagProfil'
                let h7 = document.createElement('h7')
                h7.textContent = msg.name
                messagProfil.appendChild(h7)
                newMsg.appendChild(messagProfil)

                let msgDiv = document.createElement('h3')
                msgDiv.textContent = `${msg.messageContent}`
                newMsg.append(msgDiv)

                let timeDiv = document.createElement('h7')
                timeDiv.textContent = `${formatDate(Date.now())}`
                newMsg.append(timeDiv)
                chatBody.append(newMsg);
            }

        }
        toool.offset++;
        const el = document.getElementById('typing');
        if (el) el.remove();
        chatBody.scrollTop = chatBody.scrollHeight;
    });



    logout();
    Create();
    fetchPosts();
    categories();
    comment();
    fetchUser()
}

export let Username
export function login() {
    const body = document.querySelector('body')
    fetch('/api/authenticated')
        .then(r => r.json())
        .then(res => {
            let profile = `<i class="fa-solid fa-user"></i>`


            if (res.ok) {
                body.innerHTML =
                    ` <div id="content">
                <header>
                <div class="nav">
                <div class="notification-circle">
                <i class="fa-regular fa-bell"></i>
                   <div class="notification-badge" , id ="notification-circle">${notifications}</div>
               </div>
               <button id="Create" style="z-index: 10;"><i class="fa-solid fa-plus"></i></button>
               <button id="profile" style="z-index: 10;">
               </button>
              </div>
                </header>
                
                <div id="category">
                    <span class="category-title">Category</span>
                </div>
                <button id="showUsers"><i class="fa-solid fa-users"></i></button>
                <div id="all">
                <div id="user">
                <div id="users"></div>
                </div>
                <div id="postsContainer">
                <span class="posts-title">Posts</span>
                </div>
                </div>
                </div>
            
            <script type="module" src="static/main.js"></script>`

                document.getElementById('profile').innerHTML = `${profile}`


                const div = document.createElement('div');
                div.innerHTML = `
                    <button id="logout">Logout</button>
                `;
                body.append(div);

                div.style.position = 'absolute';
                div.style.top = '8vh';
                div.style.height = '20vh'
                div.style.right = '0';
                div.style.background = 'rgba(26, 35, 50, 0.8)';
                div.style.padding = '10px';
                div.style.boxShadow = '0 2px 8px rgba(0,0,0,0.2)';
                div.style.zIndex = '1000';
                div.style.display = 'none';
                const logoutBtn = document.getElementById('logout');

                logoutBtn.style.margin = '5px';

                document.getElementById('profile').addEventListener('click', () => {
                    isAuthenticated().then(auth => {
                        if (!auth) {
                            triggerUserLogout()
                            main()
                        } else {
                            if (div.style.display === 'none') {
                                div.style.display = 'flex';
                            } else {
                                div.style.display = 'none';
                            }
                        }
                    })

                });

                logoutBtn.addEventListener('click', () => {
                    isAuthenticated().then(auth => {
                        if (!auth) {
                            triggerUserLogout()
                            main()
                        } else {
                            logout()
                        }
                    })
                })
                Username = res.nickname
                const show = document.getElementById('showUsers')

                const user = document.getElementById('user')
                show.addEventListener('click', () => {
                    if (!document.getElementById('message')) {

                        isAuthenticated().then(auth => {
                            if (!auth) {
                                triggerUserLogout()
                                main()
                            } else {

                                user.style.display = (user.style.display === 'none') ? 'block' : 'none';

                            }
                        })
                    }
                })
                window.addEventListener('resize', () => {
                    if (window.innerWidth > 1370) {
                        user.style.display = 'block'
                    }
                })

                handleUserLogin(res.id);
                return true
            } else {

                body.innerHTML = `
                    <script type="module" src="static/main.js"></script>
                    `
                logindiv()
                return false
            }
        }).catch(err => {
            const existingPopup = document.querySelector(".content");
            if (existingPopup) {
                existingPopup.remove();
            }
            const ErrorDiv = document.createElement('div');
            ErrorDiv.className = 'error-container';
            ErrorDiv.innerHTML = `<div class="content">${err.message}</div>`;
            document.querySelector('body').append(ErrorDiv);
            setTimeout(() => {
                ErrorDiv.remove()
            }, 1000)

        });
}

export function isAuthenticated() {
    return fetch('/api/authenticated')
        .then(r => r.json())
        .then(res => {
            return res.ok
        }).catch(err => {
            const existingPopup = document.querySelector(".content");
            if (existingPopup) {
                existingPopup.remove();
            }
            const ErrorDiv = document.createElement('div');
            ErrorDiv.className = 'error-container';
            ErrorDiv.innerHTML = `<div class="content">${err.message}</div>`;
            document.querySelector('body').append(ErrorDiv);
            setTimeout(() => {
                ErrorDiv.remove()
            }, 1000)
            return false
        });
}
