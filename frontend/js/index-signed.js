
  let allPosts = [];
  let allCategories = [];
  const container = document.getElementById('postsContainer');
  const categoriesList = document.getElementById('categoriesList');

  // Load categories for sidebar
  async function loadCategories() {
    try {
      const res = await fetch('/api/categories');
      allCategories = await res.json();
      renderCategories();
    } catch (e) {
      console.error('Failed to load categories', e);
      categoriesList.innerHTML = '<div class="text-muted p-3">Failed to load categories</div>';
    }
  }

  // Render categories in sidebar
  function renderCategories() {
    categoriesList.innerHTML = '';

    // Add "All Posts" option
    const allPostsItem = document.createElement('button');
    allPostsItem.className = 'list-group-item list-group-item-action category-item active';
    allPostsItem.innerHTML = `
            <i class="bi bi-collection me-2"></i>
            <span>All Posts</span>
        `;
    allPostsItem.addEventListener('click', () => {
      setActiveCategory(allPostsItem);
      setActiveTopbarTab('questions');
      loadPosts();
    });
    categoriesList.appendChild(allPostsItem);

    // Add categories
    allCategories.forEach(category => {
      const categoryItem = document.createElement('button');
      categoryItem.className = 'list-group-item list-group-item-action category-item';
      categoryItem.innerHTML = `
                <i class="bi bi-tag me-2"></i>
                <span>${category.name}</span>
            `;
      categoryItem.addEventListener('click', () => {
        setActiveCategory(categoryItem);
        setActiveTopbarTab('questions');
        loadPosts('categories', category.name);
      });
      categoriesList.appendChild(categoryItem);
    });
  }

  // Set active category
  function setActiveCategory(activeItem) {
    document.querySelectorAll('.category-item').forEach(item => {
      item.classList.remove('active');
    });
    activeItem.classList.add('active');
  }

  // Set active topbar tab
  function setActiveTopbarTab(filter) {
    document.querySelectorAll('[data-filter]').forEach(a => a.classList.remove('active'));
    document.querySelector(`[data-filter="${filter}"]`)?.classList.add('active');
  }

  async function loadPosts(filter = "", value = "") {
    try {
      let url = '/api/posts';

      // Handle authenticated user filters
      if (filter === 'my-posts' || filter === 'my-likes') {
        url = '/api/posts/filtered';
        url += `?filter=${encodeURIComponent(filter)}`;
      } else if (filter === 'categories' && value) {
        url += `?filter=categories&value=${encodeURIComponent(value)}`;
      } else if (filter && filter !== 'questions') {
        url += `?filter=${encodeURIComponent(filter)}`;
      }

      const res = await fetch(url);
      if (!res.ok && (filter === 'my-posts' || filter === 'my-likes')) {
        // If unauthorized, redirect to login
        if (res.status === 401) {
          window.location.href = '/login';
          return;
        }
      }

      allPosts = await res.json();
      renderPosts(allPosts);
    } catch (e) {
      console.error('Failed to load posts', e);
    }
  }

  function renderPosts(posts) {
    container.innerHTML = '';

    if (posts.length === 0) {
      container.innerHTML = `
                <div class="col-12 text-center py-5">
                    <div class="text-muted">
                        <i class="bi bi-chat-square-dots fs-1 mb-3"></i>
                        <h4>No posts found</h4>
                        <p class="mb-3">Be the first to start a discussion!</p>
                        <a href="/new-post" class="btn btn-outline-light">
                            <i class="bi bi-plus-circle me-2"></i>Create Post
                        </a>
                    </div>
                </div>
            `;
      return;
    }

    posts.forEach(post => {
      const col = document.createElement('div');
      col.className = 'col-12';

      // Use thumbnailUrl first, fallback to imageUrl, then no image
      const imageURL = post.thumbnailUrl || post.imageUrl;
      const imageHTML = imageURL ? `
                <div class="post-thumbnail-container">
                    <img src="${imageURL}" 
                         alt="Post thumbnail" 
                         class="post-thumbnail"
                         loading="lazy"
                         onerror="this.parentElement.style.display='none'">
                </div>
            ` : '';

      col.innerHTML = `
                <div class="card shadow-sm mb-3 post-card" onclick="viewPost(${post.id})">
                    <div class="card-body">
                        <!-- Post Header -->
                        <div class="d-flex justify-content-between align-items-start mb-3">
                            <div class="flex-grow-1 clickable-area" onclick="viewPost(${post.id})" style="cursor: pointer;">
                                <h5 class="card-title mb-1 text-white">${post.title}</h5>
                                <small class="text-light">
                                    <i class="bi bi-person me-1"></i>by <strong>${post.author}</strong>
                                    <span class="ms-2"><i class="bi bi-clock me-1"></i>${post.timeAgo}</span>
                                </small>
                            </div>
                        </div>
                        
                        <!-- Post Image Thumbnail -->
                        <div class="clickable-area" onclick="viewPost(${post.id})" style="cursor: pointer;">
                            ${imageHTML}
                        </div>
                        
                        <!-- Post Excerpt -->
                        <div class="clickable-area" onclick="viewPost(${post.id})" style="cursor: pointer;">
                            <p class="card-text text-light mb-3">${post.excerpt}</p>
                        </div>
                        
                        <!-- Tags -->
                        <div class="mb-3 clickable-area" onclick="viewPost(${post.id})" style="cursor: pointer;">
                            ${post.tags.map(tag => `
                                <span class="badge me-1" style="background-color: rgba(109, 214, 40, 0.8); color: white;">
                                    #${tag}
                                </span>
                            `).join('')}
                        </div>
                        
                        <!-- Post Actions and Stats -->
                        <div class="post-actions border-top border-light border-opacity-25">
                            <!-- Left side: Voting buttons -->
                            <div class="post-voting">
                                <button class="vote-btn-sm like-btn ${post.userVote === 1 ? 'active-like' : ''}" 
                                        onclick="voteOnPost(event, ${post.id}, 1)" 
                                        data-post-id="${post.id}" data-vote-type="like">
                                    <i class="bi bi-hand-thumbs-up"></i>
                                </button>
                                <span class="vote-count like-count">${post.likes || 0}</span>
                                
                                <button class="vote-btn-sm dislike-btn ${post.userVote === -1 ? 'active-dislike' : ''}" 
                                        onclick="voteOnPost(event, ${post.id}, -1)" 
                                        data-post-id="${post.id}" data-vote-type="dislike">
                                    <i class="bi bi-hand-thumbs-down"></i>
                                </button>
                                <span class="vote-count dislike-count">${post.dislikes || 0}</span>
                            </div>
                            
                            <!-- Right side: Stats and read more -->
                            <div class="post-stats">
                                <small class="text-light">
                                    <i class="bi bi-chat me-1"></i>${post.comments || 0} comments
                                </small>
                                <small class="text-light clickable-area" onclick="viewPost(${post.id})" style="cursor: pointer;">
                                    <i class="bi bi-arrow-right-circle me-1"></i>See more
                                </small>
                            </div>
                        </div>
                    </div>
                </div>
            `;
      container.appendChild(col);
    });
  }

  // Vote on post from index page
    async function voteOnPost(event, postId, vote) {
      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();

      try {
        const response = await fetch('/api/posts/like', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: `post_id=${postId}&vote=${vote}`
        });

        if (!response.ok) {
          if (response.status === 401) {
            alert('Please login to vote on posts');
            return;
          }
          throw new Error('Failed to vote');
        }

        const result = await response.json();

        console.log('Vote response:', result); // Debug

        if (result.success !== false) {
          // Update the counts in the UI
          updatePostVoteCounts(postId, result.likeCount, result.dislikeCount, result.userVote);
        } else {
          throw new Error('Vote failed');
        }

      } catch (error) {
        console.error('Error voting on post:', error);
        alert('Failed to vote. Please try again.');
      }
    }

    // Update vote counts and button states for a specific post
    function updatePostVoteCounts(postId, likeCount, dislikeCount, userVote) {
        // Find the post card by looking for buttons with the specific post ID
        const likeBtn = document.querySelector(`[data-post-id="${postId}"].like-btn`);
        const dislikeBtn = document.querySelector(`[data-post-id="${postId}"].dislike-btn`);

        if (!likeBtn || !dislikeBtn) {
          console.error(`Could not find vote buttons for post ${postId}`);
          return;
        }

        // Find the count elements (they should be siblings of the buttons)
        const likeCountElement = likeBtn.parentElement.querySelector('.like-count');
        const dislikeCountElement = dislikeBtn.parentElement.querySelector('.dislike-count');

        // Update counts
        if (likeCountElement) likeCountElement.textContent = likeCount || 0;
        if (dislikeCountElement) dislikeCountElement.textContent = dislikeCount || 0;

        // Update button states
        likeBtn.classList.remove('active-like');
        dislikeBtn.classList.remove('active-dislike');

        if (userVote === 1) {
          likeBtn.classList.add('active-like');
        } else if (userVote === -1) {
          dislikeBtn.classList.add('active-dislike');
        }

        console.log(`Updated post ${postId}: likes=${likeCount}, dislikes=${dislikeCount}, userVote=${userVote}`);
      }

  function viewPost(postId) {
    window.location.href = `/view-post?id=${postId}`;
  }

  // Menu tab switching
  document.querySelectorAll('[data-filter]').forEach(link => {
    link.addEventListener('click', e => {
      e.preventDefault();
      setActiveTopbarTab(e.currentTarget.dataset.filter);

      // Reset category selection when switching to user-specific filters
      const filter = e.currentTarget.dataset.filter;
      if (filter === 'my-posts' || filter === 'my-likes') {
        document.querySelectorAll('.category-item').forEach(item => {
          item.classList.remove('active');
        });
      }

      loadPosts(filter);
    });
  });

  // Live text filter
  document.getElementById('filterInput').addEventListener('input', e => {
    const term = e.target.value.toLowerCase().trim();
    renderPosts(allPosts.filter(p =>
      p.title.toLowerCase().includes(term) ||
      p.excerpt.toLowerCase().includes(term) ||
      p.author.toLowerCase().includes(term)
    ));
  });

  // Initial load
  document.addEventListener('DOMContentLoaded', () => {
    loadCategories();
    loadPosts();
  });