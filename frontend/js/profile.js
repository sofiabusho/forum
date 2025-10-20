document.addEventListener('DOMContentLoaded', () => {
    // Cache
    const cache = { posts: null, comments: null, likes: null, dislikes: null };

   // Load profile basics (+ dislikes stats)
    fetch("/api/user/profile")
      .then(res => res.json())
      .then(data => {
        document.getElementById("username-label").textContent = "@" + data.username;
        document.getElementById("joined-date").textContent = data.joinDate;
        document.getElementById("stat-posts").textContent = data.postCount;
        document.getElementById("stat-comments").textContent = data.commentCount;
        document.getElementById("stat-likes-given").textContent = data.likesGiven;
        document.getElementById("stat-likes-received").textContent = data.likesReceived;

        // NEW: fill dislikes stats (fallback 0)
        document.getElementById("stat-dislikes-given").textContent = (data.dislikesGiven ?? 0);
        document.getElementById("stat-dislikes-received").textContent = (data.dislikesReceived ?? 0);

        document.getElementById("bioText").textContent = data.bio || "No bio set yet.";
        if (data.profileImage && data.profileImage.trim() !== "") {
          document.querySelector("aside.profile-box img").src = data.profileImage + "?t=" + new Date().getTime();
        }
      });

    // Change name
    document.getElementById("btn-change-name").addEventListener("click", () => {
      new bootstrap.Modal(document.getElementById('changeNameModal')).show();
    });
    document.getElementById("changeNameForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      const newName = document.getElementById("newDisplayName").value.trim();
      const res = await fetch("/profile", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ username: newName })
      });
      if (res.ok) {
        const data = await res.json();
        if (data.success) {
          document.getElementById("username-label").textContent = "@" + data.username;
          bootstrap.Modal.getInstance(document.getElementById('changeNameModal')).hide();
        }
      }
    });

    // Upload image
    document.getElementById("btn-change-image").addEventListener("click", () => {
      new bootstrap.Modal(document.getElementById("uploadImageModal")).show();
    });
    document.getElementById("uploadImageForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      const formData = new FormData(e.target);
      formData.append("image_type", "profile"); // Specify image type as profile
      const res = await fetch("/upload-image", { method: "POST", body: formData });
      if (res.ok) {
        const data = await res.json();
        if (data.success && data.thumbnailUrl) {
          document.querySelector("aside.profile-box img").src = data.thumbnailUrl + "?t=" + new Date().getTime();
          bootstrap.Modal.getInstance(document.getElementById('uploadImageModal')).hide();
        }
      }
    });

    // Edit bio
    document.getElementById("btn-edit-bio").addEventListener("click", () => {
      const currentBio = document.getElementById("bioText").textContent;
      document.getElementById("newBio").value = currentBio === "No bio set yet." ? "" : currentBio;
      new bootstrap.Modal(document.getElementById("editBioModal")).show();
    });
    document.getElementById("editBioForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      const newBio = document.getElementById("newBio").value.trim();
      const res = await fetch("/profile", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ bio: newBio })
      });
      if (res.ok) {
        const data = await res.json();
        if (data.success) {
          document.getElementById("bioText").textContent = newBio;
          bootstrap.Modal.getInstance(document.getElementById('editBioModal')).hide();
        }
      }
    });

    // Helpers
    function showSection(sectionId) {
      document.querySelectorAll(".section-tab").forEach(sec => {
        sec.style.display = (sec.id === sectionId) ? "block" : "none";
      });
    }
    function activate(btnId) {
      document.querySelectorAll(".profile-sidebar .list-group-item").forEach(b => b.classList.remove("active"));
      document.getElementById(btnId).classList.add("active");
    }

    // Renderers
    function renderPosts(list, container) {
      const el = document.getElementById(container);
      el.innerHTML = "";
      if (!list || list.length === 0) {
        el.innerHTML = `<div class="text-muted px-2">Nothing here yet.</div>`;
        return;
      }
      list.forEach(p => {
        const a = document.createElement("a");
        a.className = "list-group-item list-group-item-action post-card";
        a.href = `/view-post?id=${p.id}`;
        a.innerHTML = `
          <div class="d-flex w-100 justify-content-between">
            <strong class="mb-1">${p.title || "(untitled)"}</strong>
            <small class="text-muted">${p.timeAgo || ""}</small>
          </div>
          ${p.excerpt ? `<p class="mb-1">${p.excerpt}</p>` : ""}
        `;
        el.appendChild(a);
      });
    }

    // Tabs actions
    document.getElementById("tab-bio").addEventListener("click", () => {
      activate("tab-bio");
      showSection("bioSection");
    });

    document.getElementById("tab-posts").addEventListener("click", async () => {
      activate("tab-posts");
      showSection("postsSection");
      if (!cache.posts) {
        try {
          const res = await fetch("/api/user/posts");
          cache.posts = await res.json();
        } catch { cache.posts = []; }
      }
      renderPosts(cache.posts, "userPostsContainer");
    });

    document.getElementById("tab-comments").addEventListener("click", async () => {
      activate("tab-comments");
      showSection("commentsSection");
      if (!cache.comments) {
        try {
          const res = await fetch("/api/user/comments");
          cache.comments = await res.json();
        } catch { cache.comments = []; }
      }
      const el = document.getElementById("userCommentsContainer");
      el.innerHTML = "";
      if (!cache.comments || cache.comments.length === 0) {
        el.innerHTML = `<div class="text-muted px-2">No comments yet.</div>`;
      } else {
        cache.comments.forEach(c => {
          const a = document.createElement("a");
          a.className = "list-group-item list-group-item-action post-card";
          a.href = `/view-post?id=${c.postId}`;
          a.innerHTML = `
            <div class="d-flex w-100 justify-content-between">
              <strong class="mb-1">${c.postTitle || "Post"}</strong>
              <small class="text-muted">${c.timeAgo || ""}</small>
            </div>
            <p class="mb-1">${c.content}</p>
          `;
          el.appendChild(a);
        });
      }
    });

    document.getElementById("tab-likes").addEventListener("click", async () => {
      activate("tab-likes");
      showSection("likesSection");
      if (!cache.likes) {
        try {
          const res = await fetch("/api/user/likes");
          cache.likes = await res.json();
        } catch { cache.likes = []; }
      }
      renderPosts(cache.likes, "userLikesContainer");
    });

    // Dislikes
    document.getElementById("tab-dislikes").addEventListener("click", async () => {
      activate("tab-dislikes");
      showSection("dislikesSection");
      if (!cache.dislikes) {
        try {
          const res = await fetch("/api/user/dislikes");
          cache.dislikes = await res.json();
        } catch { cache.dislikes = []; }
      }
      renderPosts(cache.dislikes, "userDislikesContainer");
    });
  });