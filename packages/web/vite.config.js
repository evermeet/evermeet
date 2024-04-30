import { sveltekit } from '@sveltejs/kit/vite';
import ViteYaml from '@modyfi/vite-plugin-yaml';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit(),
		ViteYaml()
	],
	build: {
		commonjsOptions: {
			dynamicRequireTargets: [
				'langs/*.json'
			]
		}
	}
});
