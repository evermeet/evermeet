
export function stringToSlug(str) {
    return str
        .toLowerCase()     // Convert the string to lowercase letters
        .normalize('NFD')  // Normalize the string to decompose combined letters like "é" to "e"
        .replace(/[\u0300-\u036f]/g, '') // Remove diacritical marks
        .replace(/[^a-z0-9 -]/g, '') // Remove invalid characters
        .replace(/\s+/g, '-') // Replace spaces with -
        .replace(/-+/g, '-'); // Replace multiple - with single -
}