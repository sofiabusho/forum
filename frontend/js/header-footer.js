    // header
  fetch('/frontend/templates/shared/header-signed.html')
    .then(r => r.text())
    .then(html => {
      document.getElementById('shared-header').innerHTML = html;

      // Wait a moment, then directly call the functions
      setTimeout(() => {
        // Update notifications
        fetch('/api/notifications/count')
          .then(response => response.json())
          .then(data => {
            const dot = document.getElementById('notification-dot');
            if (dot) {
              dot.style.display = data.count > 0 ? 'block' : 'none';
              console.log('🔔 Red dot updated:', data.count > 0 ? 'visible' : 'hidden');
            }
          })
          .catch(error => console.error('🔔 Notification error:', error));

        // Update profile image 
        fetch('/api/user/profile')  // ← Make sure this is singular "user", not "users"
          .then(response => {
            if (response.ok) {
              return response.json();
            }
            throw new Error(`HTTP ${response.status}`);
          })
          .then(data => {
            console.log('👤 Profile data received:', data);
            const headerImg = document.getElementById('header-profile-image');
            if (headerImg && data.profileImage && data.profileImage.trim() !== '') {
              console.log('👤 Setting header image to:', data.profileImage);
              headerImg.src = data.profileImage;
              headerImg.onerror = function () {
                console.log('👤 Image failed to load, using default');
                this.src = '/frontend/css/images/avatar.png';
              };
            } else {
              console.log('👤 No profile image found or element missing');
            }
          })
          .catch(error => console.error('👤 Profile image error:', error));

      }, 300);
    })
    .catch(error => console.error('❌ Header loading error:', error));

    // footer
    fetch('/frontend/templates/shared/footer.html').then(r => r.text())
        .then(html => document.getElementById('shared-footer').innerHTML = html);
    