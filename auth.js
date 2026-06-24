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
    var element = document.body;
    var element1 = document.getElementById("container-fluid")
    element.classList.toggle("dark-mode")
    element1.classList.toggle("dark-mode")
}
document.getElementById("themes").addEventListener("click", darkmode)
 document.getElementById("login").addEventListener("click", Auth);
 document.getElementById("register").addEventListener("click", Register) 

 