(function () {
  function injectButtons() {
    const passwordInputs = document.querySelectorAll(
      'input[type="password"]:not([data-kpass-injected])'
    );

    passwordInputs.forEach((input) => {
      input.setAttribute("data-kpass-injected", "true");

      const button = document.createElement("button");
      button.type = "button";
      button.className = "kpass-fill-button";
      button.textContent = "Open in KPass";
      button.addEventListener("click", () => {
        const host = window.location.hostname;
        window.location.href = `kpass://fill?host=${encodeURIComponent(host)}`;
      });

      input.insertAdjacentElement("afterend", button);
    });
  }

  injectButtons();

  let debounceTimer;
  const observer = new MutationObserver(() => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(injectButtons, 300);
  });

  observer.observe(document.body, { childList: true, subtree: true });
})();
