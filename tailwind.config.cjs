module.exports = {
	darkMode: "class",
	content: [
		"./views/**/*{.go.html,.html,.templ}",
	],
	theme: {
		extend: {},
	},
	plugins: [
		require('@tailwindcss/forms'),
	],
};
