import { comment } from "./comment.js";
import { fetchComments } from "./comment.js";
import { main } from "./main.js";
import { isAuthenticated } from "./login.js";
import { triggerUserLogout } from "./logout.js";


export function Create() {
    const Create = document.getElementById('Create')
    const CreateCard = document.createElement('div')
    CreateCard.innerHTML = `
  <div class="post-card" id="createPostCard" >
  <form id="createPostForm" method="get">
  <h3>Create Post</h3>
  <div class="div-title">
  <label for="title">Title :</label>
  <input type="text" name="title" id="title" required>
  </div>
  <div class="div-description">
  <label for="description">description :</label>
  <textarea name="description" id="description" rows="4" required></textarea>
  </div>
  <div class="topic-options">
  <label><input type="checkbox" id="music" name="topic" value="Music" style="margin-right: 39px;"> Music</label>
  <label><input type="checkbox" id="sport" name="topic" value="Sport"> Sport</label>
  <label><input type="checkbox" id="technology" name="topic" value="Technology"> Technology</label>
  <label><input type="checkbox" id="culture" name="topic" value="Culture"> Culture</label>
  <label><input type="checkbox" id="gcience" name="topic" value="Science"> Science</label>
  </div>
  <div id="errorMsg" style="display:none; color:red; margin: 10px 10px;"></div>
  <button type="submit">Post</button>
  </form>
  </div>
  `;
    content.appendChild(CreateCard);
    CreateCard.style.display = 'none';

    const deleteButton = document.createElement('button')
    deleteButton.innerHTML = `<i class="fa-solid fa-xmark"></i>`
    deleteButton.id = 'closeAddPost'

    const form = document.getElementById('createPostForm')
    let createPostDiv = document.getElementById('createPostCard')
    createPostDiv.prepend(deleteButton)

    Create.addEventListener('click', (e) => {
        isAuthenticated().then((auth) => {
            if (!auth) {
                triggerUserLogout()
                main()
            } else {
                form.style.display = "block"
                CreateCard.style.display = 'block'
            }
        })

    });
    document.getElementById('createPostForm').addEventListener('submit', async function (e) {
        e.preventDefault();
        const selectedTopics = Array.from(document.querySelectorAll('input[name="topic"]:checked')).map(el => el.value);
        const formData = new FormData();
        formData.append('title', this.title.value);
        formData.append('description', this.description.value);
        selectedTopics.forEach(topic => formData.append('topics', topic));

        fetch('/api/post', {
            method: 'POST',
            body: formData
        })
            .then(r => {
                if (!r.ok) {
                    return r.json().then(errorData => {
                        throw new Error(errorData.Text || `HTTP error! Status: ${r.status}`);
                    });
                }
                return r.json();
            })
            .then(data => {
                resetForm(form)
                form.style.display = "none"
                CreateCard.style.display = 'none'
                fetchPosts();
            })
            .catch(err => {
                resetForm(form)
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
    });


    deleteButton.addEventListener('click', () => {
        CreateCard.style.display = "none"
        form.style.display = "none"
    })

}

export function Notifications(notifs) {
    const notifications = document.getElementById('notifications')
    const CreateCard = document.createElement('div')
    CreateCard.innerHTML = `
  <div  id="notifications" > ${notifs}
  </div>
  `;
    content.appendChild(notifications);
    notifications.style.display = 'none';
}

export function fetchPosts() {
    fetch('/api/fetch_posts')
        .then(res => res.json())
        .then(posts => {
            const postsContainer = document.getElementById('postsContainer');
            postsContainer.innerHTML = '<span class="posts-title">Posts</span>';
            if (!posts) {
                return
            }
            posts.forEach(post => {
                const topics = post.interest ? post.interest.split(',') : [];
                const postCard = document.createElement('div');
                postCard.className = 'post-card1';
                postCard.innerHTML = `
                    <h3>${post.title}</h3>
                    <p>${post.content}</p>
                    <p>Topics: ${topics.join(', ')}</p>
                    <p>Posted by: #${post.nickname || "Unknown"} on ${new Date(post.created_at).toLocaleDateString()}</p>
                     <form class="commentForm">
                       <div class="inputWrapper">
                     <input type="hidden" name="post_id" value="${post.id}">
                      <input type="text" name="content" class="commentInput" placeholder="Write a comment..." required>
                      <button type="submit" class="commentButton"><i class="fa-solid fa-comment"></i></button>
                      <button type="button" class="show"><i class="fa-solid fa-arrow-down"></i></button>
                      </div>
                    </form>
                `;
                const menu = document.createElement('div')
                menu.style.display = 'none'
                menu.className = 'menu'
                postCard.prepend(menu)

                const div = document.createElement('div');
                div.className = 'comments-container';
                postCard.appendChild(div);
                postsContainer.prepend(postCard);
                div.style.display = 'none'
                let offset = 0;
                const limit = 10;
                const showBtn = postCard.querySelector('.show');
                showBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    if (div.style.display === 'none') {
                        div.style.display = 'block';
                        div.innerHTML = '';
                        offset = 0;
                        fetchComments(post.id, div, offset, limit);
                    } else {
                        div.style.display = 'none';
                    }
                });

                let lastCall = 0;
                const delay = 500;

                div.addEventListener('scroll', () => {
                    if (div.scrollTop + div.clientHeight >= div.scrollHeight-1) {
                        const now = Date.now();
                        const canCall = now - lastCall >= delay;
                        if (canCall) {
                            lastCall = now;
                            offset += limit;
                            fetchComments(post.id, div, offset, limit);
                        }
                    }
                });

                comment(div)
            });
        })
        .catch(err => console.error('Error fetching posts:', err));

}

function resetForm(form) {
    form.title.value = '';
    form.description.value = '';

    document.querySelectorAll('input[name="topic"]:checked').forEach(el => el.checked = false);

    
}