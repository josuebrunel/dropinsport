const selection = document.querySelector('.selection');

selection.addEventListener('click', (event) => {
    const links = selection.querySelectorAll('a');
    console.log(links);
    links.forEach((link) => {
        if (event.target.id === link.id) {
            link.classList.remove('outline');
        } else {
            link.classList.add('outline');
        }
    });
});