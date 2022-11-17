function doNow(fn) {
    fn();
    return fn;
}

function labelTimeElement(tag, time) {
    time = dayjs(time);

    if (tag.innerText == "")
        tag.innerText = time.fromNow();
    tag.setAttribute("data-bs-toggle", "tooltip");
    tag.setAttribute("data-bs-title", time.format("L LT"));
    new bootstrap.Tooltip(tag);
}

function dateInZone(zone) {
    if (zone)
        return new Date(new Date().toLocaleString("en-US", { timeZone: zone }));
    return new Date()
}