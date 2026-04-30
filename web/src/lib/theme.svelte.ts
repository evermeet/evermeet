export type ThemeId = 'default' | 'luma';

export interface ThemeInfo {
	id: ThemeId;
	name: string;
	description: string;
}

export const themes: ThemeInfo[] = [
	{ id: 'default', name: 'Default', description: 'Clean light theme' },
	{ id: 'luma', name: 'Luma', description: 'Dark theme inspired by lu.ma' },
];

const STORAGE_KEY = 'evermeet-theme';

function createTheme() {
	let current = $state<ThemeId>('luma');

	function apply(id: ThemeId) {
		current = id;
		document.documentElement.setAttribute('data-theme', id);
		localStorage.setItem(STORAGE_KEY, id);
	}

	function load() {
		const saved = localStorage.getItem(STORAGE_KEY) as ThemeId | null;
		const id = saved && themes.find(t => t.id === saved) ? saved : 'luma';
		apply(id);
	}

	return {
		get current() { return current; },
		apply,
		load,
	};
}

export const theme = createTheme();
