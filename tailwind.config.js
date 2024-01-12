/** @type {import('tailwindcss').Config} */
const defaultTheme = require("tailwindcss/defaultTheme");
module.exports = {
    content: ["./templates/**/*.{html,js}"],
    theme: {
        fontFamily: {
            sans: ["Open Sans", ...defaultTheme.fontFamily.sans],
            serif: ["Merriweather", ...defaultTheme.fontFamily.serif],
            mono: [...defaultTheme.fontFamily.mono],
        },
    },
    plugins: [require("@tailwindcss/forms")],
};
