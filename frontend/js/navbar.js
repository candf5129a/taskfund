const menuBtn = document.querySelector(".menu-btn");
const closeBtn = document.querySelector(".close-menu");
const mobileMenu = document.querySelector(".mobile-menu");

console.log(menuBtn);
console.log(closeBtn);
console.log(mobileMenu);

menuBtn.addEventListener("click", () => {

    mobileMenu.classList.add("active");

});

closeBtn.addEventListener("click", () => {

    mobileMenu.classList.remove("active");

});