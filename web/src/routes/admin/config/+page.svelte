<script lang="ts">
	import { intl } from '$lib/i18n.svelte.js';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount, onDestroy } from 'svelte';
	import { EditorState } from '@codemirror/state';
	import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view';
	import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
	import { StreamLanguage, syntaxHighlighting, defaultHighlightStyle, HighlightStyle } from '@codemirror/language';
	import { tags } from '@lezer/highlight';
	import { toml } from '@codemirror/legacy-modes/mode/toml';

	let editorContainer = $state<HTMLDivElement | null>(null);
	let defaultsContainer = $state<HTMLDivElement | null>(null);
	let view: EditorView | null = null;
	let defaultsView: EditorView | null = null;
	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let saveError = $state('');
	let saved = $state(false);

	const darkHighlightStyle = HighlightStyle.define([
		{ tag: tags.comment,                   color: '#6a9955' },
		{ tag: [tags.atom, tags.bool],         color: '#7cb8e8' },
		{ tag: tags.string,                    color: '#ce9178' },
		{ tag: [tags.number, tags.literal],    color: '#b5cea8' },
		{ tag: tags.meta,                      color: '#4e7a4e' },
		{ tag: tags.invalid,                   color: '#f44747' },
	]);

	function makeExtensions(readonly = false) {
		const dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		return [
			...(readonly ? [] : [history(), keymap.of([...defaultKeymap, ...historyKeymap])]),
			lineNumbers(),
			...(readonly ? [] : [highlightActiveLine()]),
			StreamLanguage.define(toml),
			syntaxHighlighting(dark ? darkHighlightStyle : defaultHighlightStyle),
			EditorState.readOnly.of(readonly),
			EditorView.theme({
				'&': { height: '100%', fontSize: '0.875rem' },
				'&.cm-focused': { outline: 'none' },
				'.cm-scroller': { fontFamily: 'monospace', lineHeight: '1.6', overflow: 'auto' },
				'.cm-content': { padding: '0.75rem 0' },
				'.cm-line': { padding: '0 0.75rem' },
			}),
			EditorView.lineWrapping,
		];
	}

	onMount(async () => {
		await auth.load();
		if (!auth.user) { goto('/auth/login?next=/admin/config'); return; }
		if (!auth.user.is_admin) { goto('/admin'); return; }

		let tomlContent = '';
		let defaultsContent = '';
		try {
			const res = await api.admin.config();
			tomlContent = res.toml;
			defaultsContent = res.defaults;
		} catch (err: any) {
			error = err.message ?? 'Failed to load config';
			loading = false;
			return;
		}

		loading = false;
		await Promise.resolve();

		if (editorContainer) {
			view = new EditorView({
				state: EditorState.create({ doc: tomlContent, extensions: makeExtensions(false) }),
				parent: editorContainer,
			});
		}

		if (defaultsContainer) {
			defaultsView = new EditorView({
				state: EditorState.create({ doc: defaultsContent, extensions: makeExtensions(true) }),
				parent: defaultsContainer,
			});
		}
	});

	onDestroy(() => {
		view?.destroy();
		defaultsView?.destroy();
	});

	async function save(e: Event) {
		e.preventDefault();
		if (!view) return;
		saving = true;
		saveError = '';
		saved = false;
		try {
			await api.admin.saveConfig({ toml: view.state.doc.toString() });
			saved = true;
		} catch (err: any) {
			saveError = err.message ?? 'Save failed';
		} finally {
			saving = false;
		}
	}
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>{intl.t('admin.nav.config')}</h1>
			<p class="muted">Editing <code>evermeet.toml</code> — the server will restart after saving.</p>
		</div>
	</div>
	<AdminNav active="config" />

	{#if loading}
		<p class="muted">Loading...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<div class="layout">
			<div class="col">
				<h2>{intl.t('admin.config.active')}</h2>
				<form onsubmit={save}>
					<div class="editor-wrap" bind:this={editorContainer}></div>
					<div class="actions">
						<button type="submit" disabled={saving}>{saving ? 'Saving…' : intl.t('admin.config.saveRestart')}</button>
						{#if saveError}<span class="error">{saveError}</span>{/if}
						{#if saved}<span class="success">Saved — server is restarting.</span>{/if}
					</div>
				</form>
			</div>
			<div class="col">
				<h2>Defaults <span class="filename">evermeet.defaults.toml</span></h2>
				<div class="editor-wrap readonly" bind:this={defaultsContainer}></div>
			</div>
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 72rem;
		margin: 2.5rem auto;
		padding: 0 1.5rem 4rem;
		font-family: system-ui, sans-serif;
	}
	.page-header { margin-bottom: 0.75rem; }
	.eyebrow {
		margin: 0 0 0.35rem;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	h1 { margin: 0; }
	h2 { margin: 0 0 0.5rem; font-size: 1rem; display: flex; align-items: baseline; gap: 0.5rem; }
	.filename { font-size: 0.8rem; font-weight: 400; color: var(--text-muted); font-family: monospace; }
	.muted { color: var(--text-secondary); }
	.error { color: var(--text-error); }
	.success { color: var(--text-success); }

	.layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		align-items: start;
	}
	@media (max-width: 800px) {
		.layout { grid-template-columns: 1fr; }
	}

	.editor-wrap {
		min-height: 28rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
		overflow: hidden;
	}
	.editor-wrap.readonly { opacity: 0.8; }

	.editor-wrap :global(.cm-editor) { height: 100%; outline: none; }
	.editor-wrap :global(.cm-editor.cm-focused) { outline: none; }

	@media (prefers-color-scheme: dark) {
		.editor-wrap :global(.cm-cursor) { border-left-color: #fff !important; }
	}
	/* light mode gutter — one-dark overrides these in dark mode */
	.editor-wrap :global(.cm-gutters) {
		background: var(--bg-card);
		border-right: 1px solid var(--border-subtle);
		color: var(--text-muted);
	}
	.editor-wrap :global(.cm-activeLineGutter),
	.editor-wrap :global(.cm-activeLine) {
		background: color-mix(in srgb, var(--bg-raised, #f5f5f5) 60%, transparent);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-top: 0.75rem;
	}
	button {
		padding: 0.55rem 1.1rem;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-input);
		background: var(--bg-raised);
		color: var(--text);
		cursor: pointer;
	}
	button:disabled { opacity: 0.55; cursor: not-allowed; }
</style>
