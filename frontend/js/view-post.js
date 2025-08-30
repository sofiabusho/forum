   let currentPost = null;
        let currentPostId = null;
        let isLoggedIn = false;
        let currentUserId = null;
        let currentEditingCommentId = null;
        let editAllCategories = [];
        let editSelectedCategories = new Set();

        // Get post ID from URL
        function getPostIdFromUrl() {
            const urlParams = new URLSearchParams(window.location.search);
            return urlParams.get('id');
        }

        // Load post data
        async function loadPost() {
            const postId = getPostIdFromUrl();
            if (!postId) {
                showError();
                return;
            }

            currentPostId = postId;

            try {
                const response = await fetch(`/api/post?id=${postId}`);
                if (!response.ok) {
                    throw new Error('Post not found');
                }

                const post = await response.json();
                displayPost(post);
                loadComments();

                // Hide loading, show post
                document.getElementById('loading').classList.add('d-none');
                document.getElementById('post-container').classList.remove('d-none');

            } catch (error) {
                console.error('Error loading post:', error);
                showError();
            }
        }

        // Display post data
        function displayPost(post) {
            currentPost = post;

            document.getElementById('post-title').textContent = post.title;
            document.getElementById('post-author').textContent = post.author;
            document.getElementById('post-time').textContent = post.timeAgo;
            document.getElementById('post-content').innerHTML = post.content.replace(/\n/g, '<br>');
            document.getElementById('like-count').textContent = post.likes || 0;
            document.getElementById('dislike-count').textContent = post.dislikes || 0;
            document.getElementById('comment-count').textContent = post.comments || 0;

            // Display tags
            const tagsContainer = document.getElementById('post-tags');
            if (post.tags && post.tags.length > 0) {
                tagsContainer.innerHTML = post.tags.map(tag =>
                    `<span class="tag-badge">#${tag}</span>`
                ).join('');
            }

            // Display image if available
            if (post.imageUrl) {
                const imageContainer = document.getElementById('post-image-container');
                const imageElement = document.getElementById('post-image');
                imageElement.src = post.imageUrl;
                imageContainer.classList.remove('d-none');
            }

            // Add edit/delete buttons if user is the author
            if (isLoggedIn && post.isAuthor) {
                addPostActionButtons();
            }

            // Update vote buttons
            updateVoteButtons(post.userVote || 0);

            // Update page title
            document.title = `${post.title} - Plant Talk`;

            // Set share URL
            document.getElementById('share-url').value = window.location.href;
        }

        // Add post action buttons
            function addPostActionButtons() {
                const postHeader = document.querySelector('.post-content .d-flex.justify-content-between');
                if (postHeader && !document.getElementById('post-actions-dropdown')) {
                    const actionsHTML = `
            <div class="dropdown" id="post-actions-dropdown">
                <button class="btn btn-outline-light btn-sm dropdown-toggle" type="button" 
                        data-bs-toggle="dropdown" aria-expanded="false">
                    <i class="bi bi-three-dots"></i>
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li>
                        <button class="dropdown-item" onclick="editCurrentPost()">
                            <i class="bi bi-pencil me-2"></i>Edit Post
                        </button>
                    </li>
                    <li><hr class="dropdown-divider"></li>
                    <li>
                        <button class="dropdown-item text-danger" onclick="deleteCurrentPost()">
                            <i class="bi bi-trash me-2"></i>Delete Post
                        </button>
                    </li>
                </ul>
            </div>
        `;
                    postHeader.insertAdjacentHTML('beforeend', actionsHTML);
                }
            }

            // Edit current post
            function editCurrentPost() {
                    if (!currentPost) return;

                    // Reset selected categories
                    editSelectedCategories.clear();

                    // Pre-populate form fields
                    document.getElementById('edit-post-title').value = currentPost.title;
                    document.getElementById('edit-post-content').value = currentPost.content;

                    // Load categories and show modal
                    loadEditCategories().then(() => {
                        const modal = new bootstrap.Modal(document.getElementById('editPostModal'));
                        modal.show();
                    });
                }


                async function loadEditCategories() {
                        try {
                            const response = await fetch('/api/categories');
                            const categories = await response.json();
                            editAllCategories = categories;

                            // Pre-select current post categories
                            if (currentPost && currentPost.tags) {
                                currentPost.tags.forEach(tag => editSelectedCategories.add(tag));
                            }

                            renderEditCategories();
                            updateEditCategoryDisplay();
                            validateEditForm();
                        } catch (error) {
                            console.error('Error loading categories for edit:', error);
                            document.getElementById('editCategoryBubbles').innerHTML =
                                '<div class="text-danger">Failed to load categories</div>';
                        }
                    }

                    // Render category bubbles for edit modal
                    function renderEditCategories() {
                        const container = document.getElementById('editCategoryBubbles');
                        container.innerHTML = '';

                        editAllCategories.forEach(category => {
                            const isSelected = editSelectedCategories.has(category.name);
                            const bubble = document.createElement('div');
                            bubble.className = `category-bubble ${isSelected ? 'selected' : ''}`;
                            bubble.dataset.categoryName = category.name;
                            bubble.innerHTML = `
            <span class="category-name">${category.name}</span>
            <i class="bi ${isSelected ? 'bi-check-circle' : 'bi-plus-circle'} ms-1"></i>
        `;
                            bubble.addEventListener('click', () => toggleEditCategory(category.name));
                            container.appendChild(bubble);
                        });
                    }

                    // Toggle category selection for edit modal
                    function toggleEditCategory(categoryName) {
                        if (editSelectedCategories.has(categoryName)) {
                            editSelectedCategories.delete(categoryName);
                        } else {
                            editSelectedCategories.add(categoryName);
                        }

                        updateEditCategoryDisplay();
                        validateEditForm();
                    }

                    // Update visual display of categories for edit modal
                    function updateEditCategoryDisplay() {
                        // Update bubbles
                        document.querySelectorAll('#editCategoryBubbles .category-bubble').forEach(bubble => {
                            const categoryName = bubble.dataset.categoryName;
                            const icon = bubble.querySelector('i');

                            if (editSelectedCategories.has(categoryName)) {
                                bubble.classList.add('selected');
                                icon.className = 'bi bi-check-circle ms-1';
                            } else {
                                bubble.classList.remove('selected');
                                icon.className = 'bi bi-plus-circle ms-1';
                            }
                        });

                        // Update selected categories display
                        const selectedContainer = document.getElementById('editSelectedCategories');
                        selectedContainer.innerHTML = '';

                        if (editSelectedCategories.size === 0) {
                            selectedContainer.innerHTML = '<div class="category-help-text" style="color: rgba(255, 255, 255, 0.7);">No categories selected</div>';
                            return;
                        }

                        editSelectedCategories.forEach(categoryName => {
                            const tag = document.createElement('div');
                            tag.className = 'selected-category-tag';
                            tag.innerHTML = `
            <span>#${categoryName}</span>
            <i class="bi bi-x remove-btn" onclick="removeEditCategory('${categoryName}')"></i>
        `;
                            selectedContainer.appendChild(tag);
                        });
                    }

                    // Remove category from selection in edit modal
                    function removeEditCategory(categoryName) {
                        editSelectedCategories.delete(categoryName);
                        updateEditCategoryDisplay();
                        validateEditForm();
                    }

                    // Validate edit form
                    function validateEditForm() {
                        const title = document.getElementById('edit-post-title').value.trim();
                        const content = document.getElementById('edit-post-content').value.trim();
                        const hasCategories = editSelectedCategories.size > 0;

                        const saveBtn = document.getElementById('save-post-btn');
                        saveBtn.disabled = !(title && content && hasCategories);
                    }

            // Delete current post
            function deleteCurrentPost() {
                if (!currentPost) return;

                const modal = new bootstrap.Modal(document.getElementById('deletePostModal'));
                modal.show();
            }

            // Edit post API call
            async function editPost(title, content) {
                    try {
                        // Create form data with categories
                        const formData = new FormData();
                        formData.append('post_id', currentPostId);
                        formData.append('title', title);
                        formData.append('content', content);

                        // Add categories
                        editSelectedCategories.forEach(category => {
                            formData.append('categories[]', category);
                        });

                        const response = await fetch('/api/posts/edit', {
                            method: 'POST',
                            body: formData
                        });

                        if (response.ok) {
                            // Update the current display
                            currentPost.title = title;
                            currentPost.content = content;
                            currentPost.tags = Array.from(editSelectedCategories);

                            document.getElementById('post-title').textContent = title;
                            document.getElementById('post-content').innerHTML = content.replace(/\n/g, '<br>');

                            // Update tags display
                            const tagsContainer = document.getElementById('post-tags');
                            if (currentPost.tags && currentPost.tags.length > 0) {
                                tagsContainer.innerHTML = currentPost.tags.map(tag =>
                                    `<span class="tag-badge">#${tag}</span>`
                                ).join('');
                            } else {
                                tagsContainer.innerHTML = '';
                            }

                            showSuccessMessage('Post updated successfully!');
                        } else {
                            showErrorMessage('Failed to edit post');
                        }
                    } catch (error) {
                        console.error('Error editing post:', error);
                        showErrorMessage('Error editing post');
                    }
                }

            // Delete post API call
            async function deletePost() {
                try {
                    const response = await fetch('/api/posts/delete', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded',
                        },
                        body: `post_id=${currentPostId}`
                    });

                    if (response.ok) {
                        showSuccessMessage('Post deleted successfully!');
                        setTimeout(() => {
                            window.location.href = '/'; // Redirect to home
                        }, 1500);
                    } else {
                        showErrorMessage('Failed to delete post');
                    }
                } catch (error) {
                    console.error('Error deleting post:', error);
                    showErrorMessage('Error deleting post');
                }
            }

            // Display comments with edit/delete dropdowns
            function displayComments(comments) {
                const commentsList = document.getElementById('comments-list');

                const commentCount = comments ? comments.length : 0;
                const commentText = commentCount === 1 ? 'Comment' : 'Comments';

                const commentsHeader = document.querySelector('.comment-section h4');
                if (commentsHeader) {
                    commentsHeader.innerHTML = `<i class="bi bi-chat-left-text"></i> ${commentCount} ${commentText}`;
                }

                if (!comments || comments.length === 0) {
                    document.getElementById('no-comments').classList.remove('d-none');
                    return;
                }

                commentsList.innerHTML = comments.map(comment => {
                    // Check if current user is the comment author
                    const isCommentAuthor = isLoggedIn && comment.isAuthor;

                    const actionsDropdown = isCommentAuthor ? `
            <div class="dropdown">
                <button class="btn btn-outline-light btn-sm dropdown-toggle" type="button" 
                        data-bs-toggle="dropdown" aria-expanded="false">
                    <i class="bi bi-three-dots-vertical"></i>
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li>
                        <button class="dropdown-item" onclick="editComment(${comment.id}, '${comment.content.replace(/'/g, "\\'")}')">
                            <i class="bi bi-pencil me-2"></i>Edit Comment
                        </button>
                    </li>
                    <li><hr class="dropdown-divider"></li>
                    <li>
                        <button class="dropdown-item text-danger" onclick="deleteComment(${comment.id})">
                            <i class="bi bi-trash me-2"></i>Delete Comment
                        </button>
                    </li>
                </ul>
            </div>
        ` : '';

                    return `
            <div class="comment-item p-3" id="comment-${comment.id}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="d-flex align-items-center mb-2">
                            <strong class="text-white">${comment.author}</strong>
                            <small class="text-muted ms-2">${comment.timeAgo}</small>
                        </div>
                        <p class="text-white mb-2" id="comment-content-${comment.id}">${comment.content.replace(/\n/g, '<br>')}</p>
                        <div class="d-flex align-items-center gap-2">
                            <button class="btn btn-sm btn-outline-light vote-comment-btn ${comment.userVote === 1 ? 'active-like' : ''}" 
                                onclick="voteComment(${comment.id}, 1)">
                                <i class="bi bi-hand-thumbs-up"></i> 
                            </button>
                            <span class="vote-count like-count">${comment.likeCount || 0}</span>
                            <button class="btn btn-sm btn-outline-light vote-comment-btn ${comment.userVote === -1 ? 'active-dislike' : ''}" 
                                onclick="voteComment(${comment.id}, -1)">
                                <i class="bi bi-hand-thumbs-down"></i> 
                            </button>
                            <span class="vote-count like-count">${comment.dislikeCount || 0}</span>
                        </div>
                    </div>
                    ${actionsDropdown}
                </div>
            </div>
        `;
                }).join('');
            }

            // Edit comment function
            function editComment(commentId, currentContent) {
                currentEditingCommentId = commentId;
                document.getElementById('edit-comment-text').value = currentContent;
                const modal = new bootstrap.Modal(document.getElementById('editCommentModal'));
                modal.show();
            }

            // Edit comment API call
            async function editCommentAPI(commentId, content) {
                try {
                    const response = await fetch('/api/comments/edit', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded',
                        },
                        body: `comment_id=${commentId}&content=${encodeURIComponent(content)}`
                    });

                    if (response.ok) {
                        // Update the comment display
                        document.getElementById(`comment-content-${commentId}`).innerHTML = content.replace(/\n/g, '<br>');
                        showSuccessMessage('Comment updated successfully!');
                    } else {
                        showErrorMessage('Failed to edit comment');
                    }
                } catch (error) {
                    console.error('Error editing comment:', error);
                    showErrorMessage('Error editing comment');
                }
            }

            // Delete comment function
            function deleteComment(commentId) {
                    currentEditingCommentId = commentId;
                    const modal = new bootstrap.Modal(document.getElementById('deleteCommentModal'));
                    modal.show();
                }


                // Delete comment API call
                    async function deleteCommentAPI(commentId) {
                        try {
                            const response = await fetch('/api/comments/delete', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/x-www-form-urlencoded',
                                },
                                body: `comment_id=${commentId}`
                            });

                            if (response.ok) {
                                // Remove comment from display
                                document.getElementById(`comment-${commentId}`).remove();

                                // Update comment count
                                const currentCount = parseInt(document.getElementById('comment-count').textContent);
                                document.getElementById('comment-count').textContent = Math.max(0, currentCount - 1);

                                showSuccessMessage('Comment deleted successfully!');
                            } else {
                                showErrorMessage('Failed to delete comment');
                            }
                        } catch (error) {
                            console.error('Error deleting comment:', error);
                            showErrorMessage('Error deleting comment');
                        }
                    }

                     // Show success message
                        function showSuccessMessage(message) {
                            const alertDiv = document.createElement('div');
                            alertDiv.className = 'alert alert-success alert-dismissible fade show position-fixed';
                            alertDiv.style.cssText = 'top: 80px; right: 20px; z-index: 9999; background-color: rgba(40, 167, 69, 0.9); border: none; color: white;';
                            alertDiv.innerHTML = `
                ${message}
                <button type="button" class="btn-close btn-close-white" data-bs-dismiss="alert"></button>
            `;
                            document.body.appendChild(alertDiv);

                            setTimeout(() => {
                                if (alertDiv.parentNode) {
                                    alertDiv.parentNode.removeChild(alertDiv);
                                }
                            }, 3000);
                        }

                        // Show error message
                        function showErrorMessage(message) {
                            const alertDiv = document.createElement('div');
                            alertDiv.className = 'alert alert-danger alert-dismissible fade show position-fixed';
                            alertDiv.style.cssText = 'top: 80px; right: 20px; z-index: 9999; background-color: rgba(220, 53, 69, 0.9); border: none; color: white;';
                            alertDiv.innerHTML = `
                ${message}
                <button type="button" class="btn-close btn-close-white" data-bs-dismiss="alert"></button>
            `;
                            document.body.appendChild(alertDiv);

                            setTimeout(() => {
                                if (alertDiv.parentNode) {
                                    alertDiv.parentNode.removeChild(alertDiv);
                                }
                            }, 3000);
                        }

        // Load comments
        async function loadComments() {
            try {
                document.getElementById('comments-loading').classList.remove('d-none');

                const response = await fetch(`/api/comments?post_id=${currentPostId}`);
                const comments = await response.json();

                document.getElementById('comments-loading').classList.add('d-none');
                displayComments(comments);

            } catch (error) {
                console.error('Error loading comments:', error);
                document.getElementById('comments-loading').classList.add('d-none');
                document.getElementById('no-comments').classList.remove('d-none');
            }
        }

        // Vote on post
        async function votePost(vote) {
            if (!isLoggedIn) {
                alert('Please login to like/dislike posts');
                return;
            }

            try {
                const response = await fetch('/api/posts/like', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: `post_id=${currentPostId}&vote=${vote}`
                });

                if (!response.ok) {
                    throw new Error('Failed to like/dislike post');
                }

                const result = await response.json();
                if (result.success) {
                    document.getElementById('like-count').textContent = result.likeCount;
                    document.getElementById('dislike-count').textContent = result.dislikeCount;
                    updateVoteButtons(result.userVote);
                } else {
                    throw new Error('Like/Dislike failed');
                }

            } catch (error) {
                console.error('Error voting on post:', error);
                alert('Failed to like/dislike. Please try again.');
            }
        }

        // Vote on comment
        async function voteComment(commentId, vote) {
            if (!isLoggedIn) {
                alert('Please login to like/dislike comments');
                return;
            }

            try {
                const response = await fetch('/api/comments/like', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: `comment_id=${commentId}&vote=${vote}`
                });

                const result = await response.json();
                if (result.success) {
                    // Reload comments to update counts
                    loadComments();
                }

            } catch (error) {
                console.error('Error liking/disliking comment:', error);
            }
        }

        // Update vote button states
        function updateVoteButtons(userVote) {
            const likeBtn = document.getElementById('like-btn');
            const dislikeBtn = document.getElementById('dislike-btn');

            // Reset classes
            likeBtn.classList.remove('active-like');
            dislikeBtn.classList.remove('active-dislike');

            // Set active state
            if (userVote === 1) {
                likeBtn.classList.add('active-like');
            } else if (userVote === -1) {
                dislikeBtn.classList.add('active-dislike');
            }
        }

        // Submit comment
        async function submitComment(event) {
            event.preventDefault();

            const commentText = document.getElementById('comment-text').value.trim();
            if (!commentText) return;

            try {
                const response = await fetch('/api/comments/create', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: `post_id=${currentPostId}&content=${encodeURIComponent(commentText)}`
                });

                const result = await response.json();
                if (result.success) {
                    document.getElementById('comment-text').value = '';
                    loadComments();

                    // Update comment count
                    const currentCount = parseInt(document.getElementById('comment-count').textContent);
                    document.getElementById('comment-count').textContent = currentCount + 1;
                }

            } catch (error) {
                console.error('Error submitting comment:', error);
                alert('Error submitting comment. Please try again.');
            }
        }

        // Show image modal
        function showImageModal(imageUrl) {
            document.getElementById('modalImage').src = imageUrl;
            const modal = new bootstrap.Modal(document.getElementById('imageModal'));
            modal.show();
        }

        // Copy URL to clipboard
        function copyToClipboard() {
            const shareUrl = document.getElementById('share-url');
            shareUrl.select();
            shareUrl.setSelectionRange(0, 99999);
            navigator.clipboard.writeText(shareUrl.value);

            // Show feedback
            const button = event.target.closest('button');
            const originalText = button.innerHTML;
            button.innerHTML = '<i class="bi bi-check"></i> Copied!';
            setTimeout(() => {
                button.innerHTML = originalText;
            }, 2000);
        }

        // Show error state
        function showError() {
            document.getElementById('loading').classList.add('d-none');
            document.getElementById('error-container').classList.remove('d-none');
        }

        // Check authentication status
            async function checkAuthStatus() {
                try {
                    const response = await fetch('/api/auth/status');
                    const data = await response.json();
                    isLoggedIn = data.loggedIn;
                    currentUserId = data.userID;

                    if (isLoggedIn) {
                        document.getElementById('comment-form-container').classList.remove('d-none');
                        document.getElementById('login-prompt').classList.add('d-none');

                        // Add post action buttons if user is the author
                        if (currentPost && currentPost.isAuthor) {
                            addPostActionButtons();
                        }
                    } else {
                        document.getElementById('login-prompt').classList.remove('d-none');
                        document.getElementById('comment-form-container').classList.add('d-none');
                    }

                } catch (error) {
                    console.error('Error checking auth status:', error);
                    document.getElementById('login-prompt').classList.remove('d-none');
                    document.getElementById('comment-form-container').classList.add('d-none');
                    isLoggedIn = false;
                }
            }

        // Initialize page
        document.addEventListener('DOMContentLoaded', async () => {
            await checkAuthStatus();
            await loadPost();

            // Setup comment form
            document.getElementById('comment-form').addEventListener('submit', submitComment);

            document.getElementById('edit-post-title').addEventListener('input', validateEditForm);
            document.getElementById('edit-post-content').addEventListener('input', validateEditForm);


            // Setup share button
            document.getElementById('share-btn').addEventListener('click', () => {
                const modal = new bootstrap.Modal(document.getElementById('shareModal'));
                modal.show();
            });
        // Modal event listeners
            // Comment modal handlers
            document.getElementById('save-comment-btn').addEventListener('click', async () => {
                if (currentEditingCommentId) {
                    const newContent = document.getElementById('edit-comment-text').value.trim();
                    if (newContent) {
                        await editCommentAPI(currentEditingCommentId, newContent);
                        bootstrap.Modal.getInstance(document.getElementById('editCommentModal')).hide();
                    }
                }
            });

            document.getElementById('confirm-delete-comment-btn').addEventListener('click', async () => {
                if (currentEditingCommentId) {
                    await deleteCommentAPI(currentEditingCommentId);
                    bootstrap.Modal.getInstance(document.getElementById('deleteCommentModal')).hide();
                }
            });

            // Post modal handlers
            document.getElementById('save-post-btn').addEventListener('click', async () => {
                const newTitle = document.getElementById('edit-post-title').value.trim();
                const newContent = document.getElementById('edit-post-content').value.trim();

                if (!newTitle || !newContent || editSelectedCategories.size === 0) {
                    alert('Please fill in all fields and select at least one category.');
                    return;
                }

                await editPost(newTitle, newContent);
                bootstrap.Modal.getInstance(document.getElementById('editPostModal')).hide();
            });

            document.getElementById('confirm-delete-post-btn').addEventListener('click', async () => {
                await deletePost();
                bootstrap.Modal.getInstance(document.getElementById('deletePostModal')).hide();
            });
        });