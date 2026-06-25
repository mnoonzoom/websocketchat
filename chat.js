const username = localStorage.getItem("username");
let currentChat = "all";

let socket = new WebSocket("ws://localhost:8080/ws?username=" + username);

function addMessage(msg) {
    const div = document.createElement("div");

    const isMe = msg.username === username;

    if (isMe) {
        div.className = "text-end mb-2";
    } else {
        div.className = "text-start mb-2";
    }

    const bubble = document.createElement("div");
    bubble.classList.add("message");

    if (isMe) {
        bubble.classList.add("my-message");
    } else {
        bubble.classList.add("other-message");
    }

    bubble.innerHTML =
        "<b>" + msg.username + "</b><br>" +
        msg.message;

    div.appendChild(bubble);

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
            all.className = "list-group-item";

            if (currentChat === "all") {
                all.classList.add("selected");
            }

            all.textContent = "All";

            all.onclick = () => {
                document.querySelectorAll("#users li").forEach(item => {
                    item.classList.remove("selected");
                });

                all.classList.add("selected");

                currentChat = "all";
                loadMessages();
            };

            list.appendChild(all);

            users.forEach(u => {
                if (u === username) return;

                const li = document.createElement("li");
                li.className = "list-group-item";
                li.textContent = u;

                if (currentChat === u) {
                    li.classList.add("selected");
                }

                li.onclick = () => {
                    document.querySelectorAll("#users li").forEach(item => {
                        item.classList.remove("selected");
                    });

                    li.classList.add("selected");

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

function darkmode(){
    
    var element1 = document.getElementById("navbar")
    var element2 =document.getElementById("body")
  var element4 =document.getElementById("chat")

    element2.classList.toggle("bg-light")
    element2.classList.toggle("dark-mode")
    element1.classList.toggle("bg-light");
    element1.classList.toggle("dark-mode");
    element1.classList.toggle("navbar-light")
    element1.classList.toggle("navbar-dark")
    element4.classList.toggle("darkmode")

      const icon = document.querySelector("#themes i");

if (element2.classList.contains("dark-mode")) {
    icon.classList.remove("bi-moon-fill");
    icon.classList.add("bi-sun-fill");
} else {
    icon.classList.remove("bi-sun-fill");
    icon.classList.add("bi-moon-fill");
}


}
document.getElementById("themes").addEventListener("click", darkmode)
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