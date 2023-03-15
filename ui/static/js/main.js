function getSelectedPlaylists() {
    return $("#table input[type=checkbox]:checked").map(function() {
        const name = document.getElementById(this.id)
            .parentElement
            .parentElement
            .getElementsByClassName("name")[0].textContent;
        return {id: this.id, name: name};
    }).get();
}

function toggleSelectAll(checked) {
    if (checked) {
        $('#table input[type=checkbox]').prop('checked', true)
    } else {
        $('#table input[type=checkbox]').prop('checked', false)
    }
    toggleSubmitButton();
}

function toggleSubmitButton() {
    if ($("#table input[type=checkbox]:checked").length == 0) {
        document.getElementById("submit").classList.add("disabled")
    } else {
        document.getElementById("submit").classList.remove("disabled")
    }
}

function setup() {
    $(".checkbox").change(function() {
        toggleSubmitButton();
    })
    document.getElementById("submit").onclick = () => {
        const selectedPlaylists = getSelectedPlaylists()
        fetch("/transfer", {
            method: "POST",
            body: JSON.stringify({"playlists": selectedPlaylists})
        })
    }
    document.getElementById("bulk").onchange = (event) => {
        toggleSelectAll(event.target.checked)
    }
}

window.onload = setup
