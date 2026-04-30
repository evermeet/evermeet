import { i18n } from '@lingui/core';
import { messages as enMessages } from '../locales/en/messages.js';
import { messages as csMessages } from '../locales/cs/messages.js';

export type Locale = 'en' | 'cs';

export const locales: Locale[] = ['en', 'cs'];

export const localeNames: Record<Locale, string> = {
	en: 'English',
	cs: 'Čeština'
};

const STORAGE_KEY = 'evermeet-locale';
const messages = {
	en: enMessages,
	cs: csMessages
};

function normalizeLocale(locale: string | null | undefined): Locale | null {
	if (!locale) return null;
	const base = locale.toLowerCase().split('-')[0];
	return locales.includes(base as Locale) ? (base as Locale) : null;
}

function createI18n() {
	let locale = $state<Locale>('en');

	function activate(nextLocale: Locale) {
		if (!i18n.messages[nextLocale]) {
			i18n.load(nextLocale, messages[nextLocale]);
		}
		i18n.activate(nextLocale);
		locale = nextLocale;
		document.documentElement.lang = nextLocale;
		localStorage.setItem(STORAGE_KEY, nextLocale);
	}

	function load() {
		const saved = normalizeLocale(localStorage.getItem(STORAGE_KEY));
		const preferred = normalizeLocale(navigator.language);
		activate(saved ?? preferred ?? 'en');
	}

	function t(id: string, values?: Record<string, unknown>) {
		locale;
		const translated = i18n._(id);
		if (!values) return translated;
		return translated.replace(/\{(\w+)\}/g, (match, key: string) => {
			const value = values[key];
			return value === undefined || value === null ? match : String(value);
		});
	}

	function dateLocale() {
		return locale === 'cs' ? 'cs-CZ' : 'en';
	}

	i18n.load('en', enMessages);
	i18n.activate('en');

	return {
		get locale() { return locale; },
		activate,
		load,
		t,
		dateLocale
	};
}

export const intl = createI18n();
