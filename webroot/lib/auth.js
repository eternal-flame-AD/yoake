async function getAuth() {
    let body = await fetch("/auth.json?" + new Date(), { method: "GET" })
    let bodyJSON = await body.json()
    return bodyJSON
}


function submitLoginForm(target, e) {
    e.preventDefault()

    console.log("submitLoginForm", target)
    let form = $(target);
    var actionUrl = form.attr('action');
    $.ajax({
        type: 'POST',
        url: actionUrl,
        data: form.serialize(),
        success: function (data) {
            window.location.reload();
        },
        error: function (data) {
            try {
                let msg = data.responseJSON.message || data.responseJSON;
                $('#login-form-error').removeClass('d-none').find('span').text(msg);
            } catch (e) {
                $('#login-form-error').removeClass('d-none').find('span').text(e.message);
            }
        }
    });

}

function signin() {
    $('#login-modal').modal('show');
}
function signout() {
    $.ajax({
        type: 'DELETE',
        url: '/auth.json',
        success: function (data) {
            window.location.reload();
        },
        error: function (data) {
            console.warn(data)
        }
    });
}
