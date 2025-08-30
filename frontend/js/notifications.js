    let notifications = { unread: [], read: [] };
    let isLoading = false;

    // Load notifications from server
    async function loadNotifications() {
        if (isLoading) return;

        try {
            isLoading = true;
            showLoading('unread');

            const response = await fetch('/api/notifications');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            notifications = data;

            hideLoading('unread');
            renderNotifications();
            updateBadges();

        } catch (error) {
            console.error('Error loading notifications:', error);
            hideLoading('unread');
            showError('Failed to load notifications. Please try again.');
        } finally {
            isLoading = false;
        }
    }

    // Render notifications in the UI
    function renderNotifications() {
        renderNotificationList(notifications.unread, 'unreadList', 'unreadEmpty', true);
        renderNotificationList(notifications.read, 'readList', 'readEmpty', false);
    }

    // Render a list of notifications
    function renderNotificationList(notifList, containerId, emptyId, isUnread) {
        const container = document.getElementById(containerId);
        const emptyState = document.getElementById(emptyId);

        if (!container) {
            console.error(`Container ${containerId} not found`);
            return;
        }

        container.innerHTML = '';

        if (!notifList || notifList.length === 0) {
            if (emptyState) emptyState.style.display = 'block';
            return;
        }

        if (emptyState) emptyState.style.display = 'none';

        notifList.forEach(notification => {
            const notifElement = createNotificationElement(notification, isUnread);
            container.appendChild(notifElement);
        });
    }

    // Create a single notification element
    function createNotificationElement(notification, isUnread) {
        const div = document.createElement('div');
        div.className = `notification-item ${isUnread ? 'unread' : 'read'} p-3`;
        div.setAttribute('data-notification-id', notification.id);

        // Get icon based on notification type
        const iconInfo = getNotificationIcon(notification.type);

        // Create clickable link if notification has related content
        const isClickable = notification.relatedPostId || notification.relatedCommentId;
        const clickHandler = isClickable ? `onclick="handleNotificationClick(${notification.id}, ${notification.relatedPostId || 'null'})"` : '';

        div.innerHTML = `
        <div class="d-flex align-items-start" ${clickHandler}>
            <div class="notification-icon ${notification.type}">
                <i class="bi ${iconInfo.icon}"></i>
            </div>
            <div class="notification-content">
                <div class="notification-title">${notification.title}</div>
                <div class="notification-message">${notification.message}</div>
                <div class="notification-time">
                    <i class="bi bi-clock me-1"></i>${notification.timeAgo}
                </div>
                ${isUnread ? `
                    <div class="notification-actions">
                        <button class="btn btn-outline-light mark-read-btn" onclick="markAsRead(${notification.id}, event)">
                            <i class="bi bi-check"></i> Mark as read
                        </button>
                    </div>
                ` : ''}
            </div>
        </div>
    `;

        return div;
    }

    // Get appropriate icon for notification type
    function getNotificationIcon(type) {
            const icons = {
                like: { icon: 'bi-heart-fill', color: '#dc3545' },
                dislike: { icon: 'bi-heartbreak-fill', color: '#6c757d' }, // NEW: Add dislike icon
                comment: { icon: 'bi-chat-fill', color: '#0d6efd' },
                system: { icon: 'bi-info-circle-fill', color: '#198754' }
            };

            return icons[type] || icons.system;
        }

    // Handle notification click (navigate to related content)
    async function handleNotificationClick(notificationId, postId) {
            if (postId) {
                try {
                    // Mark as read when clicking and wait for completion
                    await markAsRead(notificationId);
                    // Navigate to post after marking as read
                    window.location.href = `/view-post?id=${postId}`;
                } catch (error) {
                    console.error('Error marking notification as read:', error);
                    // Navigate anyway, even if mark-as-read failed
                    window.location.href = `/view-post?id=${postId}`;
                }
            }
        }

    // Mark a single notification as read
    async function markAsRead(notificationId, event) {
        if (event) {
            event.stopPropagation(); // Prevent parent click handler
        }

        try {
            const response = await fetch('/api/notifications/mark-read', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: `notification_id=${notificationId}`
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success) {
                    // Move notification from unread to read
                    moveNotificationToRead(notificationId);
                    updateBadges();
                }
            }
        } catch (error) {
            console.error('Error marking notification as read:', error);
            showError('Failed to mark notification as read.');
        }
    }

    // Mark all notifications as read
    async function markAllAsRead() {
        if (notifications.unread.length === 0) return;

        try {
            const response = await fetch('/api/notifications/mark-all-read', {
                method: 'POST'
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success) {
                    // Move all unread to read
                    notifications.read = [...notifications.unread, ...notifications.read];
                    notifications.unread = [];
                    renderNotifications();
                    updateBadges();
                }
            }
        } catch (error) {
            console.error('Error marking all notifications as read:', error);
            showError('Failed to mark all notifications as read.');
        }
    }

    // Move notification from unread to read list
    function moveNotificationToRead(notificationId) {
        const notificationIndex = notifications.unread.findIndex(n => n.id === notificationId);
        if (notificationIndex !== -1) {
            const notification = notifications.unread.splice(notificationIndex, 1)[0];
            notification.isRead = true;
            notifications.read.unshift(notification); // Add to beginning of read list
            renderNotifications();
        }
    }

   function updateBadges() {
        const unreadCount = notifications.unread.length;

        // Update the unread tab badge on notifications page
        const unreadBadge = document.getElementById('unread-badge');
        if (unreadBadge) {
            if (unreadCount > 0) {
                unreadBadge.textContent = unreadCount;
                unreadBadge.style.display = 'inline';
            } else {
                unreadBadge.style.display = 'none';
            }
        }

        // Update the header notification dot (simple red dot)
        const notificationDot = document.getElementById('notification-dot');
        if (notificationDot) {
            if (unreadCount > 0) {
                notificationDot.style.display = 'block';  // Show red dot
            } else {
                notificationDot.style.display = 'none';   // Hide red dot
            }
            console.log('Header dot updated:', unreadCount > 0 ? 'visible' : 'hidden');
        }
    }

    // Show loading state
    function showLoading(tab) {
        const loading = document.getElementById(`${tab}Loading`);
        if (loading) {
            loading.style.display = 'block';
        }
    }

    // Hide loading state
    function hideLoading(tab) {
        const loading = document.getElementById(`${tab}Loading`);
        if (loading) {
            loading.style.display = 'none';
        }
    }

    // Show error message
    function showError(message) {
        // Create a simple toast or alert
        const alertDiv = document.createElement('div');
        alertDiv.className = 'alert alert-danger alert-dismissible fade show position-fixed';
        alertDiv.style.cssText = 'top: 20px; right: 20px; z-index: 9999; max-width: 400px;';
        alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
        document.body.appendChild(alertDiv);

        // Auto remove after 5 seconds
        setTimeout(() => {
            if (alertDiv.parentNode) {
                alertDiv.parentNode.removeChild(alertDiv);
            }
        }, 5000);
    }

    // Handle tab switching
    document.addEventListener('shown.bs.tab', function (event) {
        const targetTab = event.target.getAttribute('data-bs-target');

        if (targetTab === '#read' && notifications.read.length === 0) {
            // Load read notifications if not already loaded
            showLoading('read');
            setTimeout(() => {
                hideLoading('read');
                renderNotificationList(notifications.read, 'readList', 'readEmpty', false);
            }, 500);
        }
    });

    // Event listeners
    document.addEventListener('DOMContentLoaded', () => {
        loadNotifications();

        // Auto-refresh every 30 seconds
        setInterval(loadNotifications, 30000);
    });

    // Real-time updates (if you implement WebSocket in the future)
    function handleRealTimeNotification(notification) {
        notifications.unread.unshift(notification);
        renderNotifications();
        updateBadges();

        // Show browser notification if permission granted
        if (Notification.permission === 'granted') {
            new Notification(notification.title, {
                body: notification.message,
                icon: '/frontend/css/images/cactus.png'
            });
        }
    }

    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
    }