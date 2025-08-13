import 'htmx.org';

document.body.addEventListener("htmx:responseError", function (event) {
  createToast("The server returned an unexpected error");
});

document.body.addEventListener("htmx:sendError", function (event) {
  createToast("Failed to connect to the server");
});

document.body.addEventListener("showMessage", function (event) {
  createToast(event.detail.message, event.detail.level);
});

function createToast(message, level) {
  document.getElementById("toasts").insertAdjacentHTML("beforeend", `
    <div class="alert alert-${sanitizeAlertLevel(level)} shadow-xl">
      ${sanitizeHTML(message)}
      <button class="btn btn-sm btn-circle btn-ghost" onClick="this.parentElement.remove()">✕</button>
    </div>`
  );
}

function sanitizeHTML(content) {
  const decoder = document.createElement("div");
  decoder.innerHTML = content;
  return decoder.textContent;
}

function sanitizeAlertLevel(level) {
  switch (level) {
    case "error":
      return "error";
    case "success":
      return "success";
    default:
      return "error";
  }
}
