const username = localStorage.getItem("username");

let socket = new WebSocket("ws://localhost:8080/ws?username=" + username);
  document.addEventListener("DOMContentLoaded", (event) => {
    fetch("/messages?username=" + localStorage.getItem("username"))
  .then(r => r.json())
  .then(messages => {
      messages.forEach(msg => {
          const li = document.createElement("li");
          li.textContent =
              `${msg.time} ${msg.username}` +
              (msg.recipient ? ` → ${msg.recipient}` : "") +
              `: ${msg.message}`;

          document.getElementById("messages").appendChild(li);
      });
  });
  fetch("/users")
  .then(r => r.json())
  .then(users => {
    users.forEach(user =>{
        const li = document.createElement("li");
        li.textContent=user;
       li.onclick = () => {
    document.getElementById("recipient").value = user;

    document.querySelectorAll("#users li").forEach(el => {
        el.classList.remove("selected");
    });

    li.classList.add("selected");
};
        document.getElementById("users").appendChild(li)
    })
  })
      

       socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);

    const li = document.createElement("li");

  li.textContent =
    `${msg.time} ${msg.username}` +
    (msg.recipient ? ` → ${msg.recipient}` : " → all") +
    `: ${msg.message}`;

    document.getElementById("messages").appendChild(li);
};
         document.getElementById("send").addEventListener("click", sendMessage) 
  });
 
 function sendMessage() {
            const username = document.getElementById("username").value;
            const message = document.getElementById("message").value;
            const recipient = document.getElementById("recipient").value;
            const time = Date.now()
      
         socket.send(JSON.stringify({
    username: username,
    recipient: recipient,
    message: message
}));

            document.getElementById("message").value = "";
   }
   