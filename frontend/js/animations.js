const counters = document.querySelectorAll(".counter");

const animateCounter = (counter) => {

    const target = Number(counter.dataset.target);

    let current = 0;

    const increment = target / 100;

    const update = () => {

        current += increment;

        if (current < target) {

            counter.textContent = Math.floor(current).toLocaleString();

            requestAnimationFrame(update);

        } else {

            counter.textContent = target.toLocaleString();

        }

    };

    update();

};

const observer = new IntersectionObserver((entries) => {

    entries.forEach((entry) => {

        if (entry.isIntersecting) {

            animateCounter(entry.target);

            observer.unobserve(entry.target);

        }

    });

});

counters.forEach(counter => observer.observe(counter));