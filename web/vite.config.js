import { sveltekit } from '@sveltejs/kit/vite';
import { lingui } from '@lingui/vite-plugin';
import { defineConfig } from 'vite';
import { readFileSync } from 'fs';
import { resolve } from 'path';

function readVersion() {
	const src = readFileSync(resolve('../internal/version/version.go'), 'utf-8');
	const m = src.match(/Version\s*=\s*"([^"]+)"/);
	return m ? m[1] : 'dev';
}

export default defineConfig({
	define: {
		__APP_VERSION__: JSON.stringify(readVersion()),
	},
	plugins: [sveltekit(), lingui()],
	server: {
		proxy: {
			'/api': 'http://localhost:7332',
			'/health': 'http://localhost:7332'
		}
	}
});
