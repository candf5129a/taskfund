console.log("TaskFunds login.js loaded");
const loginForm = document.getElementById("login-form");
const loginButton = document.getElementById("login-button");
const messageBox = document.getElementById("login-message");

function showMessage(message) {
    messageBox.textContent = message;
    messageBox.classList.add("show");
}

loginForm.addEventListener("submit", async function (event) {
    console.log("LOGIN FORM SUBMITTED");

    event.preventDefault();

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    loginButton.disabled = true;
    loginButton.textContent = "Logging in...";

    messageBox.classList.remove("show");

    try {
        const response = await fetch(
            "http://localhost:8080/api/v1/auth/login",
            {
                method: "POST",

                headers: {
                    "Content-Type": "application/json"
                },

                body: JSON.stringify({
                    email: email,
                    password: password
                })
            }
        );

        const result = await response.json();

        if (!result.success) {
            showMessage(result.message || "Login failed.");
            return;
        }

        // Save JWT
        localStorage.setItem(
            "access_token",
            result.data.access_token
        );

        // Save user information
        localStorage.setItem(
            "user",
            JSON.stringify(result.data)
        );

        // Login successful
        window.location.href = "dashboard.html";
        
        
    } catch (error) {
        console.error(error);

        showMessage(
            "Unable to connect to TaskFunds server."
        );
    } finally {
        loginButton.disabled = false;
        loginButton.textContent = "Login";
    }
});
