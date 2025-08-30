
      const toggleBtn = document.getElementById('togglePassword');
    const passInput = document.getElementById('password');
    const icon = toggleBtn.querySelector('i');

    // Click toggle
    toggleBtn.addEventListener('click', () => {
        const isHidden = passInput.type === 'password';
        passInput.type = isHidden ? 'text' : 'password';
        icon.classList.toggle('bi-eye');
        icon.classList.toggle('bi-eye-slash');
        toggleBtn.setAttribute('aria-label', isHidden ? 'Hide password' : 'Show password');
    });