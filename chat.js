const username = localStorage.getItem("username");
let currentChat = "all";

let socket = new WebSocket("ws://localhost:8080/ws?username=" + username);

function addMessage(msg) {
    const div = document.createElement("div");

    const isMe = msg.username === username;

    div.className = "mb-2 " + (isMe ? "text-end" : "text-start");

    div.innerHTML = `
        <div class="d-inline-block p-2 rounded ${isMe ? 'bg-primary text-white' : 'bg-light'}">
            <b>${msg.username}</b>: ${msg.message}
        </div>
    `;

    document.getElementById("messages").appendChild(div);
}

function shouldShow(msg) {

    if (currentChat === "all") {
        return !msg.recipient || msg.recipient === "all";
    }

    return (
        (msg.username === username && msg.recipient === currentChat) ||
        (msg.username === currentChat && msg.recipient === username)
    );
}

function loadMessages() {
    document.getElementById("messages").innerHTML = "";

    fetch("/messages?username=" + username)
        .then(r => r.json())
        .then(messages => {
            messages.forEach(msg => {
                if (shouldShow(msg)) addMessage(msg);
            });
        });
}

function loadUsers() {
    fetch("/users")
        .then(r => r.json())
        .then(users => {
            const list = document.getElementById("users");
            list.innerHTML = "";


            const all = document.createElement("li");
            all.className = "list-group-item active";
            all.textContent = "All";

            all.onclick = () => {
                currentChat = "all";
                loadMessages();
            };

            list.appendChild(all);

            users.forEach(u => {
                if (u === username) return;

                const li = document.createElement("li");
                li.className = "list-group-item";
                li.textContent = u;

                li.onclick = () => {
                    currentChat = u;
                    loadMessages();
                };

                list.appendChild(li);
            });
        });
}

socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    if (shouldShow(msg)) {
        addMessage(msg);
    }
};


document.getElementById("send").onclick = () => {
    const message = document.getElementById("message").value;

    socket.send(JSON.stringify({
        username,
        recipient: currentChat === "all" ? "" : currentChat,
        message
    }));

    document.getElementById("message").value = "";
};

document.addEventListener("DOMContentLoaded", () => {
    loadUsers();
    loadMessages();
});