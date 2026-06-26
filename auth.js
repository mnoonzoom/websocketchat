async function Auth(){


    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    const res = await fetch("/login", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        username,
        password
    })
})
.then(r => r.json())
.then(data => {
    if (data.success) {
        localStorage.setItem("token", data.token);
        window.location = "/chat";
    }
});

    const data = await res.json();

    if (data.success) {
        window.location.href = "/chat";
        localStorage.setItem("username", username);
    } else {
        alert("Wrong credentials");
    }
}
async function Register() {
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
    
    const res = await fetch("/register", {
        method: "POST",
        headers:{
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            username,
            password
        })
    })
    const data = await res.json();
    if (data.success){
        alert("Registration success")
    } else {
        alert("Registration fail");
    }
}
function darkmode(){
    var element1 = document.getElementById("navbar")
    var element2 =document.getElementById("body")
  var element3 = document.getElementById("login-form");
  var element4 =document.getElementById("login")
  var element5 =document.getElementById('register')
    element2.classList.toggle("bg-light")
    element2.classList.toggle("dark-mode")
    element1.classList.toggle("bg-light");
    element1.classList.toggle("dark-mode");
    element1.classList.toggle("navbar-light")
    element1.classList.toggle("navbar-dark")
    element4.classList.toggle("darkmode")
    element5.classList.toggle('darkmode')
      const icon = document.querySelector("#themes i");

        if (document.body.classList.contains("dark-mode")) {
        icon.className = "bi bi-sun-fill";
    } else {
        icon.className = "bi bi-moon-fill";
    }

}
const token = localStorage.getItem("token");

let socket = new WebSocket(
    "ws://localhost:8080/ws?token=" + token
);
document.getElementById("themes").addEventListener("click", darkmode)
 document.getElementById("login").addEventListener("click", Auth);
 document.getElementById("register").addEventListener("click", Register) 
 

 