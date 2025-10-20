        let selectedCategories = new Set();
        let allCategories = [];
        let uploadedImageId = null;

        // Load categories from API
        async function loadCategories() {
            try {
                const response = await fetch('/api/categories');
                allCategories = await response.json();
                renderCategoryBubbles();
            } catch (error) {
                console.error('Failed to load categories:', error);
                document.getElementById('categoryBubbles').innerHTML =
                    '<div class="text-danger">Failed to load categories</div>';
            }
        }


        // Render category bubbles
        function renderCategoryBubbles() {
            const container = document.getElementById('categoryBubbles');
            container.innerHTML = '';

            allCategories.forEach(category => {
                const bubble = document.createElement('div');
                bubble.className = 'category-bubble';
                bubble.dataset.categoryName = category.name;
                bubble.innerHTML = `
                    <span>${category.name}</span>
                    <i class="bi bi-plus-circle" data-action="add"></i>
                `;
                
                bubble.addEventListener('click', () => toggleCategory(category.name));
                container.appendChild(bubble);
            });
        }

        // Toggle category selection
        function toggleCategory(categoryName) {
            if (selectedCategories.has(categoryName)) {
                selectedCategories.delete(categoryName);
            } else {
                selectedCategories.add(categoryName);
            }
            
            updateCategoryDisplay();
            updateCategoryInputs();
            validateForm();
        }

        // Update visual display of categories
        function updateCategoryDisplay() {
            // Update bubbles
            document.querySelectorAll('.category-bubble').forEach(bubble => {
                const categoryName = bubble.dataset.categoryName;
                const icon = bubble.querySelector('i');
                
                if (selectedCategories.has(categoryName)) {
                    bubble.classList.add('selected');
                    icon.className = 'bi bi-check-circle';
                    icon.dataset.action = 'remove';
                } else {
                    bubble.classList.remove('selected');
                    icon.className = 'bi bi-plus-circle';
                    icon.dataset.action = 'add';
                }
            });

            // Update selected categories display
            const selectedContainer = document.getElementById('selectedCategories');
            selectedContainer.innerHTML = '';

            if (selectedCategories.size === 0) {
                selectedContainer.innerHTML = '<div class="category-help-text">No categories selected</div>';
                return;
            }

            selectedCategories.forEach(categoryName => {
                const tag = document.createElement('div');
                tag.className = 'selected-category-tag';
                tag.innerHTML = `
                    <span>#${categoryName}</span>
                    <i class="bi bi-x remove-btn" onclick="removeCategory('${categoryName}')"></i>
                `;
                selectedContainer.appendChild(tag);
            });
        }

        // Remove category from selection
        function removeCategory(categoryName) {
            selectedCategories.delete(categoryName);
            updateCategoryDisplay();
            updateCategoryInputs();
            validateForm();
        }

        // Update hidden form inputs for categories
        function updateCategoryInputs() {
            const container = document.getElementById('categoryInputs');
            container.innerHTML = '';

            selectedCategories.forEach(categoryName => {
                const input = document.createElement('input');
                input.type = 'hidden';
                input.name = 'categories[]';
                input.value = categoryName;
                container.appendChild(input);
            });
        }

        // Form validation
        function validateForm() {
            const title = document.getElementById('title').value.trim();
            const content = document.getElementById('content').value.trim();
            const hasCategories = selectedCategories.size > 0;
            
            const publishBtn = document.getElementById('btnPublish');
            publishBtn.disabled = !(title && content && hasCategories);
        }

        // Image upload functionality
        const imageUploadArea = document.getElementById('imageUploadArea');
        const imageInput = document.getElementById('imageInput');
        const imagePreview = document.getElementById('imagePreview');
        const previewImg = document.getElementById('previewImg');
        const imageInfo = document.getElementById('imageInfo');
        const removeImageBtn = document.getElementById('removeImage');
        const uploadProgress = document.getElementById('uploadProgress');
        const progressBar = document.getElementById('progressBar');

        // Click to upload
        imageUploadArea.addEventListener('click', () => {
            imageInput.click();
        });

        // Drag and drop functionality
        imageUploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            imageUploadArea.classList.add('dragover');
        });

        imageUploadArea.addEventListener('dragleave', () => {
            imageUploadArea.classList.remove('dragover');
        });

        imageUploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            imageUploadArea.classList.remove('dragover');

            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleImageFile(files[0]);
            }
        });

        // File input change
        imageInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                handleImageFile(e.target.files[0]);
            }
        });

        // Handle image file
        function handleImageFile(file) {
            // Validate file type
            const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
            if (!allowedTypes.includes(file.type)) {
                alert('Please select a JPEG, PNG, or GIF image.');
                return;
            }

            // Validate file size (20MB)
            const maxSize = 20 * 1024 * 1024;
            if (file.size > maxSize) {
                alert('File size must be less than 20MB.');
                return;
            }

            // Show preview
            const reader = new FileReader();
            reader.onload = (e) => {
                previewImg.src = e.target.result;
                imageInfo.textContent = `${file.name} (${formatFileSize(file.size)})`;

                imageUploadArea.style.display = 'none';
                imagePreview.style.display = 'block';
            };
            reader.readAsDataURL(file);

            // Upload image
            uploadImage(file);
        }

        // Upload image to server
        async function uploadImage(file) {
            const formData = new FormData();
            formData.append('image', file);

            try {
                uploadProgress.style.display = 'block';
                progressBar.style.width = '50%';

                const response = await fetch('/api/upload-image', {
                    method: 'POST',
                    body: formData
                });

                progressBar.style.width = '100%';

                if (response.ok) {
                    const result = await response.json();
                    if (result.success) {
                        uploadedImageId = result.filename;
                        document.getElementById('imageId').value = uploadedImageId;

                        setTimeout(() => {
                            uploadProgress.style.display = 'none';
                        }, 500);
                    } else {
                        throw new Error('Upload failed');
                    }
                } else {
                    throw new Error('Upload failed');
                }
            } catch (error) {
                console.error('Upload error:', error);
                alert('Failed to upload image. Please try again.');
                removeImage();
            }
        }

        // Remove image
        removeImageBtn.addEventListener('click', removeImage);

        function removeImage() {
            // If image was uploaded, delete it from server
            if (uploadedImageId) {
                fetch('/api/delete-image', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: `filename=${uploadedImageId}`
                }).catch(console.error);

                uploadedImageId = null;
                document.getElementById('imageId').value = '';
            }

            // Reset UI
            imagePreview.style.display = 'none';
            imageUploadArea.style.display = 'block';
            uploadProgress.style.display = 'none';
            imageInput.value = '';
            previewImg.src = '';
            progressBar.style.width = '0%';
        }

        // Image modal functionality
            function showImageModal(imageUrl) {
                // Create modal if it doesn't exist
                let modal = document.getElementById('imageModal');
                if (!modal) {
                    modal = document.createElement('div');
                    modal.id = 'imageModal';
                    modal.className = 'modal fade';
                    modal.innerHTML = `
            <div class="modal-dialog modal-lg modal-dialog-centered">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Image</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body text-center">
                        <img id="modalImage" src="" alt="Full size image" class="img-fluid">
                    </div>
                </div>
            </div>
        `;
                    document.body.appendChild(modal);
                }

                // Set image source and show modal
                document.getElementById('modalImage').src = imageUrl;
                const bootstrapModal = new bootstrap.Modal(modal);
                bootstrapModal.show();
            }

        // Add styles to head
            if (!document.getElementById('image-styles')) {
                const styleElement = document.createElement('div');
                styleElement.id = 'image-styles';
                // styleElement.innerHTML = imageStyles;
                document.head.appendChild(styleElement);
            }


        // Format file size
        function formatFileSize(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        // Form submission
            document.getElementById('postForm').addEventListener('submit', async (e) => {
                e.preventDefault();

                if (selectedCategories.size === 0) {
                    alert('Please select at least one category.');
                    return;
                }

                // Submit the form
                e.target.action = '/new-post';
                e.target.submit();
            });

            // Event listeners for form validation
            document.getElementById('title').addEventListener('input', validateForm);
            document.getElementById('content').addEventListener('input', validateForm);

            // Initialize page
            document.addEventListener('DOMContentLoaded', () => {
                loadCategories();
                validateForm();
            });