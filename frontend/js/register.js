    // Eye toggle handlers
    document.getElementById('togglePassword').addEventListener('click', function () {
        const input = document.getElementById('password');
        const icon = this.querySelector('i');
        if (input.type === 'password') {
            input.type = 'text';
            icon.classList.replace('bi-eye', 'bi-eye-slash');
            this.setAttribute('aria-label', 'Hide password');
        } else {
            input.type = 'password';
            icon.classList.replace('bi-eye-slash', 'bi-eye');
            this.setAttribute('aria-label', 'Show password');
        }
    });

    document.getElementById('toggleConfirmPassword').addEventListener('click', function () {
        const input = document.getElementById('confirmPassword');
        const icon = this.querySelector('i');
        if (input.type === 'password') {
            input.type = 'text';
            icon.classList.replace('bi-eye', 'bi-eye-slash');
            this.setAttribute('aria-label', 'Hide confirm password');
        } else {
            input.type = 'password';
            icon.classList.replace('bi-eye-slash', 'bi-eye');
            this.setAttribute('aria-label', 'Show confirm password');
        }
    });

    // Client-side validation
    const form = document.getElementById('registerForm');

    function validatePasswordRule(pass) {
        // 1 lowercase, 1 uppercase, 1 number, 1 symbol, length >= 8
        const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[\W_]).{8,}$/;
        return regex.test(pass);
    }

    function setValidity(el, isValid, invalidMsgEl) {
        if (isValid) {
            el.classList.remove('is-invalid');
            el.classList.add('is-valid');
            if (invalidMsgEl) invalidMsgEl.style.display = 'none';
        } else {
            el.classList.remove('is-valid');
            el.classList.add('is-invalid');
            if (invalidMsgEl) invalidMsgEl.style.display = 'block';
        }
    }

    form.addEventListener('submit', function (e) {
        const passEl = document.getElementById('password');
        const confirmEl = document.getElementById('confirmPassword');
        const pass = passEl.value;
        const confirm = confirmEl.value;

        const passOk = validatePasswordRule(pass);
        setValidity(passEl, passOk, document.getElementById('passwordInvalidMsg'));

        const matchOk = pass === confirm;
        setValidity(confirmEl, matchOk, document.getElementById('confirmInvalidMsg'));

        if (!passOk || !matchOk) {
            e.preventDefault();
        }
    });

    // Live validation feedback as user types
    document.getElementById('password').addEventListener('input', function () {
        setValidity(this, validatePasswordRule(this.value), document.getElementById('passwordInvalidMsg'));
    });
    document.getElementById('confirmPassword').addEventListener('input', function () {
        const pass = document.getElementById('password').value;
        setValidity(this, this.value === pass, document.getElementById('confirmInvalidMsg'));
    });