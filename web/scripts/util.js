function setPage(pageName, create_history = true) {
    const all_pages = document.getElementsByClassName("main-content");
    for (let i = 0; i < all_pages.length; i++) {
        all_pages[i].classList.add("hidden");
    }

    const page_element = document.getElementById(`page-${pageName}`);
    if (page_element) {
        page_element.classList.remove("hidden");
    }

    if (create_history) {
        const url = new URL(window.location.href)
        url.searchParams.set("p", pageName)
        window.history.pushState({}, "", url.href)
    }
}