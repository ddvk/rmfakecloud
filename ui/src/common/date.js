export function formatDate(dateString) {
    // TODO: support other date formats
    // TODO: option for local time instead of utc

    let date = new Date(dateString);

    let year = String(date.getUTCFullYear());
    let month = String(date.getUTCMonth() + 1).padStart(2, "0");
    let day = String(date.getUTCDate()).padStart(2, "0");

    let hour = String(date.getUTCHours()).padStart(2, "0");
    let minute = String(date.getUTCMinutes()).padStart(2, "0");
    let second = String(date.getUTCSeconds()).padStart(2, "0");

    return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}
