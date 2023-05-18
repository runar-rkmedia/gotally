<script>
	export const ssr = false
	import 'pollen-css'
	import { httpErrorStore } from '../connect-web'
</script>

<!-- content here -->
<div class="errors">
	{#each $httpErrorStore?.errors || [] as err}
		<button
			class="error"
			on:click={() =>
				httpErrorStore.update((e) => ({
					...e,
					errors: e.errors.filter((error) => error.time !== err.time && error.url !== err.url),
				}))}
		>
			<!-- content here -->
			<p>Sorry, an error occured:</p>
			<p>
				{err.error.message}
			</p>

			<p>Sorry for the inconvinience</p>
		</button>
	{/each}
</div>
<slot />

<style lang="scss">
	:root {
		--color-primary: var(--color-blue-500);
		--color-secondary: var(--color-green-500);
		--color-danger: var(--color-orange-500);
		--color-error: var(--color-red-500);
		--color-white: var(--color-grey-50);
	}
	.errors {
		position: fixed;
		z-index: 1;
		background: var(--color-red);
	}
</style>
