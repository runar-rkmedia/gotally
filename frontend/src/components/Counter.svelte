<script lang="ts">
	export let value = 0
	export let variant: 'normal' | 'error' | 'success' = 'normal'
	export let show = true
	export let label = ''
</script>

<div
	class="wrapper"
	class:show
	class:error={variant === 'error'}
	class:success={variant === 'success'}
>
	<div class="label">
		{label}:
	</div>
	{value}
</div>

<style>
	.error {
		color: var(--color-red-300);
	}
	.success {
		color: var(--color-green-700);
	}
	.wrapper {
		font-family: var(--font-mono);
		display: flex;
		align-items: center;
		transition: all 150ms var(--easing-standard);
		min-width: 18ch;
	}
	.wrapper:not(.show) {
		opacity: 0;
	}
	.label {
		width: 100%;
	}
</style>
