const list = document.querySelectorAll('.list');

function activeLink() {
    list.forEach((item) => item.classList.remove('active'));
    this.classList.add('active');
    localStorage.activated = this.id;
}

function setActive() {
    list.forEach((item) => item.classList.remove('active', 'stopped'));
    let activated = null;
    if (window.location.pathname === '/') {
        activated = document.getElementById('1');
    }
    if (window.location.pathname === '/history') {
        activated = document.getElementById('2');
    }
    if (window.location.pathname === '/settings') {
        activated = document.getElementById('3');
    }
    if (window.location.pathname === '/documentation') {
        activated = document.getElementById('4');
    }
    activated.classList.add('active');
    activated.classList.add('stopped');

}

list.forEach((item) => item.addEventListener('click', activeLink));


