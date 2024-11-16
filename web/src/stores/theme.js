import { writable } from 'svelte/store';

const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)').matches;
const initialTheme = prefersDarkScheme ? 'dark' : 'light';

export const theme = writable(initialTheme);

// Listen for changes in the user's preferred color scheme
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (event) => {
    theme.set(event.matches ? 'dark' : 'light');
});