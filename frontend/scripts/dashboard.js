const token = localStorage.getItem("access_token");


// No token = not logged in
if (!token) {
    window.location.href = "login.html";
}


// Load the logged-in user's profile
async function loadProfile() {

    try {

        const response = await fetch(
            "http://localhost:8080/api/v1/profile",
            {
                method: "GET",

                headers: {
                    "Authorization": `Bearer ${token}`
                }
            }
        );


        const result = await response.json();


        if (!response.ok || !result.success) {

            localStorage.removeItem("access_token");
            localStorage.removeItem("user");

            window.location.href = "login.html";

            return;
        }


        const user = result.data;


        document.getElementById("user-name").textContent =
            `${user.first_name} ${user.last_name}`;


        document.getElementById("user-email").textContent =
            user.email;


    } catch (error) {

        console.error("Dashboard error:", error);

        document.getElementById("dashboard-message").textContent =
            "Unable to connect to TaskFunds server.";

    }
}


// Logout
document
    .getElementById("logout-button")
    .addEventListener("click", function () {

        localStorage.removeItem("access_token");
        localStorage.removeItem("user");

        window.location.href = "login.html";
    });


loadProfile();