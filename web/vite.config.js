import { sveltekit } from '@sveltejs/kit/vite';
import { lingui } from '@lingui/vite-plugin';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit(), lingui()],
	server: {
		proxy: {
			'/api': 'http://localhost:7332',
			'/health': 'http://localhost:7332'
		}
	}
});
