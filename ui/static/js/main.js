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
        // startTransfer()
        setupProgress();
    }
    document.getElementById("bulk").onchange = (event) => {
        toggleSelectAll(event.target.checked)
    }
}

function startTransfer() {
    const socket = new WebSocket("ws://localhost:8080/transfer")
    const playlists = getSelectedPlaylists();
    socket.addEventListener('open', (event) => {
        socket.send(JSON.stringify({"playlists": playlists}));
    });
    
    socket.addEventListener('message', (event) => {
        // TODO Listen to progress messages
        console.log('Message from server ', event.data);
    });
}

function setupProgress(playlists) {
    progressContainer = document.createElement("div");
    progressContainer.classList.add("progressContainer")

    title = document.createElement("p");
    title.classList.add("progressTitle")
    title.textContent = "Transfer in Progress"

    currentPlaylist = document.createElement("p");
    currentPlaylist.classList.add("currentPlaylist");
    currentPlaylist.textContent = "-";

    playlistProgressCount = document.createElement("p")
    playlistProgressCount.classList.add("playlistProgressCount")
    // TODO get total count from selected playlists

    trackProgressCount = document.createElement("p")
    trackProgressCount.classList.add("trackProgressCount")
    // TODO get total count from selected playlists

    progressContainer.appendChild(title);
    progressContainer.appendChild(currentPlaylist);
    progressContainer.appendChild(playlistProgressCount);
    progressContainer.appendChild(trackProgressCount);

    document.getElementsByTagName("body")[0].innerHTML = progressContainer;

}

window.onload = setup
