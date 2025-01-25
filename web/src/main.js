import 'htmx.org';

document.body.addEventListener("htmx:responseError", function (event) {
  createToast("The server returned an unexpected error");
});

document.body.addEventListener("htmx:sendError", function (event) {
  createToast("Failed to connect to the server");
});

function createToast(message) {
  document.getElementById("toasts").insertAdjacentHTML("beforeend", `
    <div class="toast">
      <div class="alert alert-error shadow-xl">
        ${message}
        <button class="btn btn-sm btn-circle btn-ghost" onClick="this.parentElement.parentElement.remove()">✕</button>
      </div>
    </div>`
  );
}
