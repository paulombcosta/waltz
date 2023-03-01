function getSelectedPlaylists() {
    return $("#table input[type=checkbox]:checked").map(function() {
        return this.id;
    }).get();
}

function toggleSelectAll(checked) {
    if (checked) {
        $('#table input[type=checkbox]').prop('checked', true)
    } else {
        $('#table input[type=checkbox]').prop('checked', false)
    }
}

function setup() {
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
