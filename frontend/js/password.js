(function () {
    // Eye toggles
    const newBtn = document.getElementById('toggleNewPass');
    const newInput = document.getElementById('new-password');
    const newIcon = newBtn.querySelector('i');
    newBtn.addEventListener('click', () => {
        const hidden = newInput.type === 'password';
        newInput.type = hidden ? 'text' : 'password';
        newIcon.classList.toggle('bi-eye');
        newIcon.classList.toggle('bi-eye-slash');
        newBtn.setAttribute('aria-label', hidden ? 'Hide password' : 'Show password');
    });

    const confBtn = document.getElementById('toggleConfirmPass');
    const confInput = document.getElementById('confirm-password');
    const confIcon = confBtn.querySelector('i');
    confBtn.addEventListener('click', () => {
        const hidden = confInput.type === 'password';
        confInput.type = hidden ? 'text' : 'password';
        confIcon.classList.toggle('bi-eye');
        confIcon.classList.toggle('bi-eye-slash');
        confBtn.setAttribute('aria-label', hidden ? 'Hide confirm password' : 'Show confirm password');
    });

    // Minimal client-side match check
    const form = document.getElementById('resetForm');
    function updateConfirmValidity() {
        const match = confInput.value === newInput.value;
        if (match || confInput.value.length === 0) {
            confInput.classList.remove('is-invalid');
            document.getElementById('confirmInvalid').style.display = 'none';
        } else {
            confInput.classList.add('is-invalid');
            document.getElementById('confirmInvalid').style.display = 'block';
        }
    }
    newInput.addEventListener('input', updateConfirmValidity);
    confInput.addEventListener('input', updateConfirmValidity);
    form.addEventListener('submit', (e) => {
        if (confInput.value !== newInput.value) {
            confInput.classList.add('is-invalid');
            document.getElementById('confirmInvalid').style.display = 'block';
            e.preventDefault();
        }
    });
})();


