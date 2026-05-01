<script lang="ts">
	import { api } from '$lib/api.js';
	import { intl } from '$lib/i18n.svelte.js';

	interface Props {
		value: string;
		onchange?: (url: string) => void;
		rounded?: boolean;   // show preview as circle (for avatars)
		previewSize?: number; // preview height in px, default 200
	}

	let { value = $bindable(''), onchange, rounded = false, previewSize = 200 }: Props = $props();

	let uploading = $state(false);
	let uploadError = $state('');
	let inputEl = $state<HTMLInputElement | null>(null);

	async function handleFile(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		uploading = true;
		uploadError = '';
		try {
			const res = await api.blobs.upload(file);
			value = res.url;
			onchange?.(res.url);
		} catch (err: any) {
			uploadError = err.message;
		} finally {
			uploading = false;
		}
	}

	function clear() {
		value = '';
		onchange?.('');
		if (inputEl) inputEl.value = '';
	}
</script>

<div class="image-upload">
	{#if value}
		<div
			class="preview"
			class:preview-rounded={rounded}
			style={rounded ? `width: ${previewSize}px; height: ${previewSize}px;` : `max-height: ${previewSize}px;`}
		>
			<img src={value} alt={intl.t('upload.coverPreview')} />
			<button type="button" class="clear-btn" onclick={clear} title={intl.t('upload.removeImage')}>
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
				</svg>
			</button>
		</div>
	{:else}
		<label class="upload-area" class:uploading>
			<input
				bind:this={inputEl}
				type="file"
				accept="image/jpeg,image/png,image/gif,image/webp"
				onchange={handleFile}
				disabled={uploading}
			/>
			{#if uploading}
				<span class="upload-hint">{intl.t('upload.uploading')}</span>
			{:else}
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/>
					<polyline points="21 15 16 10 5 21"/>
				</svg>
				<span class="upload-hint">{intl.t('upload.clickImage')}</span>
				<span class="upload-sub">{intl.t('upload.acceptedImages')}</span>
			{/if}
		</label>
	{/if}
	{#if uploadError}
		<p class="upload-error">{uploadError}</p>
	{/if}
</div>

<style>
	.image-upload { display: flex; flex-direction: column; gap: 0.4rem; }

	.upload-area {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.4rem;
		padding: 1.5rem;
		border: 1px dashed var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		color: var(--text-muted);
		cursor: pointer;
		transition: border-color 0.1s, background 0.1s;
	}
	.upload-area:hover { border-color: var(--border-input-focus); background: var(--bg-hover); color: var(--text); }
	.upload-area.uploading { opacity: 0.6; cursor: default; }
	.upload-area input { display: none; }

	.upload-hint { font-size: 0.9rem; font-weight: 500; }
	.upload-sub { font-size: 0.78rem; color: var(--text-muted); }

	.preview {
		position: relative;
		display: inline-block;
		overflow: visible;
		border-radius: var(--radius-md);
	}
	.preview-rounded {
		border-radius: 50%;
		overflow: hidden;
		border: 1px solid var(--border-input);
	}
	.preview img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-input);
	}
	.preview-rounded img {
		border-radius: 50%;
		border: none;
	}
	.clear-btn {
		position: absolute;
		top: -8px;
		right: -8px;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 1px solid var(--border-input);
		background: var(--bg-raised);
		color: var(--text);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0;
		transition: background 0.1s;
	}
	.clear-btn:hover { background: var(--bg-hover); }

	.upload-error { font-size: 0.85rem; color: var(--text-error); margin: 0; }
</style>
