 // Load header based on authentication status

async function loadHeaderBasedOnAuth() {
            try {
                // First check authentication status
                const authResponse = await fetch('/api/auth/status');
                const authData = await authResponse.json();
                const isLoggedIn = authData.loggedIn;

                // Load appropriate header
                let headerFile = isLoggedIn ?
                    '/frontend/templates/shared/header-signed.html' :
                    '/frontend/templates/shared/header.html';

                const headerResponse = await fetch(headerFile);
                const html = await headerResponse.text();
                document.getElementById('shared-header').innerHTML = html;

                // If signed in, update notifications and profile image
                if (isLoggedIn) {
                    setTimeout(() => {
                        // Update notifications
                        fetch('/api/notifications/count')
                            .then(response => response.json())
                            .then(data => {
                                const dot = document.getElementById('notification-dot');
                                if (dot) {
                                    dot.style.display = data.count > 0 ? 'block' : 'none';
                                }
                            })
                            .catch(error => console.error('🔔 Notification error:', error));

                        // Update profile image 
                        fetch('/api/user/profile')
                            .then(response => {
                                if (response.ok) {
                                    return response.json();
                                }
                                throw new Error(`HTTP ${response.status}`);
                            })
                            .then(data => {
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
                }
            } catch (error) {
                // Fallback to unsigned header
                fetch('/frontend/templates/shared/header.html')
                    .then(r => r.text())
                    .then(html => document.getElementById('shared-header').innerHTML = html)
                    .catch(fallbackError => console.error('❌ Fallback header error:', fallbackError));
            }
        }

        // Call the function
        loadHeaderBasedOnAuth();


        // Load footer
        fetch('/frontend/templates/shared/footer.html').then(r => r.text())
            .then(html => document.getElementById('shared-footer').innerHTML = html);