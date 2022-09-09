<script lang="ts">
	import { go } from '../connect-web'
	import type { Giphy } from './giphy'

	const apiKey = import.meta.env.VITE_GIPHY_API_KEY
	export let tag = ''
	export let endpoint = 'random'
	export let rating: 'g' | 'r' | 'pg' | 'pg-13' = 'r'

	let response: Giphy.Response | null
	$: {
		const url =
			!!apiKey &&
			`https://api.giphy.com/v1/gifs/${endpoint}?api_key=${apiKey}&tag=${tag}&rating=${rating}`
		console.log('tag', tag)
		if (url && true) {
			fetch(url).then(async (result) => {
				if (!result.ok) {
					console.error('failed to fetch giphy', result)
					return
				}

				const [r, err] = await go<Giphy.Response>(result.json())
				if (err) {
					console.error('failed to parse giphy', err)
					return
				}
				response = r
			})
		}
	}
	$: image = !!response && response.data.images.downsized
</script>

<div>
	{#if response}
		{#if image}
			<img
				src={image.url}
				alt={'Animated gif with tag ' + tag}
				onload="this.style.opacity=1;this.style.transform='scale(1)'"
			/>
		{/if}
	{/if}
</div>

<style>
	div {
		width: min(80vw, 480px);
		height: min(calc(80vw * 9 / 16), calc(480px * 9 / 16));
		/* background-color: var(--color-grey-800); */
		margin: auto;
	}
	img {
		opacity: 0;
		transform: scale(0);
		transition-property: opacity, transform;
		transition-duration: 300ms;
		transition-timing-function: var(--easing-standard);
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
</style>
