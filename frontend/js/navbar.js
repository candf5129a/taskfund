const menuBtn = document.querySelector(".menu-btn");
const closeBtn = document.querySelector(".close-menu");
const mobileMenu = document.querySelector(".mobile-menu");

console.log(menuBtn);
console.log(closeBtn);
console.log(mobileMenu);

if (menuBtn && mobileMenu) {
    menuBtn.addEventListener("click", () => {
        mobileMenu.classList.add("active");
    });
}

if (closeBtn && mobileMenu) {
    closeBtn.addEventListener("click", () => {
        mobileMenu.classList.remove("active");
    });
}