function displayErrorToast(title, message) {
    const errorToastId = "error-toast";
    if (!document.getElementById(errorToastId)) {
        let toastDiv = document.createElement("div");
        toastDiv.classList.add("toast-container", "end-0", "position-fixed", "p-3", "bottom-0");
        toastDiv.innerHTML = `
        <div class="toast" role="alert" id="error-toast" aria-live="assertive" aria-atomic="true">
            <div class="toast-header">
            <img src="https://yumechi.jp/img/trima/en/btn_stop.gif" width="20" height="20" class="rounded me-2" aria-hidden="true">
            <strong class="me-auto" id="error-toast-title">Error</strong>
            </div>
            <div class="toast-body">
            An unknown error has occurred.
            </div>
        `
        document.body.prepend(toastDiv);
    }
    let toastDiv = document.getElementById(errorToastId);
    toastDiv.querySelector("#error-toast-title").innerText = title || "Error";
    toastDiv.querySelector(".toast-body").innerText = message || "An unknown error has occurred.";
    let toast = new bootstrap.Toast(toastDiv);
    toast.show();
}

$(document).ajaxError(function (event, jqxhr, settings, thrownError) {
    if (jqxhr.status == 401) {
        signin()
    } else if (jqxhr.status) {
        if (jqxhr.responseJSON && jqxhr.responseJSON.message) {
            displayErrorToast(`${jqxhr.status} ${jqxhr.statusText}`, jqxhr.responseJSON.message);
        } else {
            displayErrorToast(`${jqxhr.status} ${jqxhr.statusText}`, jqxhr.responseText);
        }

    } else {
        displayErrorToast("Error", event.message);
    }
})

window.onerror = function (message, source, lineno, colno, error) {
    displayErrorToast("Error", `${message} (${source}:${lineno}:${colno})`);
}