<script>
	export const ssr = false
	import 'pollen-css'
</script>

<slot />

<style lang="scss">
	:root {
		--color-primary: var(--color-blue-500);
		--color-secondary: var(--color-green-500);
		--color-danger: var(--color-orange-500);
		--color-error: var(--color-red-500);
		--color-white: var(--color-grey-50);
	}
</style>
