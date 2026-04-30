<script lang="ts">
	let { src = '', did = '', size = 100, class: className = '', rounded = true } = $props();

	function getIdenticon(did: string) {
		let hash = 0;
		for (let i = 0; i < did.length; i++) {
			hash = ((hash << 5) - hash) + did.charCodeAt(i);
			hash |= 0;
		}

		const hue = Math.abs(hash % 360);
		const saturation = 50 + Math.abs((hash >> 8) % 30);
		const color = `hsl(${hue}, ${saturation}%, 45%)`;

		let rects = [];
		for (let i = 0; i < 5; i++) {
			for (let j = 0; j < 3; j++) {
				const bit = (Math.abs(hash) >> (i * 3 + j)) & 1;
				if (bit) {
					rects.push({ x: j * 20, y: i * 20 });
					if (j < 2) {
						rects.push({ x: (4 - j) * 20, y: i * 20 });
					}
				}
			}
		}

		return { color, rects };
	}

	const identicon = $derived(did ? getIdenticon(did) : null);
</script>

<div class="avatar-box {className}" style="width: {size}px; height: {size}px; border-radius: {rounded ? '50%' : 'var(--radius-xl)'}">
	{#if src}
		<img {src} alt={did} class="avatar-img" />
	{:else if identicon}
		<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="identicon">
			<rect width="100" height="100" fill="var(--bg-avatar)" />
			{#each identicon.rects as rect}
				<rect x={rect.x} y={rect.y} width="20" height="20" fill={identicon.color} />
			{/each}
		</svg>
	{:else}
		<div class="placeholder">?</div>
	{/if}
</div>

<style>
	.avatar-box {
		display: inline-block;
		overflow: hidden;
		flex-shrink: 0;
		background: var(--bg-avatar);
		border: 1px solid var(--shadow-avatar);
	}
	.avatar-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.identicon {
		width: 100%;
		height: 100%;
		display: block;
	}
	.placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		font-weight: bold;
	}
</style>
