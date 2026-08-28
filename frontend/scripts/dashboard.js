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

        console.error("Profile error:", error);

        document.getElementById("dashboard-message").textContent =
            "Unable to connect to TaskFunds server.";
    }
}


// Load wallet
async function loadWallet() {

    try {

        const response = await fetch(
            "http://localhost:8080/api/v1/wallet",
            {
                method: "GET",

                headers: {
                    "Authorization": `Bearer ${token}`
                }
            }
        );

        const result = await response.json();

        if (!response.ok || !result.success) {
            throw new Error(result.message || "Failed to load wallet.");
        }

        const wallet = result.data;

        document.getElementById("wallet-balance").textContent =
            `₦${Number(wallet.balance).toFixed(2)}`;

    } catch (error) {

        console.error("Wallet error:", error);

        document.getElementById("wallet-balance").textContent =
            "₦0.00";

        document.getElementById("dashboard-message").textContent =
            "Unable to load wallet.";
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


// Load dashboard data
loadProfile();
loadWallet();