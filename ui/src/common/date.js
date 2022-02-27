export function formatDate(dateString) {
    // TODO: support other date formats
    // TODO: option for local time instead of utc

    let date = new Date(dateString);
    return date.toLocaleString('sv');
}
