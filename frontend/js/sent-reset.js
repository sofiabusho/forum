(function () {
    const params = new URLSearchParams(window.location.search);
    const sent = params.get('sent');

    // set up modal if 'sent' query param is present
    if (sent === '1') {
      const modalEl = document.getElementById('fpSuccessModal');
      const modal = new bootstrap.Modal(modalEl, { backdrop: true, keyboard: true });
      modal.show();

      const goHome = () => { window.location.href = '/'; };

      // Home button click
      document.getElementById('fpGoHome').addEventListener('click', goHome);

      // x button click
      modalEl.addEventListener('hidden.bs.modal', goHome);
    }
  })();
