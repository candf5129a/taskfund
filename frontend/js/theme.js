document.addEventListener("DOMContentLoaded", () => {

    const themeBtn = document.querySelector(".theme-toggle");

    if (!themeBtn) return;

    themeBtn.addEventListener("click", () => {

        document.body.classList.toggle("dark");

    });

});