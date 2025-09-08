import { main } from "./main.js";
import { isAuthenticated } from "./login.js";
import { triggerUserLogout } from "./logout.js";

export function comment(div) {
    const forms = document.querySelectorAll('.commentForm');
    forms.forEach((form) => {
        form.addEventListener("submit", (e) => {
            e.preventDefault()
            isAuthenticated().then(auth => {
                if (!auth) {
                    triggerUserLogout()
                    main()
                } else {
                    const commentInput = form.querySelector(".commentInput");
                    const comment = commentInput.value;
                    const post_id = form.querySelector("[name='post_id']").value;
                    if (!comment) return;
                    fetch("/comment", {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json",
                        },
                        body: JSON.stringify({ comment, post_id }),
                    })
                        .then(res => {
                            if (!res.ok) {
                                return res.json().then(errorData => {
                                    throw new Error(errorData.Text || `HTTP error! Status: ${res.status}`);
                                });
                            }
                            return res.json()
                        })
                        .then(data => {
                            commentInput.value = "";
                            div.innerHTML = "";
                            fetchComments(post_id, div, 0, 10)
                        })
                        .catch(err => {
                            console.error("Error  :", err);
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
            })
        });
    });
}

export function fetchComments(postId, container, offset, limit) {
    fetch(`/api/fetch_comments?offset=${offset}&limit=${limit}`)
        .then(res => res.json())
        .then(comments => {
            if (!comments) {
                return
            }
            
            comments.forEach(comment => {
                if (comment.PostID != postId) return;

                const p = document.createElement("div");
                p.className = 'comment'
                p.innerHTML = `
                            <p><strong>${comment.Name}:</strong> ${comment.Content}</p>
                            <p class="comment-date">${new Date(comment.CreatedAt).toLocaleDateString()}</p>
                `;
                container.appendChild(p);
            });
        }).catch(err=>{console.log(err);
        });
}