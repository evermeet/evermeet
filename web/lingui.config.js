import { defineConfig } from '@lingui/cli';

export default defineConfig({
	sourceLocale: 'en',
	locales: ['en', 'cs'],
	catalogs: [
		{
			path: '<rootDir>/src/locales/{locale}/messages',
			include: ['src']
		}
	]
});
