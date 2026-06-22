const username = localStorage.getItem("username");
let currentChat = "all";

let socket = new WebSocket(
    "ws://localhost:8080/ws?username=" + username
);

function addMessage(msg) {
    const li = document.createElement("li");

    li.textContent =
        `${msg.time} ${msg.username}` +
        (msg.recipient ? ` → ${msg.recipient}` : " → all") +
        `: ${msg.message}`;

    document.getElementById("messages").appendChild(li);
}

function shouldShow(msg) {

    if (currentChat === "all") {
        return !msg.recipient || msg.recipient === "";
    }

    return (
        (msg.username === username &&
            msg.recipient === currentChat) ||

        (msg.username === currentChat &&
            msg.recipient === username)
    );
}

function loadMessages() {
    document.getElementById("messages").innerHTML = "";

    fetch("/messages?username=" + username)
        .then(r => r.json())
        .then(messages => {
            messages.forEach(msg => {
                if (shouldShow(msg)) {
                    addMessage(msg);
                }
            });
        });
}

document.addEventListener("DOMContentLoaded", () => {

    loadMessages();

    fetch("/users")
        .then(r => r.json())
        .then(users => {

            const allLi = document.createElement("li");
            allLi.textContent = "All";

            allLi.onclick = () => {
                currentChat = "all";
                document.getElementById("recipient").value = "";

                document.querySelectorAll("#users li").forEach(el => {
                    el.classList.remove("selected");
                });

                allLi.classList.add("selected");

                loadMessages();
            };

            document.getElementById("users").appendChild(allLi);

            users.forEach(user => {

                if (user === username) return;

                const li = document.createElement("li");
                li.textContent = user;

                li.onclick = () => {

                    currentChat = user;
                    document.getElementById("recipient").value = user;

                    document.querySelectorAll("#users li").forEach(el => {
                        el.classList.remove("selected");
                    });

                    li.classList.add("selected");

                    loadMessages();
                };

                document.getElementById("users").appendChild(li);
            });
        });

    socket.onmessage = function (event) {
        const msg = JSON.parse(event.data);

        if (shouldShow(msg)) {
            addMessage(msg);
        }
    };

    document
        .getElementById("send")
        .addEventListener("click", sendMessage);
});

function sendMessage() {

    const message = document.getElementById("message").value;
    const recipient = document.getElementById("recipient").value;

    socket.send(JSON.stringify({
        username: username,
        recipient: recipient,
        message: message
    }));

    document.getElementById("message").value = "";
}